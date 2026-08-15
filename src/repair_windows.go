//go:build windows

package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

func runRepair(cfg repairConfig) error {
	mode := "DRY-RUN (READ-ONLY)"
	if cfg.apply {
		mode = "APPLY (MUTATING)"
	}
	fmt.Printf("odhydrate %s  |  repair  |  %s\n", version, mode)
	fmt.Printf("Root                  : %s\n", cfg.root)
	fmt.Printf("Discovery workers     : %d\n", cfg.workers)
	if cfg.limit > 0 {
		fmt.Printf("Limit                 : %d\n", cfg.limit)
	} else {
		fmt.Println("Limit                 : all SAFE_CANDIDATE files")
	}
	fmt.Printf("Redact paths          : %v\n", cfg.redact)
	if cfg.noReport {
		fmt.Println("Report                : disabled")
	} else {
		fmt.Printf("Report                : %s\n", cfg.reportPath)
	}
	fmt.Println("Mutation method       : CfUpdatePlaceholder(VERIFY_IN_SYNC | DEHYDRATE)")
	fmt.Println("ContentRead           : NONE")

	rootInfo, err := getSyncRootInfo(cfg.root)
	if err != nil {
		return fmt.Errorf("cannot read sync-root information: %w", err)
	}
	fmt.Printf("Provider              : %s %s\n", rootInfo.ProviderName, rootInfo.ProviderVersion)
	fmt.Printf("HydrationPrimary      : %d%s\n", rootInfo.HydrationPrimary, enumLabelHydrationPrimary(rootInfo.HydrationPrimary))
	if !isOneDriveProvider(rootInfo.ProviderName) {
		return fmt.Errorf("refusing repair: sync provider %q is not recognized as OneDrive", rootInfo.ProviderName)
	}
	if rootInfo.HydrationPrimary == cfHydrationAlwaysFull {
		return errors.New("refusing repair: sync root uses HydrationPrimary=ALWAYS_FULL")
	}

	if cfg.apply {
		if err := requireOneDriveStopped(); err != nil {
			return err
		}
		fmt.Println("OneDrive process      : not running (required for placeholder management)")
	} else {
		fmt.Println("OneDrive process      : not required for dry-run")
	}

	fmt.Println("\n=== Phase 1: fresh exhaustive discovery (READ-ONLY) ===")
	candidates, reviews, c, err := discoverRepairCandidates(cfg)
	if err != nil {
		return err
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].onDisk > candidates[j].onDisk })
	if cfg.limit > 0 && len(candidates) > cfg.limit {
		candidates = candidates[:cfg.limit]
	}
	var targetBytes int64
	for _, x := range candidates {
		targetBytes += x.onDisk
	}
	fmt.Printf("Files scanned         : %d\n", c.filesSeen.Load())
	fmt.Printf("CFAPI queried         : %d\n", c.queried.Load())
	fmt.Printf("Not-cloud files       : %d\n", c.notCloud.Load())
	fmt.Printf("Query errors          : %d\n", c.queryErrors.Load())
	fmt.Printf("SAFE_CANDIDATE        : %d selected, %s\n", len(candidates), humanBytes(targetBytes))
	fmt.Printf("REVIEW excluded       : %d\n", reviews)

	if len(candidates) == 0 {
		fmt.Println("RepairResult          : NOTHING_TO_DO")
		return nil
	}
	if !cfg.apply {
		fmt.Println("RepairResult          : DRY_RUN_READY")
		fmt.Println("Next                  : completely exit OneDrive, then rerun with --apply")
		return nil
	}

	// OneDrive may auto-restart during discovery. Re-check immediately before mutation.
	if err := requireOneDriveStopped(); err != nil {
		return fmt.Errorf("OneDrive state changed after discovery; no files were modified: %w", err)
	}

	fmt.Println("\n=== Phase 2: sequential repair ===")
	outcomes := make([]repairOutcome, 0, len(candidates))
	var success, skipped, failed int
	var freed int64

	for i, cand := range candidates {
		// Re-check before every file so an auto-restarted OneDrive client stops the batch.
		if err := requireOneDriveStopped(); err != nil {
			failed++
			outcomes = append(outcomes, repairOutcome{
				index: i + 1, status: "STOPPED_ONEDRIVE_RUNNING", reason: err.Error(),
				path: cand.path, logical: cand.logical, before: cand.onDisk, after: cand.onDisk,
			})
			fmt.Printf("[%d/%d] %s | STOPPED: %v\n", i+1, len(candidates), displayPath(cand.path, cfg.redact), err)
			break
		}

		label := displayPath(cand.path, cfg.redact)
		fmt.Printf("[%d/%d] %s | resident %s ... ", i+1, len(candidates), label, humanBytes(cand.onDisk))
		out := repairOneCandidate(i+1, cand, rootInfo)
		outcomes = append(outcomes, out)
		switch out.status {
		case "SUCCESS_VERIFIED":
			success++
			freed += out.freed
			fmt.Printf("SUCCESS, freed %s\n", humanBytes(out.freed))
		case "SKIPPED_STATE_CHANGED", "SKIPPED_WRONG_SYNC_ROOT", "ALREADY_DEHYDRATED", "SKIPPED_OPEN_FAILED":
			skipped++
			fmt.Printf("%s (%s)\n", out.status, out.reason)
		default:
			failed++
			fmt.Printf("FAILED (%s)\n", out.reason)
		}
		fmt.Printf("         progress: success=%d skipped=%d failed=%d | freed=%s / target=%s\n", success, skipped, failed, humanBytes(freed), humanBytes(targetBytes))
		if failed > 0 {
			fmt.Println("         safety stop: unexpected failure; no further files will be modified")
			break
		}
	}

	if !cfg.noReport {
		if err := writeRepairReport(cfg, outcomes); err != nil {
			return fmt.Errorf("repair ran, but writing the audit report failed: %w", err)
		}
	}

	fmt.Println("\n=== Repair summary ===")
	fmt.Printf("Selected              : %d files, %s\n", len(candidates), humanBytes(targetBytes))
	fmt.Printf("Success verified      : %d\n", success)
	fmt.Printf("Skipped               : %d\n", skipped)
	fmt.Printf("Failed                : %d\n", failed)
	fmt.Printf("Freed verified        : %s (%d bytes)\n", humanBytes(freed), freed)
	if !cfg.noReport {
		fmt.Printf("Report                : %s\n", cfg.reportPath)
	}
	if failed > 0 {
		fmt.Println("RepairResult          : STOPPED_ON_FAILURE")
		return errors.New("repair stopped on an unexpected failure; successfully verified files remain dehydrated")
	}
	fmt.Println("RepairResult          : SUCCESS_VERIFIED")
	fmt.Println("Next                  : restart OneDrive, wait for sync, then run scan again")
	return nil
}

