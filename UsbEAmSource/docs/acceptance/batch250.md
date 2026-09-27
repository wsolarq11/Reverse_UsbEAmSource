# 批次 250 · BootstrapService/ShellVerb 短函数落地 +4（+3 [S] +1 [S-sig]）

## 目标

延续「已存在文件短函数」策略，聚焦 `bootstrapservice.go` + `shellverb.go` 的缺失函数。

## 基线 / 收口

| 指标 | 基线（batch 249 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2808 | 2812 |
| MARKED | 2808 | 2812 |
| S | 1277 | 1280 |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2768（58.22%） | 2772（58.31%） |

构建：`bash build.sh` 成功，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe`（21,322,240 B），
SHA256 `253923d58e5b5aa7c1422bc63aa94c82f889caa52172bb5465e84dab53f3b34d`。
`go build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+4）

### backend/shellverb.go（+1 [S]）

1. `(*BootstrapService).InvokeShellVerb(verb, path string) error` `[S 0x1408a6920]`
   = 参数重排后尾调用 `invokeShellVerb(verb, path)`（该自由函数已在 shellverb.go 落地）。
   签名实证：morestack 保存 rax/rbx/rcx/rdi/rsi（receiver + 2 个 string），`invokeShellVerb("", "x")`
   与 `invokeShellVerb("open", "  ")` 测试用例（shellverb_test.go）锁定 `(verb, path string) error`。

### backend/bootstrapservice.go（+3）

2. `(*BootstrapService).emitScreenshotCaptureAccepted()` `[S-sig 0x140797080]`
   = 签名经符号表实证（单 receiver 无参无返回）；体为事件分发链（resolveLauncherWindow +
   接口 typeAssert + [rax+0x20] 回调），留待专项。
3. `(*screenshotCaptureAcceptedNotifier).Emit()` `[S 0x140796fe0]`
   = nil 检查（receiver/service）→ `once.Do(func1)`；func1(0x140797040) =
   `n.service.emitScreenshotCaptureAccepted()`。
4. `(*BootstrapService).clearStartupTrayMode()` `[S 0x140796820]`
   = nil 检查 → lock(+0x540) → startupTrayMode(+0x449)=false → unlock。

## 关键知悉

- `screenshotCaptureAcceptedNotifier`（types_screenshot.go:317）= `{service *BootstrapService, once sync.Once}`，
  [0]=service、[8]=once；Emit 闭包捕获 n，体内解引用 `n.service`。
- `BootstrapService` 字段偏移复用 batch 249 已有锚点：lock(+0x540)、startupTrayMode(+0x449)，
  均见 bootstrapservice_lifecycle.go 头部偏移表。
- 待落地的依赖链：`applyLauncherWindowSizingForShow`(0x1407968c0) 依赖
  `consumeLauncherDefaultSizeReset` + `loadLauncherUIScalePercent`（均未落地）；`LaunchApp`(0x1408a6020)
  为栈传大结构体 + `LaunchAppWithPrivilege` 尾调用。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；FUNCS 2812/4754 = 59.15%。
