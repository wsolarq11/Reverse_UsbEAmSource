# 批次 266 · windowmanagement 光标环绕线程链 +3 [S]（startCursorWrap 480B / stopCursorWrap 256B / cursorWrapLoop 1248B + 闭包）

## 目标

落地 HANDOFF batch265 尾部遗留的 windowmanagement 域三函数并完成整链闭环：
goroutine 启动 `startCursorWrap`（0x1409ecaa0，480B）、停止收口 `stopCursorWrap`
（0x1409ecce0，256B）、主循环 `cursorWrapLoop`（0x1409ecde0，1248B）及其编译闭包
func1（0x1409ed360）、deferwrap1（0x1409ed440）、deferwrap2（0x1409ed300）。
三函数原为 [S-sig]（签名已定型、体为存根），本批升档 [S]。

## 基线 / 收口

| 指标 | 基线（batch 265 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2860 | 2860 |
| MARKED | 2860 | 2860 |
| S | 1328 | 1331 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1453 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1364 | 1367 |
| USABLE | 1364 | 1367 |
| 真函数（S+S-inline+S-sig） | 2820（59.32%） | 2820（59.32%） |

`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

## G1 编译

`go1.25.12 build -tags production -trimpath ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1367  FUNCS=2860  MARKED=2860  P=40  S-eq=0  S-inline=36  S-sig=1453  S=1331  USABLE=1367
```

三函数升档 [S]（S 1328→1331，S-sig 1456→1453），FUNCS/P/S-inline/UNMARKED 持平（升档非新增）。
闭包（gowrap1/func1/deferwrap1/deferwrap2）为编译产物，不计入 count_funcs，随三函数自动生成。

## G3 行为（asm 逐地址实证 + 生命周期黄金测试）

### 3.1 startCursorWrap [S 0x1409ecaa0]（480B）

签名（morestack spill 3 寄存器实证）：`(horizontal, vertical bool)`，AX=s、BX=horizontal、
CX=vertical。无 nil 接收者检查（0x1409ecac0 直接 `lea rdx,[rax+8]`）。

1. `lock cmpxchg [rsi+8],1` 加锁 → `cmp byte[rsi+0x110],0; je 0x1409ecb2b`：cursorActive
   非 0 → 显式 `lock xadd -1` Unlock 后 return（0x1409ecb0c-0x1409ecb2a，**非 defer**）。
2. `makechan` 两次（0x1409ecb2b / 0x1409ecb3e，类型 0x1409f5e00=chan struct{}，ebx=size=0）
   得 stop/done 无缓冲 channel。
3. `[rcx+0x100]=stop`、`[rcx+0x108]=done`、`byte[rcx+0x110]=1`（cursorActive=true）→ Unlock
   （0x1409ecb93/0x1409ecb9a/0x1409ecba1）。
4. `newobject`（闭包类型 0x140ae780，5 捕获）填 `[+0]=gowrap1`、`[+8]=s`、`[+0x10]=h`、
   `[+0x11]=v`、`[+0x18]=stop`、`[+0x20]=done` → `runtime.newproc`（0x1409ecc30）。
5. gowrap1（0x1409ecc80）：`bl=[rdx+0x10] cl=[rdx+0x11] rdi=[rdx+0x18] rsi=[rdx+0x20]
   rax=[rdx+8]` → `call cursorWrapLoop`，即 `go func(){ s.cursorWrapLoop(h,v,stop,done) }()`。

### 3.2 stopCursorWrap [S 0x1409ecce0]（256B）

签名：`()`，AX=s。无 nil 检查、**无 cursorActive 前置检查**（幂等，0x1409eccf3 直接加锁）。

锁内 `rax=[rdx+0x100]`(stop)、`rsi=[rdx+0x108]`(done) → `movups xmm15,[rdx+0x100]`（清
cursorStop+cursorDone 16B）+ `byte[rdx+0x110]=0`（cursorActive=false）→ Unlock
（0x1409ecd60-0x1409ecd82）。随后 `stop!=nil → closechan`（0x1409ecda5）；`done!=nil →
chanrecv1`（0x1409ecdb9，即 `<-done` 丢弃值，阻塞至 cursorWrapLoop 收尾 close(done)）。

### 3.3 cursorWrapLoop [S 0x1409ecde0]（1248B）+ 闭包

签名（morestack spill 5 寄存器 + gowrap1 调用实证）：`(horizontal, vertical bool, stop,
done chan struct{})`，AX=s、BX=h、CX=v、DI=stop、SI=done。

defer 链（open-coded bit 0/1/2 于 `byte[rsp+0x37]`，声明序 = 装配序）：

| 声明序 | 符号 | 捕获 | 语义 |
|---|---|---|---|
| 1 | deferwrap1 0x1409ed440 | SI=done | `close(done)`（closechan） |
| 2 | func1 0x1409ed360 | AX=s, DI=stop | 见下 |
| 3 | deferwrap2 0x1409ed300 | ticker | `ticker.Stop()`（内联 stopTimer） |

执行序（LIFO，0x1409ed233/0x1409ed272）：ticker.Stop → func1 → close(done)。

func1（0x1409ed360，176B）：`rcx=[rdx+8]`(s)、`rdx=[rdx+0x10]`(stop) → lock →
`cmp [rcx+0x100],rdx`（s.cursorStop==stop）→ 相等才 `movups xmm15,[rcx+0x100]`（清
cursorStop+cursorDone）+ `byte[rcx+0x110]=0`（cursorActive=false）→ unlock。即防止旧
goroutine 收尾时误清新一代 cursorStop 的竞态守卫。

主循环：

1. `time.NewTicker(0x7a1200=8,000,000ns=8ms)`（0x1409ece84）。
2. `selectgo(cas0=[rsp+0x180], order0=[rsp+0x38], pc0=0, nsends=0, nrecvs=2, block=1)`
   （0x1409ecf76）：case0=ticker.C、case1=stop；返回 index!=0（stop 命中）→ return
   （0x1409ecf80 `test rax,rax; jne 0x1409ed272`）。
3. 锁内 `duffcopy+0x2bc` 拷 config（0xe0=224B）到栈副本 + 读 `[rax+0x118]`(cursorGuardPx)
   + `[rax+0xf0]`(moduleEnabled) → Unlock；副本 `[rsp+0xa0]`(wrapX)/`[rsp+0xa1]`(wrapY)。
4. 判断链：`!moduleEnabled → return`（0x1409ed065 je）；`!wrapX && !wrapY → return`
   （0x1409ed06c/0x1409ed06f）；`getSystemMetrics(0x50=SM_CMONITORS)<=1 → return`
   （0x1409ed09e `cmp rax,1; setle cl`，单显示器无环绕）。
5. `time.Since(lastWrap) < 0x8583b00(140ms) → continue`（0x1409ed0ef）；`time.Since(
   lastDisplayRefresh) >= 0x1dcd6500(500ms)` 或 `len(monitors)==0 → 刷新
   windowManagementDisplayRects + lastDisplayRefresh=now`（0x1409ed149/0x1409ed156/0x1409ed15b）。
6. `windowManagementMaybeWrapCursor(wrapX, wrapY, monitors, guardPx)` 真 → `lastWrap=now`
   （0x1409ed1e8/0x1409ed1ef/0x1409ed205）。

**死参数**：horizontal/vertical 经 morestack spill bl/cl 后从未读取（循环初始化
0x1409ecebf `xor ebx/ecx` 覆盖），函数体从 config 实时值取 wrapX/wrapY，故保留签名不置用。

### 3.4 生命周期黄金测试（`windowmanagement_windows_test.go`）

`TestWindowManagementCursorWrapLifecycle`：start 置 cursorActive/双 channel → 二次 start
被 cursorActive 短路不覆盖 → stop 清空三态。`TestWindowManagementStopCursorWrapIdempotent`：
未启动 stop 为 no-op 不 panic。moduleEnabled 默认 false → cursorWrapLoop 首次 tick 或
stop 关闭即 return，不触达 GetSystemMetrics/MaybeWrapCursor 系统调用，确定性安全。全 PASS。

## proc 身份内存实证

| 全局槽位 | 实测 LazyProc.Name |
|---|---|
| 0x141BC1DB0 | GetSystemMetrics（cursorWrapLoop 0x1409ed084 读 `[rip+0x11d4d25]`，Python `hex(0x1409ed08b+0x11d4d25)=0x141BC1DB0`） |

## G4 独立复核

仅改 `backend/windowmanagement.go`（新增 `time` import + 三函数体升档 [S]）、
`backend/windowmanagement_windows_test.go`（追加生命周期黄金测试，不计入 count_funcs）。
无既有函数签名/行为变更；全量测试回归 PASS。

## 下一批

继续 gap_aggregate.txt 剩余短函数落地。P=40。FUNCS 2860/4754 = 60.16%。
