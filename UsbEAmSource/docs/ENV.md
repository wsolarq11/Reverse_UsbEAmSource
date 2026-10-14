# 环境与资产（UsbEAm Launcher 1.0.3 逆向还原）

> 从 `HANDOFF.md` 拆出。目标物、锁定构建参数、关键环境事实、反汇编资产清单的唯一真相源。
> 与本文冲突的旧文档数字，以本文 + 活体实测为准。

## 1. 目标物与已锁定身份

- 目标二进制：`D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe`
- 大小：`29,965,824 B`
- **SHA256：`995A12DAD88AA43CB3ACBE92A506624EB9EDC4599BA22B0A33EAFFF0C7924364`**
- 身份（已实测）：
  - **Go 1.25.12**，PE32+ 64 位，非 .NET
  - 框架 **Wails v3**（`github.com/wailsapp/wails/v3@v3.0.0-alpha2.117`）+ WebView2
  - 前端 **Vite 打包的 Vue SPA**（内嵌字节可抽取）
  - 主模块 `changeme`；go.mod 根 = `changeme/backend`（重建时）
  - 依赖共 **25**（`go version -m` 可见全部 module+version+h1 hash）

### 锁定构建参数

`go1.25.12` · `-buildmode=exe` · `-compiler=gc` · `-tags=production` · `-trimpath=true` · `CGO_ENABLED=0` · GOOS=windows · `GOARCH=amd64` · `GOAMD64=v1`

## 2. ⚠️ 目标二进制不入库 —— 复现前必读

目标 exe **不在本仓库**（版权与体积双重原因）。这意味着：

1. **反汇编坐标（VA、偏移、长度）的复核需要自备该 exe**，且必须是上表 SHA256 对应的那一份。
   指纹不符时，所有 VA 都可能错位——**先校验 SHA256，再动手**：

   ```bash
   sha256sum /d/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe
   # 期望 995a12dad88aa43cb3acbe92a506624eb9edc4599ba22b0a33eafff0c7924364
   ```

2. **`backend/*.go` 头部注释里的 `docs/goresym/pipeline/tmp/*.asm.txt` 等路径是溯源坐标**，
   指向磁盘上的分析中间产物。这些产物**不入库**（见 `STRUCTURE.md` §9），
   但**可由目标 exe 用 `va_dump.py` 重新生成**。克隆仓库后这些路径默认不存在，属预期行为。

3. 依赖**权威数据**（入库，可直接读）做复核，不依赖反汇编产物：
   - `docs/goresym/symbols.txt`（符号 → VA 全表）
   - `docs/goresym/source_funcs.txt`（行号蓝图，4,754 个函数项）
   - `docs/goresym/all_types.txt` / `types_normalized.go`（类型蓝图）

## 3. 关键环境事实（2026-12-12 实测核准）

| 工具 | 状态 | 位置 / 说明 |
|---|---|---|
| `go` | ✅ | `C:\Users\Administrator\go\bin\go.exe`（go1.25.12）。**用绝对路径调用** |
| `git` | ✅ | `C:\Program Files\Git\cmd\git.exe` |
| `bash` | ✅ | `C:\Program Files\Git\usr\bin\bash.exe` |
| `awk` | ✅ | 同上（Git for Windows 自带，非 gawk） |
| `gh` | ✅ | `C:\Program Files\GitHub CLI\gh.exe`，已认证 `wsolarq11`，token 含 `workflow` scope |
| `node` | ✅ | `C:\Program Files\nodejs\node.exe` |
| `python` | ✅ | `C:\Python314\python.exe`（**用 `python`，`python3` 是 WindowsApps stub**） |
| `scoop` | ❌ 不在 PATH | 且 `scoop/apps/temurin-lts-jdk/current` 是**断裂链接**（JDK 已卸载） |
| `dulwich` | ❌ 未安装 | 旧的 `tmp_git_sync.py` 已不存在 —— commit/push 一律用原生 `git` |
| Ghidra | ✅ 可用 | 12.1.3 在 `external/`；JDK 21 在 `D:\_tools_\jdk\jdk-21.0.12.1+1`；`JAVA_HOME`（User 级）已指向它，新终端可原生跑 `analyzeHeadless.bat`。钉版本的入口：`tools/ghidra/ghidra_headless.ps1`。已知限制：Go 大结构体按值传参的函数反编译报 `Cannot assign parameter address`（工具限制，与脚本无关） |

其余环境事实：

1. **外网走本地代理 `127.0.0.1:7890`**（Clash/V2Ray 类，系统已配）。Go 需显式配代理：
   `export HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 ALL_PROXY=http://127.0.0.1:7890`
2. 精确工具链 Go 1.25.12 在 `C:\Users\Administrator\go\bin\go.exe`。
3. 已装逆向工具：`redress.exe`、`GoReSym.exe`、`addr_tool.exe`、`va_dump.py`（见 §5）；现成的反汇编查询入口是 `tools/index/asmtool.py`。
4. **`go` 的构建缓存与 `GOTOOLCHAIN`**：本地验证时若出现工具链自动下载，用 `GOTOOLCHAIN=local` 锁定。

## 4. 反汇编资产清单（全部不入库，可由目标 exe 重建）

```
docs/goresym/pipeline/tmp/          — 方法字节 + 反汇编（2026-12-12 实测 2955 asm.txt / 2755 bin，共 5756 文件 / 22.3 MB）
docs/goresym/disasm/                — 批次 29/31 手动核对版（同名文件可能更新更完整）
docs/goresym/disasm_assemble/       — 部分方法的跳转分析版
docs/goresym/disasm_assemble_image/ — 图片类方法跳转分析版
docs/goresym/disasm_lu/             — lookup 类方法分析版
docs/goresym/disasm_archive/        — 历史反汇编归档
docs/goresym/dump_archive/          — 历史 dump 归档
../work/disasm/dump/                — 方法级别手动 dump（windowmgt_batch36 / batch35 / oled_batch22）
../work/disasm/appicon_asm/         — 图标域 asm.txt
../work/disasm/appicon_bin/         — 图标域原始字节
tools/go-introspect/                — 抽取流水线脚本（入库）
tools/parity/bin/                   — 对拍抽取产物（不入库，由 extract_func.py 重建）
```

## 5. 现场抽取命令

```bash
EXE="D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
# 按 VA + 长度抽函数字节 + 反汇编
python tools/go-introspect/va_dump.py "$EXE" 0xBASEVA LENGTH ../work/disasm/dump/PREFIX
# 符号 → VA 查询
"tools/go-introspect/addr_tool.exe" "$EXE" <SymbolName>
```

对拍基线抽取（带门禁，只支持无 call 的叶子函数）：

```bash
python tools/parity/emu/extract_func.py <funcVA> <length> <outName>
```

Ghidra headless（唯一入口；`-scriptPath` 目录里任一脚本编译不过会让**整个目录**的脚本都装载失败）：

```bash
pwsh tools/ghidra/ghidra_headless.ps1 ../work/ghidra/project usbeam \
  -process -noanalysis -readOnly \
  -scriptPath ../work/ghidra/scripts -postScript InspectAssemble.java
```
