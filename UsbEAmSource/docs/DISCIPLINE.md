# 还原纪律与实证错误记录

> 从 `HANDOFF.md` 拆出。逐寄存器还原纪律 + 累计 77 处实证错误纠正的唯一真相源。每批开工前必读。

## 4. 关键纪律（55+ 处错误换来的，每批开工前必读）

1. **字段偏移只能由 asm 偏移反查 struct 字段序**，不可由字段名反推。
2. **签名未定型拒绝落体**，宁留 `[S-sig]`/`[P]` 并写明阻断原因，不制造伪 `[S]`。
3. **`addr_tool <substr>` 是子串匹配且同名 basename 静默覆盖** —— 不同属主同名方法必须分目录 dump，并核对 dumped bytes 与符号表 `len=` 一致。
4. **上一轮的落体必须重审，不得当作既成事实**。批次 21 发现 6 处手令字段引用偏差（`Apps`/`AppEntries`、`ID`/`AppID`、`ConfigFile`/`DefaultConfigPath` 等）。
5. **锁内读共享变量 vs 锁后读直接影响语义等价性**——asm 中 `mov [rcx+off]` 发生于 unlock（`lock xadd`）之前还是之后是界定「快照」与「无条件读」的分水岭。
6. **开工第一件事：确认 asm 实录真实存在**。asm 在 `docs/goresym/pipeline/tmp/*.asm.txt`，bin 在 `docs/goresym/pipeline/tmp/*.bin`（410 asm / 646 bin）。缺的用 `va_dump.py` 现场抽。
7. **方法覆盖丢失**：替换整文件时注意其他域方法也被删除，批次 22 因此丢失 12 个 windowManagement/mouseGesture 骨架方法。

### 纪律约束（新会话开工前必读）

1. **缓存函数行为已被精确还原**：`cachedLauncherConfigIconAssetURL` 返回单值 `string`（非 `(string, bool)`），含 LRU Touch 语义（命中刷新 Sequence）。`cacheLauncherConfigIconAssetURL` 不接受空 url（直接返回），驱逐策略为 Linear Scan 找最小 Sequence（非随机删）。
2. **分隔符 **`launcherConfigIconAssetCacheKey` 使用 `\x00`（NUL 字节），经指令字节 + `.reloc` 表 PE 字节级确证。不是 `"/"` 也不是 `"|"`。
3. **基本纪律**：asm 直译优先于语义推断。每段翻译前先列出 asm 的显式路径、跳转条件、寄存器分配，准确后再输出 Go 代码。
4. **签名未定型拒绝落体**，宁留 `[S-sig]`/`[P]` 并写明阻断原因，不制造伪 `[S]`。
5. **锁内读共享变量 vs 锁后读直接影响语义等价性**——asm 中 `mov [rcx+off]` 发生于 unlock（`lock xadd`）之前还是之后是界定「快照」与「无条件读」的分水岭。**批次 33 实证**：`ReadBytes`/`ServeAssetRequest` 自身不持锁，由 `lookup` 内部完成锁获取/释放，之后在临界区外读 `s.entries[id]`；照搬 `lock.Lock()` 包裹会**直接死锁**（`sync.Mutex` 不可重入）。
6. **开工第一件事：确认 asm 实录真实存在**。asm 在 `docs/goresym/pipeline/tmp/*.asm.txt`，bin 同目录 `*.bin`（2026-09-20 实测：**489 asm.txt / 709 bin**）。另有 `docs/goresym/disasm/`（批次 29/31 产出）、`disasm_assemble/`、`disasm_assemble_image/`、`disasm_lu/`、`disasm_archive/`、`dump_archive/` 六份资产目录，**同名文件可能内容不同，取用前先确认哪一份最新最完整**。缺的用 `va_dump.py` 现场抽（命令见本节末模板）。
7. **方法覆盖丢失**：替换整文件时注意其他域方法也被删除，批次 22 因此丢失 12 个 windowManagement/mouseGesture 骨架方法。
8. [批次 31 新增] **驱逐的实证签名两次反转**：`evictOldestLocked` 第一参从 `exceptID string`→`namespaceFilter string`（命名空间筛选而非 ex-id），`ensureCapacityLocked` 第一参同改为 `namespace string`（非 itemSize/id），三层驱逐循环代替单层交织。开工后务必先对比 disasm 确认当前代码签名再修改。
9. [批次 33 新增] **`disasm/` 目录存在指令丢失**：`ReadBytes.dis.txt`/`readFileBounded.dis.txt` 等文件的 duffcopy 段（`call 0x140081a46` 前后）**缺失参数装载指令**，只靠该目录会误判"栈槽无写入"。关键位置（尤其决定签名的寄存器装配）必须用 `va_dump.py` 重新现场抽取做字节级比对。

