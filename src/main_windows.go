//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	_, _, _ = procSetConsoleOutputCP.Call(65001)

	if len(os.Args) < 2 {
		usage(2)
	}

	switch strings.ToLower(os.Args[1]) {
	case "scan":
		cfg, err := parseScanArgs(os.Args[2:])
		if err != nil {
			fatal(fmt.Errorf("invalid arguments: %w", err))
		}
		if err := ensureCFAPI(); err != nil {
			fatal(err)
		}
		if err := runScan(cfg); err != nil {
			fatal(err)
		}
	case "inspect":
		if len(os.Args) != 3 {
			fatal(errors.New("inspect requires exactly one file path"))
		}
		if err := ensureCFAPI(); err != nil {
			fatal(err)
		}
		if err := runInspect(os.Args[2]); err != nil {
			fatal(err)
		}
	case "repair":
		cfg, err := parseRepairArgs(os.Args[2:])
		if err != nil {
			fatal(fmt.Errorf("invalid arguments: %w", err))
		}
		if err := ensureCFAPI(); err != nil {
			fatal(err)
		}
		if err := runRepair(cfg); err != nil {
			fatal(err)
		}
	case "--version", "-v", "version":
		fmt.Println("odhydrate", version)
	case "--help", "-h", "help":
		usage(0)
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		usage(2)
	}
}

func usage(code int) {
	out := os.Stdout
	if code != 0 {
		out = os.Stderr
	}
	fmt.Fprintf(out, `odhydrate %s — audit and reclaim stale OneDrive Files On-Demand data

Usage:
  odhydrate.exe inspect <file>
  odhydrate.exe scan <directory> [options]
  odhydrate.exe repair <directory> [--apply] [options]

scan options:
  --csv <path>        CSV report path (default: %%TEMP%%)
  --no-csv            Do not write a CSV report
  --workers <n>       CFAPI query workers (default: %d, max: 128)
  --progress <dur>    Progress refresh interval, e.g. 250ms or 1s (default: 500ms)
  --deep              Query every file instead of using cheap Cloud Files attribute hints
  --top <n>           Show the largest n resident candidates (default: 15, 0 disables)
  --redact            Replace paths in terminal output and CSV with short SHA-256 IDs

repair options:
  --apply             Perform dehydration; without this flag repair is read-only
  --workers <n>       Discovery workers (default: %d, max: 128)
  --progress <dur>    Discovery progress refresh interval (default: 500ms)
  --report <path>     Audit report path (default: %%TEMP%%)
  --no-report         Do not write the repair audit report
  --redact            Replace paths in terminal output and report with short SHA-256 IDs
  --limit <n>         Repair only the largest n safe candidates (0/default: all)

Safety model:
  * scan, inspect, and repair without --apply are read-only.
  * repair --apply refuses to run while OneDrive.exe is running.
  * Repair only targets UNPINNED + IN_SYNC placeholders with ModifiedDataSize=0.
  * Every file is re-checked under an exclusive CFAPI oplock before dehydration.
  * The tool uses CfUpdatePlaceholder(VERIFY_IN_SYNC | DEHYDRATE) and verifies
    OnDiskDataSize=0 plus key placeholder metadata afterward.
  * Any unexpected mutation/API failure stops the remaining batch.

Examples:
  odhydrate.exe scan "C:\Users\me\OneDrive"
  odhydrate.exe scan "C:\Users\me\OneDrive" --redact
  odhydrate.exe inspect "C:\Users\me\OneDrive\large-file.mkv"
  odhydrate.exe repair "C:\Users\me\OneDrive" --redact
  odhydrate.exe repair "C:\Users\me\OneDrive" --apply --limit 3 --redact
`, version, defaultWorkers(), defaultWorkers())
	os.Exit(code)
}

