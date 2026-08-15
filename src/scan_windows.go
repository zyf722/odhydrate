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
	"syscall"
	"time"
)

func runInspect(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return errors.New("inspect only accepts files")
	}

	info, err := getPlaceholderInfo(abs)
	if err != nil {
		return err
	}

	fmt.Printf("Path              : %s\n", abs)
	fmt.Printf("LogicalSize       : %d (%s)\n", st.Size(), humanBytes(st.Size()))
	fmt.Printf("OnDiskDataSize    : %d (%s)\n", info.OnDiskDataSize, humanBytes(info.OnDiskDataSize))
	fmt.Printf("ValidatedDataSize : %d\n", info.ValidatedDataSize)
	fmt.Printf("ModifiedDataSize  : %d\n", info.ModifiedDataSize)
	fmt.Printf("PropertiesSize    : %d\n", info.PropertiesSize)
	fmt.Printf("PinState          : %d%s\n", info.PinState, enumLabelPin(info.PinState))
	fmt.Printf("InSyncState       : %d%s\n", info.InSyncState, enumLabelSync(info.InSyncState))
	fmt.Printf("FileIdentityLength: %d\n", info.FileIdentityLen)

	if ok, _ := isSafeCandidate(info); ok {
		fmt.Println("Classification    : SAFE_CANDIDATE")
	} else if info.OnDiskDataSize > 0 && info.PinState == cfPinStateUnpinned {
		fmt.Println("Classification    : REVIEW")
	} else {
		fmt.Println("Classification    : NORMAL/NOT-TARGET")
	}
	return nil
}

