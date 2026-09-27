# 批次 177 — 小文件落地 + 体错/签名订正（8 函数新增 + 4 处订正）

## 基线 / 收口

| 指标 | 基线（批次 176 收口） | 收口（批次 177） |
|---|---|---|
| FUNCS | 2659 | **2667** |
| S | 1104 | **1111** |
| S-inline | 35 | 35 |
| S-sig | 1405 | **1406** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2544（53.5%） | **2552（53.7%）** |
| 未落地文件（差集） | 79（本轮实测，§1 旧值 100/§10 旧值 110 均过时） | **73** |

SHA256 `615C8D00CDCE7F0BB2CEB2E8C739DCF68A45B11CBFD44A7B9D109E0E52ADFB01`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（8 函数新增，6 文件新建）

| 函数 | VA | 档位 | 文件 | 关键证据 |
|---|---|---|---|---|
| `launcherSecondInstanceShouldReveal` | 0x1408a55c0 | [S] | launcherinstance.go | len(args)<=1→true；else `!launcherStartedForStartupTray(args[1:])`（rax+0x10/len/cap-1，尾调 xor eax,1） |
| `windowManagementVirtualScreenBounds` | 0x1409edbc0 | [S] | windowmanagement_windows.go | 4×GetSystemMetrics(0x4c/0x4d/0x4e/0x4f)，尾段 edx=x+w/edi=y+h → image.Rect，无回退分支 |
| `suppressLauncherWindowsAltMenuMessage` | 0x14086b0c0 | [S] | launcheraltmenu_windows.go | 叶子函数 cmp ebx,0x112 + and ecx,0xfff0 + cmp rcx,0xf100 双匹配→(0,true) |
| `(*launcherHotkeyRegistrationError).Error` | 0x14089d6e0 | [S] | launcherglobalhotkey.go | e==nil||e.Err==nil→""；else itab[0x18] 调 e.Err.Error() |
| `(*launcherHotkeyRegistrationError).Unwrap` | 0x14089d740 | [S] | launcherglobalhotkey.go | e==nil→nil；else e.Err（interface 两字透传） |
| `resolveLauncherWebViewUserDataPath` | 0x1409dd840 | [S] | webview2_options.go | TrimSpace→空 errors.New("WebView2 用户数据目录根路径不能为空")@0x140c86df3,48B→Abs→MkdirAll(abs,0o755) |
| `buildLauncherBrowserArgs` | 0x1409dd920 | [S-sig] | webview2_options.go | 返回 []string；参数 0x448B 栈结构（normalizePreferencesWithOptions 调用），类型未落地，体待还原 |
| `loadLauncherBrowserArgs` | 0x1409ddb40 | [S-sig] | webview2_options.go | (configPath string) []string；loadLauncherConfigIfExists→失败/!ok 返 nil |

## 本批订正（4 处，不改变 FUNCS 计数）

| 函数 | VA | 订正 |
|---|---|---|
| `normalizeAudioFlowName` | 0x140750ee0 | 旧体 `(string,error)` + playback/record/all 分支为臆造；订正为 `string` 单返回，input/capture→"capture"、output/render→"render"、其余→""（.rodata 常量 0x140c392eb="capture"/0x140c375ae="render"） |
| `getLauncherBackgroundMetrics` | 0x140874520 | [S-sig] 空骨架 → [S] 完整体：windowManagementVirtualScreenBounds 宽/高<=0 返零值，否则 Supported=true+VirtualX/Y/Width/Height（栈 40B 返回区展开） |
| `detectLinkBrowsers` | 0x1408d0b60 | 签名 `()` → `() []StartMenuApp`（尾声 rax/rbx/rcx=rsi/rdi/rbx 三字 slice） |
| `ScanLinkBrowsers` | 0x1408d0b40 | 签名 `()` → `() []StartMenuApp`，体 `return detectLinkBrowsers()`，从 bootstrap_desktopwidget.go 迁入 linkbrowser.go（正确文件名） |

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.421s`。
- **G2 count_funcs**：`FUNCS=2667 / S=1111 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`，P 未增，UNMARKED=0。
- **G3 行为**：`normalizeAudioFlowName` 的 switch 分支经 .rodata 常量字节级解码确证；`getLauncherBackgroundMetrics` 的栈返回区 40B 展开确证；`windowManagementVirtualScreenBounds` 的 GetSystemMetrics 四索引确证。
- **G4 review**：本轮最重要产出是**发现已落地函数存在体错/签名错**（normalizeAudioFlowName 臆造分支、detectLinkBrowsers/ScanLinkBrowsers 丢返回类型），提示历史 `[S]` 标记可能混杂臆造体，后续需抽检已落地 `[S]` 函数。

## 差集口径校准（本轮新增）

- 实测未落地文件差集 = **79**（非 §1 的 100、非 §10 的 110），本轮减 6 → **73**。
- 差集文件 79 个对应 1710 蓝图函数项（1273 顶层 / 1151 唯一名）。
- 唯一名中 **697 个真正未落地**（backend 无同名定义），454 个已在别处落地（函数按域重组导致文件名对不上，如 qrcode.go 散落 screenshot_qrcode_*.go）。
- 结论：FUNCS 缺口（→4754）的核心是 697 个未落地函数（+ 闭包/方法），非文件名对齐。
