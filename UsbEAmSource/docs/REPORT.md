# UsbEAm Launcher 1.0.3 — 源码还原交付说明

> 目标（用户确认口径）：**可编译 + 功能一致的重建源码**（非字节级同哈希）。
> 还原对象：`D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe`（29,965,824 字节）

## 一、二进制鉴定（实测证据）

| 项目 | 结论 | 证据 |
|---|---|---|
| 位数/格式 | PE32+ 64 位，非 .NET | Optional Header magic `0x20B`；CLR header RVA=0 |
| 语言 | **Go 1.25.12** 编译 | 文件头 `Go build ID` + `go1.25.12` 字符串 + pclntab 全符号 |
| 框架 | **Wails v3**（v3.0.0-alpha2.117）+ WebView2 | `go version -m`、wails symbols |
| 前端 | **Vite 打包的 Vue SPA**（内嵌于 exe）| `frontend/dist/index.html`、`index-*.js/css` 内嵌 |
| 主模块路径 | `changeme`（Wails 默认脚手架名未改）| buildinfo `path changeme` |
| 核心依赖 | modernc sqlite、lunar-go、go-pinyin、gozxing、clipboard 等 | buildinfo（见 go.mod/go.sum）|

关键判断：**官方未开源**（模块名 `changeme`、无公开仓库），但**二进制可高度还原**——前端为内嵌 Vite 产物（近完整还原），Go 后端因符号完整可还原到近源码。

## 二、本次实际产出

### 1. 前端源码（已字节级抽取）`frontend/dist/`
从 exe 内嵌数据按文件边界切割还原，**内容与目标构建完全一致**：
- `index.html` —— 应用入口（Vue，`<div id="app">`）
- `assets/index-CE9BgBo7.js`（1.5 MB 主 bundle `const __vite__mapDeps=...`
- `assets/index-Rtj8v7ar.css`（主样式）
- 12 个组件 JS + 12 个组件 CSS（AudioManager / GpuPreference / MemoryRelease / OledBlackout / WindowManagement / MouseGesture / FileLocator / TwoFactor / QRCode / Screenshot / PluginHost / DesktopWidgets）
- `_inline_wails_shim.js` —— Wails 前端运行时桥（Events/Emit 等）

> `mapDeps` 共 25 个 chunk，已 25/25 核对；唯一 `scaledScrollPosition-C56c8HVu.js` 被 Vite 内联进父 chunk（其逻辑已在主/组件 bundle 中），非独立 URL。

### 2. Go 后端 —— 重建骨架（待符号级补全）
- `main` 包已定位符号：`main.AppEntry`、`main.IndexNode`、`main.item`、`main.loadJob`、`main.walkState`、`main.LinkEntry`、`main.FileEntry`、`main.AudioState`、`main.hashWriter`、QRCode/TwoFactor 相关服务等。
- 需用 **GoReSym**（`github.com/mandiant/GoReSym`）+ Ghidra 从 pclntab 恢复完整类型/字段布局后，重建可编译的 `main` 包源码（功能等价）。

### 3. 构建配置（已生成，与目标构建完全一致）
- `go.mod` —— 模块 `changeme`，Go 1.25，27 个依赖锁定；
- `go.sum` —— 25 个精确 `h1:` 哈希（取自目标 exe 内嵌 buildinfo，非手抄）；
- 构建参数（`go version -m` 验证）：
  - `-buildmode=exe`，`-compiler=gc`，`-tags=production`
  - `-trimpath=true`，`CGO_ENABLED=0`，`GOARCH=amd64`，`GOAMD64=v1`，GOOS=windows

## 三、为何「字节级同哈希 exe」不可行（已与用户对齐）
Go 把源码编译为原生机器码。要产出与目标**逐字节相同**的 exe，重建的 `main` 源码必须编译出完全一致的指令布局/分配顺序——这要求拿到作者原始 `.go` 源码文本，而非从反编译重建。因此只能做到 **可编译 + 功能一致**，不能保证同哈希。**前端层例外**：内嵌的 html/js/css 是原样字节，若将其按原样重新打包并交给同一 Go 代码 serve，前端的字节是可复现的。

