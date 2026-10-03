# 批次 335 · 书签图标文件名清洗/Firefox 文件夹标题/legacy 浏览器判定 +3（FUNCS 3095）

## 目标

落地 3 个函数：`sanitizeBookmarkIconFilename`（书签图标文件名清洗）、
`resolveFirefoxBookmarkFolderTitle`（Firefox 文件夹标题解析）、
`shouldUseLegacyExplicitLinkBrowser`（legacy 浏览器判定）。

## 基线 / 收口

| 指标 | 基线（batch 334 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3092 | 3095 |
| MARKED | 3092 | 3095 |
| S | 1495 | 1496 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1518 | 1520 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1532 | 1533 |
| USABLE | 1533 | 1534 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1533  FUNCS=3095  MARKED=3095  P=41  S-eq=1  S-inline=37  S-sig=1520  S=1496  USABLE=1534
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1496 + 37 + 1 + 1520 + 41 = 3095 = FUNCS`。
S 1495→1496（+1）、S-sig 1518→1520（+2）、FUNCS 3092→3095（+3）、FAITHFUL 1532→1533（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 sanitizeBookmarkIconFilename [S-sig 0x140764540, 288B]

TrimSpace → 空返回空 → strings.NewReplacer(18 对替换) → Replace（替换表常量待读）。

### 3.2 resolveFirefoxBookmarkFolderTitle [S 0x14076fa20, 288B]

title=TrimSpace(title) 非空→原样；否则按 fallback 内部名
（"menu________"/"mobile______"/"toolbar_____"/"unfiled_____"）→
"Bookmarks Menu"/"Mobile Bookmarks"/"Bookmarks Toolbar"/"Other Bookmarks"；否则空串。
（返回常量实测 @0x140C506E8="Bookmarks Menu"、@0x140C52C92="Other Bookmarks"）

### 3.3 shouldUseLegacyExplicitLinkBrowser [S-sig 0x140769b20, 320B]

url TrimSpace 空→false；cleanStringList 非空→true；IndexAny(url,":/")>=0→true；
否则末尾找扩展名截断→TrimSpace 非空→true。

## G4 独立复核

- `backend/bookmarks.go`：+sanitizeBookmarkIconFilename [S-sig] +resolveFirefoxBookmarkFolderTitle [S]。
- `backend/linkbrowser.go`：+shouldUseLegacyExplicitLinkBrowser [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 3095/4754 = 65.10%。下一批：bookmarks
buildChromiumRootAncestry / filesearch 域。
