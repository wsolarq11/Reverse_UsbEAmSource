# 批次 298 · filesearch 访问器 + OLED WinRT + 调度器/工具栏入口 +8（FUNCS 2978）

## 目标

落地 8 个短函数：`nodePathCache.Get`/`VolumeIndex.EntryCount`（filesearch 访问器）、
`oledBlackoutIReference.GetInt32`（WinRT COM getter）、`desktopWidgetScheduler.Start`/`run`
（调度器入口）、`screenshotSelectionToolbarRevealWindow`/`BringToTop`（工具栏入口）。
新建 `oledBlackoutIReference` 类型。

## 基线 / 收口

| 指标 | 基线（batch 297 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2970 | 2978 |
| MARKED | 2970 | 2978 |
| S | 1445 | 1449 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1446 | 1450 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1482 | 1486 |
| USABLE | 1483 | 1487 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1486  FUNCS=2978  MARKED=2978  P=41  S-eq=1  S-inline=37  S-sig=1450  S=1449  USABLE=1487
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1449 + 37 + 1 + 1450 + 41 = 2978 = FUNCS`。
S 1445→1449（+4）、S-sig 1446→1450（+4）、FUNCS 2970→2978（+8）、FAITHFUL 1482→1486（+4）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 nodePathCache.Get [S 0x1407e1720, 192B]

`(key int32) (string, bool)`。cmp ebx,0x40000000（signed）分路：>= 0x40000000 走
deltaCache（mapaccess2_fast32 透传 (v,ok)）；< 走 sliceCache 索引（movsxd 符号扩展、
负键/越界 → ("",false)、空字符串判不存在）。

### 3.2 VolumeIndex.EntryCount [S 0x1407e7b00, 192B]

`() int`。mu.RLock（lock xadd [rax+0x10] readerCount）+ defer RUnlock →
activeEntryCountLocked() 透传返回。

### 3.3 VolumeIndex.activeEntryCountLocked [S-sig 0x1407e8340, 448B]

`() int`。readProvider(+0x520) 空则默认 provider 取读视图，defer 释放；打包
activeCount(+0xd8)/delta(+0x548..+0x588) 传 activeVolumeIndexEntryCountForView。体待读视图域。

### 3.4 oledBlackoutIReference.GetInt32 [S 0x140921060, 160B]

`() (int32, bool)`。SyscallN(vtbl[6]@0x30 get_Value, this, &value)；HRESULT 有符号 <0 →
(0,false)；否则 (value,true)。新建类型 `oledBlackoutIReference{vtbl *uintptr}`。

### 3.5 desktopWidgetScheduler.Start [S 0x1407acb80, 96B]

`()`。once(+0x20).Do(func1)；func1 newobject 打包 gowrap1 捕获 scheduler，runtime.newproc
起 goroutine；gowrap1 尾调 run。幂等。

### 3.6 desktopWidgetScheduler.run [S-sig 0x1407acec0]

`()`。单 receiver 无参无返回。体待调度循环专项。

### 3.7 screenshotSelectionToolbarRevealWindow [S-sig 0x1409ad240, 224B]

`(window application.Window, activate bool)`。activate=true 经 itab fun[0x248] 直接激活；
false 走 InvokeSync(func1 ShowWindowNoActivate)。体待 wails 句柄方法名专项。

### 3.8 screenshotSelectionToolbarBringToTop [S-sig 0x1409ad3a0, 160B]

`(window application.Window)`。nil 接口返回；InvokeSync(func1 0x1409ad440：取句柄→
IsWindowVisible→ShowWindow(8)→SetWindowPos(HWND_TOP,0x53))。体待 wails 句柄方法名专项。

## G4 独立复核

- `backend/filesearch_windows.go`：+nodePathCache.Get [S]。
- `backend/filesearch_index_windows.go`：+EntryCount [S]、activeEntryCountLocked [S-sig]。
- `backend/oledblackout_windows.go`：+GetInt32 [S]；`types_oled.go`：+oledBlackoutIReference 类型。
- `backend/desktopwidgets_scheduler.go`：+Start [S]、run [S-sig]。
- `backend/screenshot_selection_toolbar_windows.go`：+RevealWindow/BringToTop [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

8 个函数落地（4 [S] + 4 [S-sig]）。FUNCS 2978/4754 = 62.64%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
