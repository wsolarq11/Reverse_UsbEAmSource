# 批次 191 — launcherWidgetStore 全方法 [S] 闭环（Delete + validateDesktopWidgetDocument 升档）

## 基线 / 收口

| 指标 | 基线（批次 190 收口） | 收口（批次 191） |
|---|---|---|
| FUNCS | 2732 | **2732** |
| S | 1179 | **1181** |
| S-inline | 36 | 36 |
| S-sig | 1402 | **1400** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2617 | 2617（55.0%） |

SHA256 `731D447F3326FDD4C4230A06ACD1343FF000E20C12D4E7D6BED72DBFEF97D654`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（组件存储 CRUD 五法 + 校验链全部 [S] 完整还原）

### 1. `launcherWidgetStore.Delete`（0x1407c0d20，992B）— [S-sig]→[S]

nil→"首页组件存储不可用"；加锁 defer 解锁；TrimSpace 空→"首页组件存储路径不能为空"；
os.Remove 失败且 !os.ErrNotExist→返回 err；成功→缓存重置为默认文档（cachedExists=false）。

### 2. `validateDesktopWidgetDocument`（0x1407c3940，2368B）— [S-sig]→[S]

校验链完整还原：版本/修订号/组件数上限；Widgets 遍历（key==ID、Type 白名单 8 值、
Title rune 计数 ≤160、Revision>0、Config 序列化 ≤0x10000）；Notes 遍历（key==WidgetID、
Body 字节 ≤0x40000、rune ≤0x10000、累计 ≤0x800000）；StopwatchLaps 遍历（单条 ≤0x3e8、
累计 ≤0x2710）；终 Marshal ≤0x2000000 后 validateLauncherConfigJSONStructure。全部错误字符串逐条 asm 对齐。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.916s`。
- **G2 count_funcs**：`FUNCS=2732 / S=1181 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`。
- **G3 行为**：launcherWidgetStore 五法（Read/Ensure/Update/ReplaceWithRollback/Delete）+ loadUnlocked/writeUnlocked +
  writeDesktopWidgetDocumentFile + validateDesktopWidgetDocument 全部 [S]；白名单 8 值（note/clock/timer/weather/
  calendar/reminder/stopwatch/worldClock）与所有错误字符串/阈值逐条 asm 对齐。
- **G4 review**：改 desktopwidgets_store.go（升档 validate + import unicode/utf8）与 bootstrapservice_config.go（升档 Delete）；无跨文件写重叠。

## 遗留（下一批）

- `bytesMatchWildcardFold` 0x1407e3e60（176 行，sync.Mutex + 闭包）。
- `TestReminderNotification` 0x1407a7320。
- `validateLauncherConfigIconData`（by-value 2352B 签名）。
- `desktopCalendarStringList` 0x1407aa820；`pluginid.go`（2 funcs）。
- 继续落地 truly-missing 顶层函数（678 个）与 §10 差集 60 文件。
