# 批次 271 · buildWorkspaceLayoutWithConfig + importLauncherBackgroundImage 升档 + 2 helper（+2 [S-sig]→[S]，+2 新 [S]，-2 [S-sig]，+2 FUNCS）

## 目标

1. **升档 `buildWorkspaceLayoutWithConfig`（0x1407a1ac0, 1696B）**：`[S-sig]` 骨架
   `(cfg LauncherConfig, target string)` → `[S]` 8×string 签名全还原。
2. **升档 `importLauncherBackgroundImage`（0x140798e40, 1632B）**：`[S-sig]` 骨架
   `(sourcePath, ws)` → `[S]` 单参 `(sourcePath string)(string,error)` 全还原。
3. **补齐两个未落地 helper**：`normalizeLauncherBackgroundImageExtension`（0x1407997a0）、
   `copyLauncherBackgroundImageFile`（0x1407994a0）。
4. **订正被调方签名**：`resolveLauncherConfigFilePath`（0x1407a1920）
   `(WorkspaceLayout)string`→`(string)string`。

## 基线 / 收口

| 指标 | 基线（batch 270 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2871 | 2873 |
| MARKED | 2871 | 2873 |
| S | 1349 | 1353 |
| S-inline | 36 | 36 |
| S-sig | 1446 | 1444 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1385 | 1389 |
| USABLE | 1385 | 1389 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1389  FUNCS=2873  MARKED=2873  P=40  S-eq=0  S-inline=36  S-sig=1444  S=1353  USABLE=1389
```

分项自洽：`S + S-inline + S-sig + P = 1353 + 36 + 1444 + 40 = 2873 = FUNCS`。
S 1349→1353（+4）、S-sig 1446→1444（-2）、FUNCS 2871→2873（+2）、P=40 持平、UNMARKED=0 保持。

账目：`buildWorkspaceLayoutWithConfig`/`importLauncherBackgroundImage` [S-sig]→[S]（-2 [S-sig] +2 [S]）；
新增 `normalizeLauncherBackgroundImageExtension`/`copyLauncherBackgroundImageFile`（+2 [S] +2 FUNCS）。

## G3 行为（asm 逐地址实证）

### 3.1 buildWorkspaceLayoutWithConfig [S 0x1407a1ac0, 1696B]（buildWorkspaceLayoutWithConfig.asm.txt）

签名 8×string：`root/configFile/pluginDir`（寄存器 RAX-R8）+ `dataRoot/iconDir/indexDir/
screenshotDir/webView2Dir`（栈）。返回 WorkspaceLayout（栈 160B）。

- rootEff（line 20-36 + 55-66）：`TrimSpace(root)` 空则 `"."`（len1 @0x1411cac28）；
  `normalizeStorageConfig`（line 42-48）后 `DataRoot!=0` 用 `cmovne` 覆盖 rootEff。
- ConfigFile（line 37-39 + 25）：`TrimSpace(configFile)` 空则
  `resolveLauncherConfigFilePath(resolveProcessWorkingDirectory())`。
- IconDir/IndexDir/ScreenshotDir/WebView2Dir（line 77-158）：各自 `TrimSpace` 非空优先，
  空则 `filepath.Join(rootEff, 子目录)`。子目录名 rodata 解码（单数）：
  `icon`@0x140c347be(4) / `index`@0x140c35c09(5) / `screenshot`@0x140c43f87(10) /
  `webview2`@0x140c3c12c(8)。
- BackgroundDir（line 165-172）：`filepath.Join(rootEff, "background"@0x140c4408b)`。
- LanguageDir/AppLanguageDir（line 182-187）：`filepath.Join(filepath.Dir(pluginDir),
  "language"@0x140c3c134)`；PluginDir 原样（未 TrimSpace）。
- 返回区组装（line 188-223）：Root=rootEff、ConfigFile、LanguageDir、AppLanguageDir、
  PluginDir、IconDir、IndexDir、ScreenshotDir、WebView2Dir、BackgroundDir 十字段。

三处调用点寄存器映射交叉验证：commit...WithWidgetsImpl.asm 0x140778f40（参1-3 =
ws.Root/ws.ConfigFile/ws.PluginDir，栈参数 = normalize 输出 5×string）；MigrateConfig.asm
0x14077c8d6（参1 = staged.targetPath，参2-3 = ws.ConfigFile/ws.PluginDir，栈参数 duffzero+0x150
清零）；PreviewInitializationImportConfig.asm 0x14077c1e2（先 filepathlite.Dir 后装配参2-3）。

### 3.2 importLauncherBackgroundImage [S 0x140798e40, 1632B]（importLauncherBackgroundImage.asm.txt）

签名 `(sourcePath string)(string,error)`。ws 无参——line 138 `mov rax,[rsp+0x14c0]`（receiver）
→ `workspaceSnapshot`（0x1407a10c0）。ChooseLauncherBackgroundImage.asm 0x14077b3c9 只传
`(bs, source)` 两值交叉验证。

- 空路径（line 20-23 → 0x140798fc3）：`TrimSpace(source)` 空 →
  `newExtractError("背景图片路径不能为空"@0x140c6eded, 30B)`。
- Abs（line 25-29 → 0x140798f83）：`filepath.Abs` err → 透传 err。
- Stat（line 31-33 → 0x140798f4d）：`os.Stat` err → 透传 err。
- IsDir（line 34-38 → 0x140798ef8）：`info.IsDir()` 真 →
  `newExtractError("背景图片不能是目录"@0x140c69c18, 27B)`。
- 扩展名（line 39-56 + 110-132）：从 len-1 往前扫 `.`/`\`/`/`（与 filepath.Ext 等价）→
  `normalizeLauncherBackgroundImageExtension` 空 →
  `newExtractError("不支持的背景图片格式"@0x140c6ee0b, 30B)`。
- MkdirAll（line 133-137）：`os.MkdirAll(ws.BackgroundDir, 0x1ed=0o755)` err → 透传。
- dest（line 113-132）：`concatstring2("custom-background"@0x140c572a8 17B, ext)` →
  `filepath.Join(BackgroundDir, 结果)`。
- copy（line 139-141 → 0x14079918a）：`copyLauncherBackgroundImageFile(abs, dest)` err → 透传。
- pref 装配（line 194-258）：静态模板（0x1411dea30 112B：Enabled=1，余零）拷贝 →
  ImagePath=dest；newobject bool=true（ReadabilityOverlayEnabled@0x38）、float 0.18
  （0x3fc70a3d70a3d70a，ReadabilityOverlayOpacity@0x40）、float 1.0（0x3ff0000000000000，
  ImageOpacity@0x48）→ `normalizeBackgroundPreference` → 构造 LauncherConfig（0x126 qword 清零，
  Preferences.Background@0x120 = pref）→ `attachLauncherBackgroundURL`。
- 返回（line 350-362）：`filepath.Base(abs)` + nil error。

### 3.3 normalizeLauncherBackgroundImageExtension [S 0x1407997a0, 256B]

`TrimSpace`→`ToLower` 后按 4/5 字节立即数分派：`.bmp`(0x706d622e)/`.gif`(0x6669672e)/
`.jpg`(0x67706a2e)/`.png`(0x676e702e) 原样；`.jpeg`(0x65706a2e+'g')→`.jpg`；
`.webp`(0x6265772e+'p')→`.webp`；其余 `("", "")`（line 59-61 xor 双清）。

### 3.4 copyLauncherBackgroundImageFile [S 0x1407994a0, 672B]

`filepath.Clean(src)==Clean(dst)`（line 14-27 memequal）→ nil（0x1407996ce）；
`os.OpenFile(src, 0, 0)`（O_RDONLY）err→透传，defer Close（line 37-42 deferprocStack）；
`os.OpenFile(dst, 0x241, 0x1a4)`（O_WRONLY|O_CREATE|O_TRUNC, 0o644）err→透传（defer 关 src）；
`io.Copy`（io.copyBuffer，静态 32KB buf）err→关 dst 返 copy err（line 59-101）；
成功→`dst.Close()` 返错（line 102-113）。

## G4 独立复核

`bootstrapservice_migration.go`：两骨架升 [S] + 新增 2 helper；`main.go`
resolveLauncherConfigFilePath 签名订正；`bootstrapservice.go` 两处 buildWorkspaceLayoutWithConfig
调用点改 8 参（MigrateConfig 处补接 configStoreSnapshot 第二返回值 ws）。全量 build/vet/test
回归 PASS，无新增测试。

## 移交（本轮收尾）

workspace migration 域**全部落地 [S]**（批次 270 的 12 函数 + 本批 buildWorkspaceLayoutWithConfig
收尾）；启动器背景图导入链 `importLauncherBackgroundImage` + 2 helper 落地 [S]。
`attachLauncherBackgroundURL`（0x140798600, 387 行 asm）仍 [S-sig]（launcher asset 背景域，
本批仅作为 import 的收口调用保留，非本批范围）。P=40 持平。FUNCS 2873/4754 = 60.43%。
