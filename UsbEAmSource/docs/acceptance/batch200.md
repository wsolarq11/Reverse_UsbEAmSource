# 批次 200 — buildDesktopCalendarMonth [S] 完整还原 + GetDesktopCalendarMonth 签名修正

## 基线 / 收口

| 指标 | 基线（批次 199 收口） | 收口（批次 200） |
|---|---|---|
| FUNCS | 2763 | **2763** |
| S | 1222 | **1223** |
| S-inline | 36 | 36 |
| S-sig | 1391 | **1390** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2649 | 2649（55.7%） |
| §10 差集 | 56 | **55** |

SHA256 `42CDAC7DB025558CC0D0717ABC9D8F1883112263150BA994BB3A017E9C05283D`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

## 本批内容

`desktopwidgets_calendar.go` 域完整闭环（差集 -1），新增 lunar-go v1.4.6 依赖。

### `buildDesktopCalendarMonth`（0x1407a9f60，2240B）— [S-sig]→[S]

`func buildDesktopCalendarMonth(year int, month int, now time.Time) (DesktopCalendarMonth, error)`：

1. `year∈[1901,2100]`（`rax-0x76d` 无符号 ≤0xc7）、`month∈[1,12]`（`rbx-1` ≤0xb），
   否则 `errors.New("desktopWidgets.calendarRange")`（28B @0x140c6b990）。
2. `now.IsZero()`（sec==0 && nsec==0）→ `now = time.Now()`（asm bt/wallToInternal 0xdd7b17f80
   提取 sec 后判零）。
3. `dayCount = time.Date(year, month+1, 0, 12, 0, 0, 0, time.UTC).Day()`（下月第 0 天 = 本月天数）。
4. `serverNow = now.Format(time.RFC3339Nano)`（格式串 `2006-01-02T15:04:05.999999999Z07:00`
   35B @0x140c7708d）。
5. `make([]DesktopCalendarDay, 0, dayCount)`（元素 144B=0x90，makeslice et@0x140bebe00）。
6. 逐日：`NewSolar(year, month, day, 0, 0, 0)` → `NewLunarFromSolar`；
   - `Date = fmt.Sprintf("%04d-%02d-%02d")`（14B @0x140c50704）；
   - `Day = day`；
   - `Weekday = time.Date(...,12,...).Weekday()`（absSec/86400 + 3 对 7 取余，
     magic 0xc22e450672894ab7 除以 86400 + 0x2492492492492493 除以 7 逐位模拟验证）；
   - `LunarYear = GetYearInGanZhi()`（GAN[yearGanIndex+1]+ZHI[yearZhiIndex+1]，字段 +0x30/+0x38）；
   - `LunarMonth = GetMonthInChinese()`（month<0→"闰"+MONTH[-month]，字段 +0x08）；
   - `LunarDay = GetDayInChinese()`（DAY[day]，字段 +0x10）；
   - `SolarTerm = GetJieQi()`；
   - `SolarFestivals = desktopCalendarStringList(solar.GetFestivals())`；
   - `LunarFestivals = desktopCalendarStringList(lunar.GetFestivals())`。
7. 返回 `(DesktopCalendarMonth{Year,Month,ServerNow,Days}, nil)`（7 word 结构值 + nil error =
   rax..r9 + r10/r11=0）。

### `GetDesktopCalendarMonth`（0x1407a9ac0）— 签名修正

`func (bs *BootstrapService) GetDesktopCalendarMonth(year int, month int) DesktopCalendarMonth`：
调用 `buildDesktopCalendarMonth(year, month, time.Now())`，忽略 error 返回 `DesktopCalendarMonth`。
asm 实证接收 rbx=year、rcx=month，经 time.Now 后传 now，返回 7 寄存器（56B 结构值，非 interface{}）。

### 依赖新增

`github.com/6tail/lunar-go v1.4.6`（h1:APCXi1PC3Q7gZt6RJyug/ZdZcwX2qOkzIsZIcjCQdHY=，
经 buildinfo `go version -m` 实证），走 127.0.0.1:7890 代理 `go get`。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（0.960s 真实跑）。
- **G2 count_funcs**：`FUNCS=2763 / S=1223 / S-inline=36 / S-sig=1390 / P=114 / UNMARKED=0`。
- **G3 行为**：临时 go test 实证 2024-02 月视图：29 天；02-01 为癸卯年腊月廿二、周四（Weekday=4）；
  02-10 正月初一 + 春节；02-24 正月十五 + 元宵节；month=13 / year=1800 均返回 calendarRange error。
  测试后已删（不 ship）。
- **G4 review**：改动 desktopwidgets_calendar.go（新增主函数）、bootstrapservice_callees.go（删 [S-sig]
  存根）、bootstrap_desktopwidget.go（签名修正 + import time）、go.mod/go.sum（+lunar-go）；无跨文件
  写重叠；vet/test 复验通过。

## 遗留（下一批）

- §10 差集 55 文件。desktopwidgets_clock.go 还剩 buildDesktopWorldClockSnapshot（0x1407aa960，1984B）
  主函数 [S-sig]，体含世界时钟/时区计算，独立成批；其调用点 GetDesktopWorldClockSnapshot（0x1407a9be0）
  同样需签名修正（同模式：接收参数 + 返回结构值）。
- 其余候选：`matchNodeNameTerms`（0x1407e3ac0）、qrcode.go 2 个 [S-sig] 剪贴板存根、
  `pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、`transport.go`。
