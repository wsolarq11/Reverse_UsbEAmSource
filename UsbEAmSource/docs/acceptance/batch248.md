# 批次 248 — 已存在文件的短函数批量 [S] 落地（+11 函数）

## 基线 / 收口

| 指标 | 基线（批次 247 收口） | 收口（批次 248） |
|---|---|---|
| FUNCS | 2788 | **2799** |
| S | 1257 | **1268** |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1455 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |

SHA256（`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B）
`aea815511cc97fadaeffe66cd6321c668a752325cae5b606ee7155c716d8a11b`。

## 本批内容

全部为**已存在 backend 文件**的缺失函数补全（无需新建文件），11 个函数全部 asm 直译 `[S]`：

`backend/filesearch_index_windows.go`：

- `(*IndexNode).IsDirectory`（`[S] 0x1407e78a0`）＝ `Flags&1 != 0`（叶子函数 movzx word [rax+0x16]）。
- `(*IndexNode).IsDeleted`（`[S] 0x1407e78c0`）＝ `Flags&2 != 0`（test 2 setne）。

`backend/oledblackout.go`：

- `(*oledBlackoutService).beginBackgroundActivity`（`[S] 0x1408ff340`）＝ lock→defer Unlock→
  ensureLifecycleLocked→shuttingDown||generation 不匹配返回 false→backgroundActivities.Add(1) 返回 true。
- `(*oledBlackoutService).endBackgroundActivity`（`[S] 0x1408ff4c0`）＝ backgroundActivities(+0x10).Add(-1)。

`backend/main.go`：

- `closeLauncherStartupDebugLog`（`[S] 0x1408d2f00`）＝ nil 检查后 `_ = f.Close()`（[rax]=file 内部指针）。

`backend/oledblackout_windows.go`：

- `(*windowsOLEDBlackoutCursorController).request`（`[S] 0x140913f20`）＝ nil 检查→lock+defer Unlock→
  closed 返回 closeErr→make(chan error,1)+commands<-命令+<-响应；nil receiver 错误
  `"鼠标控制器不可用"`（asm 0x140c65370 直译）。
- `(*windowsOLEDBlackoutCursorController).Hide`（`[S] 0x1409140a0`）＝ request(true,false,false,nil)。
- `(*windowsOLEDBlackoutCursorController).Show`（`[S] 0x1409140e0`）＝ request(false,true,false,nil)。

`backend/screenshot_cursor_windows.go`：

- `screenshotCursorCurrentlyShowing`（`[S] 0x140973c40`）＝ GetCursorInfo；失败→true（保守可见）；
  否则 flags&1(CURSOR_SHOWING)。
- `adjustScreenshotCursorVisibility`（`[S] 0x140973e00`）＝ 循环 64 次 ShowCursor 直至达到目标状态；
  错误串 `"恢复截图覆盖层鼠标指针失败: %w"` / 无参版（asm 0x140c82182 直译）。
- `ensureScreenshotCursorVisible`（`[S] 0x140973ba0`）＝ 已可见则返回，否则 adjust(true)。
- `procShowCursor` LazyProc 变量（user32.ShowCursor）。

## asm 证据

- `IsDirectory/IsDeleted`：Flags 字段偏移 +0x16（types_misc.go IndexNode），bit0=目录、bit1=删除。
- `beginBackgroundActivity`：morestack 存 rax+rbx（receiver+generation uint64）；[0xa8]=lifecycleGeneration
  与 rbx 比较；[0x88]=shuttingDown；lock=[0x00]、backgroundActivities=[0x10]（types_oled.go 对齐）。
- `request`：morestack 存 5 寄存器；response 形参 asm 未引用（内部总是 makechan），保留仅为签名对齐。
- `screenshotCursorCurrentlyShowing`：CURSORINFO cbSize@0=0x18、flags@4（screenshotCursorNativeInfo 对齐）。
- `adjustScreenshotCursorVisibility`：ShowCursor 计数器语义（visible 要求 r>=0，隐藏要求 r<0）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build/vet ./backend` EXIT=0；`test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2799 / S=1268 / S-inline=36 / S-sig=1455 / P=40 / UNMARKED=0`。
- **G3 行为**：纯补全既有文件缺失函数，无既有函数签名/行为变更，既有测试全量 PASS。
- **G4 review**：上述 7 个文件新增方法/函数。

## 遗留（下一批）

- P=40（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 / mousegestures 1 /
  nativedrag_windows 1）。
- FUNCS 缺口约 1955（2799/4754 = 58.88%）：继续按 `gap_aggregate.txt`（长度升序）落地已存在文件的
  短函数 + 新建文件完整落地。
- `restoreScreenshotCursorAfterOverlay`（0x140973be0）、`Close`（0x140914120）等 cursor/oled 域
  剩余函数待续。
