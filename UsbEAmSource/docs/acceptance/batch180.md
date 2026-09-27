# 批次 180 — 文件搜索归一化链升档 + 提醒音频播放链落地（4 函数新增 + 3 升档）

## 基线 / 收口

| 指标 | 基线（批次 179 收口） | 收口（批次 180） |
|---|---|---|
| FUNCS | 2703 | **2707** |
| S | 1144 | **1151** |
| S-inline | 35 | 35 |
| S-sig | 1409 | **1406** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2588（54.4%） | **2592（54.5%）** |
| 未落地文件（差集） | 63 | **62** |

SHA256 `EFABF9C3CDD7511242BA319809E0A655D6B180F9F798E1A9F3F7287C8FD8B5BE`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（4 函数新增 + 3 升档，1 文件新建）

### 提醒音频播放链 · desktopwidgets_audio_windows.go（3 [S]，新建）

| 函数 | VA | 关键证据 |
|---|---|---|
| `playDesktopReminderSound` | 0x1407a7b20 | winmm.dll PlaySoundW（LazyProc 名 "PlaySoundW"@0x140c43f7d,10B）；UTF16PtrFromString→PlaySound(ptr,0,flags)；r1!=0→nil；err==nil 或 errors.Is(err,ERROR_SUCCESS) → "Windows 无法播放提醒音频"@0x140c72466,32B |
| `playDesktopReminderSystemSound` | 0x1407a79e0 | 逐条 playDesktopReminderSound(s,0x210002)，s∈{"Notification.Default","SystemNotification","SystemAsterisk"}（全局 slice@0x141bc67a0）；全失败回退 user32.dll MessageBeep(1)（LazyProc 名 "MessageBeep"@0x140c47376）；错误 "Windows 系统提醒音频不可用"@0x140c770f6,35B |
| `playDesktopReminderAudioSequence` | 0x1407a78e0 | 循环 count 次；name=="custom"（len==6+"cust"@0x74737563+"om"@0x6d6f）→ playDesktopReminderSound(path,0x220002)；否则系统音；err 即返回；间隔 time.Sleep(0xee6b280=250ms) |

### 音频定义转换 · desktopwidgets_audio.go（1 [S] + 1 类型）

| 函数/类型 | 说明 |
|---|---|
| `DesktopReminderAudioDefinition` | 字段偏移从 asm 0x1407a6780 确证：Name@0x88、Path@0x98、Volume@0xa8；前置 0x88 字节布局未触及用填充占位 |
| `desktopReminderAudioFromDefinition` | [S] 0x1407a6780：def==nil→normalizeDesktopReminderAudio("","",0)；否则 normalize(def.Name,def.Path,def.Volume) |

### 归一化链升档（3 处，filesearch_normalize.go）

| 函数 | VA | 订正 |
|---|---|---|
| `normalizeFileSearchIgnoredDirectoryRules` | 0x1407e0440 | identity stub → [S]：make 输出+map 去重；绝对路径规则去重键=normalizeFileSearchIgnoredDirectoryMatchPath，否则 ToLower；键空/重复跳过；append 原 norm |
| `defaultFileSearchIgnoredDirectoryRules` | 0x1407e03c0 | nil stub → [S]：normalizeFileSearchIgnoredDirectoryRules(["node_modules"@0x140c4b9a8,12B, TrimSpace(os.TempDir()), "system volume information"@0x140c66ad8,25B]) |
| `normalizeFileSearchTypeFilters` | 0x140815220 | identity stub → [S]：nil→default；空→空；上限 32；ID 空或 "all" 跳过；seen 去重；Rules 归一化失败跳过；Label TrimSpace 超 64 rune 截断；enabled 非内置且 label 空→false，enabled 且 rules 空→false |

## 关键订正与知悉

- **errors.Is 哨兵确证**：playDesktopReminderSound/SystemSound 两处 errors.Is 目标均为同一静态错误 `windows.ERROR_SUCCESS`（接口 data@0x1411cd520=0x0），非先前误判的 ERROR_INSUFFICIENT_BUFFER（0x7a 为相邻变量）。
- **PlaySound flags 语义**：system 用 `0x210002`（SND_ALIAS 0x10000|SND_NODEFAULT 0x2|0x200000），custom 用 `0x220002`（SND_FILENAME 0x20000|SND_NODEFAULT 0x2|0x200000）——系统音走别名、自定义音走文件路径。
- **MessageBeep/PlaySound DLL 归属**：playDesktopReminderSound=winmm.dll PlaySoundW，playDesktopReminderSystemSound 回退=user32.dll MessageBeep，均读 LazyProc.Name 字节级确证。
- **`DesktopReminderAudioDefinition` 前置 0x88 字节**：本轮仅确证 Name/Path/Volume 偏移，前置字段布局未触及（真相单一出口，不臆造字段名），用 `_ [0x88]byte` 占位。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.350s`。
- **G2 count_funcs**：`FUNCS=2707 / S=1151 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`。
- **G3 行为**：normalizeFileSearchIgnoredDirectoryRules 的绝对路径/相对路径双键去重、defaultFileSearchIgnoredDirectoryRules 的 "node_modules"/"system volume information" 常量、playDesktopReminderAudioSequence 的 "custom" 分支与 250ms 间隔、playDesktopReminderSound 的 "Windows 无法播放提醒音频" 均 .rodata 字节级确证。
- **G4 review**：本批无并行写重叠；升档与新增函数分域清晰（filesearch_normalize.go 升档、desktopwidgets_audio*.go 新增）。