func runScan(cfg scanConfig) error {
	mode := "FAST"
	if cfg.deep {
		mode = "DEEP"
	}
	fmt.Printf("odhydrate %s  |  READ-ONLY  |  mode=%s  |  workers=%d\n", version, mode, cfg.workers)
	fmt.Printf("Root: %s\n", cfg.root)
	if cfg.noCSV {
		fmt.Println("CSV : disabled")
	} else {
		fmt.Printf("CSV : %s\n", cfg.csvPath)
	}
	fmt.Printf("Redact paths: %v\n", cfg.redact)
	fmt.Println("Only file attributes and CFAPI metadata are queried; file contents are not read.")
	fmt.Println()

	var csvFile *os.File
	var csvWriter *csv.Writer
	var bw *bufio.Writer
	if !cfg.noCSV {
		if err := os.MkdirAll(filepath.Dir(cfg.csvPath), 0o755); err != nil {
			return err
		}
		f, err := os.Create(cfg.csvPath)
		if err != nil {
			return err
		}
		csvFile = f
		bw = bufio.NewWriterSize(f, 1<<20)
		if _, err := bw.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			_ = f.Close()
			return err
		}
		csvWriter = csv.NewWriter(bw)
		if err := csvWriter.Write([]string{
			"status", "reason", "path", "logical_size", "on_disk_size", "validated_size",
			"modified_size", "properties_size", "pin_state", "in_sync_state", "attributes_hex", "reparse_tag_hex",
		}); err != nil {
			_ = f.Close()
			return err
		}
	}

	c := &counters{}
	tasks := make(chan fileTask, cfg.workers*128)
	results := make(chan scanResult, cfg.workers*16)
	doneCollector := make(chan struct{})
	var collectorErr error
	var collectorErrMu sync.Mutex
	var topMu sync.Mutex
	top := make([]candidate, 0, cfg.topN+1)

	go func() {
		defer close(doneCollector)
		for r := range results {
			if r.Status != "ERROR" {
				c.resident.Add(1)
				c.residentBytes.Add(r.Info.OnDiskDataSize)
				if r.Status == "SAFE_CANDIDATE" {
					c.safe.Add(1)
					c.safeBytes.Add(r.Info.OnDiskDataSize)
				} else {
					c.review.Add(1)
					c.reviewBytes.Add(r.Info.OnDiskDataSize)
				}
				if cfg.topN > 0 {
					topMu.Lock()
					top = append(top, candidate{status: r.Status, path: r.Path, logical: r.Logical, onDisk: r.Info.OnDiskDataSize})
					sort.Slice(top, func(i, j int) bool { return top[i].onDisk > top[j].onDisk })
					if len(top) > cfg.topN {
						top = top[:cfg.topN]
					}
					topMu.Unlock()
				}
			}

			if csvWriter != nil {
				if err := csvWriter.Write(scanResultCSV(r, cfg.redact)); err != nil {
					collectorErrMu.Lock()
					if collectorErr == nil {
						collectorErr = err
					}
					collectorErrMu.Unlock()
				}
			}
		}
		if csvWriter != nil {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				collectorErrMu.Lock()
				if collectorErr == nil {
					collectorErr = err
				}
				collectorErrMu.Unlock()
			}
		}
		if bw != nil {
			if err := bw.Flush(); err != nil {
				collectorErrMu.Lock()
				if collectorErr == nil {
					collectorErr = err
				}
				collectorErrMu.Unlock()
			}
		}
	}()

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
					results <- scanResult{Status: "ERROR", Reason: err.Error(), Path: t.path, Logical: t.logical, Attributes: t.attributes, ReparseTag: t.reparseTag}
					continue
				}
				if info.OnDiskDataSize <= 0 {
					continue
				}

				if ok, _ := isSafeCandidate(info); ok {
					results <- scanResult{
						Status: "SAFE_CANDIDATE", Reason: "UNPINNED + IN_SYNC + ModifiedDataSize=0 + OnDiskDataSize>0",
						Path: t.path, Logical: t.logical, Info: info, Attributes: t.attributes, ReparseTag: t.reparseTag,
					}
					continue
				}

				// Only surface non-safe resident placeholders as REVIEW when their attributes
				// indicate a cloud/offline/recall state. Intentionally PINNED local files are
				// not repair targets and should not flood the report in --deep mode.
				if shouldReviewResident(t.attributes, info) {
					reason := fmt.Sprintf("resident placeholder: PinState=%d InSyncState=%d ModifiedDataSize=%d", info.PinState, info.InSyncState, info.ModifiedDataSize)
					results <- scanResult{Status: "REVIEW", Reason: reason, Path: t.path, Logical: t.logical, Info: info, Attributes: t.attributes, ReparseTag: t.reparseTag}
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
				printScanProgress(c, start)
			case <-stopProgress:
				clearProgressLine()
				return
			}
		}
	}()

	enumErr := enumerate(cfg.root, cfg.deep, c, tasks)
	close(tasks)
	wg.Wait()
	close(results)
	<-doneCollector
	close(stopProgress)
	<-doneProgress

	if csvFile != nil {
		if err := csvFile.Close(); err != nil && collectorErr == nil {
			collectorErr = err
		}
	}
	if collectorErr != nil {
		return fmt.Errorf("failed to write CSV report: %w", collectorErr)
	}

	elapsed := time.Since(start)
	fmt.Println("=== Summary ===")
	fmt.Printf("Elapsed              : %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("Files scanned        : %d\n", c.filesSeen.Load())
	fmt.Printf("Directories scanned  : %d\n", c.dirsSeen.Load())
	fmt.Printf("Cloud-hint files     : %d\n", c.placeholders.Load())
	fmt.Printf("UNPINNED enum        : %d\n", c.unpinnedEnum.Load())
	fmt.Printf("CFAPI queried        : %d\n", c.queried.Load())
	fmt.Printf("Resident candidates  : %d\n", c.resident.Load())
	fmt.Printf("  SAFE_CANDIDATE     : %d\n", c.safe.Load())
	fmt.Printf("  REVIEW             : %d\n", c.review.Load())
	fmt.Printf("Resident bytes       : %s (%d bytes)\n", humanBytes(c.residentBytes.Load()), c.residentBytes.Load())
	fmt.Printf("Safe reclaimable     : %s (%d bytes)\n", humanBytes(c.safeBytes.Load()), c.safeBytes.Load())
	fmt.Printf("Review resident bytes: %s (%d bytes)\n", humanBytes(c.reviewBytes.Load()), c.reviewBytes.Load())
	fmt.Printf("Not-cloud hints      : %d\n", c.notCloud.Load())
	fmt.Printf("Errors               : query=%d enumeration=%d\n", c.queryErrors.Load(), c.enumErrors.Load())
	fmt.Printf("Skipped reparse dirs : %d\n", c.skippedReparseDirs.Load())
	if !cfg.noCSV {
		fmt.Printf("CSV                  : %s\n", cfg.csvPath)
	}

	if cfg.topN > 0 && len(top) > 0 {
		fmt.Printf("\nTop %d by on-disk data:\n", len(top))
		topMu.Lock()
		for i, x := range top {
			fmt.Printf("%2d. %-14s %10s / logical %-10s  %s\n", i+1, x.status, humanBytes(x.onDisk), humanBytes(x.logical), displayPath(x.path, cfg.redact))
		}
		topMu.Unlock()
	}

	if enumErr != nil {
		return fmt.Errorf("scan completed with an enumeration error: %w", enumErr)
	}
	return nil
}

