# 批次 156 验收（gpu_windows.go registry/DXGI 入口：7 函数 [S]）

日期：2026-09-25
子批次：gpu_windows.go 注册表/DXGI 入口链（GPU 偏好读取/新增/保存/删除/清理 + DXGI 适配器名解析）

## 目标

落地 gpu_windows.go 蓝图（source_funcs.txt 1798-1824）中**剩余 registry/DXGI 入口**，共 7 个具名
函数全量 `[S]`（gpu_pick_windows.go 全 11 函数留待后续批次）：

`resolveGPUPreferenceAdapterInfo`、`resolveDXGIAdapterNameByPreference`、`getGPUPreferenceState`、
`addGPUPreferenceEntry`、`saveGPUPreferenceEntry`、`removeGPUPreferenceEntry`、
`cleanupMissingGPUPreferenceEntries`。

同时：删除 `bootstrapservice_callees.go` 中 5 个 GPU 域 `[S-sig]` 存根（getGPUPreferenceState /
addGPUPreferenceEntry / saveGPUPreferenceEntry / removeGPUPreferenceEntry /
cleanupMissingGPUPreferenceEntries），并对齐 5 个 BootstrapService 方法包装器签名。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test ./backend` | ok (0.394s) |
| 黄金用例 | `go1.25.12 test -run 'TestAddGPUPreferenceEntryInvalidPath\|TestSaveGPUPreferenceEntryInvalidPath\|TestRemoveGPUPreferenceEntryEmptyPath' -v` | 全 PASS |

黄金用例说明：三条 registry 入口（add/save/remove）的「空/非法路径前置校验」分支在任何 Win32 I/O
之前早退，可安全实测且不触注册表；三者均验证错误串 `显卡调度目标路径不能为空` 逐字一致。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2650 / S=1070 / S-inline=35 / S-sig=1348 / P=197 / UNMARKED=0`。

**真函数 = 1070 + 35 + 1348 = 2453 / 4754 = 51.6%**（相对批次 155 的 2451 增 +7 [S]、-5 [S-sig] 存根）。

重建产物 SHA256：`87922B609F2EACDE51401043D47A09EEBCF244EC8F86D77D9E68E6F675785CFE`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（7 具名函数 [S]，`backend/gpu_windows.go`）

| 函数 | 蓝图行 | VA | 说明 |
|---|---|---|---|
| resolveGPUPreferenceAdapterInfo | 108-125 | 0x14085d240 | DXGI 工厂枚举节能/高性能适配器名（4 字返回 AX/BX/CX/DI） |
| resolveDXGIAdapterNameByPreference | 125-227 | 0x14085d340 | IDXGIFactory6 EnumAdapterByGpuPreference + GetDesc 名解析 |
| getGPUPreferenceState | 230-280 | 0x14085dda0 | OpenKey→ReadValueNames→去重 map→sort.Slice（Name 升序，PathKey 决胜） |
| addGPUPreferenceEntry | 284-317 | 0x14085e920 | validate→CreateKey→删别名→序列化 {GpuPreference,"0"}→SetStringValue→刷新 |
| saveGPUPreferenceEntry | 320-355 | 0x14085f340 | validate→ensure→CreateKey→删别名→序列化→SetStringValue→刷新 |
| removeGPUPreferenceEntry | 358-394 | 0x14085fde0 | normalize 空报错→OpenKey→ErrNotExist 视为已删→删别名→刷新 |
| cleanupMissingGPUPreferenceEntries | 397-432 | 0x140860760 | OpenKey→ReadValueNames→逐名 shouldCleanup→DeleteValue→Close→刷新 |

## G4 关键实证结论

1. **5 个 BootstrapService 方法包装器返回 GPUPreferenceState（7 字宽），丢弃 error**：
   `GetGPUPreferenceEntry`（0x14077e1a0）→ `GetGPUPreferenceState() GPUPreferenceState`；
   `AddGPUPreferenceEntry`（0x14078fc40）、`RemoveGPUPreferenceEntry`（0x14078fe60）、
   `CleanupMissingGPUPreferenceEntries`（0x14078ff60）均取 7 字宽结构、弃 r10/r11 error 透传。
2. **SaveGPUPreferenceEntry 真实签名**：`(path string, settings []GPUPreferenceSetting)`，非旧存根的
   `(entry interface{})`。方法包装器（0x14078fd40）寄存器重排后传 5 字（string 2 字 + slice 3 字）。
3. **addGPUPreferenceEntry 默认写入 `{GpuPreference, "0"}`**（字面量 VA 0x140c4e04e/13B +
   0x1411ca3c8/1B），经 serializeGPUPreferenceRegistrySettings 生成 `GpuPreference=0;`。
4. **CreateKey/OpenKey 访问掩码 = 3**（`registry.QUERY_VALUE|registry.SET_VALUE`，asm `edi=3`）。
5. **getGPUPreferenceState 排序**：主键 `ToLower(TrimSpace(Name))`，决胜 `gpuPreferencePathKey(Path)`
   双 cmpstring <0 直译；ErrNotExist 分支返回 `[]GPUPreferenceEntry{}`（zerobase 非 nil 数据指针）。
6. **removeGPUPreferenceEntry 空路径早退**：`normalizeGPUPreferencePath` 空 → `errors.New("显卡调度目标路径不能为空")`
   （36B），先于任何注册表 I/O；OpenKey 命中 ErrNotExist → 返 (空 Entries + AdapterInfo, nil)。
7. **cleanupMissingGPUPreferenceEntries 显式 Close（非 defer）**：ReadValueNames 失败 / DeleteValue 非
   ErrNotExist 失败路径 `key.Close()` 忽略错误后返；正常收尾 `key.Close()` 校验错误后 `return getGPUPreferenceState()`。
8. **resolveDXGIAdapterNameByPreference 5 参 SyscallN 异常**：args `[factory6, 0, preference, &iidIDXGIAdapter, &adapter]`
   含字面量 0 于 args[1]（间隙 [rsp+0x208] 原始字节 `4889b42408020000` 实证），忠实复刻并记录。

## 残留 [P] / 未落地（不触及）

gpu_pick_windows.go 全 11 函数（pickWindowProcessForService / resolveWindowProcessPath /
resolveProcessPathByPID 等，蓝图 1786-1797）；ShortcutInfo 布局对齐（bool@0x78 + target@0x08，
待 resolveShortcutInfoWithIconResolver 专项批次一并修）。

## 已知偏差（诚实记录）

无。本批 7 函数均为 asm 直译，registry/DXGI 调用与错误分支逐条对照。
