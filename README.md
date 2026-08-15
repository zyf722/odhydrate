# odhydrate

[![CI](https://img.shields.io/github/actions/workflow/status/zyf722/odhydrate/ci.yml?branch=main&label=CI&logo=githubactions&logoColor=white)](https://github.com/zyf722/odhydrate/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/zyf722/odhydrate?display_name=tag&logo=github)](https://github.com/zyf722/odhydrate/releases/latest)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Codex vibe coding](https://img.shields.io/badge/vibe_coded_with-Codex-412991?logo=data:image/svg%2bxml;base64,PHN2ZyB3aWR0aD0iMzg2IiBoZWlnaHQ9IjM4NiIgdmlld0JveD0iMTY1IDE2NSAzODYgMzg2IiBmaWxsPSJub25lIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgo8cGF0aCBkPSJNNTA4Ljc0OSAzMTcuMzk5QzUxNi43NzcgMjg3LjMxNCA1MDguOTkxIDI1My44ODQgNDg1LjM4OSAyMzAuMjgyQzQ2MS43ODggMjA2LjY4MSA0MjguMzYgMTk4Ljg5NSAzOTguMjczIDIwNi45MjNDMzc2LjIzMSAxODQuOTI4IDM0My4zOSAxNzQuOTU2IDMxMS4xNDggMTgzLjU5NkMyNzguOTA2IDE5Mi4yMzQgMjU1LjQ1IDIxNy4yOTIgMjQ3LjM2IDI0Ny4zNjFDMjE3LjI5MSAyNTUuNDUxIDE5Mi4yMzMgMjc4LjkxIDE4My41OTUgMzExLjE0OUMxNzQuOTU3IDM0My4zOTEgMTg0LjkyNyAzNzYuMjMyIDIwNi45MjQgMzk4LjI3NEMxOTguODk2IDQyOC4zNTkgMjA2LjY4MyA0NjEuNzg5IDIzMC4yODQgNDg1LjM5MUMyNTMuODg1IDUwOC45OTIgMjg3LjMxMyA1MTYuNzc5IDMxNy40MDEgNTA4Ljc1QzMzOS40NDIgNTMwLjc0NSAzNzIuMjg2IDU0MC43MTcgNDA0LjUyNSA1MzIuMDc5QzQzNi43NjcgNTIzLjQ0MSA0NjAuMjIzIDQ5OC4zODQgNDY4LjMxMyA0NjguMzE1QzQ5OC4zODMgNDYwLjIyNCA1MjMuNDQgNDM2Ljc2NiA1MzIuMDc4IDQwNC41MjZDNTQwLjcxNiAzNzIuMjg1IDUzMC43NDcgMzM5LjQ0MyA1MDguNzQ5IDMxNy40MDJWMzE3LjM5OVpNNDcwLjg5OSAyNDQuNzc2QzQ4Ni44OTIgMjYwLjc3IDQ5My40ODggMjgyLjYwMSA0OTAuNjg3IDMwMy40MTJMNDE1LjU3NyAyNjAuMDQ2QzQxMi40MTEgMjU4LjIxOCA0MDguNTA5IDI1OC4yMTggNDA1LjM0NSAyNjAuMDQ2TDMxNy40MDEgMzEwLjgyVjI3Ny41MjZDMzE3LjQwMSAyNzUuMTkxIDMxOC42NTIgMjczLjAwNSAzMjAuNjc2IDI3MS44MzdMMzg3LjY0NCAyMzMuMTc0QzQxNC4xNzggMjE4LjM1MyA0NDguMzQ2IDIyMi4yMjMgNDcwLjkwMSAyNDQuNzc2SDQ3MC44OTlaTTM1Ny44MzcgMzExLjE0NEwzOTguMjc1IDMzNC40OTFWMzgxLjE4NUwzNTcuODM3IDQwNC41MzJMMzE3LjM5OCAzODEuMTg1VjMzNC40OTFMMzU3LjgzNyAzMTEuMTQ0Wk0yNjQuNzc2IDI2OS42OTNDMjY1LjIwNyAyMzkuMzA1IDI4NS42NDQgMjExLjY0OSAzMTYuNDUzIDIwMy4zOTNDMzM4LjMgMTk3LjU0IDM2MC41MDUgMjAyLjc0NCAzNzcuMTI3IDIxNS41NzNMMzAyLjAxNCAyNTguOTM3QzI5OC44NDggMjYwLjc2NCAyOTYuODk4IDI2NC4xNDQgMjk2Ljg5OCAyNjcuNzk4VjM2OS4zNDZMMjY4LjA2NSAzNTIuNjk5QzI2Ni4wNDMgMzUxLjUzMSAyNjQuNzc2IDM0OS4zNTMgMjY0Ljc3NiAzNDcuMDE3VjI2OS42OTFWMjY5LjY5M1pNMjAzLjM5MSAzMTYuNDU0QzIwOS4yNDQgMjk0LjYwOCAyMjQuODU0IDI3Ny45NzggMjQ0LjI3NiAyNjkuOTk5VjM1Ni43M0MyNDQuMjc2IDM2MC4zODQgMjQ2LjIyNiAzNjMuNzYzIDI0OS4zOTIgMzY1LjU5MUwzMzcuMzM3IDQxNi4zNjVMMzA4LjUwMyA0MzMuMDEzQzMwNi40ODEgNDM0LjE4MSAzMDMuOTYxIDQzNC4xODggMzAxLjkzOSA0MzMuMDJMMjM0Ljk3MSAzOTQuMzU3QzIwOC44NjggMzc4Ljc4OSAxOTUuMTM4IDM0Ny4yNjEgMjAzLjM5MSAzMTYuNDU0Wk0yNDQuNzc1IDQ3MC45QzIyOC43ODEgNDU0LjkwNiAyMjIuMTg2IDQzMy4wNzUgMjI0Ljk4NiA0MTIuMjY0TDMwMC4wOTYgNDU1LjYzQzMwMy4yNjMgNDU3LjQ1NyAzMDcuMTY0IDQ1Ny40NTcgMzEwLjMyOCA0NTUuNjNMMzk4LjI3MyA0MDQuODU2VjQzOC4xNDlDMzk4LjI3MyA0NDAuNDg1IDM5Ny4wMjIgNDQyLjY3MSAzOTQuOTk3IDQ0My44MzlMMzI4LjAyOSA0ODIuNTAyQzMwMS40OTUgNDk3LjMyMiAyNjcuMzI3IDQ5My40NTIgMjQ0Ljc3MiA0NzAuOUgyNDQuNzc1Wk00NTAuODk3IDQ0NS45ODJDNDUwLjQ2NiA0NzYuMzcxIDQzMC4wMjkgNTA0LjAyNyAzOTkuMjIgNTEyLjI4M0MzNzcuMzczIDUxOC4xMzYgMzU1LjE2OCA1MTIuOTMyIDMzOC41NDcgNTAwLjEwMkw0MTMuNjU5IDQ1Ni43MzhDNDE2LjgyNiA0NTQuOTExIDQxOC43NzUgNDUxLjUzMiA0MTguNzc1IDQ0Ny44NzdWMzQ2LjMyOUw0NDcuNjA5IDM2Mi45NzdDNDQ5LjYzMSAzNjQuMTQ1IDQ1MC44OTcgMzY2LjMyMyA0NTAuODk3IDM2OC42NTlWNDQ1Ljk4NVY0NDUuOTgyWk01MTIuMjgyIDM5OS4yMjFDNTA2LjQyOSA0MjEuMDY4IDQ5MC44MTkgNDM3LjY5NyA0NzEuMzk3IDQ0NS42NzZWMzU4Ljk0NkM0NzEuMzk3IDM1NS4yOTIgNDY5LjQ0OCAzNTEuOTEyIDQ2Ni4yODEgMzUwLjA4NUwzNzguMzM2IDI5OS4zMTFMNDA3LjE3IDI4Mi42NjNDNDA5LjE5MiAyODEuNDk1IDQxMS43MTIgMjgxLjQ4NyA0MTMuNzM0IDI4Mi42NTVMNDgwLjcwMiAzMjEuMzE4QzUwNi44MDUgMzM2Ljg4NyA1MjAuNTM2IDM2OC40MTUgNTEyLjI4MiAzOTkuMjIxWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+Cg==)

**English** | [简体中文](README.zh-CN.md)

`odhydrate` is a conservative Windows utility for auditing and reclaiming stale local data in OneDrive Files On-Demand placeholders.

It targets files that are already `UNPINNED` and `IN_SYNC`, but for which Windows Cloud Files still reports `OnDiskDataSize > 0` after OneDrive's **Free up space**, a client restart, or a reset.

> [!CAUTION]
> `repair --apply` changes Cloud Files placeholder state. Confirm that important data is synchronized or backed up, run the read-only preview first, and start with a small `--limit`.

## Key features

- `scan` performs a read-only, concurrent audit of a OneDrive directory tree.
- `inspect` displays Cloud Files metadata for a single file.
- `repair` discovers safe candidates and remains read-only unless `--apply` is specified.
- CSV reports are written outside the scanned OneDrive tree by default.
- `--redact` replaces paths in terminal output and reports with short SHA-256 identifiers.
- No third-party Go dependencies; the release is a single Windows executable.

A repair candidate must satisfy all of the following:

- `PinState == UNPINNED`
- `InSyncState == IN_SYNC`
- `ModifiedDataSize == 0`
- `OnDiskDataSize > 0`

The repair path does not call `ReadFile` or intentionally read file contents.

## Quickstart

### Download

Download `odhydrate-windows-amd64-<version>.zip` and `SHA256SUMS.txt` from the latest [GitHub Release](https://github.com/zyf722/odhydrate/releases/latest), then verify and extract the archive.

Requirements: Windows 10 version 1709 or later, OneDrive Files On-Demand, and an x64 processor.

### Audit and preview

```powershell
.\odhydrate.exe scan "C:\Users\me\OneDrive" --redact
.\odhydrate.exe repair "C:\Users\me\OneDrive" --redact
```

Without `--apply`, `repair` is read-only.

### Exit OneDrive completely

Exit OneDrive from the notification area; pausing synchronization is not sufficient. Verify that it has stopped:

```powershell
Get-Process OneDrive -ErrorAction SilentlyContinue
```

The command should produce no output.

### Apply a small batch and verify

```powershell
.\odhydrate.exe repair "C:\Users\me\OneDrive" --apply --limit 3 --redact
```

If the result is clean, process the remaining candidates:

```powershell
.\odhydrate.exe repair "C:\Users\me\OneDrive" --apply --redact
```

Restart OneDrive, wait for synchronization, and run `scan` again to verify the result.

## Safety model

Before dehydrating each file, `odhydrate`:

1. requires the sync provider to be recognized as OneDrive;
2. rejects sync roots with `HydrationPrimary=ALWAYS_FULL`;
3. requires `OneDrive.exe` to be completely stopped;
4. re-checks all safe-candidate conditions;
5. verifies that the file still belongs to the same sync root;
6. obtains an exclusive protected handle with `CfOpenFileWithOplock(EXCLUSIVE | WRITE_ACCESS)`;
7. calls `CfUpdatePlaceholder(VERIFY_IN_SYNC | DEHYDRATE)` synchronously;
8. verifies that `OnDiskDataSize == 0` and key placeholder identity/state fields remain valid;
9. stops the remaining batch after any unexpected mutation or API failure.

Files that do not meet every safe-candidate condition are never automatically repaired.

## Commands

### `inspect <file>`

Display read-only Cloud Files metadata for one file:

```powershell
.\odhydrate.exe inspect "C:\Users\me\OneDrive\video.mkv"
```

### `scan <directory>`

```text
--csv <path>        CSV output path (default: %TEMP%)
--no-csv            Disable CSV output
--workers <n>       CFAPI query workers (default: CPU count, capped at 16)
--progress <dur>    Progress refresh interval, e.g. 250ms or 1s
--deep              Query every file instead of using attribute hints
--top <n>           Show the largest n candidates (default: 15; 0 disables)
--redact            Hash paths in output and CSV
```

### `repair <directory>`

```text
--apply             Perform changes; omitted means dry-run
--workers <n>       Discovery workers
--progress <dur>    Discovery progress refresh interval
--report <path>     Repair audit CSV path (default: %TEMP%)
--no-report         Disable the repair report
--redact            Hash paths in output and report
--limit <n>         Repair only the largest n safe candidates (0/default: all)
```

## Reports

Reports are written to `%TEMP%` by default so they do not become part of the scanned OneDrive tree. `scan` records candidate and error metadata. `repair --apply` records every attempted file, including before/after `OnDiskDataSize`, verified freed bytes, and the operation result.

## Development

Requirements: Go 1.22+. Source code lives in `src/`; repository tasks are implemented in Go under `tools/build/` and do not depend on Bash or PowerShell.

```powershell
go run ./tools/build check
go run ./tools/build actions
go run ./tools/build build
```

Run `actions` with [`actionlint`](https://github.com/rhysd/actionlint) available on `PATH`. The application itself has no third-party Go dependencies.

Available tasks:

```text
check                Check formatting, run tests, and vet Windows amd64
actions              Lint GitHub Actions workflows with actionlint
build [version]      Check and build dist/odhydrate.exe
release <tag>        Create the release ZIP and SHA256SUMS.txt
publish <tag>        Publish prepared assets through the GitHub CLI
```

## Technical basis

Microsoft documents that:

- `CF_PLACEHOLDER_STANDARD_INFO.OnDiskDataSize` is the total number of bytes present on disk;
- `CF_UPDATE_FLAG_VERIFY_IN_SYNC` rejects the update if the placeholder is no longer in sync;
- `CF_UPDATE_FLAG_DEHYDRATE` dehydrates a file and requires an exclusive handle;
- `CfOpenFileWithOplock(CF_OPEN_FILE_FLAG_EXCLUSIVE)` provides that exclusivity while minimizing conflicts with foreground applications.

References:

- [CfUpdatePlaceholder](https://learn.microsoft.com/windows/win32/api/cfapi/nf-cfapi-cfupdateplaceholder)
- [CF_UPDATE_FLAGS](https://learn.microsoft.com/windows/win32/api/cfapi/ne-cfapi-cf_update_flags)
- [CF_OPEN_FILE_FLAGS](https://learn.microsoft.com/windows/win32/api/cfapi/ne-cfapi-cf_open_file_flags)
- [CF_PLACEHOLDER_STANDARD_INFO](https://learn.microsoft.com/windows/win32/api/cfapi/ns-cfapi-cf_placeholder_standard_info)
- [File Attribute Constants](https://learn.microsoft.com/windows/win32/fileio/file-attribute-constants)

## License and disclaimer

[MIT](LICENSE)

`odhydrate` is not affiliated with or endorsed by Microsoft. Mutation is intentionally restricted to recognized OneDrive sync roots; read-only inspection may also work with other Cloud Files providers.

This project was developed primarily through vibe coding and has undergone manual end-to-end testing by the author. The code may not have received a comprehensive human review and may still contain compatibility, stability, or other issues. By using the source or a release artifact, you acknowledge and accept the associated risks.