## 9b. 实证错误记录（累计 77 处）

| 批次 | 错误 | 纠正 |
|---|---|---|
| 19 | 锁后读 config.Mode — 并发写者覆盖 | 锁内取快照再 unlock |
| 20 | — | — |
| 21 | 6 处字段引用偏差：`AppEntries→Apps`、`AppID→ID`、`DefaultConfigPath→ConfigFile`、`normalizeAppEntryType(entry)→normalizeAppEntryType(entry.EntryType)`、签名 int→string | 全部修正 |
| 22 | oledblackout.go 覆盖导致 12 个 windowManagement/mouseGesture 骨架丢失 | 补回 service_configure_stubs.go |
| 22 | GetScreens 签名 interface{}→[]OLEDBlackoutScreen | 三寄存器返回 = slice |
| 22 | ToggleProfile() 无参→ToggleProfile(string) | 入口三寄存器含 rbx/rsi = string |
| 23 | — | — |
| 24 | emitSearchCategoryShortcut 署名方法→包级函数(interface{},string) | 入口 rdi 为参4槽位(窗口 itab)，非第四参；纠正签名 |
| 24 | searchCategoryShortcutSuppressionHotkeyForLauncher 签名 bool→(string,bool) | handleHotkeyAction 的 test cl 确认返回三寄存器 |
| 24 | summon/toggleLauncherByHotkey 无参→带窗口 interface | morestack 保存 rbx/rcx 即窗口 itab/data |
| 24 | normalizeShortcutBindingWithOptions 排序字典序→权重降序 | sortShortcutModifiers.func1 setg = weight 降序 |
| 27 | prependScreenshotHistoryEntry 吞 load error | return err 透传 |
| 27 | clearScreenshotHistory 空桩→缺参缺体 | 重写为 70 行实证体 |
| 27 | ClearScreenshotHistory 伪 [S]（无 gate）| 加 beginWorkspaceDataOperation + wsSnapshot |
| 27 | loadScreenshotHistory 漏 clearCutoff + merge | 补全调用 |
| 28 | attachOLEDBlackoutConfigIconURLs 完全缺失 | 新增 [S] 体（164 行 asm → 16L Go），配套 buildLauncherConfigIconResource [P] 存根 |
| 28 | buildLauncherConfigIconResource 完全缺失 | 新增 [P] 存根（232 行 asm，5+ attach* 调用共享） |
| 29 | 分隔符推断为 `"/"`/`"|"` → 实为 `"\x00"`（NUL） | 经指令字节 + `.reloc` 表 PE 字节级双重确证 |
| 29 | `cachedLauncherConfigIconAssetURL` 返回 `(string,bool)` → 单值 `string` | asm 调用点 `test rbx,rbx` 用返回值长度判命中 |
| 29 | fail 路径写入空 URL 到缓存（不存在 asm 逻辑） | asm 无任何失败回写路径 |
| 29 | `buildLauncherIconResource` 返回 `{IconRef:ref.ID, ...}` → `{IconData:trimmed, ...}` | asm 三条出口 duffcopy 确证 |
| 29 | LRU 驱逐「随机删 entry」→「Linear Scan min Sequence」 | `mapIter` 遍历确证 |

---

