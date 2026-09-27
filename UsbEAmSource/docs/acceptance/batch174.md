# 批次 174 验收（audio_windows_runtime.go：会话应用名解析 + 会话组 ID 构建 [P]→[S-sig] 签名订正）

日期：2026-09-25
子批次：resolveSessionAppName / audioBuildSessionGroupID

## 目标

`resolveSessionAppName`（0x140757a00）、`audioBuildSessionGroupID`（0x1407588c0）两个 `[P]`
升 [S-sig]：签名逐寄存器实证，并订正旧树的错误近似（漏 bool/processID/sessionID、返回类型错误）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test -count=1 ./backend` | ok (0.4s) |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1348 / P=172 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1348 = 2487 / 4754 = 52.3%**（相对批次 173 **增 +2**；P 174→**172**，
S-sig 1346→1348，两函数由 [P] 升 [S-sig]）。

重建产物 SHA256：`1F1812B64248A6C2F2ED179B679E96B36C522DF5C5A16481D3C17C91A651513B`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数，`backend/audio_windows_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| resolveSessionAppName | 0x140757a00 | [P] (sessionID string) bool | [S-sig] (isSystemSounds bool, processPath, appName string, processID uint32) string |
| audioBuildSessionGroupID | 0x1407588c0 | [P] (appName, processPath string, processID uint32) string | [S-sig] (isSystemSounds bool, processPath string, processID uint32, sessionID, appName string) string |

## G4 关键实证结论

1. **resolveSessionAppName 四参 + string 返回**：morestack 保存 bl(bool)/rcx+rdi(string)/rsi+r8(string)/
   r9d(uint32) + rax=receiver；调用点 0x140756340 传 ebx=flag、rcx/rdi=processPath、rsi/r8=appName、
   r9d=processID。体：flag!=0→"System Sounds"(13B@0x140C6BF71)；resolveProcessDisplayName(processPath)
   非空→返回；TrimSpace(appName) 非空→返回；processID!=0→fmt.Sprintf("PID %d", processID)
   ("PID %d" 6B@0x140C595C0)；否则 "Unknown App"(11B@0x140C6A452)。旧树 (sessionID) bool 全错，已订正。
2. **audioBuildSessionGroupID 五参 + string 返回**：morestack 保存 al(bool)/rbx+rcx(string)/edi(uint32)/
   rsi+r8(string)/r9+r10(string)；调用点 0x14075646b 传 eax=flag、rbx/rcx=processPath、edi=processID、
   rsi/r8=sessionID、r9/r10=appName。体：flag!=0→"system-sounds"；audioNormalizePathKey(processPath)
   非空→"path:"+键；processID!=0→"pid:"+strconv.FormatUint(processID,10)；ToLower(sessionID) 非空→
   "session:"+值；ToLower(appName) 非空→"name:"+值；否则 ""。前缀常量逐字节解码：path:(5B@0x140C35BB9)、
   pid:(4B@0x140C3477A)、session:(8B@0x140C3C0BC)、name:(5B@0x140C35BC8)、system-sounds(13B@0x140C4DF7E)。
   旧树 (appName,processPath,processID) 漏 bool+sessionID，已订正。
3. **分组 ID 语义链**：组 ID 由「system-sounds / path:/ pid:/ session:/ name:」五级降级拼装，与
   buildSessionSnapshot 的会话字段（进程路径/进程号/会话实例标识/应用名）逐一对应，键语义闭环。

## 残留 / 未落地（不触及）

- (*audioSessionAccumulator).add（0x140757e20，[P]）：duffzero+0x142 提示按值装配 audioSessionSnapshot 累加
  结果，接收者字段 [rax+0x88] 判空分支，待专项实证。
- 两个 COM 包装方法（audioIAudioPolicyConfigFactory.SetPersistedDefaultAudioEndpoint / audioIPropertyStore.GetValue）
  待实证。
