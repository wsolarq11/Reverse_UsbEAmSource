# 批次 249 · oledblackout 域短函数批量 [S] 落地 +9 函数 + normalizeLauncherUpdatePackageRoot 常量订正

## 目标

延续「已存在文件短函数」落地策略，聚焦 `oledblackout.go` 域 + `filesearch_index_windows.go` +
`bootstrapservice.go` 的缺失函数，全部 [S] 汇编直译。同时订正 batch 246 一处常量读错。

## 基线 / 收口

| 指标 | 基线（batch 248 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2799 | 2808 |
| MARKED | 2799 | 2808 |
| S | 1268 | 1277 |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1455 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2759（58.03%） | 2768（58.22%） |

构建：`bash build.sh` 成功，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe`
（21,322,240 B），SHA256 `d2d808b497d49a1ee576605f3ce77bd20f5718535c918343d4fc7fa6c3e4313d`。

`go build ./backend` / `go vet ./backend` / `go test ./backend` 全部 EXIT=0。

## 本批落地（9 个 [S]，+9 FUNCS / +9 S）

### backend/oledblackout.go（+7，新增 import：path/filepath、strings、time）

1. `normalizeOLEDBlackoutMediaPauseExclusionPath(path string) string` `[S 0x1408fdd60]`
   = TrimSpace → 空返空 → `filepath.Clean`。
2. `oledBlackoutMediaPausePathBase(path string) string` `[S 0x1408fde40]`
   = TrimSpace → TrimRight(`\/`) → 空返空 → Replace(`\`→`/`) → LastIndex(`/`) <0 返回整串，否则 `path[idx+1:]`。
3. `normalizeOLEDBlackoutMediaPauseExclusionProcessName(name string) string` `[S 0x1408fddc0]`
   = TrimSpace → 空返空 → PathBase → 空/`.`/`..` 返空 → ToLower。
4. `(*oledBlackoutService).idleRunCurrentLocked(generation, idleRun uint64) bool` `[S 0x1408ff500]`
   = shuttingDown 或 generation==0 或 lifecycleGeneration!=generation 或 idleScheduleID!=idleRun 或
   !moduleEnabled → false；否则经全局函数值间接调用（fun[0]→oledBlackoutSupported）返回其结果。
5. `(*oledBlackoutService).stopFocusRetryLocked()` `[S 0x14090c200]`
   = focusRetryTimer(+0x150)!=nil → Stop()→置 nil；focusRetryScheduleID(+0x158)++。
6. `(*oledBlackoutBrowserMediaContinuity).clear()` `[S 0x1408fcc40]`
   = nil 检查 → lock.Lock → entries(+8)=nil → Unlock。
7. `(*oledBlackoutService).armInputDismissGuardLocked()` `[S 0x14090d180]`
   = inputDismissGuardUntil(+0x180) = time.Now().Add(250ms)。

### backend/filesearch_index_windows.go（+1）

8. `(*volumeIndexReadLease).Release()` `[S 0x1407e6d00]`
   = released(+0xf0) 已置位 → 返回；否则置位 → release(+0xe8)!=nil → 调用回调。

### backend/bootstrapservice.go（+1）

9. `(*BootstrapService).GetPinnedScreenshotStates() []interface{}` `[S 0x14078eb20]`
   = screenshotPin(+0x400)==nil → 返回 nil；否则 ListStates()。

## 本批订正（batch 246 常量读错）

`backend/launcherupdate_plan.go` 的 `normalizeLauncherUpdatePackageRoot`（`[S 0x1408b6b60]`）：

- 旧（错误）：`strings.Replace(root, "\\", "-", -1)` + `strings.Contains(root, "-")`。
- 新（asm 直译）：`strings.Replace(root, "\\", "/", -1)` + `strings.Contains(root, "/")`。

原因：batch 246 把 `0x140c3362f` 读成 `-`（0x2d），本轮用 `va_read.py` 精确复核为 `/`（0x2f）。
该地址同时是 PathBase 的 Replace-new 与 LastIndex 分隔符，三条证据一致指向 `/`。
`old=0x1411cac40`（`\` 0x5c）不变。

## 关键知悉

- `va_read.py` 读 RIP 相对地址时，务必把 32 位位移与 64 位指令基址相加后统一进 64 位（本轮发现
  多处「忘进位」导致 .rdata 地址算到 .text 段，误读为 UTF-8 乱码）。`0x1411cac40` vs `0x140c8ac40`
  的差即 0x40_0000_00 的高位进位，是此类错误的典型。
- `oledBlackoutService` 字段偏移复核：moduleEnabled(+0x70)、shuttingDown(+0x88)、
  lifecycleGeneration(+0xa8)、idleScheduleID(+0x130)、focusRetryTimer(+0x150)、
  focusRetryScheduleID(+0x158)、inputDismissGuardUntil(+0x180)。
- `oledBlackoutMediaPausePathBase` 语义即「取媒体暂停排除路径的最后一段」，等价于 filepath.Base
  的 `\`/`/` 双分隔符变体；`normalizeOLEDBlackoutMediaPauseExclusionPath` 才是 filepath.Clean。
- `idleRunCurrentLocked` 与 `oledBlackoutReadInputSnapshot` 经 `.data` 全局函数值表
  （0x141bc1b30..0x141bc1b48）间接调用平台函数；该表 fun 数组首项即 `oledBlackoutSupported`。
- 未落地的依赖链（后续批次）：`dismissOverlayForKeyboardInputLocked` 依赖
  `visibleOverlayForKeyboardInputLocked`(0x14090bba0, 672B) + `dismissOverlayForInputLocked`(0x14090b320,
  2176B)；`oledBlackoutReadInputSnapshot` 依赖平台快照函数（CurrentCursorPhysicalPoint 等）。

## 下一批

P=40 不变（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 /
mousegestures 1 / nativedrag_windows 1）。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；
FUNCS 2808/4754 = 59.07%。
