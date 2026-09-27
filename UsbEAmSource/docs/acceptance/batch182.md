# 批次 182 — 内存释放执行链闭环（4 函数 + 1 错误类型 + 工厂装配修正）

## 基线 / 收口

| 指标 | 基线（批次 181 收口） | 收口（批次 182） |
|---|---|---|
| FUNCS | 2707 | **2712** |
| S | 1154 | **1158** |
| S-inline | 35 | **36** |
| S-sig | 1403 | 1403 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2592 | **2597** |

SHA256 `9A2172D378A74D72F85A450F17BCE5F348FDC4CB5A9887EEA1CEB232DFAFD04E`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（memoryrelease 执行链，全部 truly-missing → [S]）

新文件 `backend/memoryrelease_exec.go`（5 个 func 声明 + 2 个匿名闭包）。

### `executeMemoryReleaseMode` 0x1408d5120（160B）

`func(mode string) error`。`parseMemoryReleaseMode(mode)` 失败 → 回退 `"standbylist"`（11B @0x140c47523）；`resolveMemoryReleaseCommand(normalized)` 出错即返回；否则 `withMemoryReleasePrivilege("SeProfileSingleProcessPrivilege" /*31B @0x140c71007*/, fn)`。

### `executeMemoryReleaseMode.func1` 0x1408d51c0（256B，匿名闭包）

`NtSetSystemInformation(0x50 /*SystemMemoryListInformation*/, &cmd, 4)`；`err==nil`→nil；`err==STATUS_PRIVILEGE_NOT_HELD`（itab 0x1411d2e00 / data 0x1411cd478=`0xc0000061`，经 ifaceeq 确证）→ `&memoryReleaseError{"当前进程没有足够权限，请使用管理员身份启动后再试" /*72B @0x140c90ede*/}`；其余 → `fmt.Errorf("执行内存释放失败: %w" /*28B @0x140c6bcf4*/, err)`。

### `resolveMemoryReleaseCommand` 0x1408d52c0（448B）

`func(mode string) (uint32, error)`。逐字节 cmp 的 SYSTEM_MEMORY_LIST_COMMAND 映射（经长度索引 `cmp rbx,0xb/0x10/0x14` 分派）：

| 规范模式 | 命令值 |
|---|---|
| `standbylist`（"standbyl"+"is"+"t"） | 4 (MemoryPurgeStandbyList) |
| `workingsets`（"workings"+"et"+"s"） | 2 |
| `modifiedpagelist`（"modified"+"pagelist"） | 3 |
| `priority0standbylist`（memequal 20B） | 5 (LowPriority) |

非法 → `(0, fmt.Errorf("不支持的内存释放模式: %s" /*34B*/, TrimSpace(mode)))`。

### `withMemoryReleasePrivilege` 0x1408d5480（1120B）

`func(name string, fn func() error) (err error)`。`TrimSpace(name)==""` → `&memoryReleaseError{"缺少内存释放所需的系统权限名" /*42B @0x140c80fe4*/}`；`fn==nil` → `&memoryReleaseError{"缺少内存释放原生操作" /*30B @0x140c6f3ab*/}`；`LockOSThread`+`defer UnlockOSThread`；`UTF16PtrFromString` 失败 → `fmt.Errorf("准备系统权限失败: %w" /*28B @0x140c6bd10*/)`；`LookupPrivilegeValue` 失败 → `fmt.Errorf("查询系统权限失败: %w" /*28B @0x140c6bd2c*/)`；`openCurrentProcessToken(0x28)` 失败 → 原样返回；`defer token.Close()`；`adjustMemoryReleaseTokenPrivileges` 失败 → `fmt.Errorf("启用系统权限失败: %w" /*28B @0x140c6bd48*/)`；`defer` 恢复权限（失败时 `errors.Join` 原错误与 `fmt.Errorf("恢复系统权限失败: %w" /*28B @0x140c6bd64*/)`）；最后 `return fn()`。

### `withMemoryReleasePrivilege.func1` 0x1408d58e0（恢复权限 defer 闭包）

捕获 token/previous/errPtr，`adjustMemoryReleaseTokenPrivileges(token, previous, nil, nil)` 失败时按 `err==nil` 直写 vs `errors.Join`（asm 0x1408d59db 循环统计非 nil 项 + makeslice + joinError 内联展开确证）。

### `adjustMemoryReleaseTokenPrivileges` 0x1408d5c00（448B）

`func(token windows.Token, newState, previousState *windows.Tokenprivileges, retLen *uint32) error`。`newState==nil` → `&memoryReleaseError{"缺少系统权限状态" /*24B @0x140c65310*/}`；`buflen = previousState!=nil ? 16 : 0`；`AdjustTokenPrivileges(token,false,newState,buflen,previousState,retLen)` 失败 → 原样返回；`GetLastError()==ERROR_NOT_ALL_ASSIGNED`（data 0x1411cd3e0，BSS 运行时填充 1300）→ `&memoryReleaseError{"当前进程不是管理员权限，无法启用内存释放所需特权" /*72B @0x140c90f26*/}`；其余 → 返回 `GetLastError()`（nil 或其余 Errno）。

### `memoryReleaseError` 类型 + `Error()`

16 字节 `struct { msg string }`（newobject 0x140b54380 确证 kind=25/size=16，字段 `[obj]=msg.ptr` / `[obj+8]=msg.len`），以 itab 0x1411d3100（`*memoryReleaseError` 实现 error，Type=0x140b25580）返回。`Error()` 标记 `[S-inline]`。

## 工厂装配修正（memoryrelease.go `newMemoryReleaseService`）

asm 0x1408d3560 逐字段对照补齐 3 处遗漏（非新函数，正确性修复）：

1. `config.Mode = "standbylist"`（汇编 `[0x8]=ptr / [0x10]=0xb`）。
2. `s.execute = executeMemoryReleaseMode`（汇编 `[0xa0]` funcval → fn 0x1408d5120）。
3. `s.now = time.Now`（汇编 `[0xb8]` funcval → fn 0x140104fc0=time.Now）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.944s`。
- **G2 count_funcs**：`FUNCS=2712 / S=1158 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`。
- **G3 行为**：5 个 SYSTEM_MEMORY_LIST_COMMAND 命令值经逐字节 cmp/memequal 确证；3 个权限错误文案 + 2 个缺参文案 + 4 个 `%w` 包装文案全部 `read_gostring.py` 字节级解码；STATUS_PRIVILEGE_NOT_HELD 经 itab ifaceeq 双参数（0x1411d2e00/0x1411cd478）确证。
- **G4 review**：新增 memoryrelease_exec.go 单文件 + memoryrelease.go 工厂函数 3 行装配补齐；无签名变更、无并行写重叠。
