# 批次 374 · GSMTC 来源/标题读取 +2（FUNCS 3248）

## 目标

落地 2 个 WinRT GSMTC 接口方法（均 `[S]` 忠实还原）：`oledBlackoutGSMTCSession.GetSourceAppUserModelId`、
`oledBlackoutGSMTCMediaProperties.GetTitle`（后者含新类型 `oledBlackoutGSMTCMediaProperties`）。

## 基线 / 收口

| 指标 | 基线（batch 373 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3246 | 3248 |
| MARKED | 3246 | 3248 |
| S | 1506 | 1508 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1661 | 1661 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1543 | 1545 |
| USABLE | 1544 | 1546 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3248  MARKED=3248  UNMARKED=0  S=1508  S-inline=37  S-eq=1  S-sig=1661  P=41  FAITHFUL=1545  USABLE=1546  TRUE=3206
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1508 + 37 + 1 + 1661 + 41 = 3248 = FUNCS`。
FUNCS 3246→3248（+2）、S 1506→1508（+2）、FAITHFUL 1543→1545（+2）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 381→379（−2），两函数均退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutGSMTCSession.GetSourceAppUserModelId [S 0x1409204a0, 480B]

`mov rax,[rdx+0x30]`（s.vtbl[0x30] GetSourceAppUserModelId）→ `syscall.SyscallN(fn, this, &out)`（2 参）→
`test eax,eax; jl` HRESULT<0 → `fmt.Errorf`；成功路径注册 open-coded defer（deferwrap1 @0x140920680，
捕获 `out`）→ `oledBlackoutHStringToString(out)` → 返 (string, nil)。
格式串 24B `%s失败: HRESULT 0x%08X` @0x140c64c08、描述串 30B `读取系统媒体会话来源` @0x140c6f5a9 均实测。

### 3.2 oledBlackoutGSMTCMediaProperties.GetTitle [S 0x140921100, 480B]

同构：`s.vtbl[0x30] GetTitle` → SyscallN → HRESULT<0 → `fmt.Errorf("%s失败: HRESULT 0x%08X", "读取系统媒体标题", hresult)`；
成功 → defer 释放 HSTRING → `oledBlackoutHStringToString(out)`。描述串 24B `读取系统媒体标题` @0x140c653a0 实测。

## G4 独立复核

- `backend/oledblackout.go`：+GetSourceAppUserModelId [S] +oledBlackoutGSMTCMediaProperties 类型 +GetTitle [S]。
- 依赖 `oledBlackoutHStringToString`/`oledBlackoutDeleteHString`（[S-sig]，签名已对齐）保持，无新增 import。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S]）。FUNCS 3248/4754 = 68.32%，FAITHFUL 1545/4754 = 32.50%。
下一批：filesearch/oled 域 480B 候选（list_missing 头部），或 P=41 → [S] 转换。
