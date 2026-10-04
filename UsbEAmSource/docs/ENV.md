# 环境与资产（UsbEAm Launcher 1.0.3 逆向还原）

> 从 `HANDOFF.md` 拆出。目标物、锁定构建参数、关键环境事实、反汇编资产清单的唯一真相源。

## 2. 目标物与已锁定身份

- 目标二进制：`D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe`（29,965,824 B）
- 身份（已实测）：
  - **Go 1.25.12**，PE32+ 64 位，非 .NET
  - 框架 **Wails v3**（`github.com/wailsapp/wails/v3@v3.0.0-alpha2.117`）+ WebView2
  - 前端 **Vite 打包的 Vue SPA**（内嵌字节可抽取）
  - 主模块 `changeme`；go.mod 根 = `changeme/backend`（重建时）
  - 依赖共 **25**（`go version -m` 可见全部 module+version+h1 hash）

### 锁定构建参数
`go1.25.12` · `-buildmode=exe` · `-compiler=gc` · `-tags=production` · `-trimpath=true` · `CGO_ENABLED=0` · GOOS=windows · `GOARCH=amd64` · `GOAMD64=v1`

## 3. 关键环境事实（务必先读）

1. **外网可用但走本地代理 `127.0.0.1:7890`**（Clash/V2Ray 类，系统已配）。
   - 必须给 Go 配代理：`export HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 ALL_PROXY=http://127.0.0.1:7890`
2. **精确工具链 Go 1.25.12 已装**在 `$(go env GOPATH)/bin/go1.25.12.exe`。用 `go1.25.12` 而非 `go`。
   ```bash
   GOBIN="$(go env GOPATH)/bin"
   "$GOBIN/go1.25.12.exe" version   # go1.25.12 windows/amd64
   ```
3. **已装工具**：`redress.exe`、`GoReSym.exe`、`go1.25.12.exe`、`addr_tool.exe`、`va_dump.py`/`disasm.py`。
   - `tools/go-introspect/` 下有全套反汇编流水线脚本。
4. 前端抽取脚本基于 Python 3.14（bash 里用 `python`，`python3` 是 WindowsApps stub）。

## 7. 反汇编资产清单

```
docs/goresym/pipeline/tmp/      — 方法字节+反汇编（2026-09-20 实测 489 asm.txt / 709 bin）
docs/goresym/disasm/            — 批次 29/31 手动核对版（同名文件可能更新更完整）
docs/goresym/disasm_assemble/   — 部分方法的跳转分析版
docs/goresym/disasm_assemble_image/ — 图片类方法跳转分析版
docs/goresym/disasm_lu/         — lookup 类方法分析版
docs/goresym/disasm_archive/    — 历史反汇编归档
docs/goresym/dump_archive/      — 历史 dump 归档
../work/disasm/dump/            — 方法级别手动 dump（windowmgt_batch36 / batch35 / oled_batch22）
../work/disasm/appicon_asm/     — 图标域 52 个 asm.txt
../work/disasm/appicon_bin/     — 图标域 108 个 bin
tools/go-introspect/            — addr_tool.exe / va_dump.py / disasm.py / read_gostring.py / resolve_lea_strings.py
```

新函数现场 dump 命令模板：
```bash
EXE="D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
python tools/go-introspect/va_dump.py "$EXE" 0xBASEVA LENGTH ../work/disasm/dump/PREFIX
```

