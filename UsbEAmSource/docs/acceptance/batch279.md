# 批次 279 · launchericonasset.go gap 差集闭合（+4 [S]，+1 [S-inline]，+5 FUNCS）

## 目标

闭合 `launchericonasset.go` 蓝图的 4 个 gap 函数（此前其余 16 个函数已散落在
`bootstrapservice_callees.go` / `bootstrapservice_state_deps.go` /
`launcherconfigiconstore.go`）：`ResolveAppIconResource`、`attachWindowManagementTargetIconURL`、
`attachWindowManagementStateIconURLs`、`persistAndRefreshWindowManagementState`。
同时修正两个连带签名问题：`AppIconOptions.CandsCap`（padding → 真实 int 字段）与
`persistWindowManagementConfig`（无参 → 带 `cfg WindowManagementConfig`）。

## 基线 / 收口

| 指标 | 基线（batch 278 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2893 | 2898 |
| MARKED | 2893 | 2898 |
| S | 1373 | 1377 |
| S-inline | 36 | 37 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1409 | 1414 |
| USABLE | 1409 | 1414 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

### G1.1 既有 flaky 测试修复（CI 阻塞项）

`TestServeAssetRequest`（launcherasset_test.go:860）原以 `RegisterBytes(..., 1)`（TTL=1ns）注册后立即
服务，条目在 lookup 处因 `now >= expiresAt` 被判过期而 404（时间分辨率相关，`-count=30` 本地复现）。
修正 TTL 为 `time.Minute`，使资产在测试窗口内有效；`-count=10 -tags production` 稳定通过。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1414  FUNCS=2898  MARKED=2898  P=41  S-eq=0  S-inline=37  S-sig=1443  S=1377  USABLE=1414
```

分项自洽：`S + S-inline + S-sig + P = 1377 + 37 + 1443 + 41 = 2898 = FUNCS`。
S 1373→1377（+4）、FUNCS 2893→2898（+5）、S-inline 36→37（+1）、FAITHFUL 1409→1414（+5）、
P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `launchericonasset.go`（4 个 [S] 方法 + `defaultAppIconOptions` [S-inline] 辅助函数）。
连带：`appicon_windows.go` 的 `AppIconOptions.CandsCap` 字段修正（不改变 ABI 布局）；
`bootstrapservice.go` 的 `persistWindowManagementConfig` 签名补全 cfg 入参。

## G3 行为（asm 逐地址实证）

### 3.1 ResolveAppIconResource [S 0x1408a17e0, 320B]

`(path string) LauncherIconResource`。asm 实证：压默认 AppIconOptions（rodata
0x141be9500，56B=7 qword；IconIndex 经 `xor ecx` 寄存器置零）→
`resolveAppIconDataWithOptions(path, opts)` → `buildLauncherIconResource("icon/app", iconData)`
（namespace 8 字符 @0x140C3C124）。

### 3.2 attachWindowManagementTargetIconURL [S 0x1408a45c0, 800B]

`(target WindowManagementTarget) WindowManagementTarget`。asm 实证：

- `normalizeWindowManagementTarget`（0x1409dfb80，identity）；
- `buildLauncherConfigIconResource("icon/window-management", target.IconRef, target.IconData)`
  （namespace 22 字符 @0x140C612AA，len=0x16）；
- 写回：`target.IconRef = res.IconRef`（@+0x60）、`target.IconURL = res.IconURL`（@+0x70）、
  `target.IconData = ""`（@+0x50 用 xmm15 清零）；
- 回退三重判断：`!launcherConfigIconSlotHasSource("", IconRef)`（data 置空 @0x1408a473d）
  && `IconURL == ""`（cmp [rsp+0x1e8], 0）&& `TrimSpace(Path) != ""`；
- 回退体：`resolveAppIconDataWithOptions(Path, 默认 opts)`（重新加载原始 Path @0x1408a477b）
  → `buildLauncherIconResource("icon/window-management", iconData)` → 写回 IconURL。

### 3.3 attachWindowManagementStateIconURLs [S 0x1408a48e0, 416B]

`(state WindowManagementState) WindowManagementState`。asm 实证：两次调用
`attachWindowManagementTargetIconURL`，分别作用于 `state.Config.Target`（target @ state+0x28）
与 `state.Target`（target @ state+0x100），逐 128B（duffcopy+0x310）写回。

### 3.4 persistAndRefreshWindowManagementState [S 0x1408a4a80, 1184B]

`(state WindowManagementState, priorErr error) (WindowManagementState, error)`。asm 实证：

- `test rbx`（priorErr.itab）非零 → 直接 `attachWindowManagementStateIconURLs(state)` 并返回 priorErr；
- `persistWindowManagementConfig(state.Config)`（cfg 经 duffcopy 自 state+0x20 入栈）err != nil
  → `attach(state)` + err；
- `windowManagement.GetState()`（bs.windowManagement @ +0x388）err != nil → `attach(state)` + err；
- 否则 `attach(st)` + nil。

### 3.5 defaultAppIconOptions [S-inline]

原 exe 无独立符号：rodata 0x141be9500 在 3.1 与 3.2 两处被逐 qword 内联复制，提取为 DRY 辅助函数。
默认语义：IconIndex=0、Namespace=""、Size=256、ImageList=4（SHIL_JUMBO）、
CandsPtr=&defaultAppIconSizes[0]（0x141965580）、CandsLen=5、CandsCap=5。

### 3.6 连带实证

- `AppIconOptions.CandsCap`：`extractAppIconDataForLookup` [0x14074a420] 的 opts 栈复制确证
  CandsCap 是真实 int 字段（+0x38），非 padding；修正后 ABI 布局不变（仍是 64B）。
- `persistWindowManagementConfig`：`persistAndRefresh` 处 `duffcopy state.Config → [rsp]` 后 call
  （persist.asm 0x1408a4b00-0xb13）确证签名带 cfg 入参。

## G4 独立复核

`launchericonasset.go`：新建，4 方法 `[S]` + 1 辅助函数 `[S-inline]`。
`appicon_windows.go`：`AppIconOptions.CandsCap` padding→int（布局不变）。
`bootstrapservice.go`：`persistWindowManagementConfig` 补 cfg 入参（骨架忽略）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

launchericonasset.go 蓝图 gap 差集闭合，窗口管理/应用图标的 `Resolve → attach → persist`
调用链全部落地。P=41 持平。FUNCS 2898/4754 = 60.96%。未落地文件差集 38→37。
