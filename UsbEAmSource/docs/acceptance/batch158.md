# 批次 158 验收（launcher 窗口截屏捕获：begin/endLauncherScreenshotCapture 2 函数 [S]）

日期：2026-09-25
子批次：启动器窗口截屏捕获域（窗口移出屏幕 + 内容保护开关 + 状态恢复）

## 目标

落地 bootstrapservice.go 蓝图（source_funcs.txt 3723-3749）中**截屏捕获对函数**，共 2 个
具名方法全量 `[S]`，并新建 `launcherScreenshotCaptureState` 状态结构体：

`(*BootstrapService).beginLauncherScreenshotCapture`、`(*BootstrapService).endLauncherScreenshotCapture`。

本批为 `pickWindowProcessForService` 依赖链的前置域（截屏前隐藏窗口 / 截屏后恢复窗口），
其调用点已在 `pickWindowProcessForService` 0x14085bfc6（begin）与 `.func1` 0x14085c4f0
（end）实证。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.406s) |
| 黄金用例 | `go1.25.12 test -run 'TestBeginLauncherScreenshotCapture\|TestEndLauncherScreenshotCapture' -v` | 6 条全 PASS |

黄金用例说明：以 `launcherWindowControllerMock` 替身实测完整路径（可见窗口走
Position→SetContentProtection(true)→SetPosition(x-100000,y)→记态）与三条早退路径
（!enabled / nil window / !IsVisible），均不触 Wails runtime，可安全实测。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1079 / S-inline=35 / S-sig=1348 / P=197 / UNMARKED=0`。

**真函数 = 1079 + 35 + 1348 = 2462 / 4754 = 51.8%**（相对批次 157 的 2460 增 +2 [S]）。

重建产物 SHA256：`7686A3B8FFA228CAD58802A42B062D3F1BCC8352D6ADDD5BEB87147EF64C0CAD`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 方法 [S]，`backend/bootstrapservice_window.go`）

| 函数 | 蓝图行 | VA | 说明 |
|---|---|---|---|
| beginLauncherScreenshotCapture | 3723-3740 | 0x140796aa0 | IsVisible 守卫→Position→SetContentProtection(true)→SetPosition(x-100000,y)→记态 |
| endLauncherScreenshotCapture | 3740-3749 | 0x140796c20 | SetPosition(x,y)→SetContentProtection(false) 恢复 |
| launcherScreenshotCaptureState | — | — | 5 字宽状态结构（window 接口 2 字 + enabled bool + x,y int） |

## G4 关键实证结论

1. **接口确认为 `launcherWindowController`**（types_launcher.go:570）：itab 方法序
   fun[6]=IsVisible / fun[7]=Position / fun[9]=SetContentProtection / fun[10]=SetPosition，
   与 asm 读取的 `[itab+0x48]`/`[itab+0x50]`/`[itab+0x60]`/`[itab+0x68]` 逐项对位
   （Go itab fun[0] 起于 +0x18，fun[i]=+0x18+8i）。
2. **SetPosition 首参减字面量 100000**（`add rbx,-0x186a0`）：截屏时把窗口沿 x 轴左移
   100000 像素移出屏幕，`launcherWindowScreenshotHideOffset=100000` 复刻。
3. **返回序 5 字对齐**：begin 返回 `(rax=itab, rbx=data, rcx=enabled, rdi=x, rsi=y)`，
   与 `launcherScreenshotCaptureState{window,enabled,x,y}` 逐字对齐（调用点
   pickWindowProcessForService 0x14085bfcb-0x14085bfea 存 5 字实证）。
4. **三条早退统一返零态**：`!enabled` / `win==nil` / `!IsVisible()` 均返回
   `launcherScreenshotCaptureState{}`（asm 三处 je 汇于 0x140796baa，rax/rcx/rdi/rsi 清零）。
5. **end 恢复顺序**：`SetPosition(x,y)` 先于 `SetContentProtection(false)`（asm 0x140796c63
   与 0x140796c75 两次 call 顺序实证），与 begin 的 set 顺序（SetContentProtection(true)
   先于 SetPosition）相反，符合"先隐藏→再保护，先还原位置→再解保护"语义。

## 残留 [P] / 未落地（不触及）

`pickWindowProcessForService` 主体仍为 `[S-sig]`，待 `pickWindowProcessFromScreenshotSelection`
（screenshot_windows.go `[P]`）及 `resolveLauncherWindow`/`showLauncherWindow` 升档后转 `[S]`。

## 已知偏差（诚实记录）

无。本批 2 方法均为 asm 直译，接口方法序、常量偏移、返回寄存器、早退分支逐条对照。
