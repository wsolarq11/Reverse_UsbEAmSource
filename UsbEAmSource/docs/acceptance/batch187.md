# 批次 187 — 组件存储读取链底层（3 函数 → [S]）

## 基线 / 收口

| 指标 | 基线（批次 186 收口） | 收口（批次 187） |
|---|---|---|
| FUNCS | 2721 | **2724** |
| S | 1169 | **1172** |
| S-inline | 36 | 36 |
| S-sig | 1401 | 1401 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2606 | **2609** |

SHA256 `A1D0AD9BAB7AC3693CCFFCB127C3F9149D490E7C603D32CFE0F2F4B8E63BC7AA`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（desktopwidgets_store.go，loadUnlocked 依赖链底三层）

### 1. `readDesktopWidgetStoreBytes`（0x1407c4280，1200B）

`func(path string) ([]byte, error)`。TrimSpace 空→"首页组件存储路径不能为空"（36B @0x140c782eb）；
os.OpenFile(path, O_RDONLY=0, 0) 失败→err；defer Close；Stat 失败→err；
Mode()&os.ModeType!=0（test eax,0x8f280000）→"路径不是普通文件"（24B @0x140c64e90）；
Size()>0x2000000→fmt.Sprintf("文件超过 %d 字节"@0x140c611fa,22B, 0x2000000@0x1411cd548)；
io.ReadAll(io.LimitReader(f, 0x2000001)) 失败→err；len>0x2000000→同超限错误；否则 (data, nil)。

### 2. `ensureDesktopWidgetJSONEOF`（0x1407c4e20，256B）

`func(dec *json.Decoder) error`。dec.Decode(&interface{} /*type size=0x10 @0x140b1a580*/)；
errors.Is(err, io.EOF)→nil；err==nil→"JSON 包含多个顶层值"（26B @0x140c6829f）；
否则 fmt.Errorf("JSON 尾部数据无效: %w"@0x140c69cf0,27B, err)。

### 3. `cloneDesktopWidgetDocument`（0x1407c34e0，1120B）

`func(doc DesktopWidgetDocument) DesktopWidgetDocument`。doc=normalizeDesktopWidgetDocument(doc)；
json.Marshal(doc)（convT 转 interface{}）失败→normalizeDesktopWidgetDocument(零值)；
json.Unmarshal(data,&out) 失败→同零值回退；成功→normalizeDesktopWidgetDocument(out)。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.939s`。
- **G2 count_funcs**：`FUNCS=2724 / S=1172 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`。
- **G3 行为**：所有字符串常量与 os.OpenFile/io.LimitReader/errors.Is/convT/json.Marshal/json.Unmarshal 调用
  经 asm 逐条对齐；0x2000000（32MB）上限与 0x2000001（LimitReader N）常量从立即数与 .rdata 全局字节级解码。
- **G4 review**：仅改 desktopwidgets_store.go（追加 3 函数 + import encoding/json/errors/io/os）；无跨文件写重叠。

## 遗留（下一批）

- `validateDesktopWidgetDocument`（0x1407c3940，2368B，多 map 遍历校验 + memequal + decoderune）——loadUnlocked 最后一块依赖。
- `launcherWidgetStore.Read`（0x1407c0420）+ `loadUnlocked`（0x1407c2440）：依赖已基本就绪，待 validate 落地后闭环。
- `bytesMatchWildcardFold` 0x1407e3e60（176 行，通配符匹配）。
- 继续落地 truly-missing 顶层函数。
