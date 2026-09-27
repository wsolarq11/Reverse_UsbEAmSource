# 批次 190 — 组件存储 CRUD 完整闭环（Update + ReplaceWithRollback 升档 [S]，writeDesktopWidgetDocumentFile 升档 [S]）

## 基线 / 收口

| 指标 | 基线（批次 189 收口） | 收口（批次 190） |
|---|---|---|
| FUNCS | 2730 | **2732** |
| S | 1176 | **1179** |
| S-inline | 36 | 36 |
| S-sig | 1403 | **1402** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2615 | **2617** |

SHA256 `060D52BE6A3B13E7BB62592874C27AB278DDD148805343443167B00F418078E7`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（desktopwidgets_store.go，CRUD 闭环 + 原子写完整还原）

### 1. `Update`（0x1407c1100，1600B）— [S]

`func(s *launcherWidgetStore, mutate func(*DesktopWidgetDocument) error) (DesktopWidgetDocument, error)`。
深拷贝现有文档→可选 mutate 回调改副本→Revision+1 并盖 UTC 时间戳→writeUnlocked 落盘→返回副本。

### 2. `ReplaceWithRollback`（0x1407c17a0，3136B + 闭包 func1@0x1407c1e60）— [S]

`func(s *launcherWidgetStore, newDoc DesktopWidgetDocument) (func() error, error)`。归一化+时间戳+校验新文档→
直接 writeDocument 落盘→缓存→返回回滚闭包（恢复旧文档或删除文件）。loadUnlocked 错误时旧文档清零、exists=false
但不中断；闭包捕获 s/exists/w/oldDoc 四元组。

### 3. `writeDesktopWidgetDocumentFile`（0x1407c47c0，1504B）— [S-sig]→[S]

原子写完整还原：TrimSpace 空检查→normalize→validate→MkdirAll(0o755)→MarshalIndent("  ")+换行→
CreateTemp(".launcher-widgets-*.tmp")→defer Close→Write/Sync/Close→rename 重试 ≤5 次 + sleep 50ms。
重试条件两 error 变量 @0x141c10970/0x141c10980 紧邻 os.ErrNotExist@0x141c10990，按 io/fs error 声明顺序
确定为 os.ErrPermission 与 os.ErrExist。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.900s`。
- **G2 count_funcs**：`FUNCS=2732 / S=1179 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`。
- **G3 行为**：错误字符串/时间戳 RFC3339Nano/回滚闭包捕获集（s/exists/w/oldDoc）/os.ErrNotExist 判断逐条 asm 对齐；
  launcherWidgetStore CRUD 五法（Read/Ensure/Update/ReplaceWithRollback/Delete）+ loadUnlocked/writeUnlocked 底座全部落地。
- **G4 review**：仅改 desktopwidgets_store.go（追加 Update/ReplaceWithRollback + writeDesktopWidgetDocumentFile 升档 + import path/filepath）；无跨文件写重叠。

## 遗留（下一批）

- `validateDesktopWidgetDocument` 升档 [S-sig]→[S]（多 map 校验体逐条字节级还原）。
- `launcherWidgetStore.Delete` 升档 [S-sig]→[S]（bootstrapservice_config.go）。
- `bytesMatchWildcardFold` 0x1407e3e60（176 行）。
- 继续落地 truly-missing 顶层函数（678 个）。
