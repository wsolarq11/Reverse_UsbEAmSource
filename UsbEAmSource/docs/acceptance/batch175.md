# 批次 175 验收（audio_windows_runtime.go：会话累加器 add [P]→[S-sig] 签名订正）

日期：2026-09-25
子批次：(*audioSessionAccumulator).add

## 目标

`(*audioSessionAccumulator).add`（0x140757e20）由 `[P]` 升 [S-sig]：签名逐寄存器实证，
订正旧树指针近似（`*AudioSession` → `AudioSession` 按值）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test -count=1 ./backend` | ok (0.4s) |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1349 / P=171 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1349 = 2488 / 4754 = 52.4%**（相对批次 174 **增 +1**；P 172→**171**，
S-sig 1348→1349，累加器 add 由 [P] 升 [S-sig]）。

重建产物 SHA256：`F9C186E777528DFFFACB9DC407AA90F3D240420859FD8C65631189C27CBB08F4`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（1 函数，`backend/audio_windows_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| (*audioSessionAccumulator).add | 0x140757e20 | [P] (s *AudioSession) | [S-sig] (s AudioSession) |

## G4 关键实证结论

1. **按值入参、无返回**：调用点 0x140754878 前 `duffcopy+0x310` 把 12 字段结构从 `[rsp+0xa8]` 拷到调用栈
   （按值传参），`rax=receiver`；morestack 仅保存 rax（无寄存器入参）。返回无寄存器载荷。
2. **体语义**：`cmp [rax+0x88],0` 判「是否首次累加」；首次走 duffzero+0x142 清零 session 字段后 duffcopy
   把入参拷入 receiver.session（12 字段与 AudioSession 逐一对齐），再写 volumeSum(0x80)/sessionCount=1(0x88)/
   allMuted(0x90)/anyActive(0x91)；非首次走累加路径（0x140757ff2：累加 volumeSum、递增 sessionCount、逻辑与）。
3. **旧树指针订正**：`(s *AudioSession)` 实为按值传结构（栈传参 + duffcopy），非指针。签名已订正，无 Go 调用点
   受影响（build/vet/test 全 EXIT=0）。

## 残留 / 未落地（不触及）

- 两个 COM 包装方法（audioIAudioPolicyConfigFactory.SetPersistedDefaultAudioEndpoint / audioIPropertyStore.GetValue）
  待实证。至此 audio_windows_runtime.go 域内 [P] 存根已清（剩余 [P] 主要在 screenshot/window/oledblackout/
  filelocator 等域）。
