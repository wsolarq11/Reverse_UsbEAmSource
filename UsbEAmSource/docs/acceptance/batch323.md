# 批次 323 · IPropertyStore.SetValue/IPersistFile.Load+Save +3（FUNCS 3055）

## 目标

落地 3 个 COM 接口方法：`launcherAppIdentityIPropertyStore.SetValue`、
`launcherAppIdentityIPersistFile.Load`、`launcherAppIdentityIPersistFile.Save`。

## 基线 / 收口

| 指标 | 基线（batch 322 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3052 | 3055 |
| MARKED | 3052 | 3055 |
| S | 1478 | 1481 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1495 | 1495 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1515 | 1518 |
| USABLE | 1516 | 1519 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1518  FUNCS=3055  MARKED=3055  P=41  S-eq=1  S-inline=37  S-sig=1495  S=1481  USABLE=1519
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1481 + 37 + 1 + 1495 + 41 = 3055 = FUNCS`。
S 1478→1481（+3）、FUNCS 3052→3055（+3）、FAITHFUL 1515→1518（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 launcherAppIdentityIPropertyStore.SetValue [S 0x14086de40, 256B]

SyscallN(vtbl[0x30] SetValue, this, key, propvariant) → HRESULT <0 → fmt.Errorf。

### 3.2 launcherAppIdentityIPersistFile.Load [S 0x14086dc40, 256B]

SyscallN(vtbl[0x28] Load, this, fileName, mode) → HRESULT <0 → fmt.Errorf。

### 3.3 launcherAppIdentityIPersistFile.Save [S 0x14086dd40, 256B]

SyscallN(vtbl[0x30] Save, this, fileName, remember) → HRESULT <0 → fmt.Errorf。

## G4 独立复核

- `backend/launcherappidentity_windows.go`：+launcherAppIdentityIPersistFile 类型
  +SetValue +Load +Save [S]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S]）。FUNCS 3055/4754 = 64.26%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
