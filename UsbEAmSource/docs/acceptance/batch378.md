# 批次 378 · 鼠标手势暂停 delegate + 截图预览结果清洗 +2（FUNCS 3257）

## 目标

1. 升级 `mouseGestureService.SetCaptureSuspended` `[S-sig]`→`[S]`（0x1408e0840，480B）。
2. 新增 `sanitizeScreenshotPreviewResult` `[S]`（0x14099b3c0，448B）。
3. 修正类型 `ScreenshotCaptureResult`：删除推断错误的 `Truncated`/`TruncationReason`
   两字段（8 字段，与二进制 JS 空结果 `{imageData,imageUrl,thumbnailUrl,width,height,mode,path,cancelled}`
   及 sanitize 汇编逐字段一致）。

## 基线 / 收口

| 指标 | 基线（batch 377 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3256 | 3257 |
| MARKED | 3256 | 3257 |
| S | 1519 | 1521 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1658 | 1657 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1556 | 1558 |
| USABLE | 1557 | 1559 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3257  MARKED=3257  UNMARKED=0  S=1521  S-inline=37  S-eq=1  S-sig=1657  P=41  FAITHFUL=1558  USABLE=1559  TRUE=3215
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1521 + 37 + 1 + 1657 + 41 = 3257 = FUNCS`。
FUNCS 3256→3257（+1）、S 1519→1521（+2）、S-sig 1658→1657（−1）、FAITHFUL 1556→1558（+2）、
TRUE 3214→3215（+1）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 371→370（−1），`SetCaptureSuspended`/`sanitizeScreenshotPreviewResult` 退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 SetCaptureSuspended [S 0x1408e0840, 480B]

结构锁定（types_gesture.go `mouseGestureService`）：`lock sync.Mutex`@+0x0、
`capturePaused bool`@+0x150、`lastError string`@+0x118/+0x120。

- `lock.Lock`（cmpxchg dword [rcx],1 → sete/jne，失败走 lockSlow）。
- `mov byte [rcx+0x150], bl` → capturePaused = suspended。
- `mov [rcx+0x120], 0`（+ 写屏障清 [rcx+0x118]）→ lastError = ""。
- `lock.Unlock`（xadd −1 → dec/jne → unlockSlow）。
- `syncPlatformRuntime()`（0x1408f0c80）→ `buildState()`（0x1408e1e40，返回 MouseGestureState 丢弃）。
- `xor eax,ebx; ret` → return nil。

### 3.2 sanitizeScreenshotPreviewResult [S 0x14099b3c0, 448B]

- 入参/返回均为 8 字段 struct（97B 区域 [rsp+0x70..0xd1]，返回槽 [rsp+0xe8..0x149]）。
- 逐字段：ImageData/ImageURL/ThumbnailURL `TrimSpace`；Width/Height `cmovl` 钳制 <0→0；
  Mode `normalizeScreenshotMode`；Path `TrimSpace`；Cancelled `movzx` 透传。

### 3.3 ScreenshotCaptureResult 类型修正

二进制 JS 空结果字面量（0x…`return{imageData:"",imageUrl:"",thumbnailUrl:"",width:0,height:0,mode:"",path:"",cancelled:!1}`）
与 sanitize 汇编逐字段一致 → 删除推断错误的 `Truncated`/`TruncationReason`。
`Truncated`/`TruncationReason` 在 backend 无任何引用点（grep 仅 types 定义自身），删除无涟漪。

## G4 独立复核

- `backend/service_configure_stubs.go`：`SetCaptureSuspended` 体升 [S]。
- `backend/screenshot_preview_windows.go`：+`sanitizeScreenshotPreviewResult` [S]（import strings）。
- `backend/types_screenshot.go`：`ScreenshotCaptureResult` 收敛为 8 字段。
- `go build/vet/test` 全仓通过；`go fmt` 无差异。

## 移交（本轮收尾）

2 函数落地（1 升级 [S] + 1 新增 [S]）+1 类型修正。FUNCS 3257/4754 = 68.51%，
FAITHFUL 1558/4754 = 32.77%，TRUE 3215/4754 = 67.63%。
下一批：`buildScreenshotPreviewWindowUpdate`(0x14099ae00)、`TestHotCorner`/`PickAppTarget` delegate 体、
`attachMouseGestureAppProfileIconURL` 0x1407848c0、filesearch 域 480B 候选、`P=41 → [S]` 转换。
