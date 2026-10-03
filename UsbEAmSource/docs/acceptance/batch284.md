# 批次 284 · 光标定位 + 更新只读句柄 +3（全 [S]，FUNCS 2933）

## 目标

还原截图滚动的鼠标定位（setScreenshotCursorPosition）与启动器更新辅助域的两个只读句柄
打开函数（openLauncherUpdateReadOnlyHandle / openLauncherUpdateReadOnlyInheritedHandle，
新建 launcherupdate_helper_windows.go）。

## 基线 / 收口

| 指标 | 基线（batch 283 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2930 | 2933 |
| MARKED | 2930 | 2933 |
| S | 1407 | 1410 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1444 | 1447 |
| USABLE | 1445 | 1448 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1447  FUNCS=2933  MARKED=2933  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1410  USABLE=1448
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1410 + 37 + 1 + 1444 + 41 = 2933 = FUNCS`。
S 1407→1410（+3）、FUNCS 2930→2933（+3）、FAITHFUL 1444→1447（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 setScreenshotCursorPosition [S 0x1409a2e20, 256B]

`(x, y int) error`。SetCursorPos(x,y) 返回非 0 → nil；否则 err 为 nil 或
`errors.Is(err, syscall.Errno(0))` → `errors.New("设置鼠标位置失败")`，其余透传 err。
（procSetCursorPos 复用 windowmanagement_windows.go 既有声明，不再重复定义。）

### 3.2 openLauncherUpdateReadOnlyHandle [S 0x1408c4d20, 128B]

`(path string) (windows.Handle, error)`。UTF16PtrFromString 失败 → (0,err)；
CreateFile(GENERIC_READ, SHARE_READ, nil, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL|
FILE_FLAG_SEQUENTIAL_SCAN, 0)。

### 3.3 openLauncherUpdateReadOnlyInheritedHandle [S 0x1408c4da0, 160B]

`(path string) (windows.Handle, error)`。openLauncherUpdateReadOnlyHandle 失败 → (0,err)；
SetHandleInformation(handle, HANDLE_FLAG_INHERIT, HANDLE_FLAG_INHERIT) 失败 →
CloseHandle(handle) 后 (0,err)；成功 → (handle,nil)。

## G4 独立复核

- `screenshot_cursor_windows.go`：+`setScreenshotCursorPosition`（复用既有 procSetCursorPos）。
- `launcherupdate_helper_windows.go`（NEW）：+`openLauncherUpdateReadOnlyHandle` +
  `openLauncherUpdateReadOnlyInheritedHandle`（import golang.org/x/sys/windows）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（全 [S]），新落地 1 个蓝图文件（launcherupdate_helper_windows.go 26→25 未落地）。
FUNCS 2933/4754 = 61.70%。下一批：launcherupdate_helper 短函数（terminateProcess/
restartRolledBack/imagePath/sameIdentity 等 128-384B）+ 截图滚动几何。