func parseScanArgs(args []string) (scanConfig, error) {
	cfg := scanConfig{workers: defaultWorkers(), progress: 500 * time.Millisecond, topN: 15}
	var positional []string
	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "--csv":
			i++
			if i >= len(args) {
				return cfg, errors.New("--csv requires a path")
			}
			cfg.csvPath = args[i]
		case "--no-csv":
			cfg.noCSV = true
		case "--workers":
			i++
			if i >= len(args) {
				return cfg, errors.New("--workers requires a value")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 || n > 128 {
				return cfg, errors.New("--workers must be an integer from 1 to 128")
			}
			cfg.workers = n
		case "--progress":
			i++
			if i >= len(args) {
				return cfg, errors.New("--progress requires a duration")
			}
			d, err := time.ParseDuration(args[i])
			if err != nil || d < 100*time.Millisecond {
				return cfg, errors.New("--progress must be a Go duration >=100ms, e.g. 500ms")
			}
			cfg.progress = d
		case "--deep":
			cfg.deep = true
		case "--top":
			i++
			if i >= len(args) {
				return cfg, errors.New("--top requires a value")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 || n > 1000 {
				return cfg, errors.New("--top must be an integer from 0 to 1000")
			}
			cfg.topN = n
		case "--redact":
			cfg.redact = true
		case "--help", "-h":
			usage(0)
		default:
			if strings.HasPrefix(a, "--") {
				return cfg, fmt.Errorf("unknown option %s", a)
			}
			positional = append(positional, a)
		}
	}
	if len(positional) != 1 {
		return cfg, errors.New("scan requires exactly one directory path")
	}
	root, err := filepath.Abs(positional[0])
	if err != nil {
		return cfg, err
	}
	st, err := os.Stat(root)
	if err != nil {
		return cfg, fmt.Errorf("cannot access scan root: %w", err)
	}
	if !st.IsDir() {
		return cfg, errors.New("scan target is not a directory")
	}
	cfg.root = root
	if !cfg.noCSV && cfg.csvPath == "" {
		cfg.csvPath = filepath.Join(os.TempDir(), "odhydrate-scan-"+time.Now().Format("20060102-150405")+".csv")
	}
	return cfg, nil
}

func parseRepairArgs(args []string) (repairConfig, error) {
	cfg := repairConfig{workers: defaultWorkers(), progress: 500 * time.Millisecond}
	var positional []string
	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "--apply":
			cfg.apply = true
		case "--workers":
			i++
			if i >= len(args) {
				return cfg, errors.New("--workers requires a value")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 || n > 128 {
				return cfg, errors.New("--workers must be an integer from 1 to 128")
			}
			cfg.workers = n
		case "--progress":
			i++
			if i >= len(args) {
				return cfg, errors.New("--progress requires a duration")
			}
			d, err := time.ParseDuration(args[i])
			if err != nil || d < 100*time.Millisecond {
				return cfg, errors.New("--progress must be a Go duration >=100ms, e.g. 500ms")
			}
			cfg.progress = d
		case "--report":
			i++
			if i >= len(args) {
				return cfg, errors.New("--report requires a path")
			}
			cfg.reportPath = args[i]
		case "--no-report":
			cfg.noReport = true
		case "--redact":
			cfg.redact = true
		case "--limit":
			i++
			if i >= len(args) {
				return cfg, errors.New("--limit requires a value")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return cfg, errors.New("--limit must be an integer >= 0")
			}
			cfg.limit = n
		case "--help", "-h":
			usage(0)
		default:
			if strings.HasPrefix(a, "--") {
				return cfg, fmt.Errorf("unknown option %s", a)
			}
			positional = append(positional, a)
		}
	}
	if len(positional) != 1 {
		return cfg, errors.New("repair requires exactly one directory path")
	}
	root, err := filepath.Abs(positional[0])
	if err != nil {
		return cfg, err
	}
	st, err := os.Stat(root)
	if err != nil {
		return cfg, fmt.Errorf("cannot access repair root: %w", err)
	}
	if !st.IsDir() {
		return cfg, errors.New("repair target is not a directory")
	}
	cfg.root = root
	if !cfg.noReport && cfg.reportPath == "" {
		cfg.reportPath = filepath.Join(os.TempDir(), "odhydrate-repair-"+time.Now().Format("20060102-150405")+".csv")
	}
	return cfg, nil
}

func defaultWorkers() int {
	n := runtime.NumCPU()
	if n < 2 {
		return 2
	}
	if n > 16 {
		return 16
	}
	return n
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
