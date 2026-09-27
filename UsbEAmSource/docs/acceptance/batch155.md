# 批次 155 验收（gpu_windows.go 纯逻辑助手链：8 函数 [S]）

日期：2026-09-25
子批次：gpu_windows.go 纯逻辑助手链（路径规范化/条目构造/替换判定/目标校验/别名清理/文件存在判定）

## 目标

落地 gpu_windows.go 蓝图（source_funcs.txt 1798-1824）中**无 Win32 I/O 依赖的纯逻辑链**，
共 8 个具名函数全量 `[S]`（其余 registry/DXGI/进程枚举入口留待后续专项批次）：
`shouldCleanupMissingGPUPreferencePath`、`buildGPUPreferenceEntry`、
`gpuPreferencePathMayAccessFile`、`shouldReplaceGPUPreferenceEntry`、
`validateGPUPreferenceTargetPath`（含编译器生成 func1 闭包）、
`validateGPUPreferenceTargetPathWithResolver`、`deleteGPUPreferenceValueAliases`、
`gpuPreferenceFileExists`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test ./backend` | ok (0.352s) |
| 助手链 test | `go1.25.12 test -run 'TestShouldCleanupMissingGPUPreferencePath\|TestGPUPPreferencePathMayAccessFile\|TestShouldReplaceGPUPreferenceEntry\|TestBuildGPUPreferenceEntryEarlyExit\|TestValidateGPUPreferenceTargetPathWithResolver\|TestGPUPPreferenceFileExists' -v` | 全 PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2648 / S=1063 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`。

**真函数 = 1063 + 35 + 1353 = 2451 / 4754 = 51.6%**（相对批次 154 的 2443 增 +8 [S]）。

重建产物 SHA256：`616A2BEBA63A60A66DFCB3954871F64D0A11DF5867A3DCD7750CC4B592BC88CA`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（8 具名函数 [S]，`backend/gpu_windows.go`）

| 函数 | 蓝图行 | VA | asm 行 |
|---|---|---|---|
| shouldCleanupMissingGPUPreferencePath | 294-304 | 0x140860b40 | 47 |
| buildGPUPreferenceEntry | 304-330 | 0x140860c00 | 224 |
| gpuPreferencePathMayAccessFile | 330-344 | 0x140860f40 | 111 |
| shouldReplaceGPUPreferenceEntry | 344-355 | 0x1408610c0 | 45 |
| validateGPUPreferenceTargetPath | 355-371 | 0x140861180 | 94 |
| validateGPUPreferenceTargetPathWithResolver | 371-429 | 0x1408611e0 | 384 |
| deleteGPUPreferenceValueAliases | 429-448 | 0x1408617e0 | 384 |
| gpuPreferenceFileExists | 448-461 | 0x140861960 | 118 |
| validateGPUPreferenceTargetPath.func1（闭包） | 357-365 | 0x1409f7200 | 123 |

## G4 关键实证结论

1. **buildGPUPreferenceEntry 结果槽位映射**：duffcopy 零值模板后逐槽回填，Path@0x00/Name@0x10/
   IconData@0x30/PreferenceMode@0x40/PreferenceValue@0x50/Settings@0x58/RawValue@0x70 全部落位，
   **Icon@0x20 槽位刻意不写**（保留零值模板的 ""），故重建体 Icon 恒为空串。
2. **特殊路径前缀魔数（gpuPreferencePathMayAccessFile）**：四个前缀小端 dword 还原为
   `\\`(0x5c5c)、`\??\`(0x5c3f3f5c)、`\\?\`(0x5c3f5c5c)、`\\.\`(0x5c2e5c5c)，
   即 UNC / NT 命名空间 / 长路径前缀 / 设备路径四类拒绝。末段扩展名扫描等价于 `filepath.Ext`。
3. **shouldReplaceGPUPreferenceEntry 读 Settings.cap**：末位比较读结构体偏移 0x68（slice 第三字 =
   cap，非 0x60 的 len），故按 `cap(existing.Settings) < cap(candidate.Settings)` 直译，与
   buildGPUPreferenceEntry 中 settings 槽位（ptr@0x58/len@0x60/cap@0x68）交叉印证。
4. **validateGPUPreferenceTargetPathWithResolver 依赖注入**：func1 闭包 = resolveShortcutInfoWithIconResolver，
   第二参 filepath.EvalSymlinks，第三参 os.Stat（三个 8B 函数指针重叠存储，非 16B funcval）。
   stat 回调返回 os.FileInfo，IsDir 判定目录拒绝。链式 .lnk（解析后仍为 .lnk）拒绝。
5. **deleteGPUPreferenceValueAliases 忽略 ErrNotExist**：`errors.Is(err, registry.ErrNotExist)`，
   ReadValueNames 入参 n=0（asm `xor ebx,ebx`）。
6. **gpuPreferenceFileExists 双通道**：UTF16PtrFromString + GetFileAttributes 成功则按
   `FILE_ATTRIBUTE_DIRECTORY`（bt eax,4 + setae）判定；失败回退 os.Stat → `!info.IsDir()`。

## 残留 [P] / 未落地（不触及）

gpu_windows.go 剩余 registry/DXGI/进程枚举入口：getGPUPreferenceState 0x14085dda0、
addGPUPreferenceEntry 0x14085e920、saveGPUPreferenceEntry 0x14085f340、
removeGPUPreferenceEntry 0x14085fde0、cleanupMissingGPUPreferenceEntries 0x140860760、
resolveGPUPreferenceAdapterInfo 0x14085d240、resolveDXGIAdapterNameByPreference 0x14085d340、
createDXGIFactory1 0x14085da00、queryDXGIFactory6 0x14085db80、releaseDXGIUnknown 0x14085dd20；
以及 gpu_pick_windows.go 全 11 函数。这些入口最终调用本批纯逻辑助手链。

## 已知偏差（诚实记录）

func1 闭包在真实二进制读取 ShortcutInfo 结构体偏移 0x78 的 bool 字段（"有效目标"标志）与
偏移 0x08 的 target 字符串；当前 `shortcutinfo.go` 的 ShortcutInfo 存根字段序（TargetPath@0x10、
无 bool 字段）与此不符。本批 func1 以「错误 + TrimSpace(TargetPath) 空」合并判定（两分支同
错误串"快捷方式没有有效目标"），因 resolveShortcutInfoWithIconResolver 当前为 COM 存根恒返错，
行为一致；ShortcutInfo 真实布局将在 resolveShortcutInfoWithIconResolver 专项批次一并对齐。
