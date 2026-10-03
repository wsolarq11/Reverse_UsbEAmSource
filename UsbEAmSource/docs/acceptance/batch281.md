# 批次 281 · filesearch 持久化访问器 + screenshot 帧比对/预览显示（+7 [S]，FUNCS 2921）

## 目标

闭合 filesearch VolumeIndex 的持锁脏标记/持久化访问器（markDirtyLocked / markPersistedAtLocked /
ClearDirty）与拼音索引关闭，还原 screenshot 滚动的网格帧平均差异比对链
（sampleScreenshotFrameAverageDiffWithGrid / screenshotFramesAreSimilar）及原生预览 Show 消息。
附带修复 TestReadBytes 的 TTL 秒级单位 flaky。

## 基线 / 收口

| 指标 | 基线（batch 280 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2914 | 2921 |
| MARKED | 2914 | 2921 |
| S | 1392 | 1399 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1429 | 1436 |
| USABLE | 1430 | 1437 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1436  FUNCS=2921  MARKED=2921  P=41  S-eq=1  S-inline=37  S-sig=1443  S=1399  USABLE=1437
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1399 + 37 + 1 + 1443 + 41 = 2921 = FUNCS`。
S 1392→1399（+7）、FUNCS 2914→2921（+7）、FAITHFUL 1429→1436（+7）、USABLE 1430→1437（+7）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 markPersistedAtLocked [S 0x1407e9040, 224B]

`(v *VolumeIndex, t time.Time)`。sec()/nsec() 判零（wall bit 内联展开，同 timeFromWindowsTick）
→ 零值则 `t = time.Now()`；`dirty(+0xe0)=false`、`changeCaught(+0xe4)=0`、
`LastSavedAt(+0xc0/+0xc8/+0xd0)=t`。

### 3.2 markDirtyLocked [S 0x1408072e0, 192B]

`(v *VolumeIndex, n int32)`。n==0→return；`dirty(+0xe0)=true`、`changeCaught(+0xe4)+=uint32(n)`、
`LastMutationAt(+0xa8/+0xb0/+0xb8)=time.Now()`、`runtimeVersion(+0x590)` 自增（旧值 -1 时覆盖为 1，
即跳过 0）。

### 3.3 ClearDirty [S 0x1407e8e00, 160B]

`mu.Lock + defer Unlock` → `markPersistedAtLocked(time.Now())`。

### 3.4 (*volumePinyinIndex)Close [S 0x14080e240, 192B]

nil 或 `mappedFile(+0x58)==nil` → nil；否则 `mappedFile.Close()` → 透传 error，清空
mappedFile/records(+0x28..+0x38)/aliases(+0x40..+0x50)。

### 3.5 sampleScreenshotFrameAverageDiffWithGrid [S 0x1409a6040, 768B]

`(a *image.RGBA, aMinY int, b *image.RGBA, bMinY int, height int, gridX, gridY int) int`。
width=min(a.Dx,b.Dx)、height=min(a.Max.Y-aMinY, b.Max.Y-bMinY, height)；stepY=max(1,height/gridX)、
stepX=max(1,width/gridY)；双循环 RGBAAt 取 `(|ΔR|+|ΔG|+|ΔB|)/3`（8 位分量、曼哈顿差整数除 3）
累加；无采样点返回 `0x3fffffffffffffff`，否则 `sum/count`。

### 3.6 screenshotFramesAreSimilar [S 0x1409a4140, 160B]

`(a, b *image.RGBA) bool`。nil→false；Dx 或 Dy 不等→false；
`sampleScreenshotFrameAverageDiffWithGrid(a, a.Rect.Min.Y, b, b.Rect.Min.Y, a.Rect.Dy(), 42, 80) <= 2`。

### 3.7 (*screenshotNativePreviewWindow)Show [S 0x14099cea0, 192B]

handle()==0 → `errors.New("原生截图预览窗口不可用")`（33B UTF-8，`e58e9f...e794a8`）；
否则 `PostMessageW(hwnd, 0x86a1, 0, 0)` → nil。

## G4 独立复核

- `filesearch_index_windows.go`：+`markPersistedAtLocked`、`markDirtyLocked`、`ClearDirty`。
- `filesearch_pinyin_windows.go`：+`(*volumePinyinIndex)Close`。
- `screenshot_scroll_windows.go`：+`sampleScreenshotFrameAverageDiffWithGrid`、
  `screenshotFramesAreSimilar`（import image/math）。
- `screenshot_preview_native_windows.go`：+`Show`（import errors，常量 0x86a1）。
- `launcherasset_test.go`：TestReadBytes `RegisterBytes(..., 5)` → `time.Minute`（TTL 5ns
  秒级单位 flaky，与 TestServeAssetRequest 的修复同源）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

7 个琐碎函数落地（全 [S]），零新增外部依赖；连带 1 处测试 TTL 单位修复。
FUNCS 2921/4754 = 61.47%。下一批：screenshot preview 剩余（decodeImage 依赖
screenshotNativePreviewDecodeImage、Close.func1 已还原）+ screenshotNativePinWindow 布局修正
（service/pin any→指针，同 preview 的 8B 错位问题）。
