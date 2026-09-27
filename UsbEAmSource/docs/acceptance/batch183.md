# 批次 183 — 首页组件提醒音播放/下发链闭环（2 方法 + 闭包）

## 基线 / 收口

| 指标 | 基线（批次 182 收口） | 收口（批次 183） |
|---|---|---|
| FUNCS | 2712 | **2714** |
| S | 1158 | **1160** |
| S-inline | 36 | 36 |
| S-sig | 1403 | 1403 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2597 | **2599** |

SHA256 `D7DF6AC7C96C9E517DBDF644FA087F91705DDE0A9F211CBBAA7B9B5F0B090E71`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（desktopwidgets_audio.go，2 个 truly-missing 方法 → [S]）

### `(*platformDesktopWidgetAudioPlayer)Play` 0x1407a6ae0（544B）

`func(audio desktopReminderAudio) error`。`normalizeDesktopReminderAudio(Mode,Path,RepeatCount)` → `validateDesktopReminderAudio` 失败即返回；`name=="custom"`（len==6 + `"cust"`@0x74737563 + `"om"`@0x6d6f）时 `os.Stat(path)` 失败或 `info.Mode()&os.ModeType!=0`（test eax,0x8f280000=ModeType 逐位确证）→ `"desktopWidgets.reminderAudioUnavailable"`（39B @0x140c7dc2b）；否则 `go func1` 异步播放并立即返回 nil。

### `(*platformDesktopWidgetAudioPlayer).Play.func1` 0x1407a6d00（匿名闭包）

`p.playback.Lock()`（lock cmpxchg）+ `defer p.playback.Unlock()` + `playDesktopReminderAudioSequence(name,path,vol)`，失败 `log.Printf("播放首页组件提醒音频失败: %v" /*40B @0x140c7ef12*/, err)`。

### `(*desktopWidgetService)deliverNotification` 0x1407a6f20（736B）

`func(title, message string, audio *desktopReminderAudio) error`。s==nil → `"desktopWidgets.notificationFailed"`（33B）；`s.notifier!=nil` → `Notify(title,message,audio!=nil)`，否则 `"通知服务不可用"`（21B）；`audio!=nil` 时 `s.audio!=nil` → `Play(*audio)`，否则 `"音频服务不可用"`（21B）；notifErr 非空 → `log.Printf("发送首页组件通知失败: %v" /*34B @0x140c75551*/)` + notificationFailed；playErr 非空 → `log.Printf("播放首页组件提醒音频失败: %v" /*40B*/)` + `"desktopWidgets.reminderAudioFailed"`（34B @0x140c75573）；否则 nil。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.909s`。
- **G2 count_funcs**：`FUNCS=2714 / S=1160 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`。
- **G3 行为**：Play 的 custom 文件类型检查经 `os.ModeType=0x8f280000` 逐位分解确证（ModeDir|ModeSymlink|ModeNamedPipe|ModeSocket|ModeDevice|ModeCharDevice|ModeIrregular）；deliverNotification 的 notifErr/playErr 分派顺序与 3 个最终错误串（notificationFailed/reminderAudioFailed）+ 2 个缺服务串（通知/音频服务不可用）全部 `read_gostring.py` 字节级解码；Notify 第三个参数为 `audio!=nil`（setne 确证）。
- **G4 review**：单文件 desktopwidgets_audio.go 追加两方法 + import log/os；无签名变更、无并行写重叠。

## 遗留（下一批）

- `TestReminderNotification` 0x1407a7320（依赖 `launcherWidgetStore.Read` 0x1407c0420 与 `normalizeDesktopWidgetTitle` 0x1407bef80，均未落地）；`shouldDisplay` 亦未落地。
