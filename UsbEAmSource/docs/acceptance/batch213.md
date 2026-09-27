# 批次 213 — filelocator 运行时两个 [P] 升档 [S]（签名实证 + 完整翻译）

## 基线 / 收口

| 指标 | 基线（批次 212 收口） | 收口（批次 213） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1254 | **1256** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | **111** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2675 | **2677**（56.31%） |

SHA256 `00EFA6EA8B4E4BB78EC7509303EE08752BF9E873C2C0A2BC8371A10FF3281B93`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go` 两个 [P] 存根经汇编逐寄存器实证后升档 [S]（签名修正 +
完整翻译）。此前分析 `searchPinyinContextWithTombstones`（0x140810e60）时确认其闭包 `.func1`
的 48B 栈参实为 `pinyinFull+pinyinInitials`（非 plan 前半段），印证批次 211 的 b1/b2 拼音别名
签名无误；该大函数（98 行 + 84 行闭包，3398B 汇编，含排序/合并/去重三段）留待后续专项。

### 升档 [S]（签名修正 + 体完整翻译）

- `(*fileLocatorService).acquireFileLocatorContentScan`（0x1407d23e0，512B）：
  签名 `bool` → `error`。汇编尾声 `xor eax/ebx` 双寄存器返回即 error 接口，非单 bool；
  `s==nil` 分支返回的全局 error 经 .data 0x141bc4520 → "context canceled"（16B）实证为
  `context.Canceled`。体：`s.lock.Lock()` 懒初始化单缓冲信号量 `s.contentScan`，取 ch 后
  Unlock；`ctx==nil` 阻塞发送返回 nil；否则 `select { ch<-struct{}{}: nil; <-ctx.Done(): ctx.Err() }`。
- `(*fileLocatorService).waitIfPaused`（0x1407d09a0，768B）：
  签名 `(ctx) bool` → `(ctx context.Context, generation uint64, startedAt time.Time) error`。
  汇编实证 7 寄存器参数（recv+ctx 2word+generation+time.Time 3word），返回 error 双寄存器。
  体：`s.lock.Lock()` + defer Unlock；循环条件
  `generation==gen && state.Running && state.Paused && ctx.Err()==nil` 时 `pauseCond.Wait()`
  并 `time.Since(startedAt)`；退出后 `ctx.Err()!=nil` 或 generation 变 → `context.Canceled`，
  `state.Running` → nil，否则 `context.Canceled`。

## 关键知悉

- `fileLocatorService` 字段偏移经汇编对齐验证与 `types_filelocator.go` 定义一致：
  `generation`@+0x20、`state.Running`@+0x118（`FileLocatorConfig` 实测 0xf0=240B）、
  `contentScan`@+0x1e8。此前怀疑的结构体布局偏差不成立。
- `acquireFileLocatorContentScan` 的 `s==nil` 返回 `context.Canceled` 是防御路径；
  正常路径信号量语义：单缓冲 chan，首次 acquire 占满，二次 acquire 阻塞至 release 或 ctx 取消。
- `waitIfPaused` 的 `time.Since(startedAt)` 结果（duration 秒）存 `[rsp+0x158]`，dump 范围内
  未见下游读取，按 `_ =` 保守处理，进度发布链待确证。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1385 / P=111 / UNMARKED=0`。
- **G3 行为**：新增 `filelocator_runtime_test.go` 四个黄金用例（nil receiver→Canceled、
  占满信号量后已取消 ctx→Canceled、waitIfPaused 未运行→Canceled、generation 变→Canceled），
  均 PASS；全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（2 函数升档 + time import）+ 测试；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。`searchPinyinContextWithTombstones`（0x140810e60，98 行）与其闭包
  `.func1`（0x140811c60，84 行）依赖已全部就绪，但含排序/合并/去重三段（3398B 汇编），
  需专项完整翻译。filelocator 域剩余 [P]：runSearch（0x1407d0260，返回值 error 存疑→实为
  void，prepared 值传 + limit 参数待精确）、walkRoot、processFile、finishSearch（8 参数）、
  prepareFileLocatorSearch、buildFileLocatorPathFilter 等。
