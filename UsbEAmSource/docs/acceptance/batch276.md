# 批次 276 · desktopwidgets_secrets_windows.go 整文件落地（+2 [S]，+2 FUNCS）

## 目标

整文件落地 `desktopwidgets_secrets_windows.go`（41 个未落地文件差集之一）：桌面小部件凭据
DPAPI 保护/解除保护 2 函数全 `[S]`。

## 基线 / 收口

| 指标 | 基线（batch 275 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2885 | 2887 |
| MARKED | 2885 | 2887 |
| S | 1365 | 1367 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1401 | 1403 |
| USABLE | 1401 | 1403 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run DesktopWidgetSecret` `ok changeme/backend`（含 DPAPI roundtrip）。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1403  FUNCS=2887  MARKED=2887  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1367  USABLE=1403
```

分项自洽：`S + S-inline + S-sig + P = 1367 + 36 + 1443 + 41 = 2887 = FUNCS`。
S 1365→1367（+2）、FUNCS 2885→2887（+2）、FAITHFUL 1401→1403（+2）、P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `desktopwidgets_secrets_windows.go` 2 函数 `[S]`（+2 S +2 FUNCS +2 FAITHFUL）。
6 个 deferwrap 闭包（clear/LocalFree/clear ×2 函数）内联于父函数 defer 语义，不计独立 FUNCS。

## G3 行为（asm 逐地址实证）

### 3.1 protectDesktopWidgetSecret [S 0x1407b1660, 960B]

`(secret string) (string, error)`（rax/rbx=secret；返回 rax/rbx=base64、rcx/rdi=error）。

- 空串（len==0）→ `errors.New("凭据不能为空")`（18B UTF-8 @0x140C59CD2）。
- `data := []byte(secret)`（stringtoslicebyte）；`defer clear(data)`（deferwrap1 0x1407B1AE0，
  memclrNoHeapPointers 清零敏感字节）。
- `CryptProtectData(&DataBlob{Size:len,Data:&data[0]}, nil,nil,0,nil, CRYPTPROTECT_UI_FORBIDDEN=1,
  &out)`（@0x1401953E0，flags r8d=1）；失败 → `("", err)`。
- 成功 → `defer windows.LocalFree(Handle(out.Data))`（deferwrap2 0x1407B1A80，@0x140198400）；
  复制出加密字节（growslice+memmove，长度 out.Size）→ `defer clear(encrypted)`（deferwrap3）；
  `base64.StdEncoding.EncodeToString(encrypted)` → `(base64, nil)`。

### 3.2 unprotectDesktopWidgetSecret [S 0x1407b1b40, 1024B]

`(encrypted string) (string, error)`。

- `base64.StdEncoding.DecodeString`；`err != nil || len == 0` → `errors.New("受保护凭据格式无效")`
  （27B UTF-8 @0x140C69C84）。
- `CryptUnprotectData(&DataBlob{Size:len,Data:&data[0]}, nil,nil,0,nil, CRYPTPROTECT_UI_FORBIDDEN=1,
  &out)`（@0x1401955A0，flags r8d=1）；失败 → `("", err)`。
- 成功 → `defer LocalFree`；复制出明文字节（growslice+memmove）→ `defer clear`；
  `string(decrypted)`（slicebytetostring）→ `(plaintext, nil)`。

### 3.3 依赖实证

- `windows.DataBlob` 布局 Size(+0x0 uint32)/Data(+0x8 *byte)，asm DATA_BLOB 同构（Size@+0xb0、
  Data@+0xb8）。
- `CryptUnprotectData` name 参数为 `**uint16`（nil 透传）。
- 3 个 defer 逆序：clear(复制) → LocalFree(out.Data) → clear(输入)。

## G4 独立复核

`desktopwidgets_secrets_windows.go`：新建，2 函数 `[S]`，base64 StdEncoding + DPAPI + clear/LocalFree
defer 链全 asm 实证。
`desktopwidgets_secrets_windows_test.go`：新建，空串错误、无效格式错误（空/非法字符/padding-only）、
DPAPI roundtrip（本机 + CI Windows runner 均可运行）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

桌面小部件凭据 DPAPI 域整文件闭环：2 函数全 `[S]`，CryptProtectData/CryptUnprotectData 双调用、
flags=CRYPTPROTECT_UI_FORBIDDEN、DataBlob 布局、base64 StdEncoding、敏感字节 clear/LocalFree defer
链、两条错误文案全部 asm/rodata 实证，并带 DPAPI roundtrip 行为测试。P=41 持平。
FUNCS 2887/4754 = 60.73%。未落地文件差集 41→40。
