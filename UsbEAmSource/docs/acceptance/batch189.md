# 批次 189 — 组件存储写入链闭环（3 函数：2 [S] + 1 [S-sig]）

## 基线 / 收口

| 指标 | 基线（批次 188 收口） | 收口（批次 189） |
|---|---|---|
| FUNCS | 2727 | **2730** |
| S | 1174 | **1176** |
| S-inline | 36 | 36 |
| S-sig | 1402 | **1403** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2612 | **2615** |

SHA256 `9F2C5732CA904E77EB533197669BCF4C1699AAF9D056E3EAA38276FC287B5D3E`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（desktopwidgets_store.go，存储写入链闭环）

### 1. `writeDesktopWidgetDocumentFile`（0x1407c47c0，1504B）— [S-sig]

`func(path string, doc DesktopWidgetDocument) error`。签名 + 结构已还原：TrimSpace 空检查 →
normalize → validate → MkdirAll(0o755) → MarshalIndent(doc,"","  ")+换行 → CreateTemp(".launcher-widgets-*.tmp")
→ Write/Sync/Close → rename 重试 ≤5 次 + sleep 50ms。体待还原：重试条件两个 errors.Is 目标为 .data 段
运行时初始化的全局 error（@0x141c10970/0x141c10980），静态文件无法读值，待动态确认。

### 2. `writeUnlocked`（0x1407c3260，640B）— [S]

`func(s *launcherWidgetStore) error`。normalize → validate → writeDocument（nil 时取默认
writeDesktopWidgetDocumentFile，全局函数指针 @0x141096dd0=0x1407c47c0）落盘 → 缓存（cached/cachedExists/
cachedLoadErr/loaded 四字段写入点逐一对齐 asm）。

### 3. `Ensure`（0x1407c06e0，1504B）— [S]

`func(s *launcherWidgetStore) (DesktopWidgetDocument, error)`。nil→"首页组件存储不可用"；加锁 defer 解锁；
loadUnlocked 取 exists/doc/err；已存在→(doc,nil)；不存在→零值文档 + UpdatedAt=UTC RFC3339Nano → writeUnlocked
→ clone 返回。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.950s`。
- **G2 count_funcs**：`FUNCS=2730 / S=1176 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`。
- **G3 行为**：错误字符串（"首页组件存储不可用"@0x140c69cd5 / "首页组件存储路径不能为空"@0x140c782eb /
  "2006-01-02T15:04:05.999999999Z07:00"@0x140c7708d / "  "@0x140c3366b / ".launcher-widgets-*.tmp"@0x140c62ea1）
  逐条 asm 对齐；MkdirAll 权限 0x1ed=0o755、Sleep 0x2faf080=50ms、重试上限 5 次、MarshalIndent 缩进 2 空格均实证。
- **G4 review**：仅改 desktopwidgets_store.go（追加 3 函数 + import time）；无跨文件写重叠。

## 遗留（下一批）

- `writeDesktopWidgetDocumentFile` 升档 [S-sig]→[S]（重试条件 e1/e2 动态确认后）。
- `validateDesktopWidgetDocument` 升档 [S-sig]→[S]。
- `launcherWidgetStore.Update`（0x1407c1100，1600B）/`ReplaceWithRollback`（0x1407c17a0，3136B）——写入侧。
- `bytesMatchWildcardFold` 0x1407e3e60（176 行）。
- 继续落地 truly-missing 顶层函数（678 个）。
