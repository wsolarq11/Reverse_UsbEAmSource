# 批次 169 验收（launcherconfig_runtime.go：三个加载链 [P]→[S-sig] 签名坐实）

日期：2026-09-25
子批次：配置加载链三函数签名逐寄存器实证（loadOrCreate / loadOrDefault / readFile）

## 目标

`loadOrCreateLauncherConfig`（0x140875b40）、`loadLauncherConfigOrDefaultIfMissing`（0x140875e00）、
`readLauncherConfigFile`（0x1408788c0）三个 `[P]`（签名不确定）升 `[S-sig]`（签名逐寄存器实证、体为零值）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.4s) |
| 调用点回归 | 无（三函数当前无 Go 调用者，签名修正不破坏编译） | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1100 / S-inline=35 / S-sig=1343 / P=181 / UNMARKED=0`。

**真函数 = 1100 + 35 + 1343 = 2478 / 4754 = 52.1%**（相对批次 168 **增 +3**；P 184→**181**，S-sig 1340→1343，
三函数由「签名不确定 [P]」升为「签名实证 [S-sig]」，真函数口径 +3）。

重建产物 SHA256：`6A08A36B9E2B83C053CEB2E740CC6884101EAE99C9FB26CD1ED99813944ECF19`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（3 函数 [S-sig]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 旧签名 | 新签名（实证） |
|---|---|---|---|
| loadOrCreateLauncherConfig | 0x140875b40 | (path, lang string)(LauncherConfig,error) | (path string, opts []LauncherConfigOptions)(LauncherConfig,error) |
| loadLauncherConfigOrDefaultIfMissing | 0x140875e00 | (path, lang string) LauncherConfig | (path string, opts []LauncherConfigOptions)(LauncherConfig,bool) |
| readLauncherConfigFile | 0x1408788c0 | (path string)(LauncherConfig,error) | (path string, sourceTag string)(LauncherConfig,error) |

## G4 关键实证结论

1. **第二入参非 lang 而是 `[]LauncherConfigOptions`**：调用点 0x1409c8661（loadOrCreate）传
   `rcx=0/rdi=0/rsi=0` 空切片；调用点 0x14076478a（loadOrDefault）传 `rcx=&opts[0]/rdi=1/rsi=1`
   单元素切片。callee 侧 `rdx=[rcx]`、`rsi=[rdx+8]`、`rdi=[rdx+0x10]` 读取 opts[0] 的
   `DefaultTagCatalog` 切片（ptr,len,cap）并传给 `defaultLauncherConfigWithOptions`（3 寄存器切片）。
   旧重建树「(path,lang)」误把切片三寄存器当成了 lang 单串。
2. **LauncherConfigOptions 为 24B 单切片结构体**：`{DefaultTagCatalog []TagCatalogItem}`（types_launcher.go），
   `launcherConfigOptions(locale)` 返回该结构体，调用方栈上落单元素后以切片传入。
3. **readLauncherConfigFile 第二入参 sourceTag**：调用点 0x1407ab554 传 (rax,rbx)=path、(rcx,rdi)=
   sourceTag；callee 侧 TrimSpace 后嵌入 "解析配置失败"（18B UTF-8）错误消息，属错误上下文标签非 lang。
4. **返回约定**：三函数均经栈返回 LauncherConfig（调用方 `rep movsq 0x126` qword 复制 2352B）；
   loadOrCreate/readFile 经 `test rax` 判 error，loadOrDefault 经 `eax=0/1` 判 bool
   （加载成功 true / 缺省 false，与批次 61 记录的「按 bool 分支」一致）。

## 残留 / 未落地（不触及）

- 三函数体仍为零值（[S-sig]），待后续按汇编落体（loadOrCreate 含 LoadOrCreate 落盘、readFile 含
  json.Unmarshal + fmt.Errorf 包装链，均属大块头）。
- `defaultLauncherConfigWithOptions` / `normalizeLauncherConfigWithOptions` 当前重建签名疑似仍为
  `(lang string)` / `(cfg *LauncherConfig, lang string)`，汇编显示其入参实为 `[]TagCatalogItem` 切片
  （3 寄存器），待专项校正。
