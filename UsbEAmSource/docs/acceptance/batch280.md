# 批次 280 · 琐碎长尾拉满 + screenshot 访问器/释放链（+15 [S] +1 [S-eq]，FUNCS 2914）

## 目标

沿 gap_aggregate 按 asm 长度升序落地零依赖琐碎纯函数，并闭合 screenshot native preview
与 scrolling canvas 的基础访问器、隐藏/关闭与释放链。附带修正 `screenshotNativePreviewWindow`
结构布局（service 字段宽度）。

## 基线 / 收口

| 指标 | 基线（batch 279 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2898 | 2914 |
| MARKED | 2898 | 2914 |
| S | 1377 | 1392 |
| S-inline | 37 | 37 |
| S-eq | 0 | 1 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1414 | 1429 |
| USABLE | 1414 | 1430 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1429  FUNCS=2914  MARKED=2914  P=41  S-eq=1  S-inline=37  S-sig=1443  S=1392  USABLE=1430
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1392 + 37 + 1 + 1443 + 41 = 2914 = FUNCS`。
S 1377→1392（+15）、S-eq 0→1（+1）、FUNCS 2898→2914（+16）、FAITHFUL 1414→1429（+15）、
USABLE 1414→1430（+16）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizePinyinSyllable [S 0x14080b1e0, 160B]

`(s string) string`。TrimSpace → ToLower → `Replace("u:","v",-1)` → `Replace("ü","v",-1)`。
字符串字节：old1=`75 3a`("u:")、new1=`76`("v")、old2=`c3 bc`("ü")、new2=`76`("v")。

### 3.2 normalizeInputMonitorOwner [S 0x140861ea0, 160B]

`(s string) (string, error)`。TrimSpace → `len==0` 或 `len>0x100` → ("", errors.New("输入监测 owner 无效", 25B))；
否则 (trimmed, nil)。

### 3.3 resolveRemoteIconCollectionName [S 0x1409661a0, 160B]

`(name string, collections map[string]remoteIconifyCollectionRef) string`。mapaccess2_faststr 查 name；
未命中→name；命中→TrimSpace(value.Name) 非空→trimmed，否则 name（调用者 0x14095fe87 确认参数顺序）。

### 3.4 createRemoteIconCacheDirectory [S 0x140965120, 160B]

`(path string) error`。os.Mkdir(path,0o755)；err 且非 fs.ErrExist→err；否则 validateRemoteIconCacheDirectory。

### 3.5 validateRemoteIconCacheDirectory [S 0x1409651c0, 224B]

`(path string) error`。Lstat → IsDir(itab+0x18) → Mode(itab+0x28) bt 0x1b(ModeSymlink) →
remoteIconPathHasReparsePoint；失败返回全局错误"远程图标缓存路径不安全"(33B，@0x141bc3d70/0x78)。

### 3.6 (*desktopWidgetScheduler)Wake [S 0x1407accc0, 96B]

nil→ret；`[rax+0x08]` wake chan；selectnbsend 非阻塞发送（select default 分支）。

### 3.7 timeFromWindowsTick [S-eq 0x1408697a0, 128B]

`() int64`（unix 毫秒）。wall 位操作 sec()/nsec() 内联展开 ≡ time.Now().UnixMilli()。

### 3.8 screenshotNativePreviewWindow 访问器/窗口链（+6 [S]）

- `setHandle` [S 0x14099db40, 160B]：lock(+0x10).Lock → hwnd(+0x18)=h → Unlock。
- `setThreadID` [S 0x14099dd00, 160B]：lock.Lock → threadID(+0x20)=id → Unlock。
- `handle` [S 0x14099dbe0, 192B]：lock.Lock + defer Unlock → 读 hwnd 返回。
- `Hide` [S 0x14099cf60, 160B]：handle()==0→nil；PostMessageW(hwnd,0x86a2,0,0)（LazyProc 名 "PostMessageW"，12B）→nil。
- `Close` [S 0x14099d000, 128B]：handle()==0→nil；closeOnce(+0x14 对齐 +0xdc).Do(PostMessageW(hwnd,0x86a3,0,0))→nil。
- `hideOnThread` [S 0x14099d320, 128B]：handle()==0→return；SetWindowPos(hwnd,TOPMOST(-1),X+100000,Y,W,H,0x50=SWP_NOACTIVATE|SWP_SHOWWINDOW)；visible(+0xd8)=false。

### 3.9 screenshotScrollingCanvas 释放链（+3 [S]）

- `releaseMemory` [S 0x14099f840, 160B]：nil→return；release!=nil→调用+置 nil；image!=nil→Pix 清零。
- `Release` [S 0x14099f7c0, 128B]：nil→return；遍历 chunks 逐个 releaseMemory。
- `ReleaseRow` [S 0x14099f720, 160B]：遍历 chunks 找 Rect.Min.Y<=row<Max.Y；命中且 row==Max.Y-1 时 releaseMemory。

## G4 独立复核

- `filesearch_pinyin_windows.go`：+`normalizePinyinSyllable`（import strings）。
- `inputmonitor.go`：+`normalizeInputMonitorOwner`（import errors/strings）。
- `inputmonitor_windows.go`：+`timeFromWindowsTick`（import time）。
- `remoteicons.go`：+`resolveRemoteIconCollectionName`。
- `remoteicons_cache_windows.go`：+`errRemoteIconCachePathUnsafe`、`createRemoteIconCacheDirectory`、
  `validateRemoteIconCacheDirectory`（import os）。
- `desktopwidgets_scheduler.go`：新建，+`Wake`。
- `screenshot_preview_native_windows.go`：新建，+`setHandle/setThreadID/handle/Hide/Close/hideOnThread`
  （import w32，常量 0x86a2/0x86a3）。
- `screenshot_scroll_canvas_windows.go`：+`releaseMemory/Release/ReleaseRow`。
- `types_screenshot.go`：`screenshotNativePreviewWindow.service` `any`→`*screenshotPreviewWindowService`
  （asm 实证 lock@+0x10，原 any 16B 布局错位 8B；修正后 bounds@+0x28、visible@+0xd8、closeOnce@+0xdc 全对齐）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

16 个琐碎函数落地（15 [S] + 1 [S-eq]），零外部依赖或依赖已落地符号；附带 1 处结构布局修正。
FUNCS 2914/4754 = 61.30%。未落地文件差集 37→35（desktopwidgets_scheduler.go、
screenshot_preview_native_windows.go 新建）。下一批继续 size 升序长尾（VolumeIndex 访问器需先落地
activeEntryCountLocked/markPersistedAtLocked，或转 screenshot 剩余几何/光标函数）。
