# 批次 268 · workspaceMigrationPathIsReparse 签名订正 + validateWorkspaceMigrationExistingChain +1 [S]（320B）

## 目标

1. **订正 batch 267 签名缺陷**：`workspaceMigrationPathIsReparse` 真实签名为
   `(path string) (bool, error)`，batch 267 误落单 `bool`（当批无调用点，编译器不报错）。
2. **落地** `workspacemigration_identity.go` 的 `validateWorkspaceMigrationExistingChain`
   （0x1409f42a0，320B），补工作区迁移「祖先链无 reparse」校验。

## 基线 / 收口

| 指标 | 基线（batch 267 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2864 | 2865 |
| MARKED | 2864 | 2865 |
| S | 1335 | 1336 |
| S-inline | 36 | 36 |
| S-sig | 1453 | 1453 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1371 | 1372 |
| USABLE | 1371 | 1372 |

`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

## G1 编译

`go1.25.12 build -tags production -trimpath ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1372  FUNCS=2865  MARKED=2865  P=40  S-eq=0  S-inline=36  S-sig=1453  S=1336  USABLE=1372
```

`validateWorkspaceMigrationExistingChain` 新增 [S]（S 1335→1336，FUNCS 2864→2865）。
签名订正不改指标（该函数已 [S]）。

## G3 行为（asm 逐地址实证）

### 3.1 workspaceMigrationPathIsReparse 签名订正

完整 dump 0x1409f4660（44 行）实证四个返回路径：

| 路径 | asm | 返回值 |
|---|---|---|
| UTF16PtrFromString err | 0x1409f4678 `test rbx,rbx; jne 0x1409f46a3`；0x1409f46a3 `xor eax,eax; ret` | `(false, err)` |
| GetFileAttributes err | 0x1409f4685 `test rbx,rbx; je 0x1409f4692`；0x1409f468a `xor eax,eax; ret` | `(false, err)` |
| 成功 | 0x1409f4692 `bt eax,0xa; setb al`；0x1409f4699 `xor ebx,ebx; xor ecx,ecx` | `(attrs&0x400!=0, nil)` |

调用点佐证：`validateWorkspaceMigrationExistingChain` 0x1409f42ea `test rbx,rbx; jne`（检查
err）→ 0x1409f42f3 `test al,al; jne`（检查 bool），证明双返回而非单 bool。订正为
`(bool, error)`，err 分支不再吞错。

### 3.2 validateWorkspaceMigrationExistingChain [S 0x1409f42a0]（320B）

签名（morestack spill 2 寄存器实证）：`(path string) error`，AX=ptr、BX=len。

循环体（0x1409f42bd 为循环头）：

1. `cleaned = filepath.Clean(path)`（0x1409f42b8，存 `[rsp+0x38]`/`[rsp+0x28]`）。
2. `os.Lstat(cleaned)`（0x1409f42c7）；`test rcx,rcx; jne 0x1409f4398` err 非 nil →
   `rax=rcx; rbx=rdi; ret`（直接返回 err）。
3. `workspaceMigrationPathIsReparse(cleaned)`（0x1409f42e5）；`test rbx,rbx; jne
   0x1409f438c` err 非 nil → 返回 err；`test al,al; jne 0x1409f4342` true → 构造错误。
4. reparse 错误：`convTstring(cleaned)` + `fmt.Errorf(格式串, cleaned)`（0x1409f4342-0x1409f4381）。
5. `parent = filepath.Dir(cleaned)`（0x1409f4302）；`parent == cleaned`（len 相等 +
   memequal，0x1409f430c/0x1409f4323）→ `xor eax,eax; xor ebx,ebx; ret`（return nil）。
6. 否则 `path = parent` 续环（0x1409f432c-0x1409f4336）。

字符串内存实证：格式串 `迁移路径不能经过符号链接、目录联接或 reparse point: %s`
@0x140c90f6e（48B，`0x1409f436f+0x29cbff` Python 显式计算）。

## proc 身份内存实证

| 调用 | 地址 |
|---|---|
| internal/filepathlite.Clean | 0x1401154c0 |
| internal/filepathlite.Dir | 0x140116440 |
| os.Lstat | 0x14012d9a0 |
| fmt.Errorf | 0x140133280 |
| workspaceMigrationPathIsReparse | 0x1409f4660 |

## G4 独立复核

仅改 `backend/workspacemigration_identity_windows.go`（签名订正 + 注释）、
`backend/workspacemigration_identity.go`（新增 `fmt`/`os` import + validate 函数体）。
`workspaceMigrationPathIsReparse` backend 内零调用点（grep 证实），签名订正无破坏面；
validate 新增 [S] 无既有签名/行为变更。全量测试回归 PASS。

## 移交（本轮收尾）

本批同时完成一次「全仓整洁」：工作树干净（git status 空，13 个忽略项均为
pipeline/tmp、disasm、__pycache__、backend.exe 等合法临时产物），修正 batch 267 遗留
签名缺陷。**下一批**：`inspectWorkspaceMigrationPath`（1696B）+
`sameWorkspacePathInspection`（384B）依赖 `workspacePathInspection`（104B）布局 +
`workspaceMigrationIdentityForExisting`（1472B），须先落地
`workspaceMigrationIdentityForExisting` 定型结构后再三函数同批。P=40。FUNCS 2865/4754 = 60.26%。
