# 批次 287 · launcherupdate 父 PID 枚举 + 握手证明 +2（全 [S]，FUNCS 2940）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：进程快照枚举父 PID（queryLauncherUpdateParentPID）
与握手证明 SHA-256 十六进制计算（launcherUpdateHandshakeProof）。

## 基线 / 收口

| 指标 | 基线（batch 286 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2938 | 2940 |
| MARKED | 2938 | 2940 |
| S | 1415 | 1417 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1452 | 1454 |
| USABLE | 1453 | 1455 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1454  FUNCS=2940  MARKED=2940  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1417  USABLE=1455
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1417 + 37 + 1 + 1444 + 41 = 2940 = FUNCS`。
S 1415→1417（+2）、FUNCS 2938→2940（+2）、FAITHFUL 1452→1454（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 queryLauncherUpdateParentPID [S 0x1408c46c0, 576B]

`(pid uint32) (uint32, error)`。CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS,0)
失败 → (0,err)；defer CloseHandle；Process32First 失败 → (0,err)；循环比对
ProcessID，命中 → (ParentProcessID,nil)；Process32Next 失败 → "未找到进程父 PID"。

### 3.2 launcherUpdateHandshakeProof [S 0x1408c3b60, 512B]

`(nonce string) string`。sha256.Sum256("UsbEAm launcher update handshake v2\x00"+nonce)
→ hex.EncodeToString（小写）。常量 36 字节含 \x00 终止（movabs 立即数实证，
第 5 qword 最高字节 0x00）。hex 表 "0123456789abcdef" 已按地址解码确认。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`queryLauncherUpdateParentPID` +`launcherUpdateHandshakeProof`
  （import +crypto/sha256 +encoding/hex）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（全 [S]）。FUNCS 2940/4754 = 61.84%。下一批：launcherupdate 短函数
（namedPipe/frameWithTimeout/validateHelperExecutable 288-768B）+ 截图滚动几何。
