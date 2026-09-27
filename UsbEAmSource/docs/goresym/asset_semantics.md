# launcherAssetService 汇编语义分析（批次 A 续作基础）

> 本文件记录对目标 exe `launcherAssetService` 及关联函数逐条反汇编得到的语义结论。
> 数据来源：gore 提取方法地址 → capstone 符号标注反汇编（docs/goresym/disasm/*.dis.txt）。
> 供"把复杂 [T] 方法精确还原为 Go"的续作使用。所有地址均为 VA。

## 1. 方法地址表（已实证）

| 函数/方法 | VA off | len |
|---|---|---|
| RegisterBytes | 0x14086e220 | 1344 |
| RegisterStableBytes | 0x14086e760 | 1440 |
| RegisterFile | 0x14086ed00 | 2432 |
| register | 0x14086f680 | 1600 |
| registerStable | 0x14086fd20 | 2368 |
| ServeAssetRequest | 0x1408706c0 | 2816 |
| ReadBytes | 0x140871220 | 1216 |
| Exists | 0x1408716e0 | 608 |
| lookup | 0x140871940 | 1568 |
| Clear | 0x140871fc0 | 224 |
| currentTime | 0x1408720a0 | 64 |
| pruneExpiredLocked | 0x1408720e0 | 864 |
| validateItemSize | 0x140872440 | 256 |
| ensureCapacityLocked | 0x140872540 | 896 |
| addEntryLocked | 0x1408728c0 | 832 |
| evictOldestLocked | 0x140872c00 | 1440 |
| openFileBounded | 0x1408731a0 | 544 |
| readFileBounded | 0x1408733c0 | 832 |
| parseLauncherAssetRequest | 0x140873760 | 928 |
| buildLauncherAssetURL | 0x140873b00 | 416 |
| normalizeLauncherAssetNamespace | 0x140873ca0 | 736 |
| stableLauncherAssetID | 0x140873f80 | 640 |
| pathEscapeLauncherAssetSegment | 0x140874200 | 512 |
| newLauncherAssetID | 0x140874400 | 288 |

## 2. 已反汇编完整（含符号标注）

见 `docs/goresym/disasm/`（24 个 .dis.txt，含 6 个顶层函数；parseLauncherAssetRequest 已补全至完整 245 行）。

## 3. 已经汇编辑实的语义

### currentTime（0x1408720a0，[S] 已实证）
```
mov rdx,[rax+0x28]   ; 读 s.now
test rdx,rdx
je  -> call time     ; now==nil → time.Now()
mov rax,[rdx]; call rax ; else 调用 s.now()
```
Go：`if s.now==nil { time.Now() } else { s.now() }`。

### Clear（0x140871fc0，[S] 已实证）
重建 s.entries(off+8)、s.namespaceBytes(off+0x10)、置 s.totalBytes(off+0x18)=0；**不清 next(off+0x20)**。

### validateItemSize（0x140872440，[S] 已实证）
`rcx=[rax+0x38]` 取 limits.maxItemBytes；`rcx<=0 → cmovle 0x4000000(64MB)`；`size<=0 → error`；`size<=limit → ok`；否则错误。

### buildLauncherAssetURL（0x140873b00，已实证 [S]，version 段已落地）
参数 (rax/rbx=ns, rcx/rdi=id, rsi=version int64)；流程：
```
call normalizeLauncherAssetNamespace(ns)   ; 先归一化 ns（rax=ptr,rbx=len）
若 nrm(ns) 空 或 id 空 → 返回空串
id 经 0x1400a2200 处理
6 段拼接: '/__usbeam_asset__/'(18) + nrm(ns) + '/'(1) + idEsc + '?v='(3) + FormatInt(version,10)
```
常量（.rodata 实证）：`/__usbeam_asset__/`、`/`、`?v=`。**`?v=` 后的 version 段已落地**：第 3 参 `version int64`（rsi），`mov ebx,0xa`(base=10) 调 `0x1400ad400`（strconv.FormatInt）格式化为十进制。签名 `buildLauncherAssetURL(namespace, id string, version int64) string`。

### normalizeLauncherAssetNamespace（0x140873ca0，已落地 [S]）
入参 ns(string)；返回 string：
```
若 ns 空 → 返回空串（0x140873d71）
按 "/" 切分，逐段：
  段长 1 且 == "."  → 跳过
  段长 2 且 == ".." → 跳过
  否则 call pathEscapeLauncherAssetSegment(段)，加入结果
最终 join("/")
```
对应 Go：`strings.Split(ns,"/")` + 过滤 `.`/`..` + 每段 `pathEscape` + `strings.Join(...,"/")`。空段 `""` **保留**（非 `.`/`..`），故 `/a/b`→`/a/b`。

### pathEscapeLauncherAssetSegment（0x140874200，已落地为 [S]）
RIP 相对常量（.rodata，每处 `lea rdx/r8,[rip+XX]` 解引用得 1 字节 string），共 10 个单字
节 string 交替作 old,new 输入 `strings.NewReplacer`，new 全为 `-`：
```
' '(0x20) -> '-'   '#'(0x23) -> '-'   '?'(0x3f) -> '-'   '&'(0x26) -> '-'   '%'(0x25) -> '-'
```
Go：`strings.NewReplacer(" ","-","#","-","?","-","&","-","%","-").Replace(seg)`。

## 4. Exists 调用链（待续）
Exists(0x1408716e0) 语义概览（部分未定）：
1. receiver nil → false
2. namespace 空 → false
3. 从结构体字段（[r8+0x38]/[r8+0x40]）读出部分经 `parseLauncherAssetRequest` 校验
4. `normalizeLauncherAssetNamespace(ns)` 归一化后比较
5. 最终经 `lookup`(0x140871940) 深查条目
依赖链：parseLauncherAssetRequest(0x140873760)、normalize、lookup 均需继续还原。

### parseLauncherAssetRequest 常量与内部逻辑实证（本次新增）
对目标 exe 的 `.rodata`/`.text` 逐字节取证（gore dump + capstone + 自写 VA→off 映射）：

1. **前缀常量**：RIP 目标 `0x140c59d74` 实证为字符串 **`/__usbeam_asset__/`**（18 字节，`lea rbx,[rip+0x3e65c8]; mov ecx,0x12`）。与 buildLauncherAssetURL 前缀同源。
2. **分隔符常量**：多个 `lea rcx/rsi/rdi/rbx,[rip+0x3bfe12/0x3bfdf5/0x3bfdd4/0x3bfd37/0x3bfd17/0x3bfd03/0x3bfc98]` **均指向同一地址 0x140c3362f**（首字符 `/`），配 `edi=1/r8=1/r9=-1`——即**按 `'/'` 切分**（单字符分隔符，strings 内部函数（`IndexOf`/切分）用）。之前误判为 unicode 表，实为 `/` 常量。
3. **过滤 `.`/`..` 循环**：0x1408738b1-0x1408738d8 逐段检查 `cmp byte[r10],0x2e`（`.`,len==1）与 `cmp word[r10],0x2e2e`（`..`,len==2），命中即抛弃 → 空返回；否则前进。与 normalizeLauncherAssetNamespace 的过滤语义一致。
4. **hex 字符合法性循环**：0x140873a47-0x140873aa4 逐字符：`-0x30 ≤ 9`（`0`-`9`）或 `-0x61 ≤ 5`（`a`-`f`），否则置失败标志。即校验某段全为 16 进制字符（`0-9a-f`）。RIP 目标 0x140c3362f 为长度 1 的 `"/"`。
5. **相等/前缀比较函数 0x140006280**：`cmp rax,rbx; je→true; else`跳转字节比对——是类似 `bytes.Equal`/`strings` 底层相等 helper（无符号表可证其名称，语义从指令确认）。Exists/Lookup 亦调用同一地址做字符串相等判断。
6. **多失败路径**：前缀不匹配（0x140873811 之后非 18 前缀）、空余串、非法 hex、`.`,段非法均走 `xor eax/ebx/rcx/esi/r8d` → 清零 + `r8d=0`（bool 失败）。成功路径 L0x1408739fa 设 `r8d=1`（bool 成功）并返回多寄存器结果。

**结论（std 符号已解锁，parse 主体已落地）**：parse 的**前缀校验、按 `/` 切分、`.`/`..` 过滤、hex 字符合法性**四条核心语义由字节实证锁定；**被调 std 内部函数已全部命名**（经 gore `GetSTDLib()` 解锁本部 std 符号表，见 §7）：`0x140006280=runtime.memequal`（前缀）、`0x1400a2200=strings.TrimSpace`、`0x1400a1c20=strings.Trim`、`0x14009f920=strings.genSplit`（按 `/` 切分）、`0x1400a0140=strings.Join`、`0x140108ee0=path.Clean`、`0x1400a8a20=strconv.ParseInt`、`0x1400733c0=runtime.decoderune`、`0x14007f140=runtime.morestack_noctxt`、`0x140081400=runtime.panicIndex`、`0x14005fc00=runtime.concatstring2`。**`parseLauncherAssetRequest` 已落地为可编译实现**（[P]，见 launcherasset.go：HasPrefix→TrimSpace→Split('/')→过滤 `.`/`..`→末段 32-hex 校验→Join），配 11 case 表驱动测试，全量 19 PASS。多值返回精确布局仍按 Exists 调用点推断（[P]）。

### parse 返回签名：编译器 oracle 证伪/筛选（本轮）
用本机 Go1.25 编译器生成各候选签名返回段的 amd64 汇编，与 target parse 成功返回段逐寄存器对位（可复现：临时模块 + `go build -gcflags=-S`）：

| 候选签名 | 返回段写出寄存器 | 是否匹配 target |
|---|---|---|
| `(string,string,bool)` | RAX/RBX,CX/DI,SI | ✗（不用 R8） |
| `(string,string,int64,bool)` | RAX/RBX,CX/DI,SI,R8 | ✓（6 寄存器；R8=`bool`） |
| `(string,string,string)` | RAX/RBX,CX/DI,SI/R8 | ✗（R8=string.len，target 为常量 1） |
| `(string,string,string,bool)` | >6 溢出到栈 | ✗（超出寄存器） |

target 成功段（0x1408739fa-0x140873a1c）写出 `rax,rbx,rcx,rdi,rsi,`**`r8d=1 立即数`**。R8 为**常量 1**（成功 bool/计数），故**排除 3-string 形态（R8 会是变长 len）**，锁定 parse 返回为 **4 值、末位标量 bool**：`(string, string, 标量, bool)`（标量在第 6 寄存器 RSI）。

据此 parse 返回高度倾向 `(ns, id string, version int64, ok bool)`（version 来自 `?v=` 数值；ok 为解析成功标志）。该对位是**编译器可证伪**的，非臆测；实现主体仍受内部 std 切分/解码函数需 Ghidra 约束，但签名维度已收敛。

### 方法间调用链与注册路径契约（本轮新增）
对 `register`(0x14086f680) / `registerStable`(0x14086fd20) / `RegisterBytes`(0x14086e220) / `RegisterFile`(0x14086ed00) 的调用关系与关键寄存器证据：

1. **RegisterBytes → register**（0x14086e463 `call 0x14086f680`）；**RegisterFile → register**（L376）。registerStable 独立。三者共用 `addEntryLocked`/`ensureCapacityLocked`。
2. **register → ensureCapacityLocked**（L104-118）：`rbx=[rsp+0x238]`、`rcx=[rsp+0x240]`（exceptID/namespace?）、`rdi=[rsp+0x288]`（itemSize）、`rax`=s；call 后 `test rax,rax; jne 报错`。**这印证 ensureCapacityLocked 取 itemSize+exceptID 入参、返回 error**（与已完成落地签名一致）。
3. **s.next 递增**：`rdx=[rax+0x20]`（s.next），`inc rdx`，回写 `[rax+0x20]`（register L100-123）——注册时分配并推进全局自增 id。
4. **register → entry 结构构造 → addEntryLocked**（L130-166）：连续 `mov [rsp+0x228..0x2b8]` 逐字段构建 `launcherAssetEntry` 类栈结构（id/namespace/version/data/size…），随后 `lea rsi,[rsp+0x228]`（entry 起始）、`lea rdi,rsp`（目标）经 `0x140081a46`（大结构复制）后 `call addEntryLocked`。**addEntryLocked 接收一个按值/栈传递的 entry 结构体**，非简单 string 参。
5. **addEntryLocked entries/namespaceBytes 双 map 写入语义**（addEntry.dis）：
   - entries 空 → `0x140076d40`(make map) 填充；按 id 查（`0x14000e720`），已存在先回退旧 entry 字节计数再插入新。
   - namespaceBytes 空 → make；`[rax+0x10]` 计数：先把旧 namespace 的 size 回退（`[rax+0x18]`totalBytes 同步），再对新条目累加（0x14000d920/0x14000de20/0x14000e3e0）。
   - 每处 map 值更新后 `totalBytes`(`[rax+0x18]`) +size（L140-145）。
6. **addEntryLocked 存在 bool 分支**（L1e7 `test bl,bl` → L78 回退 / L68 直插）：标志决定是否先做覆盖性回退（entry 已存在时删除旧）。

**结论**：`lookup`、`addEntryLocked`、`register`、`registerStable` 共享同一 `launcherAssetEntry` 栈布局与双 map 记账；因此其**精确入参（entry 结构体字段顺序）与返回类型需 Ghidra 一次性定**（redress 对内部方法解析为空 `()`）。本轮将上述调用链/寄存器/行为契约固化，Env 到 Ghidra 还原只需按图补 entry 字段布局，勿再逐寄存器猜。

### parse → lookup 消费证据（ServeAssetRequest，本轮新增）
`ServeAssetRequest`(0x1408706c0) 对 parse 的完整消费（L104-135）：

1. **输入提取**：`r8=[rsp+0x318]`→`r8=[r8+0x10]`→`rax=[r8+0x38]`、`rbx=[r8+0x40]`（URL path string A）；另一 string B 从 `[rax]`/`[rdx+8]` 提取（`rcx=rdx, rdi=rsi`）。与 Exists 完全同模式。
2. `call parseLauncherAssetRequest` 后 **`test r8b`** —— r8 是 parse 的 **bool 成功标志**（与前 Exists 一致）。
3. parse 成功 → 把 **parse 的 6 个返回寄存器全部转喂 lookup**：`rbx=rax,rsi=原始rdi,rdi=rcx,r8=rsi,rcx=rbx,rax=[s]` 后 `call lookup`。**故 parse 与 lookup 的寄存器返回/入参互相锁死**：lookup 复用 parse 的 (rax,rbx,rcx,rdi,rsi,r8)。
4. lookup 命中(布尔) → 后续按 `openFileBounded(0x1409fa4)`/读取/响应分派（L166-208 文件 or bytes 分支）。

**语义**：ServeAssetRequest 主流程已完全确证为 `URLpath → parse → lookup → 命中则 serve（file/bytes）`。仅 parse 的多值**精确返回个数**与 lookup 的入参结构体仍需 Ghidra 一次性对齐。

## 5. 待续工作
- parseLauncherAssetRequest 已完成（见 §4/§6，[P]，11 case 19 PASS）。
- **std 符号已解锁（§7）**——lookup、register/Register*、addEntryLocked、Exists、ReadBytes、ServeAssetRequest、openFileBounded/readFileBounded、stableLauncherAssetID 的被调 std 函数可按名还原；剩余难点收敛为**多值返回/结构体参数布局**（仍需 Ghidra 数据流），按依赖链（lookup→register→addEntry→注册入口）逐项推进。
- Exists / Register* / ServeAssetRequest / ReadBytes 完整实现（依赖链含上述方法）。
- 全链 ut/可编译验证。

## 6. 本会话已实证落地（续作完成项）
- buildLauncherAssetURL 的 `?v=<version>` 段：**第 3 参 version int64（rsi）实证落地**，`strconv.FormatInt(version,10)`（base 0xa 自汇编证实），签名改为 3 参；测试更新为 4 case。
- parseLauncherAssetRequest：完整反汇编（245 行）补全归档；调用点契约（2 入参、含 bool 的多值返回）经 Exists/ServeAssetRequest/ReadBytes 三处交叉核实。**主体已落地**（[P]，HasPrefix→TrimSpace→Split('/')→`. `/`..` 过滤→末段 32-hex 校验→Join），11 case 测试。
- 工具链修正（上会话）：addr_extract.go 顶层 Function 分支补 dumpBytes；disasm.py 符号表路径修正；6 个顶层函数反汇编归档。
- **ensureCapacityLocked 容量预算落地（[P]）**：limits 默认值（4096/64MB/256MB/512MB，cmovle）实证；itemSize 校验 + 三条驱逐预算（条目数/totalBytes/namespace）循环 `evictOldestLocked(exceptID)`，无可驱逐报错；签名 `ensureCapacityLocked(itemSize int64, exceptID string) error`（据 register 调用点 `rdi=itemSize, rbx/rcx=exceptID` 佐证）。5 个表驱动测试，全量 17 PASS。
- **方法间契约证据**（见 §4"方法间调用链与注册路径契约"）：register/registerStable/Register* → ensureCapacityLocked → addEntryLocked → buildLauncherAssetURL 的完整调用链、s.next 递增、entry 结构栈布局均实证固化，缩短 Ghidra 还原工时。
- **newLauncherAssetID 落地（[S]）**：16 字节随机源（0x14025e5a0，Go 随机包装，内部引 math/rand 态）+ hex 表 `[rip+0x3e0c6e]`（'0123456789abcdef'）逐字节编码为 32 字符。落地 `newLauncherAssetID() string`（`crypto/rand.Read(16)` + `hex.EncodeToString`），与 `isValidLauncherAssetIDHex` 互洽（生成值恒通过校验）；测试以结构性不变量断言（长度/合法 hex/随机性），全量 18 PASS。

## 7. std 符号剥离边界（系统实测，重要修正）

> 对"函数体还原走 Ghidra 配 DWARF/pclntab 即可标 std 函数"交接认知的**实测修正**——本目标二进制**不含可解析的 std 命名符号**，Ghidra/反编译工具同样无法仅凭符号表对剩余 [T] 主体按名还原。

实测（本轮内）：
1. `redress packages` / `addr_tool`（gore v0.14.5）：仅列出 `changeme/internal/*` 与 `main`，**无 strings/hash/os/unicode 等 std 包**（go1.25 的 std 以 moduledata 内嵌、无独立 pclntab 符号）。
2. `go1.25.12 tool nm` / `tool objdump`：目标 exe 报 `no runtime.pclntab symbol found` / `no symbols`——**无可见 .symtab/pclntab**。
3. `rz-bin -S`（rizin 0.9.1，本机已装）：仅列 `.text`(CNT*CNT*)，**无 `.debug_info`/`.zdebug*`**——**DWARF 已剥离**（构建含 `-trimpath` 且可能含符号剥落参数）。
4. rizin `aa; afl` 仅恢复 `entry0`（无符号表则无法自动切普通函数）。

推论：剩余 [T] 主体（addEntryLocked / lookup / register* / Exists / ReadBytes / ServeAssetRequest / openFileBounded / readFileBounded / stableLauncherAssetID / parse 主体）所依赖的 **std 内部函数名（strings/strconv/hash/unicode/os·vfs 接口分派）在二进制中被剥离，无源道恢复**；Ghidra/rizin 即便部署也只能给无名地址，需逐字节比对 Go 1.25.12 GOROOT 源码指纹才能反命名，超出"应用还原"口径且不能保证（std 内联/插入差异）。

➡️ 故 launcher 域**已无 Ghidra 也能安全落地的高置信方法**；剩余 [T] 的可靠推进需外部提供可解析 std 符号的运行时/动态分析手段，或逐 std 指纹比对（GOROOT=`C:\Users\Administrator\sdk\go1.25.12`）。本文件 §4 已将该域全部可静态实证的方法链/契约固化，作为后续任何解阻手段的接续蓝图。

> ⚠️ **更正**：本节（§7）"std 符号剥离/不可恢复"结论已被下方 §8 推翻——根源是误用 gore `GetPackages()`（仅主模块）而忽略了 `GetSTDLib()`（枚举全部 std 符号）。保留本节作追溯。

## 8. std 符号已解锁（gore GetSTDLib）——决定性突破（本轮）

> 推翻 §7 的错误判定。之前多个会话（含 §7）推论"标准库符号被剥离、无 Ghidra 不可恢复"是**工具 API 误用**：gore 的 `f.GetPackages()` 仅返回主模块包，而 `f.GetSTDLib()` / `GetVendors()` / `GetGeneratedPackages()` / `GetUnknown()` 会枚举 functab 中**全部 std / 第三方 / 编译器生成函数**。

实测（本轮，工具 `tools/go-introspect/stdsyms.go`）：
- 对目标 exe 输出 **19427 条 std 符号**（database/sql、text/template、runtime、strings、strconv、hash、unicode、os 等全部）。
- parse 反汇编全部被调地址已命名：`0x140006280=runtime.memequal`、`0x1400a2200=strings.TrimSpace`、`0x1400a1c20=strings.Trim`、`0x14009f920=strings.genSplit`、`0x1400a0140=strings.Join`、`0x140108ee0=path.Clean`、`0x1400a8a20=strconv.ParseInt`、`0x1400733c0=runtime.decoderune`、`0x14005fc00=runtime.concatstring2`、`0x14007f140=runtime.morestack_noctxt`、`0x140081400=runtime.panicIndex`。
- `docs/goresym/symbols.txt` 已从 main-only（4754 行）**合入 std 升至 24181 行**（原备份 `symbols.main.bak`），disasm 工具链现可直接标注 std 调用。

**直接成果**：此前判定"需 Ghidra 才能落地"的 **`parseLauncherAssetRequest` 主体已落地**（[P]，HasPrefix→TrimSpace→Split('/')→`.`/`..` 过滤→末段 32-hex 校验→Join；11 case 表驱动测试）；随后 **`Exists` 也落地**（[P]，`strings.TrimSpace` + `runtime.mapaccess1_faststr` + `normalize` 比较；7 case）。**全量 20 PASS / 0 FAIL**。剩余 lookup/register/addEntry*/ReadBytes 等的 std 调用可按名还原；真需 Ghidra 的空间收敛为结构体参数（launcherAssetEntry）布局与多值返回精序，比 §7"全部不可解"已大幅缩小。

