# 批次 215 — filelocator 运行时三 [P] 升档 [S-sig]（publishProgress + prepareFileLocatorSearch + buildFileLocatorPathFilter）

## 基线 / 收口

| 指标 | 基线（批次 214 收口） | 收口（批次 215） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1387 | **1390** |
| P | 109 | **106** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2679 | **2682**（56.41%） |

SHA256 `94FBC3CC507E6F3541692DBF597377CB4FF3D11DA717FB93C41D197160FA3F66`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go` 两个 [P] 存根经汇编逐寄存器实证后升档 [S-sig]（签名修正，
体未逐条翻译保持零值）。

### 升档 [S-sig]（签名实证修正）

- `(*fileLocatorService).publishProgress`（0x1407d2620）：签名 `()` → `(generation uint64, startedAt time.Time)`。
  汇编实证：recv(rax)+generation(uint64)+startedAt(time.Time 3 word) 共 5 寄存器，返回 void。
  体：generation 校验（cmp [recv+0x20]）→ Running 校验（[recv+0x118]）→ 复制 state 快照
  （duffcopy recv+0x28）→ time.Since(startedAt) 算 elapsed 秒，未逐条翻译。
- `prepareFileLocatorSearch`（0x1407d2ea0）：签名 `(request FileLocatorConfig) (*fileLocatorPreparedSearch, error)`
  → `(request FileLocatorConfig) (fileLocatorPreparedSearch, error)`。
  汇编实证：request 值经栈传递（morestack 无寄存器保存）；返回 prepared 为**值**（返回值结构体
  @[rsp+0x3f8]，request 字段 duffcopy 自 normalizeFileLocatorConfig 结果），非指针；错误路径
  返回 nil prepared + error（newobject 错误 + type 指针）。体：normalizeFileLocatorConfig →
  splitFileLocatorRoots → newFileLocatorStringMatcherInternal(fileName/content) →
  newFileLocatorPathFilter 组装，未逐条翻译。
- `buildFileLocatorPathFilter`（0x1407d3400）：签名 `(cfg FileLocatorConfig) *fileLocatorPathFilter`
  → `(cfg FileLocatorConfig, remark, value, include, typeStr string) (*fileLocatorPathFilter, error)`。
  汇编实证：cfg 值经栈传递 + remark/value/include/typeStr 4 个 string（8 寄存器）；返回
  (filter 指针 + error 2 寄存器) 非单指针（正常路径 xor eax/ebx/ecx 返回 nil,nil,nil）。remark 未在
  体内使用（dead 参数，由 FileLocatorFilter.Remark 命名推断）；value 用于错误消息 convTstring，
  include 走 splitFileLocatorFilterEntries，typeStr 走 normalizeFileLocatorFilterType。体：
  parseFileLocatorFilterRule→newFileLocatorPathMatcher 循环，未逐条翻译。

## 关键知悉

- `prepareFileLocatorSearch` 返回值结构体 @[rsp+0x3f8] 基址：request@+0x00（duffcopy 自
  normalized request）、roots@+0xf0、fileNameMatcher@+0x108、contentMatcher@+0x110、
  pathFilter@+0x118，与 fileLocatorPreparedSearch 定义逐字段对齐，印证批次 213 的 240B
  FileLocatorConfig 口径。
- `publishProgress` 与 `waitIfPaused`/`finishSearch` 共享同一 generation+startedAt 参数模式，
  filelocator 执行链的参数口径已一致。
- `MatchContent`（0x1407d4ee0）初勘确认 recv+8 参数（非 content string 2 参数），体内检查
  recv+0x20/+0x28 为 mode 字符串（"boolean" 小端 0x6c6f6f62 实证），并尾调
  matchFileLocatorBooleanAcrossFile（8 参数），完整签名待专项。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1390 / P=106 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（2 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filelocator 域剩余 [P]：MatchContent（recv+8 参数）、
  matchFileLocatorContentStreamContext、fileLocatorStreamingLineMatch、
  fileLocatorBooleanAcrossFile 族（0x1407d7dc0/0x1407d81e0/0x1407d86c0/0x1407d8ce0）、
  readFileLocatorPreview 等。
- walkRoot/processFile 的 2 个指针实参类型需沿 WalkDir 闭包 func1（0x1407d10c0）捕获链确证。
- searchPinyinContextWithTombstones（0x140810e60）含排序/合并/去重三段仍待专项。
