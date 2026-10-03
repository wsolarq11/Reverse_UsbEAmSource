# 批次 277 · desktopwidgets_configio.go 整文件落地（+4 [S]，+4 FUNCS）

## 目标

整文件落地 `desktopwidgets_configio.go`（39 个未落地文件差集之一）：配置 I/O 提取/判断/
可移植转换 4 函数全 `[S]`。其蓝图同文件的另 3 函数（read/write/exportSelfContained）已先行
落地于 `bootstrapservice_config.go`，故本批只补 gap 中 4 个未落地函数。

## 基线 / 收口

| 指标 | 基线（batch 276 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2887 | 2891 |
| MARKED | 2887 | 2891 |
| S | 1367 | 1371 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1403 | 1407 |
| USABLE | 1403 | 1407 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run Configio/ConfigDocument/Portable/ExtractDesktopWidget/
LauncherConfigDocument/WidgetDocumentForImport` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1407  FUNCS=2891  MARKED=2891  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1371  USABLE=1407
```

分项自洽：`S + S-inline + S-sig + P = 1371 + 36 + 1443 + 41 = 2891 = FUNCS`。
S 1367→1371（+4）、FUNCS 2887→2891（+4）、FAITHFUL 1403→1407（+4）、P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `desktopwidgets_configio.go` 4 函数 `[S]`（+4 S +4 FUNCS +4 FAITHFUL）。

## G3 行为（asm 逐地址实证）

### 3.1 extractDesktopWidgetDocument [S 0x1407ab8c0, 1344B]

`(data []byte) (DesktopWidgetDocument, bool, error)`（rax/rbx/rcx=data ptr/len/cap；栈返 doc、
rax=ok、rbx/rcx=err）。

- `json.Unmarshal(data, &map[string]json.RawMessage)` 失败 → `(零值,false,err)`。
- 无 `"desktopWidgets"`（14B @0x140C616BE）→ `(零值,false,nil)`；TrimSpace(raw) 空或
  `== "null"`（4B @0x140C4EC8B）→ `(零值,false,nil)`。
- `json.NewDecoder(bytes.NewReader(raw)).Decode(&doc)` 失败 →
  `desktopWidgetStoreError("desktopWidgets", fmt.Sprintf("解析失败: %v"@0x140C6315A,16B, err))`。
- `ensureDesktopWidgetJSONEOF(dec)` 失败 → `desktopWidgetStoreError("desktopWidgets", err.Error())`。
- `normalizeDesktopWidgetDocument(doc)` → `validateDesktopWidgetDocument(doc)` 失败 → `(零值,false,err)`；
  成功 → `(doc, true, nil)`。

### 3.2 launcherConfigDocumentIsWidgetLibrary [S 0x1407abe00, 768B]

`(data []byte) bool`。

- `json.Unmarshal(data, &map[string]json.RawMessage)` 失败 → false；无 `"widgets"`（7B）→ false；
- 遍历 15 白名单键 {initialized, storage, preferences, apps, speedDial, bookmarks, fileSearch,
  fileLocator, files, memoryRelease, oledBlackout, windowManagement, mouseGestures, twoFactor,
  desktopWidgets}，任一存在 → false；全部不存在 → true。

### 3.3 portableDesktopWidgetDocument [S 0x1407ac4a0, 384B]

`(doc DesktopWidgetDocument) DesktopWidgetDocument`。

- `cloneDesktopWidgetDocument(doc)` 后 3×`makemap_small` 覆盖返回结构体偏移 +0x50/+0x58/+0x60
  三字段（WeatherCache/NotificationDeliveries/ProtectedSecrets）→ 清空设备相关运行时缓存。

### 3.4 desktopWidgetDocumentForImport [S 0x1407ac620, 448B]

`(doc DesktopWidgetDocument) DesktopWidgetDocument`。

- `portableDesktopWidgetDocument(doc)` 后取 `cloneDesktopWidgetDocument(doc)` 结果的 +0x60
  （ProtectedSecrets）覆盖回可移植副本 → 导入保留 DPAPI 受保护密钥，其余设备字段仍清空。

### 3.5 依赖实证

- `desktopWidgetStoreError(op,path)=fmt.Errorf("%s: %s: %s","DESKTOP_WIDGET_STORE_INVALID",…)`
  （已落地 0x1407c4f20）；`ensureDesktopWidgetJSONEOF`（0x1407c4e20）、`normalizeDesktopWidgetDocument`
  （0x1407bfec0）、`validateDesktopWidgetDocument`（0x1407c3940）、`cloneDesktopWidgetDocument`
  （0x1407c34e0）均已在 `desktopwidgets_store.go` 落地。

## G4 独立复核

`desktopwidgets_configio.go`：新建，4 函数 `[S]`，Unmarshal 类型 map[string]json.RawMessage、
"desktopWidgets" 键 + "null" 哨兵、15 白名单键、3 字段偏移清空、ProtectedSecrets 恢复全部
asm/rodata 实证。
`desktopwidgets_configio_test.go`：新建，缺键/哨兵/有效提取/解码错误、白名单判定、可移植清空、
导入保留密钥 6 组行为测试。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

桌面小部件配置 I/O 域整文件闭环：4 函数全 `[S]`，提取/判定/可移植/导入四条链路全部
asm/rodata 实证，并带 6 组行为测试。P=41 持平。
FUNCS 2891/4754 = 60.81%。未落地文件差集 40→39。
