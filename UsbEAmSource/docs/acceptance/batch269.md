# 批次 269 · workspacemigration_identity 域三函数收口：identityForExisting + inspect + same（+3 [S]）

## 目标

1. **定型结构体**：复用 `types_workspace.go` 既有 `workspacePathIdentity` /
   `workspacePathInspection`，用 `.eq` 方法（0x140a0ace0 / 0x140a0ac20）逐字段实证布局。
2. **落地** `workspaceMigrationIdentityForExisting`（0x1409f46e0，1472B），为已存在路径
   构建 Windows 文件身份（卷序列号 + 文件索引）。
3. **落地** `inspectWorkspaceMigrationPath`（0x1409f3c00，1696B），向上找现存祖先、
   拼接缺失段、产出 `workspacePathInspection`。
4. **落地** `sameWorkspacePathInspection`（0x1409f43e0，384B），判定两个检查结果是否同目标。

## 基线 / 收口

| 指标 | 基线（batch 268 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2865 | 2868 |
| MARKED | 2865 | 2868 |
| S | 1336 | 1339 |
| S-inline | 36 | 36 |
| S-sig | 1453 | 1453 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1372 | 1375 |
| USABLE | 1372 | 1375 |

`go build/vet/test -tags production ./backend` 全 EXIT=0。

## G1 编译

`go build -tags production ./backend` EXIT=0；
`go vet -tags production ./backend` EXIT=0；
`go test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1375  FUNCS=2868  MARKED=2868  P=40  S-eq=0  S-inline=36  S-sig=1453  S=1339  USABLE=1375
```

三函数均新增 [S]（S 1336→1339，FUNCS 2865→2868），无 [P]/[S-sig] 新增，P=40 持平。

## G3 行为（asm 逐地址实证）

### 3.1 结构体布局定型（`.eq` 方法逐字段）

`.eq.main.workspacePathIdentity`（0x140a0ace0）：比较 canonical（len@+8 相等后
memequal@+0/+8）、volume（qword@+0x10）、file（qword@+0x18）、valid（byte@+0x20），
40B 与 `types_workspace.go` 声明一致。

`.eq.main.workspacePathInspection`（0x140a0ac20）：比较 canonical（memequal@+0/+8）、
exists（byte@+0x10）、identity（`.eq.workspacePathIdentity`@+0x18）、
ancestorIdentity（`.eq.workspacePathIdentity`@+0x40），104B 一致；exists 后 Go 自动 pad 7B。

### 3.2 workspaceMigrationIdentityForExisting [S 0x1409f46e0]（0x5C0=1472B）

签名（morestack spill 4 寄存器实证）：`(path string, _ os.FileInfo) (workspacePathIdentity, error)`。

流程（逐地址）：

1. `windows.UTF16PtrFromString(path)`（0x1409f4738）；err 非 nil → 0x1409f4993 返回。
2. `windows.CreateFile(ptr, 0x80, 7, nil, 3, 0x2200000, 0)`（0x1409f4763）：
   access=FILE_READ_ATTRIBUTES、share=READ|WRITE|DELETE、disposition=OPEN_EXISTING、
   attrs=FILE_FLAG_BACKUP_SEMANTICS|FILE_FLAG_OPEN_REPARSE_POINT。err → 0x1409f492b 返回。
3. `defer windows.CloseHandle(handle)`（0x1409f4779-0x1409f47a0 构造 deferwrap1 funcval，
   deferwrap1 0x1409f4ca0 = `CloseHandle([rdx+8])`）。
4. `windows.GetFileInformationByHandle(handle, &info)`（0x1409f47d1）；err → 0x1409f48ad 返回。
5. `bt edx,0xa`（FileAttributes bit10=REPARSE_POINT，0x1409f47e6）命中 →
   `newExtractError("迁移路径最终句柄指向 reparse point")`（0x1409f480a-0x1409f486a 后返）。
6. `GetFinalPathNameByHandle` 循环：初始栈 buf 1024（0x1409f47fb/0x1409f4800），
   `size<=n`（0x1409f4a64 `jle 0x1409f49fb`）→ `make([]uint16, n+1)` 重试（0x1409f49fb-0x1409f4a18）。
7. 成功：`normalizeWorkspaceWindowsFinalPath(string(utf16.Decode(buf[:n])))`
   （0x1409f4ab1 unicode/utf16.decode + 0x1409f4ac1 runtime.slicerunetostring +
   0x1409f4ac6 normalize）。
8. 身份字段（0x1409f4af6-0x1409f4b1b）：volume=VolumeSerialNumber（info+0x1c，dword 零扩展）、
   file=FileIndexHigh<<32|FileIndexLow（info+0x2c/0x30）、valid=true。

**Filetime align=4 偏移实证**：x/sys `ByHandleFileInformation` 中 `Filetime{uint32,uint32}`
对齐为 4，故 VolumeSerialNumber@+0x1c、FileSizeHigh@+0x20、FileSizeLow@+0x24、
NumberOfLinks@+0x28、FileIndexHigh@+0x2c、FileIndexLow@+0x30（info 基址 0x834，asm 读
0x850/0x860/0x864）。

**info 死参数实证**：序言（0x1409f46fd-0x1409f4738）只存 rax（path.ptr），不存 rcx/rdi；
rcx 被 `mov ecx,7`（0x1409f474e）、rdi 被 `xor edi,edi`（0x1409f4753）直接覆盖，此前零读取。

