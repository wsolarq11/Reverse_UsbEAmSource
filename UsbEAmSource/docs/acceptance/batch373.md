# 批次 373 · 启动器更新通知门禁/显示 +2（FUNCS 3246）

## 目标

落地 2 个函数（均 `[S]` 忠实还原）：`BootstrapService.beginLauncherUpdateNotification`、
`BootstrapService.ShowLauncherUpdateNotification`（Wails 服务方法）。

## 基线 / 收口

| 指标 | 基线（batch 372 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3244 | 3246 |
| MARKED | 3244 | 3246 |
| S | 1504 | 1506 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1661 | 1661 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1543 |
| USABLE | 1542 | 1544 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3246  MARKED=3246  UNMARKED=0  S=1506  S-inline=37  S-eq=1  S-sig=1661  P=41  FAITHFUL=1543  USABLE=1544  TRUE=3204
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1506 + 37 + 1 + 1661 + 41 = 3246 = FUNCS`。
FUNCS 3244→3246（+2）、S 1504→1506（+2）、FAITHFUL 1541→1543（+2）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 383→381（−2），两函数均退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 beginLauncherUpdateNotification [S 0x140795fc0, 256B]

`lock cmpxchg [rcx+0x540]`（bs.lock 快路径，失败走 `sync.Mutex.lockSlow`）→ 注册 defer unlock →
`cmp byte [rcx+0x4f1]`（launcherUpdateNotificationShown）：非零→返 false（先 unlock）；零→
`mov byte [rcx+0x4f1],1` → 返 true。字段偏移经 `unsafe.Offsetof` 实测：lock=0x540、
launcherUpdateNotificationShown=0x4f1，与 asm 逐字对齐。

### 3.2 ShowLauncherUpdateNotification [S 0x140795ec0, 256B]

`test rax,rax`（bs==nil）→ 早返；spill rbx/rcx/rdi/rsi（title/message 四寄存器）→
`call beginLauncherUpdateNotification` → `test al,al` 为 false→早返；否则 reload 四寄存器 →
`call showLauncherUpdateNotification(title,message)` → `test rax,rax`（error itab）非 nil 时
`fmt.Fprintf(os.Stderr, "显示启动更新通知失败: %v\n", err)`。格式串 35B 实测于 0x140c770b0；
写入目标为全局 `*os.File`（`mov rbx,[rip+0x1479921]`→.data@0x141c0f868，经 itab@0x1411d2cc0
装箱 io.Writer）。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+beginLauncherUpdateNotification [S] +ShowLauncherUpdateNotification [S]。
- `backend/bootstrapservice_test.go`：`TestBootstrapServiceLayout` 增 launcherUpdateNotificationShown=0x4f1 锁定；
  +TestBeginLauncherUpdateNotificationOnce（首调 true / 次调 false）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S]）。FUNCS 3246/4754 = 68.28%，FAITHFUL 1543/4754 = 32.46%。
下一批：filesearch/oled 域 448B/480B 候选（list_missing 头部），或 P=41 → [S] 转换。
