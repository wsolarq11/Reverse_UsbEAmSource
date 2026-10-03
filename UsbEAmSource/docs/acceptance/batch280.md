# 批次 280 · 琐碎长尾拉满（+6 [S]，FUNCS 2904）

## 目标

沿 gap_aggregate 按 asm 长度升序落地零依赖琐碎纯函数，拉满 FUNCS：
拼音音节规范化、输入监视 owner 规范化、远程图标集合名解析、远程图标缓存目录
创建/校验、桌面小部件调度器唤醒。

## 基线 / 收口

| 指标 | 基线（batch 279 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2898 | 2904 |
| MARKED | 2898 | 2904 |
| S | 1377 | 1383 |
| S-inline | 37 | 37 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1414 | 1420 |
| USABLE | 1414 | 1420 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1420  FUNCS=2904  MARKED=2904  P=41  S-eq=0  S-inline=37  S-sig=1443  S=1383  USABLE=1420
```

分项自洽：`S + S-inline + S-sig + P = 1383 + 37 + 1443 + 41 = 2904 = FUNCS`。
S 1377→1383（+6）、FUNCS 2898→2904（+6）、FAITHFUL 1414→1420（+6）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizePinyinSyllable [S 0x14080b1e0, 160B]

`(s string) string`。asm：TrimSpace → ToLower → `Replace("u:","v",-1)` → `Replace("ü","v",-1)`。
字符串字节实证：old1=`75 3a`("u:")、new1=`76`("v")、old2=`c3 bc`("ü" UTF-8)、new2=`76`("v")；
两组 old len=2（edi=2）、new len=1（r8d=1）、n=-1（r9=0xffffffffffffffff）。

### 3.2 normalizeInputMonitorOwner [S 0x140861ea0, 160B]

`(s string) (string, error)`。asm：TrimSpace → `len==0`(test rbx) 或 `len>0x100`(cmp rbx,0x100)
→ newobject error("输入监测 owner 无效"，25B UTF-8 `e8 be 93 e5 85 a5 e7 9b 91 e6 b5 8b 20 6f 77 6e 65 72 20 e6 97 a0 e6 95 88`)
→ 返回 ("", err)；否则返回 (trimmed, nil)。

### 3.3 resolveRemoteIconCollectionName [S 0x1409661a0, 160B]

`(name string, collections map[string]remoteIconifyCollectionRef) string`。
asm：mapaccess2_faststr 查 name；bl=found==0→返回 name；命中→TrimSpace(value)（value=ref.Name 16B）
非空→返回 trimmed，否则返回 name。调用者 0x14095fe87（searchRemoteIcons.func1）以
`ecx=[rsp+0x98]`（collections map，源自 +0x50 捕获字段）确认参数顺序 name 在前、map 在后。

### 3.4 createRemoteIconCacheDirectory [S 0x140965120, 160B]

`(path string) error`。asm：`os.Mkdir(path, 0x1ed=0o755)`；err!=nil 且非 `errors.Is(err, fs.ErrExist)`
→ 返回 err；否则 `validateRemoteIconCacheDirectory(path)`。

### 3.5 validateRemoteIconCacheDirectory [S 0x1409651c0, 224B]

`(path string) error`。asm：`os.Lstat` → `fi.IsDir()`(itab+0x18) → `fi.Mode()`(itab+0x28) `bt eax,0x1b`
(ModeSymlink) → `remoteIconPathHasReparsePoint`；任一失败/非目录/symlink/reparse 返回同一全局错误
"远程图标缓存路径不安全"（33B UTF-8 `e8 bf 9c e7 a8 8b e5 9b be e6 a0 87 e7 bc 93 e5 ad 98 e8 b7 af e5 be 84 e4 b8 8d e5 ae 89 e5 85 a8`，
全局 error itab/data @ 0x141bc3d70/0x141bc3d78）。

### 3.6 (*desktopWidgetScheduler)Wake [S 0x1407accc0, 96B]

asm：nil receiver→ret；`[rax+0x08]` 取 wake chan（desktopWidgetScheduler.wake @ +0x08）；
`runtime.selectnbsend` 非阻塞发送（返回值被忽略 → select default 分支）。

## G4 独立复核

- `filesearch_pinyin_windows.go`：+`normalizePinyinSyllable`（import 加 strings）。
- `inputmonitor.go`：+`normalizeInputMonitorOwner`（import 加 errors/strings）。
- `remoteicons.go`：+`resolveRemoteIconCollectionName`。
- `remoteicons_cache_windows.go`：+`errRemoteIconCachePathUnsafe`、`createRemoteIconCacheDirectory`、
  `validateRemoteIconCacheDirectory`（import 加 os）。
- `desktopwidgets_scheduler.go`：新建，+`Wake`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

6 个琐碎函数全部 [S] 落地，均零外部依赖或依赖已落地符号。FUNCS 2904/4754 = 61.09%。
未落地文件差集 37→36（desktopwidgets_scheduler.go 新建）。下一批继续沿 gap_aggregate
size 升序落地（desktopWidgetScheduler.Start/Stop + oledBlackoutProfileKey + timeFromWindowsTick 等）。
