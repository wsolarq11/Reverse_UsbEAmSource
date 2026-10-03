# 批次 296 · remoteicons 缓存/校验短函数批量 +16 [S]（FUNCS 2967）

## 目标

remoteicons.go 全文件推进：落地 Windows 保留名判定、ID 段校验、URL 目标匹配、缓存路径
解析、失败缓存（LRU 10 分钟 TTL）、SVG 本地引用判定、受限读取等 13 个函数，另含
filesearch USN 跟随器 2 个与截图滚动候选排序 1 个。

## 基线 / 收口

| 指标 | 基线（batch 295 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2951 | 2967 |
| MARKED | 2951 | 2967 |
| S | 1428 | 1444 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1465 | 1481 |
| USABLE | 1466 | 1482 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1481  FUNCS=2967  MARKED=2967  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1444  USABLE=1482
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1444 + 37 + 1 + 1444 + 41 = 2967 = FUNCS`。
S 1428→1444（+16）、FUNCS 2951→2967（+16）、FAITHFUL 1465→1481（+16）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 remoteicons.go 13 函数 [S]

- `isWindowsReservedRemoteIconSegment` [S 0x140961d60, 288B]：ToLower(TrimSpace) 后
  `aux|con|nul|prn|clock$` 直判；4 字节 `com[1-9]|lpt[1-9]` 前缀判。
- `isRemoteIconIDSegment` [S 0x140961c80, 224B]：空/超长/首尾 `-` 判否；每字节
  a-z/0-9/`-`。
- `remoteIconURLMatchesTarget` [S 0x1409623c0, 320B]：nil 守卫；target.User 非 nil
  判否；EqualFold Scheme/Host；EscapedPath 与 RawQuery 精确相等。
- `remoteIconResponseMatchesTarget` [S 0x140962300, 192B]：resp/Request/URL nil 守卫后
  url.Parse 委托。
- `remoteIconPathWithinRoot` [S 0x140965480, 288B]：Clean 双侧 → Rel → `..` 逃逸或
  IsAbs 判否。
- `splitRemoteIconID` [S 0x1409619e0, 672B]：TrimSpace 一致性守卫；`SplitN(id,":",2)`
  （sep 0x3a 实证）；ToLower 双侧；provider 段 64 / icon 段 256；provider 白名单
  `lucide|tabler|ph|carbon|material-symbols|mingcute|simple-icons`（.data 7 元素 slice）。
- `remoteIconCachePath` [S 0x140961880, 352B]：TrimSpace 空/ split 失败判空串；
  `Join(root,"remote-icons",provider,icon+".svg")`。
- `humanizeRemoteIconName` [S 0x140965ec0, 736B]：`NewReplacer("-"," ","_"," ","/"," ")`
  → Fields → 首字母大写 → Join(" ")。
- `readRemoteIconContent` [S 0x140961e80, 352B]：r==nil → errRemoteIconReaderNil
  ("远程图标内容为空")；limit<=0 → errRemoteIconContentTooLarge
  ("远程图标内容超过大小限制")；`io.ReadAll(&io.LimitedReader{N:limit+1})`；len 超限判。
- `remoteIconSVGURLReferencesLocal` [S 0x140963a00, 384B]：循环 `Index("url(")` →
  `IndexByte(')')` → `Trim(TrimSpace(after[:j]), \`"'\`)` → 非 `#` 开头判否。
- `cacheRemoteIconFailure` [S 0x1409628e0, 1184B]：split 失败直返；加锁；map 命中删
  list 旧项；len>=2048 删最老 + mapdelete；mapassign value=timestamp + append。
- `isRemoteIconFailureCached` [S 0x140962500, 992B]：加锁；map 未命中 false；IsZero 或
  now.Sub>=10min（0x8bb2c97000=600s 实证）→ mapdelete + list 删 + false；否则 true。
- `deleteRemoteIconFailure` [S 0x140962d80, 480B]：加锁；mapdelete + list 线性删除。

### 3.2 filesearch USN 跟随器 + 截图滚动 3 函数 [S]

- `shouldUseLightweightUSNFollowerForResourceMode` [S 0x1408186c0]
- `shouldUseLightweightUSNFollowerForFileSearchRuntime` [S 0x140818760]
- `rankScreenshotScrollingAppendCandidate` [S 0x1409a5e40]：纯算术候选排序。

## G4 独立复核

- `backend/remoteicons.go`：+13 函数 [S]、+2 全局 error、+失败缓存三态（map/list/mutex）。
- `backend/filesearch_usn_follower_windows.go`：+2 函数 [S]。
- `backend/screenshot_scroll_windows.go`：+1 函数 [S]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

16 个函数落地 [S]。FUNCS 2967/4754 = 62.40%。下一批：remoteicons 剩余复杂函数
（validateRemoteIconCachePath 路径校验 / readRemoteIconCache / writeRemoteIconCache /
isRemoteIconSVG XML 判定）+ filesearch_index_windows.go 短函数。