func enumerate(root string, deep bool, c *counters, tasks chan<- fileTask) error {
	stack := []string{root}
	var firstErr error

	for len(stack) > 0 {
		dir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		c.dirsSeen.Add(1)

		pattern := filepath.Join(dir, "*")
		p, err := syscall.UTF16PtrFromString(win32Path(pattern))
		if err != nil {
			c.enumErrors.Add(1)
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", dir, err)
			}
			continue
		}
		var fd syscall.Win32finddata
		h, err := syscall.FindFirstFile(p, &fd)
		if err != nil {
			if errno, ok := err.(syscall.Errno); !ok || errno != syscall.ERROR_FILE_NOT_FOUND {
				c.enumErrors.Add(1)
				if firstErr == nil {
					firstErr = fmt.Errorf("%s: %w", dir, err)
				}
			}
			continue
		}

		for {
			name := syscall.UTF16ToString(fd.FileName[:])
			if name != "." && name != ".." && name != "" {
				full := filepath.Join(dir, name)
				attrs := fd.FileAttributes
				if attrs&fileAttributeDirectory != 0 {
					// Cloud Files directories may themselves be reparse points. Only skip
					// actual mount points/junctions and symbolic links to avoid escaping the root.
					if attrs&fileAttributeReparse != 0 &&
						(fd.Reserved0 == ioReparseTagMountPoint || fd.Reserved0 == ioReparseTagSymlink) {
						c.skippedReparseDirs.Add(1)
					} else {
						stack = append(stack, full)
					}
				} else {
					c.filesSeen.Add(1)
					if attrs&fileAttributeUnpinned != 0 {
						c.unpinnedEnum.Add(1)
					}
					cloudHint := hasCloudHint(attrs)
					if deep || cloudHint {
						c.placeholders.Add(1)
						logical := int64(uint64(fd.FileSizeHigh)<<32 | uint64(fd.FileSizeLow))
						tasks <- fileTask{path: full, logical: logical, attributes: attrs, reparseTag: fd.Reserved0}
					}
				}
			}

			err = syscall.FindNextFile(h, &fd)
			if err != nil {
				break
			}
		}
		_ = syscall.FindClose(h)
		if errno, ok := err.(syscall.Errno); ok && errno == syscall.ERROR_NO_MORE_FILES {
			continue
		}
		if err != nil {
			c.enumErrors.Add(1)
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", dir, err)
			}
		}
	}
	return firstErr
}

func hasCloudHint(attrs uint32) bool {
	return attrs&(fileAttributeUnpinned|fileAttributeOffline|fileAttributeRecallOnOpen|fileAttributeRecallOnDataAccess|fileAttributeReparse) != 0
}

func shouldReviewResident(attrs uint32, info placeholderInfo) bool {
	if info.PinState == cfPinStatePinned || attrs&fileAttributePinned != 0 {
		return false
	}
	if info.PinState == cfPinStateUnpinned {
		return true
	}
	return attrs&(fileAttributeUnpinned|fileAttributeOffline|fileAttributeRecallOnOpen|fileAttributeRecallOnDataAccess) != 0
}

func scanResultCSV(r scanResult, redact bool) []string {
	return []string{
		r.Status,
		r.Reason,
		displayPath(r.Path, redact),
		strconv.FormatInt(r.Logical, 10),
		strconv.FormatInt(r.Info.OnDiskDataSize, 10),
		strconv.FormatInt(r.Info.ValidatedDataSize, 10),
		strconv.FormatInt(r.Info.ModifiedDataSize, 10),
		strconv.FormatInt(r.Info.PropertiesSize, 10),
		strconv.FormatUint(uint64(r.Info.PinState), 10),
		strconv.FormatUint(uint64(r.Info.InSyncState), 10),
		fmt.Sprintf("0x%08X", r.Attributes),
		fmt.Sprintf("0x%08X", r.ReparseTag),
	}
}

func printScanProgress(c *counters, start time.Time) {
	elapsed := time.Since(start)
	files := c.filesSeen.Load()
	rate := float64(files) / maxSeconds(elapsed.Seconds())
	line := fmt.Sprintf(
		"[%8s] files %s | dirs %s | hints %s | queried %s | resident %s (safe %s) | %s | %.0f files/s | err %s",
		elapsed.Round(time.Second),
		groupUint(files),
		groupUint(c.dirsSeen.Load()),
		groupUint(c.placeholders.Load()),
		groupUint(c.queried.Load()),
		groupUint(c.resident.Load()),
		groupUint(c.safe.Load()),
		humanBytes(c.safeBytes.Load()),
		rate,
		groupUint(c.queryErrors.Load()+c.enumErrors.Load()),
	)
	fmt.Printf("\r%-220s", line)
}

func clearProgressLine() {
	fmt.Printf("\r%-220s\r", "")
}