func discoverRepairCandidates(cfg repairConfig) ([]repairCandidate, int, *counters, error) {
	c := &counters{}
	tasks := make(chan fileTask, cfg.workers*128)
	type discoveryResult struct {
		candidate *repairCandidate
		review    bool
	}
	results := make(chan discoveryResult, cfg.workers*16)

	var wg sync.WaitGroup
	wg.Add(cfg.workers)
	for i := 0; i < cfg.workers; i++ {
		go func() {
			defer wg.Done()
			for t := range tasks {
				info, err := getPlaceholderInfo(t.path)
				c.queried.Add(1)
				if err != nil {
					var he *hresultError
					if errors.As(err, &he) && he.HR == hresultNotACloudFile {
						c.notCloud.Add(1)
						continue
					}
					c.queryErrors.Add(1)
					continue
				}
				if info.OnDiskDataSize <= 0 {
					continue
				}
				if ok, _ := isSafeCandidate(info); ok {
					results <- discoveryResult{candidate: &repairCandidate{path: t.path, logical: t.logical, onDisk: info.OnDiskDataSize}}
					continue
				}
				if shouldReviewResident(t.attributes, info) {
					results <- discoveryResult{review: true}
				}
			}
		}()
	}

	start := time.Now()
	stopProgress := make(chan struct{})
	doneProgress := make(chan struct{})
	go func() {
		defer close(doneProgress)
		ticker := time.NewTicker(cfg.progress)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start)
				files := c.filesSeen.Load()
				rate := float64(files) / maxSeconds(elapsed.Seconds())
				line := fmt.Sprintf("[discover %6s] files %s | queried %s | not-cloud %s | errors %s | %.0f files/s",
					elapsed.Round(time.Second), groupUint(files), groupUint(c.queried.Load()), groupUint(c.notCloud.Load()), groupUint(c.queryErrors.Load()), rate)
				fmt.Printf("\r%-220s", line)
			case <-stopProgress:
				clearProgressLine()
				return
			}
		}
	}()

	var candidates []repairCandidate
	reviews := 0
	doneCollect := make(chan struct{})
	go func() {
		defer close(doneCollect)
		for r := range results {
			if r.candidate != nil {
				candidates = append(candidates, *r.candidate)
			}
			if r.review {
				reviews++
			}
		}
	}()

	// Repair discovery is deliberately exhaustive: every file is queried. The repair path
	// is run rarely, and completeness matters more than the scan command's fast prefilter.
	enumErr := enumerate(cfg.root, true, c, tasks)
	close(tasks)
	wg.Wait()
	close(results)
	<-doneCollect
	close(stopProgress)
	<-doneProgress

	if enumErr != nil {
		return candidates, reviews, c, fmt.Errorf("directory enumeration failed: %w", enumErr)
	}
	return candidates, reviews, c, nil
}

