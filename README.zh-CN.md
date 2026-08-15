# odhydrate

[![CI](https://img.shields.io/github/actions/workflow/status/zyf722/odhydrate/ci.yml?branch=main&label=CI&logo=githubactions&logoColor=white)](https://github.com/zyf722/odhydrate/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/zyf722/odhydrate?display_name=tag&logo=github)](https://github.com/zyf722/odhydrate/releases/latest)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Codex vibe coding](https://img.shields.io/badge/vibe_coded_with-Codex-412991?logo=data:image/svg%2bxml;base64,PHN2ZyB3aWR0aD0iMzg2IiBoZWlnaHQ9IjM4NiIgdmlld0JveD0iMTY1IDE2NSAzODYgMzg2IiBmaWxsPSJub25lIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgo8cGF0aCBkPSJNNTA4Ljc0OSAzMTcuMzk5QzUxNi43NzcgMjg3LjMxNCA1MDguOTkxIDI1My44ODQgNDg1LjM4OSAyMzAuMjgyQzQ2MS43ODggMjA2LjY4MSA0MjguMzYgMTk4Ljg5NSAzOTguMjczIDIwNi45MjNDMzc2LjIzMSAxODQuOTI4IDM0My4zOSAxNzQuOTU2IDMxMS4xNDggMTgzLjU5NkMyNzguOTA2IDE5Mi4yMzQgMjU1LjQ1IDIxNy4yOTIgMjQ3LjM2IDI0Ny4zNjFDMjE3LjI5MSAyNTUuNDUxIDE5Mi4yMzMgMjc4LjkxIDE4My41OTUgMzExLjE0OUMxNzQuOTU3IDM0My4zOTEgMTg0LjkyNyAzNzYuMjMyIDIwNi45MjQgMzk4LjI3NEMxOTguODk2IDQyOC4zNTkgMjA2LjY4MyA0NjEuNzg5IDIzMC4yODQgNDg1LjM5MUMyNTMuODg1IDUwOC45OTIgMjg3LjMxMyA1MTYuNzc5IDMxNy40MDEgNTA4Ljc1QzMzOS40NDIgNTMwLjc0NSAzNzIuMjg2IDU0MC43MTcgNDA0LjUyNSA1MzIuMDc5QzQzNi43NjcgNTIzLjQ0MSA0NjAuMjIzIDQ5OC4zODQgNDY4LjMxMyA0NjguMzE1QzQ5OC4zODMgNDYwLjIyNCA1MjMuNDQgNDM2Ljc2NiA1MzIuMDc4IDQwNC41MjZDNTQwLjcxNiAzNzIuMjg1IDUzMC43NDcgMzM5LjQ0MyA1MDguNzQ5IDMxNy40MDJWMzE3LjM5OVpNNDcwLjg5OSAyNDQuNzc2QzQ4Ni44OTIgMjYwLjc3IDQ5My40ODggMjgyLjYwMSA0OTAuNjg3IDMwMy40MTJMNDE1LjU3NyAyNjAuMDQ2QzQxMi40MTEgMjU4LjIxOCA0MDguNTA5IDI1OC4yMTggNDA1LjM0NSAyNjAuMDQ2TDMxNy40MDEgMzEwLjgyVjI3Ny41MjZDMzE3LjQwMSAyNzUuMTkxIDMxOC42NTIgMjczLjAwNSAzMjAuNjc2IDI3MS44MzdMMzg3LjY0NCAyMzMuMTc0QzQxNC4xNzggMjE4LjM1MyA0NDguMzQ2IDIyMi4yMjMgNDcwLjkwMSAyNDQuNzc2SDQ3MC44OTlaTTM1Ny44MzcgMzExLjE0NEwzOTguMjc1IDMzNC40OTFWMzgxLjE4NUwzNTcuODM3IDQwNC41MzJMMzE3LjM5OCAzODEuMTg1VjMzNC40OTFMMzU3LjgzNyAzMTEuMTQ0Wk0yNjQuNzc2IDI2OS42OTNDMjY1LjIwNyAyMzkuMzA1IDI4NS42NDQgMjExLjY0OSAzMTYuNDUzIDIwMy4zOTNDMzM4LjMgMTk3LjU0IDM2MC41MDUgMjAyLjc0NCAzNzcuMTI3IDIxNS41NzNMMzAyLjAxNCAyNTguOTM3QzI5OC44NDggMjYwLjc2NCAyOTYuODk4IDI2NC4xNDQgMjk2Ljg5OCAyNjcuNzk4VjM2OS4zNDZMMjY4LjA2NSAzNTIuNjk5QzI2Ni4wNDMgMzUxLjUzMSAyNjQuNzc2IDM0OS4zNTMgMjY0Ljc3NiAzNDcuMDE3VjI2OS42OTFWMjY5LjY5M1pNMjAzLjM5MSAzMTYuNDU0QzIwOS4yNDQgMjk0LjYwOCAyMjQuODU0IDI3Ny45NzggMjQ0LjI3NiAyNjkuOTk5VjM1Ni43M0MyNDQuMjc2IDM2MC4zODQgMjQ2LjIyNiAzNjMuNzYzIDI0OS4zOTIgMzY1LjU5MUwzMzcuMzM3IDQxNi4zNjVMMzA4LjUwMyA0MzMuMDEzQzMwNi40ODEgNDM0LjE4MSAzMDMuOTYxIDQzNC4xODggMzAxLjkzOSA0MzMuMDJMMjM0Ljk3MSAzOTQuMzU3QzIwOC44NjggMzc4Ljc4OSAxOTUuMTM4IDM0Ny4yNjEgMjAzLjM5MSAzMTYuNDU0Wk0yNDQuNzc1IDQ3MC45QzIyOC43ODEgNDU0LjkwNiAyMjIuMTg2IDQzMy4wNzUgMjI0Ljk4NiA0MTIuMjY0TDMwMC4wOTYgNDU1LjYzQzMwMy4yNjMgNDU3LjQ1NyAzMDcuMTY0IDQ1Ny40NTcgMzEwLjMyOCA0NTUuNjNMMzk4LjI3MyA0MDQuODU2VjQzOC4xNDlDMzk4LjI3MyA0NDAuNDg1IDM5Ny4wMjIgNDQyLjY3MSAzOTQuOTk3IDQ0My44MzlMMzI4LjAyOSA0ODIuNTAyQzMwMS40OTUgNDk3LjMyMiAyNjcuMzI3IDQ5My40NTIgMjQ0Ljc3MiA0NzAuOUgyNDQuNzc1Wk00NTAuODk3IDQ0NS45ODJDNDUwLjQ2NiA0NzYuMzcxIDQzMC4wMjkgNTA0LjAyNyAzOTkuMjIgNTEyLjI4M0MzNzcuMzczIDUxOC4xMzYgMzU1LjE2OCA1MTIuOTMyIDMzOC41NDcgNTAwLjEwMkw0MTMuNjU5IDQ1Ni43MzhDNDE2LjgyNiA0NTQuOTExIDQxOC43NzUgNDUxLjUzMiA0MTguNzc1IDQ0Ny44NzdWMzQ2LjMyOUw0NDcuNjA5IDM2Mi45NzdDNDQ5LjYzMSAzNjQuMTQ1IDQ1MC44OTcgMzY2LjMyMyA0NTAuODk3IDM2OC42NTlWNDQ1Ljk4NVY0NDUuOTgyWk01MTIuMjgyIDM5OS4yMjFDNTA2LjQyOSA0MjEuMDY4IDQ5MC44MTkgNDM3LjY5NyA0NzEuMzk3IDQ0NS42NzZWMzU4Ljk0NkM0NzEuMzk3IDM1NS4yOTIgNDY5LjQ0OCAzNTEuOTEyIDQ2Ni4yODEgMzUwLjA4NUwzNzguMzM2IDI5OS4zMTFMNDA3LjE3IDI4Mi42NjNDNDA5LjE5MiAyODEuNDk1IDQxMS43MTIgMjgxLjQ4NyA0MTMuNzM0IDI4Mi42NTVMNDgwLjcwMiAzMjEuMzE4QzUwNi44MDUgMzM2Ljg4NyA1MjAuNTM2IDM2OC40MTUgNTEyLjI4MiAzOTkuMjIxWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+Cg==)

[English](README.md) | **简体中文**

`odhydrate` 是一个保守的 Windows 工具，用来审计并释放 OneDrive Files On-Demand 占位文件中异常残留的本地数据。

它针对这样一种情况：文件已经是 `UNPINNED`、`IN_SYNC`，但在 OneDrive 执行“释放空间”、重启客户端或 reset 后，Windows Cloud Files 仍报告 `OnDiskDataSize > 0`。

> [!CAUTION]
> `repair --apply` 会修改 Cloud Files 占位文件状态。请先确认重要数据已完成同步或备份，先运行只读预演，并用较小的 `--limit` 开始处理。

## 主要特性

- `scan`：并发、只读地审计 OneDrive 目录树；
- `inspect`：查看单个文件的 Cloud Files 元数据；
- `repair`：发现安全候选；除非指定 `--apply`，否则保持只读；
- CSV 报告默认写到被扫描的 OneDrive 目录之外；
- `--redact` 会把终端和报告中的路径替换为短 SHA-256 标识；
- 无第三方 Go 依赖，发布包仅包含单个 Windows 可执行程序及文档。

修复候选必须同时满足：

- `PinState == UNPINNED`
- `InSyncState == IN_SYNC`
- `ModifiedDataSize == 0`
- `OnDiskDataSize > 0`

修复路径不会调用 `ReadFile`，也不会主动读取文件正文。

## 快速开始

### 下载

从最新的 [GitHub Release](https://github.com/zyf722/odhydrate/releases/latest) 下载 `odhydrate-windows-amd64-<版本>.zip` 和 `SHA256SUMS.txt`，校验后解压。

运行要求：Windows 10 1709 或更高版本、OneDrive Files On-Demand、x64 处理器。

### 扫描与只读预演

```powershell
.\odhydrate.exe scan "C:\Users\me\OneDrive" --redact
.\odhydrate.exe repair "C:\Users\me\OneDrive" --redact
```

不加 `--apply` 时，`repair` 完全只读。

### 完全退出 OneDrive

从通知区域退出 OneDrive；仅暂停同步并不够。然后确认进程已经停止：

```powershell
Get-Process OneDrive -ErrorAction SilentlyContinue
```

该命令应当没有输出。

### 处理小批量并复查

```powershell
.\odhydrate.exe repair "C:\Users\me\OneDrive" --apply --limit 3 --redact
```

确认结果正常后，再处理剩余候选：

```powershell
.\odhydrate.exe repair "C:\Users\me\OneDrive" --apply --redact
```

重新启动 OneDrive，等待同步完成，再次运行 `scan` 复查结果。

## 安全边界

每个文件脱水前，`odhydrate` 都会：

1. 确认同步提供程序是 OneDrive；
2. 拒绝 `HydrationPrimary=ALWAYS_FULL` 的同步根目录；
3. 要求 `OneDrive.exe` 已完全退出；
4. 重新确认全部安全候选条件；
5. 确认文件仍属于同一个同步根目录；
6. 使用 `CfOpenFileWithOplock(EXCLUSIVE | WRITE_ACCESS)` 获取独占 protected handle；
7. 同步调用 `CfUpdatePlaceholder(VERIFY_IN_SYNC | DEHYDRATE)`；
8. 验证 `OnDiskDataSize == 0`，并确认关键占位文件 identity/state 仍然有效；
9. 遇到任何意外修改或 API 失败时，停止剩余批量操作。

不满足全部安全候选条件的文件绝不会被自动处理。

## 命令

### `inspect <file>`

只读查看单个文件的 Cloud Files 元数据：

```powershell
.\odhydrate.exe inspect "C:\Users\me\OneDrive\video.mkv"
```

### `scan <directory>`

```text
--csv <path>        CSV 路径（默认 %TEMP%）
--no-csv            不生成 CSV
--workers <n>       CFAPI 查询并发数（默认 CPU 数，最多 16）
--progress <dur>    进度刷新间隔，例如 250ms 或 1s
--deep              不使用属性预筛，查询所有文件
--top <n>           显示驻留数据最大的 n 个候选（默认 15，0 关闭）
--redact            对终端和 CSV 中的路径做匿名化
```

### `repair <directory>`

```text
--apply             真正修改；缺省为只读预演
--workers <n>       发现阶段并发数
--progress <dur>    发现阶段进度刷新间隔
--report <path>     修复审计 CSV 路径（默认 %TEMP%）
--no-report         不生成修复报告
--redact            对终端和报告中的路径做匿名化
--limit <n>         只处理驻留空间最大的 n 个安全候选（0/缺省：全部）
```

## 报告

报告默认写到 `%TEMP%`，避免成为被扫描 OneDrive 目录树的一部分。`scan` 记录候选和错误元数据；`repair --apply` 记录每个尝试处理的文件，包括操作前后的 `OnDiskDataSize`、确认释放的字节数和操作结果。

## 开发

需要 Go 1.22+。源码位于 `src/`；仓库任务使用 `tools/build/` 下的 Go 程序实现，不依赖 Bash 或 PowerShell。

```powershell
go run ./tools/build check
go run ./tools/build actions
go run ./tools/build build
```

执行 `actions` 时需要确保 [`actionlint`](https://github.com/rhysd/actionlint) 位于 `PATH`。应用本身没有第三方 Go 依赖。

可用任务：

```text
check                检查格式、运行测试并 vet Windows amd64 目标
actions              使用 actionlint 检查 GitHub Actions workflow
build [version]      检查并构建 dist/odhydrate.exe
release <tag>        创建 Release ZIP 和 SHA256SUMS.txt
publish <tag>        通过 GitHub CLI 发布已准备的产物
```

## 技术依据

微软文档说明：

- `CF_PLACEHOLDER_STANDARD_INFO.OnDiskDataSize` 表示磁盘上实际存在的数据总字节数；
- `CF_UPDATE_FLAG_VERIFY_IN_SYNC` 会在占位文件不再同步时拒绝更新；
- `CF_UPDATE_FLAG_DEHYDRATE` 会对文件执行脱水，并要求调用方取得独占 handle；
- `CfOpenFileWithOplock(CF_OPEN_FILE_FLAG_EXCLUSIVE)` 可取得这种独占性，同时尽量减少与前台应用的冲突。

参考资料：

- [CfUpdatePlaceholder](https://learn.microsoft.com/windows/win32/api/cfapi/nf-cfapi-cfupdateplaceholder)
- [CF_UPDATE_FLAGS](https://learn.microsoft.com/windows/win32/api/cfapi/ne-cfapi-cf_update_flags)
- [CF_OPEN_FILE_FLAGS](https://learn.microsoft.com/windows/win32/api/cfapi/ne-cfapi-cf_open_file_flags)
- [CF_PLACEHOLDER_STANDARD_INFO](https://learn.microsoft.com/windows/win32/api/cfapi/ns-cfapi-cf_placeholder_standard_info)
- [File Attribute Constants](https://learn.microsoft.com/windows/win32/fileio/file-attribute-constants)

## 许可证及免责声明

[MIT](LICENSE)

`odhydrate` 与 Microsoft 无关联，也未获得其认可。修改操作有意限制在可识别的 OneDrive 同步根目录；只读检查也可能适用于其他 Cloud Files 提供程序。

本项目主要采用 Vibe Coding 方式开发，并由作者完成人工端到端测试；但代码可能未经全面人工审查，仍可能存在兼容性、稳定性或其他问题。使用源代码或发布产物，即视为你已知悉并接受相关风险。