### 3.3 inspectWorkspaceMigrationPath [S 0x1409f3c00]（0x6A0=1696B）

签名：`(path string) (workspacePathInspection, error)`（104B 结构经栈返回区 [rsp+0x200]）。

流程：

1. `filepath.Abs(path)`（0x1409f3c54）；err 非 nil → 0x1409f3cac 返回 err。
2. `cleaned = filepath.Clean(abs)`（0x1409f3c60，存 [rsp+0x150]/[rsp+0x80]）。
3. 循环 `os.Lstat(current)`（0x1409f3d2b）：
   - err==nil → break 得 info（0x1409f3f89）。
   - `errors.Is(err, os.ErrNotExist)` 假 → 0x1409f3f42 返回 err。
   - `parent = filepath.Dir(current)`（0x1409f3d80）；`parent==current`（len+memequal）→
     `fmt.Errorf("找不到迁移路径的现有父目录: %s", cleaned)`（0x1409f3ebb-0x1409f3f0c）。
   - 否则 `missing = append(missing, filepath.Base(current))`（0x1409f3dd7-0x1409f3eb6），
     `current = parent` 续环。
4. `validateWorkspaceMigrationExistingChain(current)`（0x1409f3fa3）；err → 0x1409f4082 返回。
5. `workspaceMigrationIdentityForExisting(current, info)`（0x1409f3fcb）；err → 0x1409f404b 返回。
6. 从尾到头 `fullPath = filepath.Join(fullPath, missing[i])`（0x1409f40bc 循环，
   join 0x1409f4118），再 `filepath.Clean`（0x1409f4140）。
7. `exists = (len(missing)==0)`（0x1409f418b `sete`）；identity 仅 exists 时填
   （0x1409f41cd），ancestorIdentity 恒填（0x1409f4193-0x1409f41bb）。

### 3.4 sameWorkspacePathInspection [S 0x1409f43e0]（0x180=384B）

签名：`(a, b workspacePathInspection) bool`（两 104B 结构经栈，A@rsp+0x80、B@rsp+0xe8）。

流程：

1. `a.exists != b.exists`（0x1409f4400 `cmp [rsp+0xf8], dl`）→ 0x1409f4545 返 false。
2. exists 真：`!a.identity.valid || !b.identity.valid`→false；`volume` 不等→false；
   `file` 不等→false；否则 true（0x1409f4463-0x1409f448e）。
3. exists 假：`!a.ancestorIdentity.valid || !b.ancestorIdentity.valid`→false；
   `volume`/`file` 不等→false；否则 `workspacePathsEqual(a.canonical, b.canonical)`
   （0x1409f44e4-0x1409f452e）。

## 错误类型身份内存实证

reparse 错误：newobject type `0x140b54380`（`0x1409f4811+0x15fb6f`）+ itab `0x1411d3100`
（`0x1409f484a+0x7de8b6`），与 `openWorkspaceMigrationRegularFileWindows` 三处错误同址
（0x1409f5026+0x15f35a / 0x1409f5044+0x7de0bc），复用 backend 既有 `newExtractError`。

字符串内存实证：`迁移路径最终句柄指向 reparse point` @0x140c832b9（44B）；
`找不到迁移路径的现有父目录: %s` @0x140c822af（43B）。

## proc 身份内存实证

| 调用 | 地址 |
|---|---|
| windows.UTF16PtrFromString | 0x140194220 |
| windows.CreateFile | 0x140195d20 |
| windows.GetFileInformationByHandle | 0x140197060 |
| windows.CloseHandle | 0x140195860 |
| windows.GetFinalPathNameByHandle | 0x140197160 |
| unicode/utf16.decode | 0x1400e08e0 |
| runtime.slicerunetostring | 0x1400607c0 |
| os.Lstat | 0x14012d9a0 |
| errors.Is | 0x1400904a0 |
| internal/filepathlite.Clean/Dir/Base | 0x1401154c0 / 0x140116440 / 0x140116320 |
| path/filepath.abs / path/filepath.join | 0x1401cd5e0 / 0x1401cd660 |
| fmt.Errorf | 0x140133280 |

## G4 独立复核

仅改 `backend/workspacemigration_identity.go`（新增 `errors` import + 2 函数）、
`backend/workspacemigration_identity_windows.go`（新增 `unicode/utf16` import + 1 函数）。
三函数 backend 内零调用点（grep 证实），无既有签名/行为变更。复用 `types_workspace.go`
既有结构体声明，未新建类型、未触碰 types_config.go 的 `stagedWorkspaceData` 字段。全量测试回归 PASS。

## 移交（本轮收尾）

workspacemigration_identity 域（两文件 10 函数）**全部落地 [S]**。**下一批**：
workspace migration 域剩余骨架升档 —— `validateWorkspaceMigrationRoots`（0x1409efac0）、
`workspacePathContains`（0x1409f0160）、`waitWorkspaceMigrationDrain`（0x1409ef500）、
`removeOwnedWorkspaceMigrationDirectory`（0x1409f34c0）四函数未落地 +
`stageWorkspaceDataMigration`（0x1409f0240）/`stagedWorkspaceData.Commit`（0x1409f2460）/
`stagedWorkspaceData.Rollback`（0x1409f3100）三 [S-sig] 骨架待体。P=40。FUNCS 2868/4754 = 60.33%。
