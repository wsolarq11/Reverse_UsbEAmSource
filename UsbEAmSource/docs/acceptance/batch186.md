# 批次 186 — 首页组件归一化链 + 存储文档归一化（4 函数 → [S]）

## 基线 / 收口

| 指标 | 基线（批次 185 收口） | 收口（批次 186） |
|---|---|---|
| FUNCS | 2717 | **2721** |
| S | 1165 | **1169** |
| S-inline | 36 | 36 |
| S-sig | 1401 | 1401 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2602 | **2606** |

SHA256 `131E08CCA0653D9B376B6CC69880FB7DC05E13209AEFBCAB9E8AD1078AF40C97`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

### 1. `normalizeDesktopReminderScheduleKind`（0x1407bf740，160B）— desktopwidgets_service.go

`func(kind string) string`。TrimSpace 后 == "water"（len5 + "wate"@0x65746177+'r'）或
=="sedentary"（len9 + "sedentar"@0x7261746e65646573+'y'）→ "interval"（8B @0x140c3c164）；
否则原样返回 TrimSpace(kind)。

### 2. `normalizeDesktopReminderTimezone`（0x1407bf220，896B）— desktopwidgets_service.go

`func(kind string, cfg DesktopWidgetConfig) DesktopWidgetConfig`。仅 kind=="reminder"（len8 +
"reminder"@0x7265646e696d6572）且 cfg.Reminder 非空才归一化；def=*cfg.Reminder 拷贝（newobject
type size=0xb0 @0x140bfe260 = DesktopReminderDefinition）；ScheduleKind 走 #1；Preset=="water"/"sedentary"
→"custom"（6B @0x140c37632）；TimezoneID="Local"（5B @0x140c35c3b）；音频走 normalizeDesktopReminderAudio；
cfg.Reminder=&def。

### 3. `normalizeDesktopWidgetConfig`（0x1407bf5a0，416B）— desktopwidgets_service.go

`func(kind string, cfg DesktopWidgetConfig) DesktopWidgetConfig`。cfg=normalizeDesktopReminderTimezone(kind,cfg)；
kind=="clock"（len5 + "cloc"@0x636f6c63+'k'）时 TrimSpace(cfg.HourFormat)=="12"（0x3231）→"12"，否则→"24"；
其余 kind 不改 HourFormat。

### 4. `normalizeDesktopWidgetDocument`（0x1407bfec0，1376B）— desktopwidgets_store.go

`func(doc DesktopWidgetDocument) DesktopWidgetDocument`。Version==0→1；九个 map 字段 nil→空 map
（Widgets/Notes/Reminders/Timers/Stopwatches/StopwatchLaps/WeatherCache/NotificationDeliveries/ProtectedSecrets）；
遍历 Widgets → widget.Config=normalizeDesktopWidgetConfig(widget.Type, widget.Config) 后写回；
遍历 Reminders → reminder.ScheduleKind=normalizeDesktopReminderScheduleKind(reminder.ScheduleKind) 后写回；返回 doc。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.925s`。
- **G2 count_funcs**：`FUNCS=2721 / S=1169 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`。
- **G3 行为**：所有字符串常量（"reminder"/"water"/"sedentary"/"interval"/"custom"/"Local"/"clock"/"12"/"24"）
  经 read_gostring.py 字节级解码；DesktopWidgetConfig.Reminder@0x40、DesktopReminderDefinition 字段偏移
  （ScheduleKind@0x00/TimezoneID@0x40/Preset@0x78/AudioMode@0x88）与 DesktopWidgetDocument 九个 map 字段
  从 asm 逐条对齐 types_desktopwidget.go。
- **G4 review**：改 desktopwidgets_service.go（追加 2 函数）+ desktopwidgets_store.go（追加 1 函数）；无跨文件写重叠。

## 遗留（下一批）

- `launcherWidgetStore.Read`（0x1407c0420，6 行）已解析（nil→"首页组件存储不可用"，加锁 defer 解锁调 loadUnlocked），
  但依赖 loadUnlocked（0x1407c2440，69 行，读文件+JSON 解析+缓存，依赖链含 readDesktopWidgetStoreBytes/
  cloneDesktopWidgetDocument/validateDesktopWidgetDocument/ensureDesktopWidgetJSONEOF）——留待闭环。
- `bytesMatchWildcardFold` 0x1407e3e60（176 行，通配符匹配）。
- 继续落地 truly-missing 顶层函数。