func repairOneCandidate(index int, cand repairCandidate, root syncRootInfo) repairOutcome {
	out := repairOutcome{index: index, path: cand.path, logical: cand.logical, before: cand.onDisk, after: cand.onDisk}
	protected, err := openProtectedExclusive(cand.path)
	if err != nil {
		out.status = "SKIPPED_OPEN_FAILED"
		out.reason = err.Error()
		return out
	}
	defer closeProtectedHandle(protected)

	before, err := getPlaceholderInfoByHandle(protected)
	if err != nil {
		out.status = "FAILED_PRECHECK"
		out.reason = err.Error()
		return out
	}
	out.before = before.OnDiskDataSize
	out.pinState = before.PinState
	out.inSyncState = before.InSyncState
	out.modified = before.ModifiedDataSize

	if before.OnDiskDataSize == 0 {
		out.status = "ALREADY_DEHYDRATED"
		out.after = 0
		out.reason = "OnDiskDataSize already 0"
		return out
	}
	if before.SyncRootFileID != root.SyncRootFileID {
		out.status = "SKIPPED_WRONG_SYNC_ROOT"
		out.reason = fmt.Sprintf("SyncRootFileID=%d, expected %d", before.SyncRootFileID, root.SyncRootFileID)
		return out
	}
	if ok, why := isSafeCandidate(before); !ok {
		out.status = "SKIPPED_STATE_CHANGED"
		out.reason = why
		return out
	}
	if root.HydrationPrimary == cfHydrationAlwaysFull {
		out.status = "FAILED_POLICY"
		out.reason = "HydrationPrimary=ALWAYS_FULL"
		return out
	}

	if err := updatePlaceholderDehydrate(protected); err != nil {
		out.status = "FAILED_CFUPDATE"
		out.reason = err.Error()
		return out
	}

	after, err := getPlaceholderInfoByHandle(protected)
	if err != nil {
		out.status = "FAILED_VERIFY"
		out.reason = err.Error()
		return out
	}
	out.after = after.OnDiskDataSize
	if after.OnDiskDataSize != 0 {
		out.status = "FAILED_VERIFY"
		out.reason = fmt.Sprintf("OnDiskDataSize still %d", after.OnDiskDataSize)
		return out
	}
	if after.PinState != before.PinState || after.InSyncState != before.InSyncState || after.ModifiedDataSize != before.ModifiedDataSize {
		out.status = "FAILED_VERIFY"
		out.reason = fmt.Sprintf("placeholder state changed: Pin %d->%d InSync %d->%d Modified %d->%d",
			before.PinState, after.PinState, before.InSyncState, after.InSyncState, before.ModifiedDataSize, after.ModifiedDataSize)
		return out
	}
	if after.FileID != before.FileID || after.SyncRootFileID != before.SyncRootFileID || after.FileIdentityLen != before.FileIdentityLen {
		out.status = "FAILED_VERIFY"
		out.reason = "file identity metadata changed unexpectedly"
		return out
	}

	out.freed = before.OnDiskDataSize
	out.status = "SUCCESS_VERIFIED"
	out.reason = "OnDiskDataSize=0 and placeholder identity/state preserved"
	return out
}

func writeRepairReport(cfg repairConfig, outcomes []repairOutcome) error {
	if err := os.MkdirAll(filepath.Dir(cfg.reportPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(cfg.reportPath)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 1<<20)
	if _, err := bw.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	w := csv.NewWriter(bw)
	if err := w.Write([]string{
		"index", "status", "reason", "path", "logical_size", "before_on_disk", "after_on_disk",
		"freed_bytes", "pin_state_before", "in_sync_state_before", "modified_size_before",
	}); err != nil {
		return err
	}
	for _, o := range outcomes {
		row := []string{
			strconv.Itoa(o.index), o.status, o.reason, displayPath(o.path, cfg.redact),
			strconv.FormatInt(o.logical, 10), strconv.FormatInt(o.before, 10), strconv.FormatInt(o.after, 10),
			strconv.FormatInt(o.freed, 10), strconv.FormatUint(uint64(o.pinState), 10),
			strconv.FormatUint(uint64(o.inSyncState), 10), strconv.FormatInt(o.modified, 10),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}

func requireOneDriveStopped() error {
	pids, err := findProcessPIDs("OneDrive.exe")
	if err != nil {
		return fmt.Errorf("cannot verify whether OneDrive.exe is running: %w", err)
	}
	if len(pids) > 0 {
		return fmt.Errorf("OneDrive.exe is still running (PID %v); exit OneDrive completely before --apply", pids)
	}
	return nil
}
