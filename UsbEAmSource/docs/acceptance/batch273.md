# 批次 273 · webview2_process.go 整文件落地（+6 [S]，+1 [P] 存根，+7 FUNCS）

## 目标

1. **整文件落地 `webview2_process.go`**（44 个未落地文件差集之一）：WebView2 进程快照归一化链
   6 函数全 `[S]`，依赖的平台层枚举 `inspectWebView2Processes`（0x1409debc0, 3104B）落 `[P]` 存根。
2. 纯逻辑链：kind 权重分派 → 进程类型提取 → kind 归一化 → 单进程归一化 → 快照归一化（累计+排序）
   → 入口转发方法。

## 基线 / 收口

| 指标 | 基线（batch 272 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2874 | 2881 |
| MARKED | 2874 | 2881 |
| S | 1355 | 1361 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 40 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1391 | 1397 |
| USABLE | 1391 | 1397 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run WebView2` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1397  FUNCS=2881  MARKED=2881  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1361  USABLE=1397
```

分项自洽：`S + S-inline + S-sig + P = 1361 + 36 + 1443 + 41 = 2881 = FUNCS`。
S 1355→1361（+6）、P 40→41（+1）、FUNCS 2874→2881（+7）、FAITHFUL 1391→1397（+6）、UNMARKED=0 保持。

账目：新建 `webview2_process.go` 6 函数 `[S]`（+6 S +6 FUNCS +6 FAITHFUL）；
`webview2_process_windows.go` 追加 `inspectWebView2Processes` `[P]` 存根（+1 P +1 FUNCS）。
闭包 `normalizeWebView2ProcessSnapshotfunc1` 内联于 sort.SliceStable 匿名函数（count 不统计匿名）。

## G3 行为（asm 逐地址实证）

### 3.1 webView2ProcessKindSortWeight [S 0x1409deac0, 256B]

`TrimSpace` → `ToLower` → 立即数/字节分派：browser=0、renderer=1、gpu-process=2、utility=3、其余=4。
常量逐字节实证：browser(0x776f7262/0x6573/0x72)、utility(0x6c697475/0x7469/0x79)、
renderer(0x72657265646e6572)、gpu-process(0x636f72702d757067/0x7365/0x73)。

### 3.2 extractWebView2ProcessType [S 0x1409de960, 352B]

`ToLower` → `Index("--type=")`（7B needle @0x140C395A0）。未命中返回 `"browser"`（7B @0x140C39599）。
命中则从 `i+7` 起跳过非空白（tab/LF/CR/space 判定），切到首个空白，`TrimSpace` 后再
`Trim(s, "\"")`（1B cutset @0x1411CAC88）。返回 token（含 `--type= --x` 得空串）。

### 3.3 normalizeWebView2ProcessKind [S 0x1409de840, 288B]

双参 `(kind, commandLine string) string`（调用点 0x1409de533 同时压 kind ptr/len + commandLine ptr/len）。
`TrimSpace`→`ToLower`；空串则 `extractWebView2ProcessType(commandLine)`。之后 `""`/`"browser"`
归一为 `"browser"`（7B 常量 @0x140C39599），其余原样。

### 3.4 normalizeWebView2ProcessInfo [S 0x1409de500, 832B]

`(info WebView2ProcessInfo) WebView2ProcessInfo`（80B 值传，字段偏移与 types_misc.go 一致：
ProcessID+0x00、Kind+0x08、WorkingSetBytes+0x18、PrivateBytes+0x20、WorkingSetMB+0x28、
PrivateMB+0x30、三 bool+0x38/39/3a、CommandLine+0x40）。

- `Kind = normalizeWebView2ProcessKind(Kind, CommandLine)`。
- `WorkingSetMB = float64(WorkingSetBytes)/1024/1024`、`PrivateMB = float64(PrivateBytes)/1024/1024`
  （float64 常量 0.0009765625 连乘两次 @0x1411CD608，实证为 ×1/1024 两次）。
- 三 bool「已置位保持，否则 `strings.Contains(ToLower(commandLine), 子串)`」：
  `--js-flags=`(11B @0x140C47699)、`--disable-features=`(19B @0x140C5BCED)、
  `--renderer-process-limit=`(25B @0x140C66E11)。

### 3.5 normalizeWebView2ProcessSnapshot [S 0x1409dde00, 1472B]

`(snapshot WebView2ProcessSnapshot) WebView2ProcessSnapshot`（72B：UserDataDir+0x00、Processes+0x10、
TotalWorkingSetBytes+0x28、TotalPrivateBytes+0x30、两 MB+0x38/40）。

- Processes nil → 置空切片（zerobase，0x1409dde9f lea rip+0x127cb1a）。
- 逐元素 `normalizeWebView2ProcessInfo` 并累计 WorkingSetBytes/PrivateBytes（+0x18/+0x20）。
- `sort.SliceStable`（0x1401af8e0）：闭包 func1（0x1409de3c0）按 `webView2ProcessKindSortWeight` 升序，
  权重相等时按 `ProcessID`（+0x00）升序（cmp/setg 语义实证）。
- `TotalWorkingSetMB/TotalPrivateMB = 累计量/1024/1024`（同 3.4 常量，×100.0 后 ÷100.0 的乘除抵消实证）。

### 3.6 InspectWebView2Processes [S 0x1409ddc60, 416B]

`(bs *BootstrapService) WebView2ProcessSnapshot`。`bs == nil`（test rax,rax）→ userDataDir=""；
否则 `bs.workspaceSnapshot().WebView2Dir`（WorkspaceLayout 偏移 +0x80 的 string，与
types_workspace.go WebView2Dir 字段一致）。随后 `inspectWebView2Processes(userDataDir)`。

### 3.7 inspectWebView2Processes [P 0x1409debc0, 3104B]

平台层进程枚举（CreateToolhelp32Snapshot/OpenProcess/读命令行链），签名已实证
`(userDataDir string) WebView2ProcessSnapshot`。本批落零值存根，待专项批次还原。

## G4 独立复核

`webview2_process.go`：新建，6 函数 `[S]` + sort 闭包内联。
`webview2_process_windows.go`：追加 `inspectWebView2Processes` `[P]` 存根（零值返回）。
`webview2_process_test.go`：新建，表驱动覆盖 kind 权重、类型提取（含引号/大写/空 token）、
kind 归一化、单进程 MB 换算与 bool 子串、快照排序（权重升序+ProcessID 平局）、nil 归一化。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

WebView2 进程快照归一化链（纯逻辑层）整文件闭环：6 函数全 `[S]`，常量与浮点换算全部 rodata/asm
实证。平台层枚举 `inspectWebView2Processes`（3104B）仍 `[P]`，属 webview2_process_windows.go
后续专项（依赖 Windows 进程枚举 API）。P 40→41。FUNCS 2881/4754 = 60.60%。未落地文件差集 44→43。
