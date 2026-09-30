# 批次 264 · 短函数 +3 [S]（滚动右键状态 + WebView2 检视辅助）

## 目标

从 gap_aggregate 的 96B 段落地 3 个高确定性短函数（无需新类型、无未落地依赖）。

## 基线 / 收口

| 指标 | 基线（batch 263 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2854 | 2857 |
| S | 1322 | 1325 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1358 | 1361 |
| 真函数（S+S-inline+S-sig） | 2814（59.29%） | 2817（59.34%） |

`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

## 本批落地（+3 [S]）

1. `readScreenshotScrollingRightButtonState() (bool, bool)` `[S 0x1409a3320]`
   —— `GetAsyncKeyState(VK_RBUTTON=2)`；`bt eax,0xf; setb al` 取 bit15 = 当前是否按下（第 1 返回值）；
   `and ebx,1` 取 bit0 = 自上次调用后是否按过（第 2 返回值）。新增 proc `procGetAsyncKeyState`。
   文件 `screenshot_scroll_windows.go`（新建）。

2. `resolveWebView2ProcessInspectUserDataDir(dir string) (string, error)` `[S 0x1409df7e0]`
   —— TrimSpace → 空串返回 ("", nil)；否则交 filepath.Abs（Abs 自身返回 (string, error)）。

3. `resolveWebView2ProcessInspectHostExeName() string` `[S 0x1409df840]`
   —— os.Executable 出错 → 默认名；否则 filepath.Base + TrimSpace，空串亦回落默认名（错误被吞并回落）。
   文件 `webview2_process_windows.go`（新建）。

## proc / 字符串内存实证

| 事项 | 实证结果 |
|---|---|
| 右键 proc 槽 0x141BC1CE8 | GetAsyncKeyState |
| 默认宿主名（19B） | `UsbEAm_Launcher.exe` @0x140C5BD00（两分支 strA/strB 同址，交叉验证一致） |

## 说明

本批为「整洁移交」前的最后一批函数落地。落地后即转入：过程产物移出版本控制、
CI 门禁统一 `-tags production`、push、远端 CI 验证、交接文档。
