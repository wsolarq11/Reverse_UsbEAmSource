# 批次 267 · workspacemigration 身份域 NOFOLLOW 文件打开链 +4 [S]（normalize 288B / source 96B / target 96B / regular 480B）

## 目标

落地 `workspacemigration_identity_windows.go` 剩余 4 函数（source_funcs.txt L4899-4902），
补齐工作区迁移「Windows 长路径规范化 + NOFOLLOW 文件打开」链：

- `normalizeWorkspaceWindowsFinalPath`（0x1409f4d00，288B）
- `openWorkspaceMigrationSourceFileNoFollow`（0x1409f4e20，96B）
- `openWorkspaceMigrationTargetFileNoFollow`（0x1409f4e80，96B）
- `openWorkspaceMigrationRegularFileWindows`（0x1409f4ee0，480B）

四函数此前**完全缺失**（不在 FUNCS 内，非 [S-sig] 升档），本批直接新增 [S]。

## 基线 / 收口

| 指标 | 基线（batch 266 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2860 | 2864 |
| MARKED | 2860 | 2864 |
| S | 1331 | 1335 |
| S-inline | 36 | 36 |
| S-sig | 1453 | 1453 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1367 | 1371 |
| USABLE | 1367 | 1371 |

`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

## G1 编译

`go1.25.12 build -tags production -trimpath ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1371  FUNCS=2864  MARKED=2864  P=40  S-eq=0  S-inline=36  S-sig=1453  S=1335  USABLE=1371
```

四函数为新增 [S]（S 1331→1335，FUNCS 2860→2864），S-sig/P/S-inline 持平（新增非升档）。

## G3 行为（asm 逐地址实证 + normalize 黄金测试）

### 3.1 normalizeWorkspaceWindowsFinalPath [S 0x1409f4d00]（288B）

签名（单 string 入 AX/BX 出）：`(path string) string`。

1. `call strings.TrimSpace`（0x1409f4d17）→ trimmed 存 `[rsp+0x30]`(ptr)/`[rsp+0x28]`(len)。
2. `call strings.ToLower`（0x1409f4d26）→ lower。
3. `cmp rbx,8; jl`（0x1409f4d2b/0x1409f4d2f）长度守卫；`movabs rdx,0x5c636e755c3f5c5c`
   即小端 `\\?\unc\`（8B），`cmp [rax],rdx; sete dl`（0x1409f4d3b/0x1409f4d3e）→ 前缀命中。
4. 命中：`concatstring2(a0=`\\`@0x140c3365d 2B, a1=trimmed[8:])`（0x1409f4d80，
   常量 `5c 5c`；rdi=ptr+8、rsi=len-8）→ Clean。
5. 未命中再 `cmp dword[rax],0x5c3f5c5c`（`\\?\`，4B，0x1409f4d8d）→ 命中则
   `rax=ptr+4, rbx=len-4`（0x1409f4dbf 的 `and ecx,4` 边界 trick）→ Clean。
6. 均未命中：原 trimmed 直接 Clean（`internal/filepathlite.Clean` 0x1401154c0）。

语义：`\\?\C:\x`→`C:\x`；`\\?\UNC\srv\share`→`\\srv\share`（保留 `\\` 头，前缀判断大小写不敏感但拼接用原始 trimmed）。

### 3.2 openWorkspaceMigrationSourceFileNoFollow [S 0x1409f4e20]（96B）

签名：`(path string) (*os.File, error)`。`mov ecx,0x80000000`(GENERIC_READ)、
`mov edi,3`(OPEN_EXISTING) → 直接 `call openWorkspaceMigrationRegularFileWindows`
（0x1409f4e40），返回值透传。

### 3.3 openWorkspaceMigrationTargetFileNoFollow [S 0x1409f4e80]（96B）

签名：`(path string) (*os.File, error)`。`mov ecx,0x40000000`(GENERIC_WRITE)、
`mov edi,1`(CREATE_NEW) → `call openWorkspaceMigrationRegularFileWindows`（0x1409f4ea0）。

### 3.4 openWorkspaceMigrationRegularFileWindows [S 0x1409f4ee0]（480B）

签名（morestack spill 4 寄存器实证）：`(path string, access uint32, disposition uint32)
(*os.File, error)`，AX=path.ptr、BX=path.len、CX=access、DI=disposition。

1. `call windows.UTF16PtrFromString`（0x1409f4f10）；err 非 nil → `return nil, err`
   （0x1409f4f18→0x1409f507f）。
2. `cmp esi,3; mov ecx,0; mov edx,7; cmove ecx,edx`（0x1409f4f25-0x1409f4f32）：
   sharemode = OPEN_EXISTING ? 7(READ|WRITE|DELETE) : 0。
3. `call windows.CreateFile(ptr16, access, sharemode, sa=nil, disposition,
   attrs=0x200080, template=0)`（0x1409f4f47）。`0x200080 =
   FILE_ATTRIBUTE_NORMAL|FILE_FLAG_OPEN_REPARSE_POINT`（NOFOLLOW 语义）。err 非 nil →
   return nil, err（0x1409f4f4f→0x1409f5077）。
4. `call windows.GetFileInformationByHandle`（0x1409f4f77，info @`[rsp+0x2c]`）；err 非 nil →
   CloseHandle + return nil, err（0x1409f4f83→0x1409f504f）。
5. `mov edx,[rsp+0x2c]; test edx,0x410`（0x1409f4f89/0x1409f4f8d）：FileAttributes 含
   FILE_ATTRIBUTE_REPARSE_POINT(0x400)|FILE_ATTRIBUTE_DIRECTORY(0x10) → CloseHandle +
   错误 A（0x1409f4f93→0x1409f5015）。
6. `cmp rax,-1`（InvalidHandle 守卫，0x1409f4fa0）→ `os.newFile(handle, path, 0)`
   （0x1409f4fc0）；file==nil → CloseHandle + 错误 B（0x1409f4fd0→0x1409f4fe0）。
7. 成功 `rax=file, rbx=0, rcx=0`（0x1409f4fd2-0x1409f4fd9）。

### 3.5 错误类型身份实证（复用 extractError）

错误对象 = `newobject(0x140b54380)`（16B `struct{msg string}`，ptrdata=8）+
`[+0]=msg ptr`、`[+8]=msg len`；itab 0x1411d3100，Error 方法（fun[0]=0x140090240）为
`mov rcx,[rax]; mov rbx,[rax+8]; ret`（返回 (ptr,len)）。

与 `openArchiveRegularFileNoFollow`（0x140750d20）错误构造**同址**：newobject type
`0x140750dcc+0x4035b4=0x140B54380`、itab `0x140750dea+0xa82316=0x1411D3100`。故复用
backend 既有 `newExtractError`。

字符串内存实证：

| 分支 | 消息 | 地址 | 长度 |
|---|---|---|---|
| A（reparse/directory） | `迁移文件最终句柄不是普通文件` | 0x140c81380 | 14 字符=42B=0x2a |
| B（os.NewFile nil） | `无法包装迁移文件句柄` | 0x140c6fb0d | 10 字符=30B=0x1e |

### 3.6 normalize 黄金测试（`workspacemigration_identity_windows_test.go`）

`TestNormalizeWorkspaceWindowsFinalPath` 8 用例覆盖：普通路径、`\\?\` 设备前缀、
`\\?\UNC\` 前缀（保留 `\\` 头）、大小写不敏感前缀、首尾空白、仅前缀无余量、空串
（Clean 得 `.`）。纯字符串无系统调用，确定性安全，全 PASS。

## proc 身份内存实证

| 调用 | 地址 | 备注 |
|---|---|---|
| internal/filepathlite.Clean | 0x1401154c0 | normalize 最终规范化 |
| windows.UTF16PtrFromString | 0x140194220 | regular 入参转换 |
| windows.CreateFile | 0x140195d20 | NOFOLLOW 打开 |
| windows.GetFileInformationByHandle | 0x140197060 | 普通文件校验 |
| windows.CloseHandle | 0x140195860 | 失败路径收口 |
| os.newFile | 0x140129ea0 | 包装 *os.File |

## G4 独立复核

仅改 `backend/workspacemigration_identity_windows.go`（新增 `os`/`path/filepath`/`strings`
import + 四函数 [S]），新增 `backend/workspacemigration_identity_windows_test.go`
（normalize 黄金测试，不计入 count_funcs）。无既有函数签名/行为变更；四函数此前无
backend 内调用点（grep 证实），全量测试回归 PASS。

## 下一批

`workspacemigration_identity.go` 剩余 3 函数（inspectWorkspaceMigrationPath 1696B /
validateWorkspaceMigrationExistingChain 320B / sameWorkspacePathInspection 384B）依赖本批
已落地的 normalize + NOFOLLOW 打开链，可成批落地。P=40。FUNCS 2864/4754 = 60.24%。
