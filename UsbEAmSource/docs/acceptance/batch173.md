# 批次 173 验收（audio_windows_runtime.go：快照构造 + 会话设备持久化 [P]→[S-sig] 签名订正）

日期：2026-09-25
子批次：buildSessionSnapshot / setPersistedSessionDevice

## 目标

`buildSessionSnapshot`（0x140756220）、`setPersistedSessionDevice`（0x140756920）两个 `[P]`
升 [S-sig]：签名逐寄存器实证，并订正旧树的错误近似（指针→值返回、漏 processID 参数）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test -count=1 ./backend` | ok (0.4s) |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1346 / P=174 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1346 = 2485 / 4754 = 52.3%**（相对批次 172 **增 +2**；P 176→**174**，
S-sig 1344→1346，两函数由 [P] 升 [S-sig]）。

重建产物 SHA256：`4B7824C9C8AA6E85EE7B94E47B49E0BACE3748FC39BC188BC62CBCA8368A225C`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数，`backend/audio_windows_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| buildSessionSnapshot | 0x140756220 | [P] (sessionKey string)(*audioSessionSnapshot, error) | [S-sig] (sessionKey string)(audioSessionSnapshot, error) |
| setPersistedSessionDevice | 0x140756920 | [P] (sessionID, deviceID string) error | [S-sig] (processID uint32, deviceType string, deviceID string) error |

## G4 关键实证结论

1. **buildSessionSnapshot 按值返回结构**：序言 spill rax=receiver、rbx/rcx=sessionKey(ptr,len)；
   0x140756252 `lea rdi,[rsp+0xe0]` 隐藏返回指针 + `duffzero+0x142` 清零返回结构——按值返回，非指针。
   字段装配 [rsp+0xe0..0x158] 与 `audioSessionSnapshot` 12 字段逐字节对齐（GroupID/AppName/ProcessID/
   ProcessPath/IconData/VolumePercent/Muted/Active/SystemSounds/OutputDeviceID/InputDeviceID/
   DeviceRoutingSupported）。成功路径 `xor eax,eax; xor ebx,ebx` 置 error=nil。旧树指针近似已订正为值返回。
2. **setPersistedSessionDevice 三参 + error 返回**：调用点 0x140753132 传 rax=receiver、
   ebx=dword(processID)、(rcx,rdi)=deviceType、(rsi,r8)=deviceID；体 0x140756a04 比对
   deviceType=="capture"（`0x74706163 0x7275 0x65` 共 7B，即 "capture"），`strings.TrimSpace(deviceID)`
   后 `runtime.concatstring3` 拼 WinRT 持久化键 → `audioCreateHString` →
   `SetPersistedDefaultAudioEndpoint`。旧树 (sessionID,deviceID) 漏 processID、误 deviceType 为
   sessionID，已订正。
3. **会话持久化语义链**：getPersistedSessionDevice(deviceType)→(deviceID,ok) 与
   setPersistedSessionDevice(processID,deviceType,deviceID)→error 与
   resolveSessionDeviceRouting(processID,flag)→(render,capture,ok) 三者键语义已闭环——
   按（进程号, 渲染/捕获方向）读写持久化音频设备，与批次 172 结论一致。

## 残留 / 未落地（不触及）

- resolveSessionAppName（0x140757a00，[P]）：asm 序言含 rcx/rdi/rsi/r8/r9d/bl 多寄存器（疑似多 string+
  uint32+bool），非旧树 (sessionID string) bool，待专项逐寄存器实证。
- audioBuildSessionGroupID（0x1407588c0，[P]）、(*audioSessionAccumulator).add（0x140757e20，[P]）待实证。
