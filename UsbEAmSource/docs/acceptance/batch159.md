# 批次 159 验收（ensureLauncherWindowForShow 体修正 [S] + createLauncherWindow 签名实证修正 [S-sig]）

日期：2026-09-25
子批次：启动器窗口创建域签名/体对齐（无净计数变化，纯正确性修正）

## 目标

修正两处与 asm 相悖的既有存根：

1. `(*BootstrapService).ensureLauncherWindowForShow`：旧体 `return nil` 丢弃已解析窗口、且
   `createLauncherWindow(true)` 调用签名错误——均与 asm 0x140795c00 不符。本批按 asm 逐条重建，
   返回类型由 `interface{}` 收窄为 `application.Window`（返 (rax,rbx) 两字，非 (window,error) 四字）。
2. `createLauncherWindow`：旧签名 `func(asMainWindow bool)`（void）与 asm 0x1408d0380 的
   `(app *application.App, bs *BootstrapService, asMainWindow bool) application.Window` 相悖，修正签名。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.375s) |
| 黄金用例 | `go1.25.12 test -run 'TestEnsureLauncherWindowForShowNilApp' -v` | PASS |

黄金用例说明：`app==nil` 早退路径（resolveLauncherWindow 返 nil → app 判空 → 返 nil），
不触 Wails runtime，可安全实测。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1079 / S-inline=35 / S-sig=1348 / P=197 / UNMARKED=0`。

**真函数 = 1079 + 35 + 1348 = 2462 / 4754 = 51.8%**（与批次 158 持平，本批为纯正确性修正）。

重建产物 SHA256：`406D97A9D927800705B85412D498AE8105EE09B5256BB4EDD53C41D4304A2AD0`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细

| 函数 | VA | 档位 | 修正点 |
|---|---|---|---|
| ensureLauncherWindowForShow | 0x140795c00 | [S]（原误标体） | 返 `win` 而非 nil；`return createLauncherWindow(app, bs, true)`；返 application.Window |
| createLauncherWindow | 0x1408d0380 | [S-sig]（签名修正） | `(app, bs, asMainWindow) application.Window` 三参 + 2 字返回 |

## G4 关键实证结论

1. **ensureLauncherWindowForShow 返回 application.Window（2 字）**：asm 三条返回路径均只返
   (rax,rbx) 两字——`jne 0x140795cc1`（透传 resolveLauncherWindow 的窗口）、`xor eax/xor ebx`
   （app nil 返零）、`call createLauncherWindow` 后 `ret`（透传新窗口）。无 error 第四字。
2. **旧体 `return nil` 是笔误**：asm 0x140795c25-28 `test rax / jne` 明确「窗口非 nil 即返窗口」，
   旧体却丢弃 `win` 返 nil，导致 ensure 永不短路。
3. **createLauncherWindow 三参签名**：asm 前导 `test rax / je`、`test rbx / je` 判两指针非空，
   `cl` 为 bool 第三参；`[arg1+0x2d0]` 读 app.WindowManager 字段、`arg2` 作 attachWindow 接收者
   ——故签名 `(app, bs, asMainWindow)`，非旧体的单 bool 无返。
4. **体仍留 [S-sig]**：createLauncherWindow 体依赖 loadLauncherUIScalePercent /
   buildLauncherWindowOptions / WindowManager.NewWithOptions / attachWindow /
   registerLauncherCloseToTrayHook / registerLauncherFileDropHandler 六个未落地/签名待修项，
   暂以 `return nil` 占位（与 asm 的 `app==nil || bs==nil` 早退同构），留待窗口装配域专项批次。

## 残留 [P] / 未落地（不触及）

`resolveLauncherWindow` 仍为 [S-sig] 存根（返 (nil,nil)），`attachWindow` 签名待修（asm 三参
`(bs, app, launcherWindowData)`），`registerLauncherCloseToTrayHook`/`registerLauncherFileDropHandler`
未落地——均属启动器窗口装配域后续批次。

## 已知偏差（诚实记录）

createLauncherWindow 体为 [S-sig] 占位（依赖未落地），签名已按 asm 实证修正；ensureLauncherWindowForShow
体为 [S]（逐条对位），但其正确性受 resolveLauncherWindow/createLauncherWindow 两个 [S-sig] 依赖约束。