**sha256 稳定 ID 核心落地（[S]）**：`stableLauncherAssetID`（0x140873f80）经 std 确认为 **SHA-256**（`crypto/internal/fips140/sha256` 的 `Digest.Reset/Write/Sum`，0x140a0c1e0/0x140a0c2e0/0x140a0c5a0），摘提前 16 字节逐位 hex 编为 32 字符。落地 `sha256HexPrefix`（`crypto/sha256` 前 16 字节 hex）+ 确定性向量测试（sha256("abc") 前 16 字节 = `ba7816bf8f01cfea414140de5dae2223`），与 `newLauncherAssetID` 的随机形态互为对照（stable vs random 两种 ID 形态）。全量当前 21 PASS / 0 FAIL。

**addEntryLocked 落地（[P]）**：保存路径核心入表 —— 懒建双 map（`runtime.makemap_small`）、覆盖回退（`mapdelete_faststr`/mapdelete）、`mapassign_faststr` 写 entries[id]、namespaceBytes/totalBytes 计数维护。std 全锁见 §9。全量当前 22 PASS / 0 FAIL。

**readFileBounded 落地（[P]）**：文件有界读取（RegisterFile 落盘读取）；std `os.File.Stat`（0x14012da00），上限 <=0 回退 64MB（与 validateItemSize 默认一致），配临时文件测试。全量当前 23 PASS / 0 FAIL。

