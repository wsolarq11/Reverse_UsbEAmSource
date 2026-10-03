# 批次 291 · launcherupdate 帧读写 +2 [S]（FUNCS 2945）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：帧写入（writeLauncherUpdateFrame）
与帧读取（readLauncherUpdateFrame）。

## 基线 / 收口

| 指标 | 基线（batch 290 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2943 | 2945 |
| MARKED | 2943 | 2945 |
| S | 1420 | 1422 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1457 | 1459 |
| USABLE | 1458 | 1460 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1459  FUNCS=2945  MARKED=2945  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1422  USABLE=1460
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1422 + 37 + 1 + 1444 + 41 = 2945 = FUNCS`。
S 1420→1422（+2）、FUNCS 2943→2945（+2）、FAITHFUL 1457→1459（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 writeLauncherUpdateFrame [S 0x1408c31e0, 288B]

`(conn io.Writer, payload []byte) error`。len==0 或 len>0x810000 → "更新认证消息大小无效"；
binary.Write(conn,LittleEndian,uint32(len)) 失败透传；否则 conn.Write(payload) 透传 err。

### 3.2 readLauncherUpdateFrame [S 0x1408c3300, 320B]

`(conn io.Reader, maxSize int) ([]byte, error)`。binary.Read(conn,LittleEndian,&size) 失败透传；
size==0 或 int(size)>maxSize → "更新认证消息大小无效"；io.ReadAtLeast(conn,buf,int(size))
失败透传；成功返回 (buf,nil)。order 常量与 writeFrame 同址（LittleEndian 实证），
data itab 分别为 uint32 值 / *uint32 指针。错误字符串 hex 解码 "更新认证消息大小无效"。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`writeLauncherUpdateFrame` +`readLauncherUpdateFrame`
  （import +encoding/binary +io）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地 [S]。FUNCS 2945/4754 = 61.95%。下一批：launcherupdate 短函数
（createNamedPipe/openNamedPipeClient/frameWithTimeout 352-768B）+ 截图滚动几何。
