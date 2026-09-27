# 批次 201 — buildDesktopWorldClockSnapshot [S] 完整还原 + GetDesktopWorldClockSnapshot 签名修正

## 基线 / 收口

| 指标 | 基线（批次 200 收口） | 收口（批次 201） |
|---|---|---|
| FUNCS | 2763 | **2763** |
| S | 1223 | **1224** |
| S-inline | 36 | 36 |
| S-sig | 1390 | **1389** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2649 | 2649（55.7%） |
| §10 差集 | 55 | **54** |

SHA256 `ABD47A9F955BB60D6C60C68B67C9750869159B68B50B885268AD89D73A46E5BE`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

## 本批内容

`desktopwidgets_clock.go` 域完整闭环（差集 -1）。

### `buildDesktopWorldClockSnapshot`（0x1407aa960，1984B）— [S-sig]→[S]

`func buildDesktopWorldClockSnapshot(timezones []DesktopWorldClockZone, now time.Time) (DesktopWorldClockSnapshot, error)`：

1. `len(timezones) > 24` → `errors.New("desktopWidgets.tooManyTimezones")`（31B @0x140c70d7c）。
2. `now.IsZero()` → `time.Now()`；`serverNow = now.Format(time.RFC3339Nano)`。
3. `make([]DesktopWorldClockTime, 0, len)` + `make(map[string]struct{}, len)` 去重。
4. 逐时区：`id = TrimSpace(zone.TimezoneID)`，空或已见 → 跳过；
   `time.LoadLocation(id)` 失败 → `fmt.Errorf("desktopWidgets.invalidTimezone: %s"@0x140c75595, id)`；
   `localNow = now.In(loc)`（loc==UTC 内部转 nil，asm cmove）；`_, offset = localNow.Zone()`；
   `LocalTime = localNow.Format(time.RFC3339)`（`2006-01-02T15:04:05Z07:00` 25B）；
   `Offset = desktopTimezoneOffset(offset)`；`OffsetSeconds = offset`；`DST = localNow.IsDST()`；
   `Label = TrimSpace(zone.Label)`。
5. 返回 `(DesktopWorldClockSnapshot{ServerNow, Clocks}, nil)`（5 word 结构值 + nil error）。

### `GetDesktopWorldClockSnapshot`（0x1407a9be0）— 签名修正

`func (bs *BootstrapService) GetDesktopWorldClockSnapshot(timezones []DesktopWorldClockZone) DesktopWorldClockSnapshot`：
调用 `buildDesktopWorldClockSnapshot(timezones, time.Now())`，忽略 error 返回 `DesktopWorldClockSnapshot`
（5 word = ServerNow string 2 word + Clocks slice 3 word，原误标 `() interface{}`）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（0.970s 真实跑）。
- **G2 count_funcs**：`FUNCS=2763 / S=1224 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`。
- **G3 行为**：临时 go test 实证 Asia/Shanghai 返回 Offset=UTC+08:00 / OffsetSeconds=28800 /
  DST=false / LocalTime RFC3339 格式正确；重复时区去重 + 空时区跳过（2 输入 → 1 输出）；
  无效时区 → invalidTimezone error；25 时区 → tooManyTimezones error。测试后已删。
- **G4 review**：改动 desktopwidgets_clock.go（新增主函数）、bootstrapservice_callees.go（删 [S-sig]
  存根）、bootstrap_desktopwidget.go（签名修正，import time 已存在）；无跨文件写重叠；vet/test 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。desktopwidgets 系列（calendar/clock）已闭环，其余候选：`matchNodeNameTerms`
  （0x1407e3ac0）、qrcode.go 2 个 [S-sig] 剪贴板存根、`pluginupdate.go`、`pluginwindow.go`、
  `twofactor.go`、`transport.go`、`TestReminderNotification`（0x1407a7320）。
