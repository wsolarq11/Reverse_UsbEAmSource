# 批次 172 验收（audio_windows_runtime.go：两个会话设备路由函数 [P]→[S-sig] 签名订正）

日期：2026-09-25
子批次：音频会话设备路由链（resolveSessionDeviceRouting / getPersistedSessionDevice）

## 目标

`resolveSessionDeviceRouting`（0x140756860）、`getPersistedSessionDevice`（0x140756f20）
两个 `[P]` 升 [S-sig]：签名逐寄存器实证，并订正旧树的错误近似。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (1.7s) |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1344 / P=176 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1344 = 2483 / 4754 = 52.2%**（相对批次 171 **增 +2**；P 178→**176**，
S-sig 1342→1344，两函数由 [P] 升 [S-sig]）。

重建产物 SHA256：`21E7E4D55C027E8F1817E2F5BB0DE438DA764765EE9EA1B49A63594929EFDA29`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数，`backend/audio_windows_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| resolveSessionDeviceRouting | 0x140756860 | [P] (sessionID string)(string,string) | [S-sig] (processID uint32, flag bool)(render, capture string, ok bool) |
| getPersistedSessionDevice | 0x140756f20 | [P] (sessionID string) string | [S-sig] (deviceType string)(string, bool) |

## G4 关键实证结论

1. **resolveSessionDeviceRouting 双参 + 三返回**：调用点 0x140756417 传 `rax=receiver`、
   `ebx=dword [rsp+0x48]`(uint32=processID)、`cl=byte [rsp+0x41]`(bool=flag)。序言 `test cl; jne` 与
   `test ebx; jne` 双守卫。体：`getPersistedSessionDevice("render")`(6B@0x140C63FAE) 与
   `getPersistedSessionDevice("capture")`(7B@0x140C658EB)，返回 (rax,rbx)=render、(rcx,rdi)=capture、
   `esi=cl|ecx`=ok（两设备任一命中）。旧树误为 (sessionID string)(string,string)，已订正。
2. **getPersistedSessionDevice 单 string 参 + (string,bool) 返回**：调用点 0x1407568a1 传 `rax=receiver`、
   `rcx=&设备类型常量`、`edi=len`；返回 `(rax,rbx)`=设备串、`cl`=命中 bool。旧树误把 deviceType 当
   sessionID、漏 bool 返回，已订正。
3. **设备类型常量**：`"render"`(6B)、`"capture"`(7B) 经 Python 逐字节解码，非 sessionID——该函数读取的是
   按渲染/捕获方向持久化的音频设备 ID。

## 残留 / 未落地（不触及）

- resolveSessionAppName（0x140757a00，[P]）：asm 序言含 rcx/rdi/rsi/r8/r9d/bl 多寄存器（疑似多 string+
  uint32+bool），非旧树 (sessionID string) bool，待专项逐寄存器实证。
- buildSessionSnapshot（0x140756220，[P]）：duffzero+0x142 提示按值返回 audioSessionSnapshot 大结构，
  旧树 `(*audioSessionSnapshot, error)` 指针近似待复核。
