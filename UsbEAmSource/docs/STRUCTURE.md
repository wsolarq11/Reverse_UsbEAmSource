# UsbEAmSource 目录结构（正向化整理后）

> 目标：把源码、分析产物、第三方工具、历史归档四类彻底分开，做到"打开任意一层都知道东西在哪"。
> 口径：可编译 + 功能一致重建（非字节级同哈希）。研究用途，版权 ©2026 DOGFIGHT360。

## 一、顶层布局

```
Reverse_UsbEAmSource/
├── UsbEAmSource/       # 主工程（Wails/Go 编译单元，唯一源码根）
├── work/               # 分析工作区（反汇编、Ghidra 工程、一次性脚本）
├── external/           # 第三方工具（Ghidra 12.1.3 安装 + 原始压缩包）
├── archive/            # 历史归档（旧脚本、旧产物、依赖备份、源码备份）
├── README.md           # 仓库入口说明
└── .gitignore
```

顶层只保留这四块。任何 `tmp_*`、一次性脚本、安装包、构建产物都不再散落在源码根。

## 二、主工程 `UsbEAmSource/`

```
UsbEAmSource/
├── backend/            # Go 后端（package main）
│   ├── main.go
│   ├── types_<domain>.go
│   └── *_test.go
├── frontend/
│   └── dist/           # 字节级还原的 Vue SPA（Wails 内嵌资源）
├── docs/               # 交付文档 + goresym 蓝图
│   ├── REPORT.md / HANDOFF.md / REBUILD_EXECUTION.md / STRUCTURE.md / STRATEGY.md
│   ├── BACKEND_SYMBOLS.md
│   ├── acceptance/     # 各批次质检验收记录
│   └── goresym/        # redress 符号、行号蓝图、按批反汇编资产
├── tools/              # 工程内工具
│   ├── go-introspect/  # addr_tool / va_dump / disasm / read_gostring 等
│   ├── dump_scripts/   # 批次批量 dump 与常量解码脚本
│   ├── split_types.py  # 类型拆分脚本
│   ├── parity/         # 对拍 harness（行为等价验证，见 tools/parity/README.md）
│   └── count_funcs.sh / count_funcs.awk
├── artifacts/          # 构建产物（不参与源码逻辑）
│   ├── UsbEAm_Launcher_rebuilt.exe
│   ├── backend_root.exe
│   └── backend_service.exe
├── build.sh            # 可复现构建（代理 + go1.25.12）
├── go.mod / go.sum
└── wails.json
```

## 三、分析工作区 `work/`

```
work/
├── disasm/             # 方法级反汇编资产
│   ├── dump/           # 批次手动 dump（含 batch35 / oled_batch22 / windowmgt_batch36）
│   ├── appicon_asm/    # 图标域 asm.txt
│   ├── appicon_bin/    # 图标域原始字节
│   ├── reparse/        # reparse point 相关 dump
│   ├── lease/          # bookmark icon lease 域 dump
│   ├── runnow/         # memory release runnow 域 dump
│   └── dump_batch1_shortcut/
├── ghidra/
│   ├── project/        # Ghidra 工程 usbeam.rep
│   ├── scripts/        # RunDecompile / AssembleDecompile / NextDiag 等
│   └── decompiled/     # Ghidra C 伪代码输出
└── scripts/            # 一次性分析脚本（原 tmp_json_decode.py 等）
```

## 四、第三方工具 `external/`

```
external/
├── ghidra_12.1.3_PUBLIC/            # 已解压安装
└── ghidra_12.1.3_PUBLIC_20260817.zip # 原始安装包
```

Ghidra 是外部工具，不归属于本项目源码，也不参与编译。

## 五、历史归档 `archive/`

```
archive/
├── legacy/             # 原顶层 _archive：早期前端抽取脚本与产物
├── gomod_backup/       # 原 UsbEAmSource/_archive/gomod：go.mod/go.sum 备份
├── source_backup/      # types.go.bak 等拆分前原始文件
└── scripts/            # 原 tmp_fallback_ct.py 等一次性脚本
```

## 六、路径书写约定

- 本目录下文档默认相对 `UsbEAmSource/` 书写，例如 `backend/main.go`、`artifacts/...`。
- 跨到顶层目录时显式写 `../work/...`、`../external/...`、`../archive/...`。
- 历史文档里的旧路径已尽量同步；若看到未同步的 `tmp_dump/`、`tmp_tools/`、`_archive/`，按第七节映射换算。

## 七、本次整理映射（可逆）

| 旧路径 | 新路径 |
|---|---|
| `_archive/`（顶层） | `archive/legacy/` |
| `tmp_fallback_ct.py` | `archive/scripts/tmp_fallback_ct.py` |
| `tmp_tools/ghidra_12.1.3_PUBLIC/` | `external/ghidra_12.1.3_PUBLIC/` |
| `tmp_tools/ghidra_12.1.3_PUBLIC_20260817.zip` | `external/ghidra_12.1.3_PUBLIC_20260817.zip` |
| `tmp_tools/ghidproj/` | `work/ghidra/project/` |
| `tmp_tools/myscripts/` + `tmp_tools/RunDecompile*` | `work/ghidra/scripts/` |
| `tmp_tools/decompiled/` | `work/ghidra/decompiled/` |
| `UsbEAmSource/tmp_appicon/` | `work/disasm/appicon_bin/` |
| `UsbEAmSource/tmp_appicon_disasm/` | `work/disasm/appicon_asm/` |
| `UsbEAmSource/tmp_dump/` | `work/disasm/dump/` |
| `UsbEAmSource/tmp_dump_batch1_shortcut/` | `work/disasm/dump_batch1_shortcut/` |
| `UsbEAmSource/tmp_reparse_dump/` | `work/disasm/reparse/` |
| `UsbEAmSource/docs/goresym/pipeline/tmp_lease/` | `work/disasm/lease/` |
| `UsbEAmSource/docs/goresym/pipeline/tmp_runnow/` | `work/disasm/runnow/` |
| `UsbEAmSource/tmp_tools/` | `UsbEAmSource/tools/dump_scripts/` |
| `UsbEAmSource/tmp_json_decode.py` | `work/scripts/tmp_json_decode.py` |
| `UsbEAmSource/backend.exe` | `UsbEAmSource/artifacts/backend_root.exe` |
| `UsbEAmSource/backend/backend.exe` | `UsbEAmSource/artifacts/backend_service.exe` |
| `UsbEAmSource/UsbEAm_Launcher_rebuilt.exe` | `UsbEAmSource/artifacts/UsbEAm_Launcher_rebuilt.exe` |
| `UsbEAmSource/backend/types.go.bak` | `archive/source_backup/types.go.bak` |
| `UsbEAmSource/_archive/gomod/` | `archive/gomod_backup/` |

反向移动即为回滚；移动过程未修改任何文件内容，仅脚本与文档引用同步更新。

## 八、构建与验证

```bash
cd UsbEAmSource
bash build.sh
```

产物：`UsbEAmSource/artifacts/UsbEAm_Launcher_rebuilt.exe`。

每批完工门禁仍按 `REBUILD_EXECUTION.md` 的 G1–G4 执行：`go build -tags production` + `go vet` + `go test` + 独立复核记录到 `docs/acceptance/<batch>.md`。
