# 批次 108 验收（路径/可执行目标判断链 + 提权判定）

日期：2026-09-23
子批次：pathclip-logic
目标：落地 pathclip 判断链 7 个新函数 + `newPathError` [S-inline]→[S] 补体。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'NewPathError|IsWindowsAbsoluteFilesystemPath|HasKnownWindowsExecutableExtension|GetTokenElevation' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1180 / MARKED=1180 / S=871 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 107（1173/863/35/274/1/0）：7 个新函数 `[S]` 落地（FUNCS +7，S +7）；
`newPathError` 由 `[S-inline]` 升格 `[S]`（S +1，S-inline -1）。`UNMARKED=0` 保持。

## G3 逻辑等价

新文件 `backend/pathclip_logic_windows.go`（7 函数）+ `bootstrap_pathclip.go` `newPathError` 补体：

- **shouldOpenPathViaExplorer**（0x1408a8f20, 96B）：`isProcessElevated()` 非真 → false；
  否则 `isWindowsExecutableOpenTarget(path)`。
- **isWindowsExecutableOpenTarget**（0x1408a8f80, 128B）：TrimSpace → 非绝对路径 → false；
  `saferIsExecutableFileType` err==nil → 返回其 bool（`test rbx`@0x1408a8fba）；err 非空 →
  回退 `hasKnownWindowsExecutableExtension`。
- **saferIsExecutableFileType**（0x1408a9000, 256B）：LazyProc=`SaferiIsExecutableFileType`
  （name 26B @0x140c68181，advapi32）；`Find()` 失败 → (false,err)；`UTF16FromString` →
  `Call(&u[0], 0)` → `ret != 0`。
- **hasKnownWindowsExecutableExtension**（0x1408a9100, 544B）：TrimSpace → 尾部扫 `.`
  （`\`/`/` 中止）→ `ToLower` → len 3/4/10 三档比较。集合（dword 立即数逐项解码）：
  `.vb .ws / .bat .cmd .com .cpl .lnk .pif .scr .url .vbe .vbs .exe .hta .jse .msc
  .msi .msp .ps1 .wsc .wsf .wsh / .appref-ms`。
- **isWindowsAbsoluteFilesystemPath**（0x1408a9320, 224B）：TrimSpace 空 → false；
  len≥2 且 `\\` 或 `//`（`cmp word 0x5c5c/0x2f2f`）→ true；len≥3 且 `[A-Za-z]:[\\/]` → true。
- **getTokenElevation**（0x1408accc0, 160B）：`GetTokenInformation(token, 0x14, &elevation,
  4, &size)` 失败 → `wrapWinError("读取令牌信息失败")`；成功 → `elevation != 0`。
- **isProcessElevated**（0x1408ad900, 224B）：`OpenProcessToken(-1, TOKEN_QUERY, &token)`
  失败 → false；defer Close；`getTokenElevation` → `elevated && err==nil`（`sete/and`）。
- **newPathError**（内联于 OpenPath）：`errors.New("路径不能为空")`（18B 消息 @0x140c59e5e，
  实证 errorString{s: 消息} 而非裸 nil）。

## G4 测试

`backend/pathclip_logic_windows_test.go` 4 用例：

- `TestNewPathError`：非 nil + 消息 `路径不能为空`。
- `TestIsWindowsAbsoluteFilesystemPath`：11 例（盘符/UNC/大小写/空白/相对/非法盘符）。
- `TestHasKnownWindowsExecutableExtension`：23 正例 + 7 负例（含大写 ToLower 命中）。
- `TestGetTokenElevation`：真实 OpenProcessToken → getTokenElevation 可调用不报错。

全 PASS。

## 判定

四路 PASS，批次 108 闭环。判断链 7 函数全 `[S]`，`newPathError` 补体为真实错误对象。
扩展名集合实测为 23 项（`.appref-ms` 为唯一 10 字符项），`saferIsExecutableFileType` 实测
调用 `SaferiIsExecutableFileType`（非 `AssocQueryString`），均已在代码注释标注。
