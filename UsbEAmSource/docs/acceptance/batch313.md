# 批次 313 · 插件包 URL/窗口边界解析 +2（FUNCS 3026）

## 目标

落地 2 个短函数：`resolvePluginPackageURL`（插件包 URL 解析）、
`resolvePluginWindowBounds`（插件窗口边界解析）。

## 基线 / 收口

| 指标 | 基线（batch 312 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3024 | 3026 |
| MARKED | 3024 | 3026 |
| S | 1469 | 1469 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1476 | 1478 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1506 | 1506 |
| USABLE | 1507 | 1507 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1506  FUNCS=3026  MARKED=3026  P=41  S-eq=1  S-inline=37  S-sig=1478  S=1469  USABLE=1507
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1469 + 37 + 1 + 1478 + 41 = 3026 = FUNCS`。
S-sig 1476→1478（+2）、FUNCS 3024→3026（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolvePluginPackageURL [S-sig 0x14092c520, 224B]

TrimSpace 分支 → resolvePluginCatalogAssetURL / normalizePluginPackageFile。

### 3.2 resolvePluginWindowBounds [S-sig 0x140931820, 224B]

覆盖值非空则透传；否则居中/钳位算术（4 字返回）。

## G4 独立复核

- `backend/bootstrapservice_state_deps.go`：+resolvePluginPackageURL +resolvePluginWindowBounds [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S-sig]）。FUNCS 3026/4754 = 63.65%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