## 四、编译验证（已完成，非网络受限）
**已实际编译通过**（本机直连外网超时，但本地代理 `127.0.0.1:7890` 走通；此前一度误判"无外网"，实为未给 Go 配代理）。

已完成的验证：
- 安装 **精确工具链 Go 1.25.12**（`golang.org/dl/go1.25.12`）；
- `go mod download`：**go.mod/go.sum 的 25 个 h1 哈希核对真实模块全通过**；
- `go build -tags production -trimpath -buildmode=exe ./backend` → 产出 `artifacts/UsbEAm_Launcher_rebuilt.exe`（8.6 MB）；
- 重编译产物 buildinfo 与目标一致：`go1.25.12`、`wails/v3@v3.0.0-alpha2.117`、`-buildmode=exe`、production、trimpath。

## 五、剩余工作（后端 main 完整逻辑）→ 已解决类型层
**pclntab 阻碍已解决**：用 **redress**（`github.com/goretk/redress`，活跃维护、支持 go1.25）取代 GoReSym v1.7.1。
- `redress types all` → **507 个 main 结构体（含完整字段 + json tag）** 全部还原，写入 `backend/types.go`；
- `redress source` → 154 个源文件/函数布局（`appdisplayname.go/audio.go/bookmarks.go/desktopwidgets_*.go/filesearch_*.go` 等）；
- **`go build -tags production -trimpath -buildmode=exe ./backend` 编译通过** → `artifacts/UsbEAm_Launcher_rebuilt.exe`（14.2 MB），dep 数 4→14。

### 仍需（后续大量工作）
- **函数/方法体**尚未还原（redress 给出函数名+行号，但无字节级函数体源码；方法体重建需逐函数从反汇编推导，属超大批量工程）。
- 依赖数 14→25：补全服务方法体后随 `image/sqlite/w32/…` 用法提升。

## 六、当前进展快照（2026-09-20 实测校准）

> 本节由活体测量得出，**取代 §二、§五 中的早期数字**（§二写于类型层阶段）。完整交接、未落地文件清单与下一批入口见 [HANDOFF.md](HANDOFF.md)。

| 层级 | 目标 | 实测 | 状态 |
|---|---|---|---|
| 前端 | 字节级一致 | Vite 打包 Vue SPA，25/25 chunk 核对 | ✅ 完成 |
| Go 类型层 | 全量结构体 | 507 个 `main` 结构体（redress） | ✅ 完成 |
| 构建配置 | 与目标一致 | go1.25.12 + wails v3.0.0-alpha2.117 + 25 依赖 | ✅ 完成 |
| 后端函数体 | 功能一致 | 969 函数：`S=551 / S-sig=100 / S-inline=1 / P=49 / UNMARKED=268` | 🔄 进行中，批次 39 闭环 |
| 蓝图覆盖 | 4,754 函数 / 145 文件 / ≈104,374 行 | 函数覆盖 ≈20%；文件覆盖 24%（35/145） | 🔄 110 个原始文件未落地 |

**门禁实测（2026-09-20）**：`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend`、`go vet ./backend`、`go test -count=1 -p=1 ./backend` 三项 **EXIT=0**（`ok changeme/backend 0.162s`）。

**规模**：backend 非测试代码 112 文件 / 25,104 行；测试 42 文件 / 5,446 行。活体统计命令：`bash tools/count_funcs.sh`。

## 七、版权/合规边界
UsbEAm Launcher（©2026 DOGFIGHT360）为闭源共享软件。以下仅限**个人学习/研究**：
- 反编译、代码还原不得用于再分发、二次开发发布、破解授权、商业化。
- 大陆《著作权法》及《计算机软件保护条例》对反向工程有严格限制；若用于合法改造，应联系作者或采用其已开源的社区数据仓库/第三方替代实现。
- 本工程源码仅作研究示例，不得对外发布为「UsbEAm 开源版」。