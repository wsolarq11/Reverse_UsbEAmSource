# 批次 384 · filesearch 写链上下文 + 桌面小部件投递键 +7（FUNCS 3278）

## 目标

落地 7 个 `[S]` 函数（5 新增 + 2 存根升级），打通 filesearch 持久化写链：

1. `desktopWidgetDeliveryKey` `[S]`（0x1407b0fa0，576B）— 新增，desktopwidgets 域。
2. `writeAllContext` `[S]`（0x1407f6420，352B）— `[S-sig]` 存根升级（修正签名与 io.EOF→io.ErrShortWrite）。
3. `writeVolumeIndexSortedIndicesContext` `[S]`（0x1407f61e0，576B）— 新增。
4. `writeVolumeNameTrigramSignaturesContext` `[S]`（0x1407f7cc0，576B）— 新增。
5. `buildVolumeIndexPersistenceMetaLocked` `[S]`（0x1407f6580，352B）— `[S-sig]` 存根升级。
6. `replaceFileAtomically` `[S]`（0x1407f8140，1280B）— 新增。
7. `saveVolumeIndexMetaContext` `[S]`（0x1407f7f00，576B）— 新增。

## 基线 / 收口

| 指标 | 基线（batch 383 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3273 | 3278 |
| MARKED | 3273 | 3278 |
| S | 1539 | 1546 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1655 | 1653 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1576 | 1583 |
| USABLE | 1577 | 1584 |
| TRUE（S+S-inline+S-sig） | 3231 | 3236 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3278  MARKED=3278  UNMARKED=0  S=1546  S-inline=37  S-eq=1  S-sig=1653  P=41  FAITHFUL=1583  USABLE=1584  TRUE=3236
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1546 + 37 + 1 + 1653 + 41 = 3278 = FUNCS`。
FUNCS 3273→3278（+5）、S 1539→1546（+7）、S-sig 1655→1653（−2，两存根升级）、
TRUE 3231→3236（+5）、UNMARKED=0/P=41 保持。`list_missing.js` 未落地数 354→349（−5）。

## G3 行为（asm 逐地址实证）

### 3.1 writeAllContext [S 0x1407f6420]

签名 `(ctx context.Context, w io.Writer, p []byte) error`（AX/BX=ctx，CX/DI=w 接口
itab/data，SI/R8/R9=p.ptr/len/cap）。@0x1407f6485 循环 `for len(p)>0`：contextErr(ctx)
（@0x1407f6499）→ err 返；`w.Write(p)`（@0x1407f64aa `mov rdx,[rcx+0x18]` 取 itab 方法槽、
@0x1407f64ae 取 data 为 receiver）→ err 返；`n<=0`（@0x1407f64c9 `jle`）→ 返回全局
`io.ErrShortWrite`（@0x1407f64de 载 {0x1411d3100,0x141bc4610}→errorString{0x140c47b7f,11}
="short write"）；`n>len` → runtime.panicSliceB；否则 `p=p[n:]`。修正存根注释 io.EOF 之误。

### 3.2 writeVolumeIndexSortedIndicesContext [S 0x1407f61e0]

签名 `(ctx context.Context, w io.Writer, indices []int32) error`。len==0（@0x1407f620e
`test r8; je`）→nil。makeslice []byte(0x20000)（@0x1407f6237 lea 元素类型 byte
@0x140ae9e80）。外层游标循环，每块 ≤0x8000 索引（@0x1407f62c0 `cmp r11,0x8000`）；
内层 `binary.LittleEndian.PutUint32(buf[i*4:], indices[cursor+i])`（@0x1407f6377
`mov r13d,[rdx+r15*4]` + @0x1407f6341 `mov [rsi+r12],r13d`）；`writeAllContext(ctx,w,buf[:chunk*4])`
（@0x1407f6320，r9d=0x20000 cap）。

### 3.3 writeVolumeNameTrigramSignaturesContext [S 0x1407f7cc0]

签名 `(ctx context.Context, w io.Writer, signatures []uint64) error`。与 3.2 同构：缓冲
0x40000、每块 ≤0x8000、每元素 8 字节小端（@0x1407f7db2 `shl r8,3`、@0x1407f7e57
`mov r13,[rdx+r15*8]`、@0x1407f7e21 `mov [rsi+r12],r13`）。

### 3.4 buildVolumeIndexPersistenceMetaLocked [S 0x1407f6580]

签名 `(v *VolumeIndex, now time.Time) volumeIndexPersistenceMeta`。Version=1、
IndexVersion=0x20003（@0x1407f6626 `movabs rdx,0x2000300000001` 写前 8 字节）；
RootFRN=v.RootFRN(+0x30 @0x1407f664f)、JournalID=v.JournalID(+0x80 @0x1407f6658)、
LastUSN=v.LastUSN(+0x88 @0x1407f6664)；三时间字段 timeToUnixNano(v.GeneratedAt@+0x90、
v.LastMutationAt@+0xa8、now @0x1407f6621)。返回经 8 寄存器（AX/BX/CX/DI/SI/R8/R9/R10）。

### 3.5 replaceFileAtomically [S 0x1407f8140]

签名 `(src, dst string) error`。8 轮循环（@0x1407f81af `cmp rdx,8`）：rename(src,dst) 成功→nil
（@0x1407f81c5）；失败 lastErr=err。CreateTemp(dir, "."+base+"-*.bak")（@0x1407f822c/821d
concatstring3：a0="." @0x1411cac28、a2="-*.bak" @0x140c37668、a1=Base(dst)）；close→remove
占位→rename(dst,tmpName) 备份；rename(src,dst) 正式替换，失败 rename(tmpName,dst) 回滚，回滚
失败返回 fmt.Errorf("替换 %q 失败且回滚旧文件失败（新文件保留在 %q，旧文件保留在 %q）: replace=%v
rollback=%w", dst, src, tmpName, err, rbErr)（@0x1407f858f format @0x140c94f4c，115B）；
成功 remove 备份返 nil。rename(dst,tmpName) 失败且非 os.ErrNotExist（@0x1407f8369 errors.Is）
时 lastErr=err。每轮退避 `time.Duration(i+1)*25ms`（@0x1407f827b `imul 0x17d7840`）。

### 3.6 saveVolumeIndexMetaContext [S 0x1407f7f00]

签名 `(ctx context.Context, path string, meta volumeIndexPersistenceMeta) error`（meta 经栈
传递，@0x1407f7f98 `lea rbx,[rsp+0x80]`）。TrimSpace(path) 空→nil（@0x1407f7f40）；contextErr；
MkdirAll(Dir(path),0o755)（@0x1407f7f7e `mov ecx,0x1ed`）；json.Marshal(meta)（@0x1407f7faf）；
CreateTemp(Dir(path),".meta-*.tmp")（@0x1407f7fe5 pattern @0x140c474d6，11B）；Write(data)
（@0x1407f802f）；Close；replaceFileAtomically(tmpName,path)（@0x1407f808f）。任一失败清临时
文件返错。

### 3.7 desktopWidgetDeliveryKey [S 0x1407b0fa0]

签名 `(entityID, variant string, due time.Time) string`。due 去单调转 UTC（bt wall,0x3f），
RFC3339Nano 格式化（layout "2006-01-02T15:04:05.999999999Z07:00" @0x140c7708d）；
sha256(entityID+"\x00"+variant+"\x00"+timeStr)；按 16 字符表（"esktopWidgets.qw"
@0x140c5711a，.rdata 实证）逐字节高低 nibble 展开 64B 指纹；返回 entityID+"\x02"+指纹。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增/升级 6 个 filesearch 函数，新增 import
  `context`/`io`/`encoding/json`/`fmt`/`path/filepath`；复用已落地 `contextErr`/
  `timeToUnixNano`/`volumeIndexPersistenceMeta`（types_filesearch.go 已存在，56B 布局与
  asm 逐字段对齐）；`writeAllContext` 依赖 `contextErr`（已 [S]）。
- `backend/desktopwidgets_scheduler.go`：新增 `desktopWidgetDeliveryKey`，复用 `crypto/sha256`/
  `time`。
- 全部 `[S]` 标记，逻辑与 asm 逐地址对齐（含 itab 方法槽读取、错误全局解码、concatstring3
  参数映射、元素类型 byte makeslice）；`go build/vet/test` 全仓通过；`gofmt -l ./backend` 无差异。

## 移交（本轮收尾）

7 函数落地（均 [S]）。FUNCS 3278/4754 = 68.95%，FAITHFUL 1583/4754 = 33.30%，
TRUE 3236/4754 = 68.07%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：`volumeIndexReadView.resolvePath` 0x1407f11e0（依赖 resolvePathWithTombstones）；
`openVolumeIndexMappedFile` 0x1407fd500（依赖 volumeIndexMappedFile.Map 签名已证）；
`loadVolumeNameTrigramIndexMappedContext` 0x1407f8d60（写链已通，可反推读链）；
remoteicons 域（`validateRemoteIconCachePath`/`validateRemoteIconTemporaryFile`/
`writeRemoteIconCache`）；`P=41 → [S]` 转换。
