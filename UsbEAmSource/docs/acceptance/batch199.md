# 批次 199 — desktopwidgets_calendar.go desktopCalendarStringList [S]

## 基线 / 收口

| 指标 | 基线（批次 198 收口） | 收口（批次 199） |
|---|---|---|
| FUNCS | 2762 | **2763** |
| S | 1221 | **1222** |
| S-inline | 36 | 36 |
| S-sig | 1391 | 1391 |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2648 | 2649（55.7%） |
| §10 差集 | 56 | **56** |

SHA256 `A3DD9DB9448D27603281E6A8FB1840464E617FFCD82CBDC6832BBDF9337D3DD9`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,312,000 B，`bash build.sh` 重建）。

## 本批内容

新建 `backend/desktopwidgets_calendar.go`，落地 1 个辅助函数 [S]（新建，非 [S-sig] 升级）。

### `desktopCalendarStringList`（0x1407aa820，320B）— [S]

`func desktopCalendarStringList(l *list.List) []string`：

1. `l == nil || l.Len() == 0` → 返回 nil slice（asm `xor eax/ebx; mov rcx,rbx`）。
2. `make([]string, 0, l.Len())`（asm `makeslice(stringType@0x140ae9e00, 0, [rax+0x28])`）。
3. 遍历 `l.Front()`（`[rax+0x28]==0 → nil`，否则 `[rax]`=root.next）至 `e.Next()`：
   `e.Value.(string)` 断言成功（`[e+0x18]==stringType`）且非空（`[e+0x20]` 指向的 string len≠0）
   → append；`e.Next()` 终止条件 `e.list==nil || e.next==e.list`（`[e+0x10]` 哨兵）。
4. 返回切片。

### 关键实证：container/list 布局精确吻合

参数 `*list.List`（48B，rtype@0x140b85860 Size=48/Kind=Struct）＝内嵌
`root Element`（40B）+ `len int`（+0x28）；`Element`（40B，rtype@0x140bb87c0）＝
`next/prev/list`（+0x00/+0x08/+0x10）+ `Value any`（+0x18 type / +0x20 data）。
`GetFestivals()`（lunar-go v1.4.6 源码实证）返回 `*list.List`，与本函数参数完全对应；
string 类型描述符经 lea 目标重算为 0x140ae9e00（Size=16/Kind=24，早前 0x140c29e00 系
手算进位错误，已纠正）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（0.958s 真实跑）。
- **G2 count_funcs**：`FUNCS=2763 / S=1222 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`。
- **G3 行为**：纯标准库 `container/list`，无 Win32/COM 依赖；类型描述符与 GetFestivals
  返回类型经源码 + rtype 双重实证，未臆造。
- **G4 review**：仅新建 desktopwidgets_calendar.go；无跨文件写重叠；vet/test 复验通过。

## 遗留（下一批）

- §10 差集 56 文件不变（desktopwidgets_calendar.go 还剩 buildDesktopCalendarMonth 主函数
  0x1407a9f60 未落地，仍在差集）。buildDesktopCalendarMonth 签名经调用点 0x1407a9ac0 实证为
  `(year int, month int, now time.Time)`，返回 `DesktopCalendarMonth`（56B=7 寄存器），
  体含 lunar-go 农历/节气/节日逐日计算，解码量大，独立成批。
- 其余候选：`matchNodeNameTerms`（0x1407e3ac0）、qrcode.go 2 个 [S-sig] 剪贴板存根、
  `pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、`transport.go`。
