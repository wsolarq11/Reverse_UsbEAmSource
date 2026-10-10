# 批次 377 · 鼠标手势配置 wrapper +3（FUNCS 3256）

## 目标

落地 BootstrapService 鼠标手势域 3 个薄 wrapper（签名 + delegate 调用链）：

- `SetMouseGestureCaptureSuspended(suspended bool) error` [S 0x140782340, 480B]
- `TestHotCorner(corner string) error` [S 0x140783a60, 512B]
- `PickMouseGestureAppTarget() (string, error)` [S 0x140783c60, 544B]

并修正 `attachMouseGestureConfigIconURLs` 签名：`(config interface{})` → `(MouseGestureConfig) MouseGestureConfig`
（与 asm 值传值返回一致；原 `interface{}` 参数为推断错误）。

## 基线 / 收口

| 指标 | 基线（batch 376 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3253 | 3256 |
| MARKED | 3253 | 3256 |
| S | 1516 | 1519 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1658 | 1658 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1553 | 1556 |
| USABLE | 1554 | 1557 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3256  MARKED=3256  UNMARKED=0  S=1519  S-inline=37  S-eq=1  S-sig=1658  P=41  FAITHFUL=1556  USABLE=1557  TRUE=3214
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1519 + 37 + 1 + 1658 + 41 = 3256 = FUNCS`。
FUNCS 3253→3256（+3）、S 1516→1519（+3）、S-sig/P 持平、FAITHFUL 1553→1556（+3）、TRUE 3211→3214（+3）、UNMARKED=0。
`list_missing.js` 未落地数 374→371（−3），三函数退出缺失清单。

## G3 行为（asm 逐地址实证）

结构锁定：`BootstrapService.mouseGestures *mouseGestureService`@+0x390（types_launcher.go:420）；
`mouseGestureService.config MouseGestureConfig`（types_gesture.go:297）。

### 3.1 SetMouseGestureCaptureSuspended [S 0x140782340]

- morestack 序仅存 `rax`(bs) + `bl`(suspended bool) → 签名 `(suspended bool) error`。
- `mov rax,[rax+0x390]` 读 mouseGestures → `mouseGestureService.SetCaptureSuspended(suspended)`。
- 返回 error 接口（2 word）→ 经栈 duffcopy 透传后原样返回。
- delegate 返回前 `attachMouseGestureConfigIconURLs(mouseGestures.config)` 副作用调用，结果丢弃。

### 3.2 TestHotCorner [S 0x140783a60]

- morestack 序存 `rax`(bs) + `rbx/rcx`(corner string 2 word) → 签名 `(corner string) error`。
- `mov rax,[rax+0x390]` → `mouseGestureService.TestHotCorner(corner)`。
- 同 3.1：attach 副作用调用 → 返回 error。

### 3.3 PickMouseGestureAppTarget [S 0x140783c60]

- morestack 序仅存 `rax`(bs) → 无入参，签名 `() (string, error)`。
- `mov rax,[rax+0x390]` → `mouseGestureService.PickAppTarget()`（返回 target+error）。
- 成功后 `attachMouseGestureAppProfileIconURL(target)`（0x1407848c0，[S] stub）→ 返回 (target, nil)。

### 3.4 attachMouseGestureConfigIconURLs 签名修正

- asm 0x140784000：`normalizeMouseGestureConfig`（值传入）→ 遍历 Apps（元素步长 0x118）→
  每项 `attachMouseGestureAppProfileIconURL` → 返回 config 值。
- 原 stub 签名 `(config interface{})`（无返回）为推断错误，改为 `(MouseGestureConfig) MouseGestureConfig`；
  体保持 normalize + 原样返回（逐 profile 图标回填待 0x1407848c0 字节级续作）。

## G4 独立复核

- `backend/bootstrapservice.go`：+3 wrapper（[S]），+`attachMouseGestureConfigIconURLs` 签名修正。
- 依赖 `mouseGestureService.SetCaptureSuspended/TestHotCorner/PickAppTarget`（[S-sig] stub，签名已锁）、
  `normalizeMouseGestureConfig`（[S-sig]）、`attachMouseGestureAppProfileIconURL`（[S] stub）——签名均已对齐。
- `attachMouseGestureConfigIconURLs` 全仓唯一定义，无其他调用点，签名修正无涟漪。
- `go build/vet/test` 全仓通过。

## 移交（本轮收尾）

3 wrapper 落地（均 [S]）+1 签名修正。FUNCS 3256/4754 = 68.49%，FAITHFUL 1556/4754 = 32.73%，TRUE 3214/4754 = 67.61%。
下一批：鼠标手势 delegate 体（SetCaptureSuspended/TestHotCorner/PickAppTarget 0x1408e0840/0x1408e12a0/0x1408e1800）；
`attachMouseGestureAppProfileIconURL` 0x1407848c0 字节级；`sanitizeScreenshotPreviewResult`/`validateRemoteIconTemporaryFile`。