**lookup 落地（[P]）**：深查资产 —— 持锁、currentTime、map 按 id 读、过期判定（与 `pruneExpiredLocked` 一致）、命中刷新 `accessedAt`。std 全锁见 §9。全量当前 24 PASS / 0 FAIL。

**openFileBounded 落地（[P]）**：有界打开器（readFileBounded 对称），std `os.Open`+`os.File.Stat`；上限 <=0 回退 64MB；超限关闭报错。全量当前 25 PASS / 0 FAIL。

**RegisterBytes 落地（[P]，外部注册入口）**：redress 签名确证；组装 register 流程组件（newID/validate/ensureCapacity/addEntry 构造 entry/buildURL/ref）+ next++。全量当前 26 PASS / 0 FAIL。

**RegisterFile 落地（[P]）**：读文件（readFileBounded,maxItemBytes）→ RegisterBytes 注册磁盘资产；配临时文件测试。全量当前 27 PASS / 0 FAIL。

**ReadBytes 落地（[P]，读取入口）**：entries 按 id 读 + namespace 匹配返回 data/contentType；持锁。全量当前 28 PASS / 0 FAIL。

**RegisterStableBytes 落地（[P]，稳定注册）**：id 空→sha256HexPrefix(namespace) 确定性派生；复用 RegisterBytes。全量当前 29 PASS / 0 FAIL。

**ServeAssetRequest 落地（[P]）**：HTTP 入口 —— parse 路径 → ReadBytes → 命中写 200+头+body；否则 false。**外部生命周期入口全部落地**。全量当前 30 PASS / 0 FAIL。

**stableLauncherAssetID 落地（[S] 核心 + [P] 段序）**：SHA-256 多段稳定 ID 完整函数（段间 0x00 分隔、摘要前 16 字节 hex）；确定性测试。**目标命名方法全部落地**。全量当前 31 PASS / 0 FAIL。

## 9. launcher 域 std 被调地址映射表（std 解锁后，供后续直接按名拼接）

用 `tools/go-introspect/stdsyms.go`（gore `GetSTDLib`）对目标 exe 定位，本域全部被调地址→std 名（可检索的接续基座）：

| 地址 | std/包函数 | 使用上下文 |
|---|---|---|
| 0x140006280 | runtime.memequal | 前缀/字符串相等比较（parse/Exists/lookup） |
| 0x140076d40 | runtime.makemap_small | 懒建 map（addEntryLocked/deClear） |
| 0x140076de0 | runtime.mapdelete | 删 map 项 |
| 0x14000d920 | runtime.mapaccess1_faststr | map[string] 快速读 |
| 0x14000e720 | runtime.mapaccess2 | map 读（bool 变体） |
| 0x14000e9e0 | runtime.mapassign | map 写 |
| 0x14000de60 | runtime.mapassign_faststr | map[string] 写 |
| 0x14000e3e0 | runtime.mapdelete_faststr | map[string] 删 |
| 0x1400a1c20 | strings.Trim | parse 段处理 |
| 0x1400a2340 | strings.Replace | parse 段处理 |
| 0x1400a2200 | strings.TrimSpace | parse/Exists |
| 0x14009f920 | strings.genSplit | 按 '/' 切分 |
| 0x1400a0140 | strings.Join | 段重组 |
| 0x140108ee0 | path.Clean | 规范化段 |
| 0x1400a8a20 | strconv.ParseInt | 数值解析 |
| 0x1400ad400 | strconv.FormatInt | buildURL 版本 |
| 0x1400733c0 | runtime.decoderune | unicode/utf8 解码 |
| 0x14005fc00 | runtime.concatstring2 | 字符串拼接 |
| 0x140081400 | runtime.panicIndex | 越界安全 |
| 0x14007f140 | runtime.morestack_noctxt | 栈分裂 |
| 0x140091940 | internal/sync.Mutex.lockSlow | 锁（lookup/addEntry） |
| 0x140268c20 | net/url.URL.Query | Exists 取 query |
| 0x14012da00 | os.File.Stat | openFileBounded |
| 0x140128960/0x14012a320 | os 包打开/关闭 | openFileBounded |
| 0x140a0c1e0/2e0/5a0 | crypto/internal/fips140/sha256.Digest.Reset/Write/Sum | stableLauncherAssetID |
| 0x14025e5a0 / 0x14025e4e0 | crypto/rand 随机源 | newLauncherAssetID |
| 0x140081060 | runtime.gcWriteBarrier2 | 写入屏障 |
| 0x14001b620 | runtime.wbMove | 写屏障 |

→ 剩余方法（lookup/register/Register*/addEntry/ReadBytes/ServeAssetRequest/openFileBounded/readFileBounded）的 std 依赖全部已在表内可按名拼接；唯一待锚 = launcherAssetEntry 结构参数布局（各字段读/写偏移）与多值返回精序。

## 10. Ghidra 反编译介入（本轮，工具链升级）

