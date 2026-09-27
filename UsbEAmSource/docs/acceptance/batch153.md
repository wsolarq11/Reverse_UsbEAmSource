# 批次 153 验收（gpu 域纯逻辑链：11 函数 [S] 反汇编实证还原）

日期：2026-09-25
子批次：gpu.go / gpu_pick_debug.go 纯逻辑层（无外部 IO / 全局态 / Win32 依赖）

## 目标

落地 gpu 域自包含纯函数层，把 `normalizeGPUPreferenceMode` 跳表、键/值/路径规范化、
注册表串解析/序列化、去重与前移语义、调试布尔解析共 11 个函数从「未落地蓝图符号」升为 `[S]`
（反汇编逐寄存器实证）。类型层（`types_gpu.go`）已在前批次落地，本批补齐逻辑层。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 ./backend` | ok (0.402s) |
| GPU 行为 test | `go1.25.12 test -run 'GPU|NormalizeGPU|ParseGPU' -v` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2637 / S=1052 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`。

**真函数 = 1052 + 35 + 1353 = 2440 / 4754 = 51.3%**（相对批次 152 的 2429 增 +11）。

重建产物 SHA256：`5189E3DFEBD9D9243A5EE017E1A7F5BBE6F89A36C66D76AC19842D20D82793E1`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（11 函数，均 [S]，`backend/gpu.go`）

| 函数 | 蓝图文件:行 | VA | asm 行 |
|---|---|---|---|
| normalizeGPUPreferenceMode | gpu.go 84-121 | 0x14085a0e0 | 141 |
| gpuPreferenceCanonicalSettingKey | gpu.go 121-136 | 0x14085a360 | — |
| gpuPreferenceNormalizeSettingValue | gpu.go 136-166 | 0x14085a4e0 | 153 |
| normalizeGPUPreferenceRegistrySettings | gpu.go 166-210 | 0x14085a700 | 273 |
| ensureGPUPreferenceRegistrySettings | gpu.go 210-221 | 0x14085ac40 | 117 |
| gpuPreferenceModeFromSettings | gpu.go 221-238 | 0x14085ade0 | — |
| parseGPUPreferenceRegistryValue | gpu.go 238-265 | 0x14085af40 | 180 |
| serializeGPUPreferenceRegistrySettings | gpu.go 265-286 | 0x14085b240 | 236 |
| normalizeGPUPreferencePath | gpu.go 286-300 | 0x14085b600 | — |
| gpuPreferencePathKey | gpu.go 300-306 | 0x14085b6a0 | — |
| parseGPUPickDebugBool | gpu_pick_debug.go 134-141 | 0x14085bd40 | 72 |

## G4 关键实证结论

1. **normalizeGPUPreferenceMode 是长度跳表**：`jmp qword ptr [rcx+rbx*8]`（表 @0x1411e0940，19 项）。
   逐项读表映射到 3 个标准串：`systemDefault`（含空串/`0`/`system`/`default`/`unspecified`/
   `systemdefault`/`letwindowsdecide`/`let-windows-decide`）、`powerSaving`（`1`/`powersaving`/
   `minimumpower`/`power-saving`/`minimum-power`）、`highPerformance`（`2`/`highperformance`/
   `high-performance`），余皆 `unknown`。长度跳表首项（len=0）指向 `systemDefault`，证得空串也归一化
   为 `systemDefault`。`resolve_lea_strings.py` 误报的 `systemDefau`/`powerSavingDXGI` 长度已用
   `read_gostring.py` 修正为 `systemDefault`(13)/`powerSaving`(11)/`highPerformance`(15)。
2. **ensureGPUPreferenceRegistrySettings 是前插而非尾插**：asm 中 `newobject` 造字面量
   `{Key:"GpuPreference", Value:"0"}`（[rax]=0x140c4e04e，[rax+8]=0xd，[rax+0x10]=0x1411ca3c8，
   [rax+0x18]=1），随后 `growslice` 以该字面量为 1 元素旧 slice 扩展，`typedslicecopy` 目标
   `lea rbx,[rax+r8]` 中 `r8=0x20`（`and r8d,0x20` 对 newCap>1 取 32），证得旧元素整体右移一位、
   新项落在 `[0]`。对应 `append([]GPUPreferenceSetting{{...}}, s...)`，而非 `append(s, ...)`。
   测试用例据此锁定前插 + 原始大小写（`GpuPreference`，未过规范化）。
3. **parseGPUPreferenceRegistryValue 先 TrimSpace 判空再 Split 原串**：asm 对 `raw` 先 `TrimSpace`
   判空（空→nil），再 `genSplit(raw,";",0,-1)` 切原串（非修剪串）；每段 `TrimSpace` 跳空、
   `Cut(part,"=")` 不命中即跳过、命中 `append({before,after})`；`len(parts)>1` 才预分配（`makeslice`），
   否则首 append 走 32 字节栈缓冲优化；收口 `normalizeGPUPreferenceRegistrySettings`。
4. **normalizeGPUPreferenceRegistrySettings 的 seen 去重键是 ToLower(key)**：`map[string]int` 值存
   result 索引，命中覆写 `result[idx]`；`IndexAny(key,"=;")>=0` 跳过滤（0x140c336a1=`=;`）；收尾
   `seen["gpupreference"]` 存在且 idx>0 时 `copy(result[1:idx+1], result[:idx])` 前移。
5. **serializeGPUPreferenceRegistrySettings 分隔符**：`Key` + `'='`(0x3d) + `Value` + `';'`(0x3b)，
   空 Key/Value 跳过，尾 `string(buf)` 转换带 panicunsafestringlen/nilptr 安全校验。
6. **parseGPUPickDebugBool 返回 (bool,bool)**：三出口寄存器对 (1,1)/(0,1)/(0,0)，第二值区分
   「已识别为假」与「未识别」。`1/on/yes/true`→(true,true)，`0/no/off/false`→(false,true)，余→(false,false)。

## 残留 [P]（本批不触及，仍依赖 Win32/注册表 IO/全局态）

`getGPUPreferenceState`/`addGPUPreferenceEntry`/`saveGPUPreferenceEntry`/`removeGPUPreferenceEntry`/
`cleanupMissingGPUPreferenceEntries`/`pickWindowProcessForService` 等 `gpu_windows.go`/
`gpu_pick_windows.go` 入口（`bootstrapservice_callees.go`），及 `gpuPickDebugLog`/
`ensureGPUPickDebugLogger`/`resolveGPUPickDebugConfiguration`（`gpu_pick_debug.go` 全局态 + 文件 IO）。
下一批 gpu_windows.go 将调用本批纯函数。
