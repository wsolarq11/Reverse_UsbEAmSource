# 批次 376 · weather 提供器错误快照/写入 +1（FUNCS 3253）

## 目标

落地 desktopWidgetWeatherService 提供器错误域 2 个方法：

- 新增 `[S]`：`desktopWidgetWeatherService.setProviderError`（0x1407c7040，544B）。
- 升级 `[S-sig]`→`[S]`：`desktopWidgetWeatherService.providerErrorSnapshot`（0x1407c72c0，448B，返回类型 `interface{}`→`map[string]string`）。

## 基线 / 收口

| 指标 | 基线（batch 375 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3252 | 3253 |
| MARKED | 3252 | 3253 |
| S | 1514 | 1516 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1659 | 1658 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1551 | 1553 |
| USABLE | 1552 | 1554 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3253  MARKED=3253  UNMARKED=0  S=1516  S-inline=37  S-eq=1  S-sig=1658  P=41  FAITHFUL=1553  USABLE=1554  TRUE=3211
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1516 + 37 + 1 + 1658 + 41 = 3253 = FUNCS`。
FUNCS 3252→3253（+1）、S 1514→1516（+2）、S-sig 1659→1658（−1）、FAITHFUL 1551→1553（+2）、TRUE 3210→3211（+1）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 375→374（−1），`setProviderError` 退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 setProviderError [S 0x1407c7040, 544B]

结构字段锁定（types_desktopwidget.go `desktopWidgetWeatherService`）：`mu sync.Mutex`@+0x30、
`providerErrors map[string]string`@+0x40。

- `test rax,rax; je` → s==nil 直接 return（0x1407c71f8）。
- `normalizeDesktopWeatherProviderID(providerID)` 后 `test rbx,rbx; je` → id 空直接 return（0x1407c71f2）。
- 加锁：`lock cmpxchg dword [rdx+0x30],1` 成功 `sete/jne` 直入，否则 `internal/sync.Mutex.lockSlow`；
  defer 闭包 @0x1407c71fe（`lea rsi,[rip+0x17c]`）持 `&s.mu` 解锁，`byte [rsp+0x27]=1`。
- `cmp [rdx+0x40],0; jne` → providerErrors nil 则 `makemap_small` 初始化（0x1407c7108）。
- `TrimSpace(errMsg)` 空 → `mapdelete_faststr`（0x1407c71be）；非空 → `mapassign_faststr` 写
  `providerErrors[id]=msg`（0x1407c715a）。

### 3.2 providerErrorSnapshot [S 0x1407c72c0, 448B]

- `makemap_small` 新建 result；`test rcx,rcx; je` → s==nil 返回空 map（0x1407c73a7）。
- 加锁同构（cmpxchg+lockSlow），defer 解锁 @0x1407c7456。
- `rbx=[rcx+0x40]`（providerErrors）→ `mapIterStart`/`mapIterNext` 遍历，
  `mapassign_faststr(result, k)=v` 复制（0x1407c73ec），返回 result。

## G4 独立复核

- `backend/desktopwidgets_weather.go`：+`setProviderError [S]`（`map[string]string` 字段语义），
  `providerErrorSnapshot` 返回类型 `interface{}`→`map[string]string` 且体升 `[S]`。
- 依赖 `normalizeDesktopWeatherProviderID`（[S]）保持；无新增 import（`strings` 已有）。
- 无其他调用点依赖旧 `interface{}` 返回（`go build` 全仓通过）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 函数落地（1 新增 [S] + 1 升级）。FUNCS 3253/4754 = 68.43%，FAITHFUL 1553/4754 = 32.67%，TRUE 3211/4754 = 67.54%。
下一批：desktopwidgets_weather 域继续（`setProviderError` 依赖的 provider 状态链），
或 bookmarks_firefox.go 剩余、filesearch 域 480B 候选、P=41 → [S] 转换。