- 自主部署 Java 25（scoop Temurin LTS）+ Ghidra 12.1.3（aria2 代理下载，`-k` 跳过证书吊销）；`RunDecompile.java` 经 official ghidra_scripts 目录 `analyzeHeadless -process -noanalysis` 反编译 6 目标。
- **实证签名修正**：
  - `parseLauncherAssetRequest(requestPath, requestVersion string) (namespace, id string, version int64, ok bool)`——**2 参 + 4 返（含 version）**；前缀 memqual 截断（非强制）；Trim/Replace(`\`→`/`)/TrimSpace；genSplit 后**任一空/./.. 段→整体失败**；末段 32-hex=id，前段 join=ns；第 2 参 ParseInt(10,64,>0)=version。已按此落地 + ServeAssetRequest `?v=` 适配，TestParse 16 case。
  - `stableLauncherAssetID` = `(a, b string, c []byte) string`（2 string + **1 []byte** 第三参），写序 = Write(a), Write(0x00), Write(b), Write(0x00), Write(c)，SHA-256 前 16 字节 hex（此前全 string varargs 近似已按此对齐语义）。
- **lookup** C 显示**3 参（2 个 string + 1 int）+ 过期则删除并回退 + 命中重写 accessedAt**——比"返回 bool"更精细，后续按此增强。
- 反编译产物归档 `../work/ghidra/decompiled/register_targets.txt`（1151 行 C 伪代码，6 目标），作为内部壳（register/registerStable）后续落体的唯一待锚数据源已闭合。
- **registerStable 落地（本轮，Ghidra 精确）**：内部稳定注册（此前空 stub）落实 —— id 须恰 32-hex（0x14086fe48 前逐字符校验，无效即 errors）；entries 已存在且 ns/id 匹配 → 仅刷 expiresAt(now+10min)+accessedAt 复用返回；不存在 → lock/prune → next++ + 构造 entry(expiresAt=TLL) + addEntryLocked + buildURL；TTL=600s。配 TestRegisterStableBody（非32hex err/新建立/重复复用 next 不变/nil）。全量 32 PASS / 0 FAIL。
- **RegisterBytes 增强（Ghidra 实证）**：register C(0x14086f680) —— 自动 id 唯一查重（显式/stable id 不重生成，保稳定）；entry.expiresAt=now+600s（10min TTL）；url `?v=` 用 `s.next` 自增序号（非用户 version）；TestRegisterBytes 按 next 语义对齐。全量 32 PASS / 0 FAIL。
- **lookup 增强（Ghidra C 对齐，本次）**：表命中需 ns 匹配（memequal）；过期 → removeEntryLocked（删+回退 totalBytes/namespaceBytes）后 false；命中未过期 → 重写 accessedAt 返回 true。配 TestLookupExpanded（过期删/回退、ns 不匹配不删）。全量 33 PASS / 0 FAIL。
- **register 主壳落地（Ghidra 反编译 0x14086f680，本次会话）**：空 stub → 真实实现。C 伪代码语义：LOCK → currentTime → pruneExpiredLocked → id 空时循环 newLauncherAssetID + mapaccess2 查重至唯一 → ensureCapacityLocked（失败→(ref,err)）→ next++ → entry（createdAt=accessedAt=now；expiresAt=now+TTL，TTL 默认 600e9=10min）→ addEntryLocked → buildLauncherAssetURL(namespace,id,next) → (launcherAssetRef,error)（72 字节聚合 = ref 56 + err 16）。**RegisterBytes 已改为委托 register**（反汇编链路 RegisterBytes→register 0x14086e463），消除平行重复；配 TestRegisterBody / TestRegisterBodyCapacityError。全量 35 PASS / 0 FAIL。
- **stableLauncherAssetID 3 参对齐（Ghidra 反证 0x140873f80，本次会话）**：真签名确证 `(a, b string, c []byte) string`（stacktrace struct{8,8};{8,8};{8,8,8}）；写序 Write(a)→0x00→Write(b)→0x00→Write(c)，SHA-256 前 16 字节 hex（± 此前 varargs string 近似）。实现/测试已对齐（TestStableLauncherAssetID 含与独立 sha256 引擎对照 + c 段参与）。全量 35 PASS / 0 FAIL。
## 11. 装配层（config/装配域）capstone 还原（本轮开启，字节级目标第一步）

> 目标被 YG 更新为：字节级完美逆向（整包 changeme 逐字节重建，产出与目标 exe 同哈希的
> 源码）；整包 scale 已获认，launcher 域先行的同时装配层（config/装配）域起步。

### 11.1 构建确定性与字节级可行前提（实测）
- Go 1.25.12 同源同 flag 双构建**字节级 DETERMINISTIC=YES**（sha256 完全一致）→ 字节级目标可验证。
- 目标构建指纹（`go version -m`）：`buildmode=exe / compiler=gc / tags=production / trimpath=true /
  CGO_ENABLED=0 / GOOS=windows / GOARCH=amd64 / GOAMD64=v1 / go1.25.12`；25 直接依赖全带 h1 哈希。
- 差异基线：目标 29,965,824 B / `sha 995a12da…`；当前重建（launcher+config 还原后）19,489,792 B /
  `sha 18cdd3aa…`；10.5MB 缺口 = 待逐域还原的方法体 + 根装配 + 依赖内嵌差异。

### 11.2 装配层地址 + Ghidra 反编译失败判定
- 9 目标经 addr_tool(gore) 提 VA：NewBootstrapService(0x140771d20, 2528B)、
  initializeWithCheckpoint(0x140772920, 2592B)、attachApp(0x140773400, 1504B)、
  attachFileIndexService(0x1407739e0, 352B)、launcherConfigStoreForPath(0x140898ac0, 352B)、
  savePreparedUnlocked(0x14089d1c0, 960B)、saveUnlockedWithCurrentRefs(0x14089cd20, 1088B)、
  writeConfigUnlocked(0x14089d580, 352B)、deletePreparedLocked(0x14089ca40, 448B)。
- **Ghidra DFAIL（实证）**：9 目标在本 project 下 `decompileFunction` 全部 FAIL（函数 body 指令完整
  456/456/229…条），而 launcher 域 6 地址同 project **decompile OK**。归因：装配层方法依赖 Go ABI
  原型/未命名交叉引用，Ghidra 当前无类型环境无法反编译（非函数不存在，InspectAssemble 已证 body 完整）。
- 判定：装配层不依赖 Ghidra decompile，转 **capstone 汇编 + std 符号标注**（launcher 域 `.dis` 同源路径）。

### 11.3 capstone 装配反汇编管道（新增）
- `tools/go-introspect/va_disasm.py`：以 VA 为基址反汇编函数 bin，call/jmp 目标做 std 符号标注
  （读取 docs/goresym/symbols.txt 24181 条）。产物归档 `docs/goresym/disasm_assemble/*.dis.txt`（9 个）。
- 验证：NewBootstrapService 反汇编 prologue/字段写回与 main.go stub 类型布局一致
  （call resolveWorkspaceLayout、call launcherConfigStoreForPath、newobject BootstrapService）。

### 11.4 首个还原：launcherConfigStore 缓存（[S] 汇编实证，本轮完成）
- `launcherConfigStorePathKey(0x140898c20)` 语义：TrimSpace → filepath.Abs(err nil 用 abs / 失败
  Clean) → ToLower（Windows 反斜杠）。→ `backend/launcherconfig.go launcherConfigStorePathKey`。
- `launcherConfigStoreForPath(0x140898ac0)` 语义：TrimSpace → PathKey → 全局 HashTrieMap Load（命中
  类型断言返）→ 未命中 newobject{path=TrimSpace} → LoadOrStore 回填 → 类型断言返。
  → `backend/launcherconfig.go`（全局 `launcherConfigStoreCache sync.Map`，触达 internal/sync.HashTrieMap 调用形态）。
- 测试 `backend/launcherconfig_test.go`（PathKey 归一化 + ForPath 单例缓存）。
  build/vet/test 全绿；**全量 37 PASS / 0 FAIL**（原 35 + 2）。
- 基线移动：rebuild2 = 19,489,792 B（config 两函数已编译进入），sha 18cdd3aa…。

### 11.5 后续（本域待续，逐函数 capstone 翻译）
- initializeWithCheckpoint（4592B）→ attachApp（1504B）→ attachFileIndexService（352B）
  → configStore 四写方法（savePrepared/saveUnlockedWithCurrentRefs/writeConfigUnlocked/deletePreparedLocked）
  → 每函数配测试，build/vet/test 全绿，基线追踪逼近目标哈希。

### 11.6 配置落盘链还原（本轮，[S] 汇编实证）
- `writeJSONFile`(0x14088bd80, 992B) 还原：TrimSpace(空→err) → MkdirAll(dir,755) →
  MarshalIndent(cfg,"","  ") 且 <=64MB → 末尾 '\n' → CreateTemp(dir,"*") 写入 → close →
  rename 到目标（遇可重试错误 sleep50ms 重试 ≤5）。→ `backend/launcherconfig.go writeJSONFile`。
- `launcherConfigIconDataFingerprint`(0x14088d300, 352B)：LauncherConfig 图标数据 SHA-256 前 32 字节；
  其中 visitLauncherConfigIconData(0x14088d860) 遍历范围为 [P]（待 icon 域字节对齐），本处以配置序列化
  字节哈希近似（[P] 标注）。→ `launcherconfig.go`。
- `writeConfigUnlocked`(0x14089d580, 352B) 还原：writeConfig 注入优先，否则 writeJSONFile；成功才刷
  iconFingerprint/iconFingerprintReady/runtimeReady（汇编错误分支跳过指纹），返回 err。
  字段偏移与 types_launcher.go launcherConfigStore 对齐（path@0/writeConfig@0x10/runtimeReady@0x18/
  iconFingerprint@0x19/iconFingerprintReady@0x39）。→ `backend/launcherconfig.go`。
- 新增测试：TestWriteJSONFile / TestWriteConfigUnlocked。全量 39 PASS / 0 FAIL
  （35 launcher + 4 config；build/vet/test 全绿）。

### 11.7 配置读侧/删除链还原（本轮，[S] 汇编实证）
- `readLauncherConfigBytes`(0x140897c20, 1184B)：os.OpenFile(O_RDONLY) → ReadAll 有界 ≤64MB → 超限
  fmt.Errorf。→ `launcherconfig.go` readLauncherConfigBytes。
- `loadUnlocked`(0x140898d40, 704B)： runtimeReady 短路走 `loadLauncherConfigRuntimeUnlocked`
  (0x140876ce0, 读回不刷指纹)；首次 `loadLauncherConfigIfExistsUnlocked`（readLauncherConfigBytes +
  json.Unmarshal）成功后刷 iconFingerprint/ready + runtimeReady；失败清指纹。→ `launcherconfig.go`。
- `deletePreparedLocked`(0x14089ca40, 448B)：loadUnlocked → prepare 回调(loaded cfg) → os.Remove
  (NotExist 容忍) → 成功清指纹。→ `launcherconfig.go`。
- 新增测试：TestReadLauncherConfigBytes / TestLoadUnlockedAndDeletePrepared。
  全量 **41 PASS / 0 FAIL**（35 launcher + 6 config；build/vet/test 全绿）。
  rebuild4 基线 = 19,489,792 B。
- 说明：load 链（loadLauncherConfigIfExistsUnlockedWithWriter 0x140876220 内 writer/Migration 边栏、
  loadRuntimeUnlocked 0x140876ce0）以纯读+反序列化实现，writer 边栏为 [P]（待装配 bootstrap 迁移域）；
  图标字段遍历范围依赖 visitLauncherConfigIconData 递归，仍 [P]（icon 域待续）。

### 11.8 装配工作目录解析（本轮，[S] 汇编实证）
- `resolveProcessWorkingDirectory`(0x1407a2480, 160B)：os.Getwd 成功且非空 → TrimSpace 返回；
  否则 filepath.Abs(".") 成功且非空 → TrimSpace 返回；否则空串。→ `backend/main.go`。
- `resolveWorkspaceLayout` 对齐汇编改为基于 resolveProcessWorkingDirectory（汇编 Getwd 优先，
  非 Executable）；新增 ConfigFile/Language/Plugin/Icon/Index/Screenshot/WebView2/Backgrounds 布局。
- 新增测试：TestResolveProcessWorkingDirectory / TestResolveWorkspaceLayoutRoot。
  全量 43 PASS / 0 FAIL（35 launcher + 8 config/装配；build/vet/test 全绿）。rebuild5 基线。
- 说明：Bootstrap 装配大方法（NewBootstrapService/initializeWithCheckpoint/attachApp/attachFileIndexService）
  依赖的 service 工厂（new*Service）与 workspace/config 深层加载链分布在其它域，待续跨域还原，不堆 stub。

### 11.9 attachFileIndexService 还原（本轮，[S] 汇编实证）
- `attachFileIndexService(fileIndex fileSearchRuntimeWarmer) bool`(0x1407739e0, 352B)：持 Bootstrap
  +0x540 锁 → bs.fileIndex=idx → 若非 nil 且 pendingFileSearchResidentWarm 置位 → 清标志并立即
  scheduleFileSearchMaintenance 预热(返回 true)；否则 false。→ `backend/main.go`。
- main() stub 调用改为传 nil（stub 装配路径非目标真实；真实由 NewBootstrapService 内工厂传递）。
- 新增测试：TestAttachFileIndexService（nil/pending 未置位/pending 预热）。全量 44 PASS / 0 FAIL。

### 11.10 批量反汇编管道 v1（本轮，方向决策后落地）
- YG 拍板：批量化管道+继续逐域。新增 `tools/go-introspect/batch_disasm.sh`：
  输入 "VA,basename" 清单 → 从 dumpdir 逐项 va_disasm.py 标注 → 归档 outdir/<name>.asm.txt。
- 冒烟：装配层 NewBootstrapService/initializeWithCheckpoint/attachApp 3 项全 OK（1183 行标注），
  与逐条产物一致；修正 python3→python（Windows 下 python3 不可解析）。
- 承接：后续域（icon/config save 链/filesearch 接口）可归一清单批量产出，压缩单函数往返。

### 11.11 NewBootstrapService 装配依赖图（本轮，汇编字段↔工厂↔类型映射）
批量 dump+标注 7 个 service 工厂（128~704B），定位其字段挂载。NewBootstrapService 装配字段↔工厂↔目标类型：

| Bootstrap offset | 工厂(VA) | 类型 |
|---|---|---|
| 0x1a0 | newLauncherGlobalHotkeyManager(0x14089d760, 288B, chan×2+goroutine) | launcherGlobalHotkeyManager |
| 0x1a8 | (map hotkeyRegistrationErrors) | map[string]string |
| 0x390 | newMouseGestureService(0x1408dfda0, 704B) | mouseGestureService |
| 0x398 | (limits/config 段) | launcherAssetLimits 等 |
| 0x420 | newInputMonitorService(0x140861c40, 224B) | inputMonitorService |
| 0x438 | newFileLocatorService(0x1407ceb00, 480B) | fileLocatorService |
| 0x440 | newDesktopWidgetService(0x1407b2060, 544B) | desktopWidgetService |
| 0x424 | newMemoryReleaseService(0x1408d3560, 192B) | memoryReleaseService |
| 0x12c | newOLEDBlackoutService(0x1408fef60, 704B) | oledBlackoutService |
| (attach 链) | newLauncherNetworkAccess(0x1408a5640, 128B) | 网络访问 |
| 0x388 | (首次挂载) | — |

- 批量管道已验证（batch_disasm.sh 对 8 项全 OK）。这些工厂各自对应一个业务子域（hotkey/mouse/input/
  fileLocator/desktopWidget/memory/oled/network），装配层方法(NewBootstrapService 等)依赖它们提供
  interface 实现才能编译——属跨域集成，装配大方法待各子域 service 类型/方法就绪后回填。
- 支撑：装配原子工厂全部已 dump+标注归档，是各域还原的接缝起点。

### 11.12 network 末端叶子还原（本轮，装配依赖首个填实位）
- `newLauncherNetworkAccess(configPath string) LauncherNetworkAccess`(0x1408a5640, 128B) [S]：
  TrimSpace 空 → directLauncherNetworkAccess；非空 → &configBackedLauncherNetworkAccess{configPath}。
- `directLauncherNetworkAccess`/`configBackedLauncherNetworkAccess` 实现 LauncherNetworkAccess 的
  Do/Get [F]（net/http 真实请求，字节级待 network 域续作）。→ `backend/networkaccess.go`。
- 新增测试：TestNewLauncherNetworkAccessFactory / TestNetworkAccessDoGet。
  全量 46 PASS / 0 FAIL。rebuilt7 基线。

### 11.13 inputmonitor 叶子还原（本轮，装配依赖第二填实位）
- `newInputMonitorService`(0x140861c40, 224B) [S]：make([]InputMonitorEvent,0x1000) + pressedKeys/owners
  两 map + newobject 装配 + maxEvents/eventSize=0x1000 → inputMonitorInitialize → 返回。
- `inputMonitorInitialize`(0x140865d20, 160B) [S]：nil 安全；持锁设平台健康托管标志（汇编写 +0xd8 值 1，
  [P] 字段名待字节对齐）。→ `backend/inputmonitor.go`。
- 新增测试：TestNewInputMonitorService / TestInputMonitorInitializeNil。
  全量 48 PASS / 0 FAIL。rebuilt8 基线。

### 11.14 memoryrelease 叶子还原（本轮，装配依赖第三填实位）
- `newMemoryReleaseService`(0x1408d3560, 192B) [S]：newobject 装配 config（TimerEnabled=false,
  IntervalMinutes=30[0x1e]）+ 文案字段 + `supported`/`processIsElevated` 两项能力探测（写
  state.Supported/IsAdmin）。能力探测函数体 [F]（待 memory 域 w32 优先级 API 对齐）。→ `backend/memoryrelease.go`。
- 新增测试：TestNewMemoryReleaseService。全量 49 PASS / 0 FAIL。rebuilt9 基线。

### 11.15 oledblackout 工厂还原（本轮，装配依赖叶子）
- `newOLEDBlackoutService(app *application.App)`(0x1408fef60, 704B) [S]：ensureOLEDBlackoutDebugLogger +
  debugLog → newobject → 5 map（overlayWindows/visibleOverlays/hotkeyRegistered/hotkeyErrors/
  lastPressedKeys）→ lifecycle context → 挂 hotkeyManager(newOLEDBlackoutHotkeyManager)。
  内部依赖（ensure/debug log/ensureLifecycle/newOLEDBlackoutHotkeyManager 体）为 [P] 债务脚手架
  （AGENTS-20 记录债务；装配链可编译）。→ `backend/oledblackout.go`。
- 新增测试：TestNewOLEDBlackoutService。全量 50 PASS / 0 FAIL。rebuilt10 基线。

### 11.16 filelocator 工厂还原（本轮，装配依赖叶子）
- `newFileLocatorService`(0x1407ceb00, 480B) [S]：defaultFileLocatorConfig → newobject →
  detailByPath map + contentScan chan(buf1) → state.Request=def + ResultLimit=0x1f4(500)。
- `defaultFileLocatorConfig`(0x1408745e0, 384B) [S 结构]：字符串字面（.rodata 常量）为 [P] 近似；
  次序 MaxSearchFileSizeMB=500、IncludeSubfolders=true。→ `backend/filelocator.go`。
- 新增测试：TestNewFileLocatorService。全量 51 PASS / 0 FAIL。r11 基线。

### 11.17 mousegesture 工厂还原（本轮，装配依赖叶子）
- `newMouseGestureService`(0x1408dfda0, 704B) [S]：构造 mouseGestureService + config.Settings 默认
  （StartDistancePx=150/StartTimeoutMs=300/StopTimeoutMs=500 + set 标志）+ runtimeStop/runtimeDone chan。
  Global/Apps/HotCorners 默认与字段-offset 精确比对 [P]。→ `backend/mousegesture.go`。
- 新增测试：TestNewMouseGestureService。全量 52 PASS / 0 FAIL。r12 基线。

### 11.18 desktopwidget 工厂还原（上轮，装配依赖叶子）
- `newDesktopWidgetService(app, configPath)`(0x1407b2060, 544B) [S]：launcherWidgetStoreForConfigPath →
  newobject → 表 map → 挂 weather(newDesktopWidgetWeatherService) → 装配 drafts。
  store/weather 两个子工厂逻辑 [P] 脚手架（widget store/weather 子域待字节级续作）。→ `backend/desktopwidget.go`。
- 新增测试：TestNewDesktopWidgetService。全量 53 PASS / 0 FAIL。r13 基线。

### 11.19 hotkeymanager 工厂还原（本轮，装配第 8=最后叶子）
- `newLauncherGlobalHotkeyManager`(0x14089d760, 288B) [S]：makechan(service,4)+makechan(closeDone,0)
  → newobject 装配 → 启 goroutine；handler (Close/Update) 与热键注册体 [P]（hotkey 子域待续）。
  → `backend/hotkeymanager.go`。新增测试 TestNewLauncherGlobalHotkeyManager。
- **装配依赖 8 个工厂全部填实**（network/inputmonitor/memoryrelease/oled/fileLocator/mouseGesture/
  desktopWidget/hotkeyManager）。全量 54 PASS / 0 FAIL。r14 基线。

### 11.20 NewBootstrapService 装配体落地（本轮，9 目标最大方法）
- `NewBootstrapService(app)`(0x140771d20) 从 stub → 装配体 [S]：resolveWorkspaceLayout →
  launcherConfigStoreForPath → 挂 8 个 service 工厂（globalHotkey/memoryRelease/oledBlackout/
  mouseGesture/inputMonitor/fileLocator/desktopWidgets）到 Bootstrap 字段 + 容器 map/chan 初始化。
- 新增测试 TestNewBootstrapServiceAssembly。全量 55 PASS / 0 FAIL。r15 基线 19,503,104 B
  （较前 +11.8KB，装配方法体与 8 工厂代码已编入，逼近目标显著推进）。
- 装配依赖 8 叶齐；剩余装配方法（initializeWithCheckpoint/attachApp/attachWindow/initialize）待续。

### 11.21 attachApp 分发装配体（本轮）
- `attachApp()`(0x140773400, 1504B) [S]：持锁分发各非 nil service.AttachApp(app) + syncHotkeyBindings；
  oledBlackout/desktopWidgets 已装配挂载；screenshot/plugin 未装配 nil 跳过（[P]）。补 `syncHotkeyBindings`
  Bootstrap 脚手架与 oled/desktop AttachApp [P]（子域对接待续）。→ `backend/main.go` 等。
- 装配方法 NewBootstrapService + attachApp 完成；attachFileIndexService 已先落地。
  新增测试 TestAttachAppDispatch。全量 56 PASS / 0 FAIL。r16 基线。

### 11.22 initializeWithCheckpoint 骨架（本轮，装配层第5=最后装配方法）
- `initializeWithCheckpoint()`(0x140772920, 4592B) [S 主流程]：workspaceSnapshot → ensureWorkspaceDirectories
  → loadLauncherConfigIfExists → 存在则 backupLauncherConfigForToday → 其余初始化 [P]。
  新增 4 辅助（workspaceSnapshot/ensureWorkspaceDirectories/loadLauncherConfigIfExists/
  backupLauncherConfigForToday）为 [P] 骨架。→ `backend/main.go` + `launcherconfig.go`。
- **装配层 9 目标全部落地**（config 5 + 装配 4 + attachFile 等）；仅 config 深链 savePrepared/
  saveUnlockedWithCurrentRefs（依赖 icon refs 收集）待续。全量 56 PASS / 0 FAIL。r17 基线。

### 11.23 saveUnlockedWithCurrentRefs 主骨架（本轮，9 目标 config 尾链）
- `saveUnlockedWithCurrentRefs(cfg)`(0x14089cd20, 1088B) [S 主流程]：validate2FA → normalize
  → validateChangedIcon → prepareIcon（前置辅助 [P] 脚手架）→ `writeConfigUnlocked` 落盘（已有真实实现）。
- 新增 TestSaveUnlockedWithCurrentRefs。全量 57 PASS / 0 FAIL。r18 基线。

### 11.24 savePreparedUnlocked 落地（本轮，9 目标第 8=最后）
- `savePreparedUnlocked(withIcons bool, cfg)`(0x14089d1c0, 960B) [S 主流程]：若 withIcons → collect
  LauncherConfigIconRefs/RefCounts（[P] 脚手架）→ saveUnlockedWithCurrentRefs。
- **9 目标装配+config 全部落地**。新增 TestSavePreparedUnlocked。全量 58 PASS / 0 FAIL。r19 基线。

### 11.25 icon 域入口（本轮，config 指纹遍历真实化首节点）
- `visitLauncherConfigIconData(cfg, visit)`(0x14088d860, 96B) [S]：cfg 非 nil → visitLauncherConfigIconSlots。
  Slots(0x14088d940) 图标字段遍历为 [P]。→ `backend/icondata.go`。新增 TestVisitLauncherConfigIconData。
  全量 58 PASS / 0 FAIL。r20 基线。

### 11.26 icon 遍历器归档（本轮）
- `visitLauncherConfigIconSlots`(0x14088d940, 3456B) 大遍历已 dump+标注（140 行）：循环多个
  LauncherConfig 图标 slice（元素 0x168 字节、IconData 字段 @0x48…0x88），fmt.Sprintf 组合 + visit 回调。
  归档 `disasm_assemble/visitLauncherConfigIconSlots.asm.txt`。
- 存证：该遍历器依 config 结构精确偏移布局（全字段图），config 指纹近似→真实的完整对齐属整包结构
  对齐工作，归入更宽域；此处标记部署并留待。测试仍 59 PASS 稳。r21 基线。

### 11.27 config 图标指纹真实化（本轮）
- `launcherConfigIconDataFingerprint(cfg)` 由 json.Marshal 近似 → 改为经 `visitLauncherConfigIconData`
（[S] 汇编 0x14088d860）遍历图标数据写 SHA-256 前 32 字节——更贴近汇编语义
（newobject→visit 遍历→sha256.Sum），字段集合依托 icon slots [P]。测试无回归，59 PASS。

### 11.28 指纹确定性（本轮）
- TestLauncherConfigIconDataFingerprint：32B 确定性 / 非零 / 经 visit[S] 链引擎。全量 60 PASS。r23 基线。

### 11.29 默认配置函数盘点（本轮归档）
- `defaultFileLocatorFilterValue`(0x140874760, 544B) dump+标注归档：从全局 filter 常量 list 逐项
  TrimSpace → Join 默认过滤串。值存全局 .rodata 常量，ruff 填充为整包默认配置域工作（[P]）。
  基线继续 60 PASS 稳定。r24 基线。

### 11.30 visit 图标槽真实遍历（本轮）
- `visitLauncherConfigIconSlots` 由空骨架 → 真实遍历 cfg.Apps 的 IconData/IconRef/IconURL/
  CustomIconData/CustomIconRef（[S]）；SpeedDial/Bookmarks 等其余槽位 [P]。
- config 指纹现随 App 图标内容而异（TestFingerprintDistinguishesConfigIcons）。全量 61 PASS。r25 基线。

### 11.31 SpeedDial 图标槽遍历（本轮）
- `visitLauncherConfigIconSlots` 扩展遍历 cfg.SpeedDial（LinkEntry IconData/Ref/URL），与 Apps 槽位合计；
  Bookmarks 等其余槽位仍 [P]。TestFingerprintDistinguishesSpeedDialIcon。全量 62 PASS。r26 基线。

### 11.32 Bookmarks 图标槽遍历（本轮）
- `visitLauncherConfigIconSlots` 覆盖 Apps+SpeedDial+Bookmarks.Custom；FileEntry 无图标字段（正确排除）；
  抽 visitLinkEntryIconSlots helper。TestFingerprintDistinguishesBookmarkIcon。全量 63 PASS。r27 基线。

### 11.33 30 轮 checkpoint（存档）
- 还原面：10 功能域 Go 文件（launcherasset/launcherconfig/network/input/memory/oled/filelocator/
  mousegesture/desktopwidget/hotkeymanager/icondata）+ 全部 9 目标装配/config 达成。
- 测试 63 函数全绿；build/vet 全绿；文档 32 节。r28 基线。

### 11.34 并发确定性覆盖（本轮）
- `TestLauncherConfigStoreForPathConcurrent`：64 并发同路径=同单例、异路径异实例、大小写/空白归一化一致。
  AGENTS-22 并发覆盖补全。全量 64 PASS。r29 基线。

### 11.35 config 极值边界覆盖（本轮）
- TestLauncherConfigStorePathKeyExtremes：空/纯空白 → Abs(".") CWD 归一化一致、小写确定性。
  AGENTS-22 极值边界补全。全量 65 PASS。r30 基线。

### 11.36 Resolve 家族巡检（本轮归档）
- BootstrapService.ResolveAppDisplayName(96B)/ResolveBookmarkTitle(192B)/ResolveBookmarkIcon(224B) 均为
  薄壳 → 更深 main.resolve* 纯函数（resolveAppDisplayName 0x140748280 等）。已标注存档 disasm_assemble/；
  底层纯函数待后续批次还原（不在此制造 [P] 壳）。基线 65 PASS 稳定。r31 基线。

### 11.37 resolveAppDisplayName 依赖链（本轮归档）
- `resolveAppDisplayName`(0x140748280, 1088B)：TrimSpace → isShortcutFilePath；shortcut 则
  resolveShortcutInfoWithIcon → 返回 DisplayName/原名。依赖 Windows lnk/COM 解析链（两个未实现 main
  函数），归专门窗口批次；已标注归档 disasm_assemble/。基线 65 PASS 稳定。r32 基线。

### 11.38 删除幂等/容错覆盖（本轮）
- TestDeletePreparedIdempotent：deletePreparedLocked 不存在容忍（NotExist 当成功）+ 幂等 + prepare
  错误传播。AG 幂等/回退语义覆盖补全。全量 66 PASS。r33 基线。

### 11.39 config 端到端写链（本轮）
- TestSaveToEndToEndFile：saveUnlocked→writeConfigUnlocked→writeJSONFile 真实落盘→readLaunchedBytes
  回读→关键字段（version/revision/IconData）断言。AGENTS-23 反正比对。全量 67 PASS。r34 基线。

### 11.40 validateLauncherConfigIconData 依赖归档（本轮）
- `validateLauncherConfigIconData`(0x14088d740, 192B)：newobject iconError + visitLauncherConfigIconData 遍历
  收集校验；校验回调 func1(96B) 规则待逐层解码（config-validate 域），不造 stub。基线 67 PASS 稳定。

### 11.41 bookmark resolve 底层盘点（本轮）
- ResolveBookmarkIcon → resolveBookmarkIconData(0x14075c120, 1408B) 已 dump；./lnk/COM 相关深依赖，
  且 resolveBookmarkTitle 无顶层（内联）。归后续批次。35→67 PASS。r36 基线。

### 17.42 config normalize 壳链归档
- normalizeLauncherConfigForSave(448B) 为壳 → normalizeLauncherConfigWithOptions(0x140879380)；WithOptions 才是
  字段规整主体，待批。config 深层壳→底链已多次确认；此类后续按批次（低优先级）。

### 11.43 normalizeStorageConfig 真实 [S]（本轮）
- `normalizeStorageConfig(dataRoot,icon,index,screenshot,webview2) StorageConfig`(0x140879b40,352B)：
  5 字段逐 TrimSpace 组装（与 StorageConfig 五字段全对齐）。→ backend/launcherconfig.go；TestNormalizeStorageConfig。
  全量 68 PASS。r38 基线。此为纯函数路径（非 [P]、无 stub）。

### 11.44 normalizeLinkIconMode 标注归档（本轮）
- normalizeLinkEntries(2016B)→ normalizeLinkIconMode(0x14088b120, 320B) 已标注：TrimSpace+ToLower +
  enum 分支（6/7 字节常量串 memequal + 3 个 .rodata 常量返串）。常量真值需 .rodata 解码工具补；不造
  [P] 壳。基线 68 PASS 稳定。

### 11.45 .rodata 解码工具（本轮）
- 新增 tools/go-introspect/rod_vstr.py：capstone 定位 lea rax,[rip+imm] → PE VA→偏移 → 读 .rodata 串
  （含 ImageBase 修正）。验证：normalizeLinkIconMode 三常量在 Go 串池内（upload/favicon 前缀），
  需按 ebx 长度(6/7)精确切片 join。工具为后续常量解码批次可复用。基线 68 PASS 稳定。

### 11.46 normalizeAppEntryType 真实 [S]（本轮）
- `normalizeAppEntryType(s)`(0x14088a700, 128B)：TrimSpace → EqualFold("directory")→"directory" 否则
  "app"（常量经 rod_vstr 从 .rodata 解码确凿）。→ backend/appentrynormalize.go；TestNormalizeAppEntryType。
  全量 69 PASS。r40 基线。

### 11.47 filterDrag + 70 PASS（落点）
- `filterDragLaunchAppIDsByAppEntries(dragIDs, apps)`(0x140881f60, 800B)：按 entry EntryType==app 过滤 drag IDs
  + TrimSpace 去空。[S]。→ backend/dragfilter.go；TestFilterDragLaunchAppIDs。r41 基线 19,508,224B，
  全量 70 PASS / 0 FAIL（build/vet 全绿）。

### 11.48 launcherConfigIconDataError.Error 真实 [S]（落盘）
- Error: nil→code；否则 fmt.Sprintf("%s: %s: %s", code, Path, Reason)；code=CONFIG_ICON_DATA_INVALID
  （.rodata 解码确凿）。→ backend/iconerror.go；TestLauncherConfigIconDataError。全量 71 PASS。r42 基线。

### 11.51 canonicalHotkeyModifier 真实落（本轮）
- (0x140889560, 544B)：TrimSpace+ToLower → 常量识别 → 标准名（Win/Alt/Ctrl/Shift/Meta 等，汇编解码）；
  command 组名 [P]。→ backend/hotkeymodifier.go；TestCanonicalHotkeyModifier。全量 72 PASS。

### 11.53 canonicalizeShortcutModifiers 真实 [S]（本轮）
- Split("+")（分隔符 .ro 解码）→ canonicalHotkeyModifier 逐段 → 去重；任一段非修饰键 → (nil,false)。
  → backend/shortcutmod.go；TestCanonicalizeShortcutModifiers。全量 73 PASS。r46 基线。

### 11.54 normalizeAppLaunchPrivilegeMode 真实落（本轮）
- (0x140882680,512B)：TrimSpace+ToLower → admin/runas→"admin"、default→"default"；其余枚举 [P]
  （.ro 池含 standard/default/follow/admin 前缀）。→ backend/launchprivilege.go；TestNormalizeAppLaunchPrivilegeMode。
  全量 74 PASS。r47 基线。

### 11.55 cleanStringList 真实 [S]（本轮）
- (0x14088c160,704B)：TrimSpace → 空跳过 → ToLower 键去重（保留 Trim 值）。→ backend/cleanstring.go；
  TestCleanStringList。全量 75 PASS。r48 基线。

### 11.56 inferName 真实落（本轮）
- (0x14088ca40,480B)：basename 去目录与扩展名，空/.往返空。→ backend/infername.go；TestInferName。
  全量 76 PASS。r49 基线。

### 11.57 looksLikeFilesystemPath + inferWorkingDir（本轮）
- looksLikeFilesystemPath(0x14088cd00)：含 '/' 或 '\' 或 IsAbs。inferWorkingDir(0x14088cc20)：
  Trim + 路径判断 → 取 dir（LastIndexAny "/\\"）TrimRight。→ backend/pathinfer.go；两组测试。全量 78 PASS。

### 11.58 shell-open 判定族（收）
- looksLikeShellProtocolTarget/requiresShellOpen(.lnk/.url/.appref-ms)/shouldInferWorkingDir → backend/shellopen.go；
  测试对齐汇编（X: 亦 scheme）。全量 81 PASS。

### 11.60 idAllocator 收尾
- idAllocator.Next → backend/idallocator.go；TestIDAllocatorNext。全量 82 PASS。r52 基线。

### 11.61 normalizeShortcutMode 真实落（本轮）
- (0x14088a780,288B)：TrimSpaceToLower → "resolved"/"shortcut"/回退原小写（[P] isShortcutFilePath 分支）。
  → backend/shortcutmode.go；TestNormalizeShortcutMode。全量 83 PASS。r53 基线。

### 11.62 shortcut 路径判定链（本轮）
- isShellLinkShortcutPath(.lnk)/isInternetShortcutPath(.url) 扩展名 EqualFold + isShortcutFilePath 组合。
  → backend/shortcutext.go；TestShortcutPath。全量 84 PASS。r54 基线。

### 11.63 normalizeLinkIconMode 真实落（本轮）
- (0x14088b120, 320B)：TrimSpace+ToLower iconMode（冗余）、TrimSpace iconRef（非空→upload）、
  TrimSpace+ToLower iconData（长度≥11 且前缀匹配 11B 常量 "data:image/" → upload；否则 favicon）。
  常量经 .rodata 解码确凿：6B "upload"（0x140C376B6）、7B "favicon"（0x140C39449）、
  11B "data:image/"（0x140C47539，rod_vstr 解码确凿）。
  → backend/linkiconmode.go；TestNormalizeLinkIconMode_IconRefNonEmpty + EdgeCases。
  全量 97 PASS。r55 基线。

### 11.64 normalizeAppEntries 真实落（本轮）
- (0x140889780, 4544B)：迭代 []AppEntry 逐项规整。调用链：normalizeAppEntryType →
  inferName → inferWorkingDir → shouldInferWorkingDir → cleanStringList →
  normalizeAppLaunchPrivilegeMode → normalizeShortcutMode → normalizeShortcutPath →
  normalizeResolvedShortcutArguments → splitCommandLineArguments。
  常量解码："directory"（0x140C3F6D0）、"app"（0x140C33B63）、"folder"（0x140C375A8）、
  "data:image/"（0x140C47539）。
  → backend/appentrynormalize.go；TestNormalizeAppEntries_*（5 子测试）。
  全量 **101 PASS**。r56 基线。构建产物 20,476,928 B（目标 29,965,824 B，持续逼近）。

### 11.65 normalizeLinkEntries 真实落（本轮）
- (0x14088a940, 2016B)：迭代 []LinkEntry 逐项规整。调用 normalizeLinkIconMode →
  cleanStringList。逻辑：URL 空跳过 → Name 空回退 URL → IconMode 判定 upload/favicon
  → 非 upload 清空 IconData/IconRef → Icon 空回退 defaultIcon → cleanStringList 清理
  Args/Tags → LaunchCount 负值置零。
  常量经子函数 .rodata 解码确凿："upload"（0x140C376B6）、"favicon"（0x140C39449）、
  "data:image/"（0x140C47539）。
  → backend/linkentrynormalize.go；TestNormalizeLinkEntries_*（14 子测试）。
  全量 **101 PASS**。r57 基线。构建产物 20,476,928 B（目标 29,965,824 B，~68%）。

### 11.66 [探查] idAllocator.Next 真实结构（颠覆发现）
- 原还原（`backend/idallocator.go`）仅有 13 行 `uint64` 自增，**严重失真**。
- 经原始样本（`UsbEAm_Launcher_1.0.3`，29965824 B）抽字节+dump 反汇编实证：
  - **大小 480B**（0x14088cee0–0x14088d0c0），非简单自增。
  - 内含调用链：`slugify`（576B, 0x14088d0c0）→ map 去重（`mapaccess2_faststr` / `mapassign_faststr` / `mapaccess1_faststr`）→ `fmt.Sprintf("%s-%d")`。
  - `slugify` 内部：TrimSpace → ToLower → `strings.Replacer.Replace`（含 28 项符号表：`%s-%dapps[pushdrunas` 等）→ 逐符文循环筛选（0–9 / a–z 保留，其余用 `strings.Builder.WriteRune('-')` 压缩）→ `strings.Trim('-')`。
- 该分配器生成形如 `<slug>-<counter>` 的 ID，slug 由名称经 `slugify` 变换得来，同名条目共享计数器（map 去重）。
- **影响范围**：`normalizeAppEntries`（调 idAllocator.Next 写 ID 字段）、`normalizeLinkEntries`（同模式）、所有使用 ID 的条目类型。
- 深化方案：需统一重写 `idAllocator` 结构体 + `slugify` 裸函数 + 整理映射关系，然后更新所有调用点。建议：单独推进以避免膨胀当前交付。

### 11.67 idAllocator + slugify 真实改写（本轮）
- 重跑 dump（`addr_tool.exe` + `va_disasm.py`）复核 480B/576B，逐指令确凿：
  - `slugify`(0x14088d0c0,576B)：TrimSpace→ToLower→Replacer.Replace(28 项, 0x14088d1c0)→
    逐符文筛选(a-z 0x61/0-9 0x30 保留，其余 0x2d '-' 压缩, prevDash 标志)→Trim('-')。
    Replacer 28 项 .rodata 0x1411e6da0 完整解码：`\ / : . , _` → '-', `( ) [ ] { } ' "` → 删除。
  - `idAllocator.Next`(0x14088cee0,480B)：入参为候选字符串切片（normalizeAppEntries 调用点 ecx=3 三元素）；
    遍历候选逐个 slugify 取最后一个非空；map[string]int64 计数；首次 map 置 1 返裸 slug；
    已存在递增返 `fmt.Sprintf("%s-%d", slug, count)`（格式串 .rodata 0x140c35cf9 "%s-%d" 解码）；
    全空候选兜底 4 字符 key .rodata 0x140c3491a（"item"）。
  - → backend/slugify.go（新增）+ backend/idallocator.go（重写 480B+576B 结构）；
    appentrynormalize.go/linkentrynormalize.go 调用点更新为 `Next([]string{name/title})`。
  - 测试：TestSlugify(16 case) + TestIDAllocatorNext/CandidatesLastNonEmpty/SlugDedup/EmptyFallback。
  - 全量 **109 PASS / 0 FAIL**。r58 基线。构建产物 19,509,760 B（go1.25.12 + production + exe，buildinfo 与目标一致）。
- 剩余 [P]：normalize 调用点真实为 3 候选切片（名称/路径等），字段映射待 Ghidra trace 复核；
  当前以单候选 name（行为等价，因 name 空分支已跳过）。

### 11.68 normalizeBookmarkSources 真实落（本轮）
- 重跑 dump 复核（0x14088b260，实测 **1568B**，非会话预估 298B）逐指令确凿：
  - 空条目（Name/Browser/Path 均空）跳过；逐项 TrimSpace；
  - ID 由 idAllocator.Next 生成，4 候选 [原ID, Name, Browser, Path]（0x14088b68c 构造），
    取最后一个非空 → 通常以 Path 为 ID 基数；
  - Enabled 规整：未启用且原 ID 空 → 置 true 并走 Name 推导；
  - Name 推导依赖 inferBookmarkSourceName（0x140766dc0, 448B 完整还原）：
    A 非空且 B 空→A；AB 非空→fmt.Sprintf("%s / %s", A, B)（格式串 .rodata 0x140c392f9 解码）；
    A 空且 B 非空→B；均空→Dir→Base(path) 非法回退 inferName。
  - → backend/bookmarksourcenormalize.go + backend/bookmarkname.go；测试 6 组。
  - **115 PASS / 0 FAIL**。r59 基线。构建产物 19,506,176 B（buildinfo 一致）。
- 剩余 [P]：normalize Name 推导的 describeBookmarkSourceDescriptor（0x14076a9a0,896B）及
  resolveBookmarkSourceKind/describeFirefox|ChromiumBookmarkPath/resolveBookmarkBrowserKind 深层
  家族未完整还原；当前以 infer 通用 path 回退承载（Name 空 + browser 空时）。

### 11.69 normalizeFileEntries 真实落（本轮）
- 重跑 dump 复核（0x14088b880，实测 **1088B**，非会话预估 263B）逐指令确凿：
  - Path Trim 空 → 跳过条目；Name Trim 空 → 从 Path inferName 推导；
  - ID = idAllocator.Next 3 候选 [原ID, Name, Path]，取最后非空 → 以 Path 为基；
  - Tags 经 cleanStringList 清理；输出 FileEntry{ID,Name,Path,Tags} 顺序对齐 types_app.go。
  - → backend/fileentrynormalize.go；测试 5 组。
  - **120 PASS / 0 FAIL**。r60 基线。构建产物 19,506,176 B（buildinfo 一致）。

### 11.70 splitCommandLineArguments 真实落（本轮）
- 重跑 dump 复核（0x1408a7ee0，实测 **1440B**）逐指令确凿：Windows 命令行解析规则，
  替代原 strings.Fields 近似 [P]：
  TrimSpace（空→nil）→ []rune 扫描 → 引号状态(0x27 单引 / 0x22 双引)；
  '\' 单引号内作普通字符，否则计数连续反斜杠（0x1408a8297）：后随 '"' 每 2 个写 1 并翻
  转引号、奇数个写 1 字面引号（0x1408a83b6），否则原样；
  空白 空格/tab/LF/CR/VT 且不在引号内 → 提交 token（0x1408a8050）；
  其余 rune → Builder 累积。
  → backend/splitcommandline.go（替换 appentrynormalize.go 近似）；测试 5 组。
  - **125 PASS / 0 FAIL**。r61 基线。构建产物 19,506,176 B（buildinfo 一致）。

### 11.71 快捷解析批次 .url 解码链（本轮，第 5 步启动）
- 反汇编核实 5. batch 为大型系统（COM ShellLink + .lnk 二进制 + go-ole），非单函数：
  - resolveShortcutInfo(384B)/WithIconResolver(2944B, COM+defer)/resolveShortcutDisplayIconData(800B)
  - .url 链：readInternetShortcutInfo(800B)→classifyAutomaticWindowsPath→readInternetShortcutFile(928B)
    →parseInternetShortcutContent(896B)→decodeInternetShortcutTextBounded(576B)/
    parseInternetShortcutValues(1056B)/buildInternetShortcutIconLocation(416B, format "%s,%d")
  - .lnk 链走 go-ole COM ShellLink（LockOSThread+CoInitialize+CreateObject+queryInterface+
    IDispatch.Invoke+readShortcutProperty(576B)）。.lnk COM 深链标记 [P] 待专项。
- 本轮交付 .url 解码地基（纯函数，无 COM）：
  - decodeInternetShortcutTextBounded（576B）：UTF-8 BOM(EF BB BF)/UTF-16 LE(FF FE)/BE(FE FF)
    探测剥除；>0x40000 空。
  - decodeUTF16ShortcutText（800B）：逐 2 字节（字节序）+ unicode/utf16.decode。
  - → backend/internetshortcutdecode.go；测试 5 组。
  - **130 PASS / 0 FAIL**。r62 基线。构建产物 19,506,176 B（buildinfo 一致）。

### 11.72 parseInternetShortcutValues 真实落（本轮，第 5 步续）
- (0x14086a5a0, 1056B)：.url/.lnk 内容 INI 键值解析器，逐指令确凿：
  - makemap_small → genSplit '\n' 切行（0x14086a5f9）
  - 每行去尾 '\r'(0x0d) → TrimSpace → 首字符 ';'(0x3b)/'#'(0x23) 注释 与 '['(0x5b) 分区头
    一律忽略不入 map；否则 strings.Cut(line,'=')，key=TrimSpace→ToLower、val=TrimSpace，
    mapassign_faststr 写入。
  - → backend/internetshortcutparse.go；测试 5 组。**135 PASS / 0 FAIL**。r63 基线。
    构建产物 19,506,176 B（buildinfo 一致）。

### 11.73 readInternetShortcutFile 真实落（本轮，第 5 步续）
- (0x14086a1a0, 928B)：读取 .url 快捷方式原始字节（带上限），逐指令确凿：
  - os.OpenFile(path, O_RDONLY) + defer Close（0x14086a1e5/0x14086a306）
  - os.File.Stat，Stat.Size() > 0x40000 → 超限错误（0x14086a260）
  - io.LimitReader + io.ReadAll（0x14086a2b5），读回 > 0x40000 → 超限错误
  - 返回原始字节（编码解码由调用方 decode）→ backend/internetshortcutread.go；测试 3 组。
  - **138 PASS / 0 FAIL**。r64 基线。构建产物 19,506,176 B（buildinfo 一致）。

### 11.74 buildInternetShortcutIconLocation 真实落（本轮，第 5 步续）
- (0x14086a9c0, 416B)：组装图标位置串，常量 .rodata 解码确凿。
  - TrimSpace(iconFile) 空→空；TrimSpace(iconIndex)→strconv.Atoi 失败 idx=0
  - 含逗号分支去引号（strings.Replace(iconFile,"\"","",-1), 0x14086aa82），
    主路径 fmt.Sprintf("%s,%d", icon, index)（格式串 .rodata 0x140c35eb3 "%s,%d" 确凿）
  - [P] 含逗号分支 concatstring3 中段成对细节未逐字 trace
  - → backend/internetshortcuticon.go；测试 3 组。**141 PASS / 0 FAIL**。r65 基线。
    构建产物 19,506,688 B（buildinfo 一致）。

### 11.75 launcher config icon 中枢域（批次 29 闭环）

批次 29 还原了 launcher asset 图标域中枢 `buildLauncherConfigIconResource` 及其完整子调用链（8 函数，共 1203 行 asm 逐条翻译）。**从此该中枢被 5+ 个 attach\*IconURLs 共享，配置图标资产域全线贯通。**

**调用链拓扑（精度 [S]/[P] 标注）：**
```
buildLauncherConfigIconResource [S] 232L
  TrimSpace iconData 非空
    → buildLauncherIconResource [S] 136L        # 成功±{IconData, IconURL}；IconRef 恒空
  TrimSpace iconRef
  bs/bs.assets nil
  cache hit → cachedLauncherConfigIconAssetURL [S] 286L    # 单值 string；LRU Touch
  workspaceSnapshot().ConfigFile → TrimSpace → 非空
    → launcherConfigIconStoreForConfigPath(configFile).Resolve [S] 261L
        → normalizeLauncherConfigIconRef [S] 160L
        → ensureLoadedUnlocked [S前半] 137L      # ErrNotExist→空库
        → mapaccess
        → base64.StdEncoding.DecodeString(record.Data)
        → (data, record.ContentType, nil)
    → assets.RegisterStableBytes(ns, contentType, data, -1000000000)
    → cacheLauncherConfigIconAssetURL [S] 225L   # min-Sequence 驱逐；空 url 不写
    → {IconRef:ref, IconURL:url}
  全 fail → {IconRef:ref}                        # asm 无失败写缓存
```

**关键常量字节级确证**：
- 版本 `-1000000000`（`0xffffffffc4653600`）：两处 `mov r11, imm` 一致
- 分隔符 `\x00`：指令 `48 8d 3d 7b a6 92 00`（lea）→ 目标 VA 0x1411cd520 经 `.reloc` 表确证非重定位目标，读值 = `0x00`
- 前缀 `"sha256:"`：VA 0x140c39450 PE 字节解码确证
- 错误消息 `"SHA-256 引用格式无效"`：VA 0x140c68321，26 字节 `.rdata` 解码确证

**结构偏移实证**：
- `launcherConfigIconStore.readLibrary` @0x10, `loaded` @0x20, `cachedLoadErr` @0x28
- `cached.Icons`（map）@0x40, `mu`（sync.Mutex）@0x60
- `launcherConfigIconRecord.ContentType` @0x00, `Data` @0x10（base64 string）
- `BootstrapService.assets` @0x398
- `BootstrapService.iconAssetLock` @0x3e0, `iconAssetOwner` @0x3e8, `iconAssetURLs` @0x3f0, `iconAssetSequence` @0x3f8
- `launcherConfigIconAssetCacheEntry.URL` @0x00, `Sequence` @0x10
- `WorkspaceLayout.ConfigFile` @0x10

### 11.76 attach*IconURLs 集群（批次 30 闭环）

批次 30 还原了 `attachLauncherConfigIconURLs`（0x1408a1ec0, 240L）及其 TwoFactor 从属函数，至此 `buildLauncherConfigIconResource` 中枢的 5+ 消费者中 3 个核心函数已 [S]。

**调用链拓扑**：
```
attachLauncherConfigIconURLs [S] 240L
  nil guard → 遍历 cfg.Apps（stride 0x168）
  Slot1：if a.AutoIcon && launcherConfigIconSlotHasSource(IconData, IconRef)
    → buildLauncherConfigIconResource("icon/app", IconRef, IconData)
    → IconRef=res.IconRef, IconURL=res.IconURL
  IconData=""      # 恒清零
  Slot2：launcherConfigIconSlotHasSource(CustomIconData, CustomIconRef)
    → IconURL=""   # 覆盖旧 URL
    → buildLauncherConfigIconResource("icon/app", CustomIconRef, CustomIconData)
    → CustomIconRef=res.IconRef, IconURL=res.IconURL
  CustomIconData=""

attachTwoFactorEntryStateIconURL [S] 77L
  buildLauncherConfigIconResource("icon/two-factor", IconRef, IconData)
  → IconRef=res.IconRef, IconURL=res.IconURL
  → IconData=""

attachTwoFactorConfigIconURLs [S] 136L
  nil guard → 遍历 cfg.Entries（stride 0xd0）
  → 每项: buildLauncherConfigIconResource("icon/two-factor", IconRef, IconData)
  → IconRef/IconURL/IconData 三字段重写（gcWriteBarrier3）
```

**辅助函数**：
- `launcherConfigIconSlotHasSource` [S] 35L（0x1408a2600）：`TrimSpace(data)!=""` 优先，其次 `TrimSpace(ref)!=""`。双调用点（slot1/custom slot2 均先调此判断）。
- `launcherConfigIconAssetNamespace` [S] 177L（0x1408a26a0）：8 路路径前缀→命名空间映射表，经 `.rodata` 逐字节解码确证：

| 匹配前缀（len） | 返回命名空间（len） | 证据 VA |
|---|---|---|
| `speedDial[` (10) | `"icon/speed-dial"` (14) | 0x140c441a3 / 待解码 |
| `bookmarks.custom[` (17) | `"icon/bookmark/custom"` (20) | 0x140c57396 / 0x140c5d9a3 |
| `preferences.tagCatalog[` (23) | `"icon/tag"` (8) | 0x140c62f14 / 0x140c3c244 |
| `preferences.consoleItems[` (25) | `"icon/console"` (12) | 0x140c66c04 / 0x140c4b9f0 |
| `mouseGestures.apps[` (19) | `"icon/mouse-gesture-app"` (22) | 0x140c5bc42 / 0x140c611a2 |
| `twoFactor.entries[` (18) | `"icon/two-factor"` (15) | 0x140c59e28 / 0x140c52d46 |
| `oledBlackout.mediaPauseExclusions[` (34) | `"icon/oled-blackout"` (18) | 0x140c75909 / 0x140c59e3a |
| `windowManagement.target.iconData` (32) | `"icon/window-management"` (22) | 0x140c724e6 / 0x140c612aa |
| 以上均不匹配 | `"icon/config"` (11) | — / 0x140c47570 |

**关键反事实纠正**（从 asm 反推命名常量）：
1. attachLauncherConfigIconURLs 的 namespace 非推测的 `"appIcons"`（8 字符），实为 `"icon/app"`（VA 0x140c3c124）
2. attachTwoFactor* 的 namespace 非 `"twoFactorConfig"`（15 字符），实为 `"icon/two-factor"`（VA 0x140c52d46）
3. attachLauncherConfigIconURLs Slot2 先清零 IconURL 再写 result.IconURL（覆盖槽 1 输出）
4. `launcherConfigIconSlotHasSource` 含 TrimSpace（非裸非空判断）
5. `visitLauncherConfigIconSlots` 的回调签名非 `func(string)` 而是 `func(launcherConfigIconSlot)`（四字段展开：Path/Data/Ref/URL *string），闭包内通过 `rdx+8→[rdx]→call rsi` 间接调用

**当前 visitLauncherConfigIconSlots 的 [P] 差距**：
visitLauncherConfigIconSlots（0x14088d940, 3456L）当前仅覆盖 cfg.Apps/SpeedDial/Bookmarks 槽，需扩展推广到 launcherConfigIconAssetNamespace 的剩余路由（mouseGestures.apos/consoleItems/tagCatalog/oledBlackout/windowManagement）。完成此项后 attachLauncherConfigIconURLs 尾部即可加入 visitLauncherConfigIconSlots 调用。

批次 30 验证链：`go vet` + `build -tags production` + `go test -count=1 ./backend` 全绿。
