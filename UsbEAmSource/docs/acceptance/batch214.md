# 批次 214 — filelocator 运行时两 [P] 升档 [S-sig]（runSearch + finishSearch 签名实证）

## 基线 / 收口

| 指标 | 基线（批次 213 收口） | 收口（批次 214） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1385 | **1387** |
| P | 111 | **109** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2677 | **2679**（56.35%） |

SHA256 `2F4D421B7D6C1BE0F6579D7290E4AAE84F4E094FF7E963FF04D47C7C66EE04F2`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go` 两个 [P] 存根经汇编逐寄存器实证后升档 [S-sig]（仅签名修正，
体未逐条翻译保持零值——体均为大函数，涉及主循环/WalkDir/错误处理链，留待专项）。

### 升档 [S-sig]（签名实证修正）

- `(*fileLocatorService).runSearch`（0x1407d0260）：签名
  `(prepared *fileLocatorPreparedSearch, ctx context.Context) error`
  → `(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64)`。
  汇编实证：recv(rax)+ctx(2 word)+generation(uint64) 共 4 寄存器，prepared 为值传（288B 栈
  @rbp+0x10，栈槽 [rsp+0x3b0]=prepared.roots.len@+0xf8）；返回 void（尾声无返回寄存器，非 error）。
  体主循环遍历 prepared.roots，对每根调 waitIfPaused(ctx,generation,time.Now()) 后 walkRoot。
- `(*fileLocatorService).finishSearch`（0x1407d2a00）：签名
  `(cancelled bool)` → `(generation uint64, startedAt time.Time, lastError string, cancelled bool)`。
  汇编实证：recv(rax)+generation(uint64)+startedAt(time.Time 3 word)+lastError(string 2 word)+
  cancelled(bool r10b) 共 8 寄存器；返回 void。体内 time.Now().Format + time.Since(startedAt) +
  TrimSpace(lastError) 组装终态并清理 recv.cancel（+0x1e0）；参数 5/6 经 strings.TrimSpace 调用
  实证为 string，非 int。

## 关键知悉

- `runSearch` 尾声调 `finishSearch(generation, startedAt, "", true)` 走取消路径（cancelled=true），
  与 defer 错误路径（cancelled=false + lastError 非空）构成两条收尾分支。
- `waitIfPaused` 的实参 startedAt 在 runSearch/processFile 内均为 `time.Now()` 的结果（3 word），
  印证批次 213 的 time.Time 参数口径。
- `finishSearch` 参数 5/6 初判为 int（runSearch 传 xor r8d/r9d 清零），经 asm 内 TrimSpace 调用
  反推为 string 的 ptr/len（空串清零），纠正本批分析初期的误判。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1387 / P=109 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（2 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filelocator 域剩余 [P]：walkRoot（0x1407d0d20，prepared 值传 + 7 寄存器 +
  startedAt 栈参，返回 error，非 void——WalkDir 尾调用返回 error）、processFile（0x1407d15a0，9
  寄存器 + prepared 值 + startedAt 栈参，含 fs.FileInfo 接口实参）、prepareFileLocatorSearch、
  buildFileLocatorPathFilter。walkRoot/processFile 的 2 个指针实参（runSearch 传 lea 局部）类型
  需沿 WalkDir 闭包 func1（0x1407d10c0）捕获链确证后落地。
- `searchPinyinContextWithTombstones`（0x140810e60）含排序/合并/去重三段，仍待专项完整翻译。
