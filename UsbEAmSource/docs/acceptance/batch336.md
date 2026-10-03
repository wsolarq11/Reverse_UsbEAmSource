# 批次 336 · 键标签/插件逻辑名/配置快照/播放类型 +4（FUNCS 3099）

## 目标

落地 4 个函数：`inputMonitorKeyLabel`（键标签）、
`normalizePluginWindowLogicalName`（插件逻辑名归一化）、
`launcherConfigSnapshotsEqual`（配置快照比较）、
`oledBlackoutGSMTCPlaybackInfo.GetPlaybackType`（播放类型）。

## 基线 / 收口

| 指标 | 基线（batch 335 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3095 | 3099 |
| MARKED | 3095 | 3099 |
| S | 1496 | 1498 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1520 | 1522 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1533 | 1535 |
| USABLE | 1534 | 1536 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1535  FUNCS=3099  MARKED=3099  P=41  S-eq=1  S-inline=37  S-sig=1522  S=1498  USABLE=1536
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1498 + 37 + 1 + 1522 + 41 = 3099 = FUNCS`。
S 1496→1498（+2）、S-sig 1520→1522（+2）、FUNCS 3095→3099（+4）、FAITHFUL 1533→1535（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 inputMonitorKeyLabel [S 0x140869820, 288B]

'A'-'Z'/'0'-'9'→单字符；0x70-0x87（VK_F1-F24）→"F"+FormatInt(vk-0x6f)；否则查全局 map，
命中返回标签，未命中 fmt.Sprintf("VK%d", vk)。

### 3.2 normalizePluginWindowLogicalName [S 0x140930f00, 288B]

TrimSpace(name) 空→TrimSpace(fallback)；仍空→默认（6B "window"）；countrunes>128→截断 128 rune。

### 3.3 launcherConfigSnapshotsEqual [S-sig 0x14089cc00, 288B]

bool 标志不等→false；皆 false→true；字段字节比较→launcherConfigRevision 两次→revision 相等
则 memequal。

### 3.4 GetPlaybackType [S-sig 0x140920ee0, 288B]

SyscallN(vtbl[0x40] GetPlaybackType)→HRESULT<0→(0,false)；ref nil→(0,false)；否则
IReference.GetInt32→(type,true)。

## G4 独立复核

- `backend/inputmonitor_windows.go`：+inputMonitorKeyLabel [S]（+fmt import）。
- `backend/bootstrapservice_state_deps.go`：+normalizePluginWindowLogicalName [S]
  +launcherConfigSnapshotsEqual [S-sig]。
- `backend/oledblackout.go`：+GetPlaybackType [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（2 [S] + 2 [S-sig]）。FUNCS 3099/4754 = 65.19%。下一批：filesearch 域
volumeIndexMappedReadProvider.releaseLease / VolumeIndex.CanReloadStaticMmapCheckpoint。
