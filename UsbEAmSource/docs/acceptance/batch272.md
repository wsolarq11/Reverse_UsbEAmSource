# 批次 272 · attachLauncherBackgroundURL 升档 + launcherBackgroundContentTypeForPath 新增（+1 [S-sig]→[S]，+1 新 [S]，-1 [S-sig]，+1 FUNCS）

## 目标

1. **升档 `attachLauncherBackgroundURL`（0x140798600, 387L asm）**：`[S-sig]` 骨架
   `(cfg *LauncherConfig)` → `[S]` 全还原（背景资源 URL 装配 + 缓存命中短路 + 重注册）。
2. **新增 `launcherBackgroundContentTypeForPath`（0x1407998a0, 352B）**：背景域内容类型分派，
   与截图域 `screenshotContentTypeForPath` 孪生。
3. **订正 launcher asset 核心签名**（本批落地 attach 的依赖链，汇编实证）：
   - `RegisterFile`（0x14086ed00）：`(namespace, id, path, version)` →
     `(namespace, path, contentType string, ttl time.Duration)`。
   - `RegisterBytes`（0x14086e220）：`(namespace, id, data, version)` →
     `(namespace, contentType string, data []byte, ttl time.Duration)`。
   - `register`（0x14086f680）：`(namespace, id, data, version)` →
     `(namespace, contentType string, data []byte, filePath string, size int64, ttl time.Duration)`。
4. **订正截图域 RegisterFile 调用点**（namespace/path/contentType/ttl 四参全错，汇编实证订正）：
   `attachScreenshotAssetURL` / `attachScreenshotCaptureAssetURL` / `attachScreenshotThumbnailAssetURL`。

## 基线 / 收口

| 指标 | 基线（batch 271 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2873 | 2874 |
| MARKED | 2873 | 2874 |
| S | 1353 | 1355 |
| S-inline | 36 | 36 |
| S-sig | 1444 | 1443 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1389 | 1391 |
| USABLE | 1389 | 1391 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1391  FUNCS=2874  MARKED=2874  P=40  S-eq=0  S-inline=36  S-sig=1443  S=1355  USABLE=1391
```

分项自洽：`S + S-inline + S-sig + P = 1355 + 36 + 1443 + 40 = 2874 = FUNCS`。
S 1353→1355（+2）、S-sig 1444→1443（-1）、FUNCS 2873→2874（+1）、P=40 持平、UNMARKED=0 保持。

账目：`attachLauncherBackgroundURL` [S-sig]→[S]（-1 [S-sig] +1 [S]）；
新增 `launcherBackgroundContentTypeForPath`（+1 [S] +1 FUNCS）。
`RegisterFile`/`RegisterBytes`/`register`/截图调用点为签名校正（均已是 [S]，无档位迁移）。

## G3 行为（asm 逐地址实证）

### 3.1 attachLauncherBackgroundURL [S 0x140798600, 387L]（attachLauncherBackgroundURL.asm.txt）

签名 `(bs *BootstrapService, cfg *LauncherConfig)`。流程：

- bs/cfg nil 守卫（line 9-13 test rax/rbx → 0x798d91）。
- `normalized := normalizeBackgroundPreference(cfg.Background)`（line 28 duffcopy → 0x14087eae0）。
- `TrimSpace(ImagePath) == ""` → 写回 normalized 返回（line 38-57）。
- `svc := bs.screenshotAssetService()`（line 60，读 bs.assets@+0x398）nil → 写回返回（line 61-76）。
- `cleaned := filepath.Clean(TrimSpace(ImagePath))`（line 77-84，TrimSpace→filepathlite.Clean）。
- `os.Stat(cleaned)`（line 85）→ err/IsDir/Size≤0 → 写回返回（line 88-121；
  IsDir 经 itab fun[0]=+0x18 `test al,al`，Size 经 fun[4]=+0x38 `test rax;jle`）。
- `modifiedAt := info.ModTime().UnixNano()`（line 122-140；itab fun[1]=+0x20 调 ModTime 后
  nano 时间变换：`bt 0x3f` / `imul 0x3b9aca00`=1e9 / `and 0x3fffffff`=nsecMask /
  常量 0xdd7b17f80=wallToInternal、0xa1b203eb3d1a0000=internalToUnix×1e9）。
- `bs.backgroundAssetLock.Lock()`（line 144 lock cmpxchg [bs+0x3a0]）+ defer Unlock（line 158）。
- 缓存命中（line 162-197）：`owner==svc && path==cleaned && size==Size()` 且
  `modifiedAt 相等`，且 `TrimSpace(URL) != ""` 且 `svc.Exists(URL, "background/custom")` → 复用
  缓存 URL 写回 normalized 返回（line 200-254）。
- 未命中重注册（line 255-265）：`RegisterFile("background/custom", cleaned,
  launcherBackgroundContentTypeForPath(cleaned), 0)`；err==nil 更新
  `backgroundAssetOwner/Path/Size/ModifiedAt/URL`（line 289-326）与 `normalized.ImageURL`。
- 写回 normalized（line 327-348）。

字段偏移实证：backgroundAssetLock@+0x3a0、Owner@+0x3a8、Path@+0x3b0(ptr)/+0x3b8(len)、
Size@+0x3c0、ModifiedAt@+0x3c8、URL@+0x3d0(ptr)/+0x3d8(len)（与 types_launcher.go 定义一致）。

### 3.2 launcherBackgroundContentTypeForPath [S 0x1407998a0, 352B]

`filepath.Ext`（内联反向扫后缀，遇 `/` `\` 中止、遇 `.` 截取）→ `strings.ToLower` →
4/5 字节立即数分派：

- `.bmp`(0x706d622e)→"image/bmp"（0x140c3f69a, 9B）
- `.gif`(0x6669672e)→"image/gif"（0x140c3f691, 9B）
- `.png`(0x676e702e)→"image/png"（0x140c3f688, 9B）
- `.jpg`(0x67706a2e)/`.jpeg`(0x65706a2e+'g')→"image/jpeg"（0x140c4404f, 10B）
- `.webp`(0x6265772e+'p')→"image/webp"（0x140c44059, 10B）
- 其余→"application/octet-stream"（0x140c64e30, 24B）

### 3.3 RegisterFile [S 0x14086ed00, 440L]（签名校正）

`(namespace, path, contentType string, ttl time.Duration)`。旧还原
`(namespace, id, path, version)` 错误：第2参实为 path（被 OpenFile 打开）、第3参实为 contentType、
第4参实为 TTL（截图调用点传 0x8bb2c97000=600s、attach 传 0）。

流程：nil 接收者→error（"资源服务尚未初始化"）；normalize ns；`TrimSpace(path)` →
`os.OpenFile(path, O_RDONLY, 0)` 失败 `fmt.Errorf("读取资源文件失败: %w", err)` → defer Close →
`f.Stat`（失败同 wrap）→ `IsDir`→"资源文件不能是目录" → `Size≤0`→"资源文件不能为空" →
`validateItemSize(size)` → contentType 派生（`TrimSpace` 空则 `mime.TypeByExtension(ToLower(Ext(path)))`
空则 "application/octet-stream"）→ `register(ns, ct, nil, path, size, ttl)`。
**不预读文件**：filePath 存 entry，ReadBytes 时经 readFileBounded 回读。

### 3.4 RegisterBytes [S 0x14086e220, 234L]（签名校正）

`(namespace, contentType string, data []byte, ttl time.Duration)`。旧还原 `(namespace, id, data, version)`
错误：第2参实为 contentType（非 id）、第4参实为 TTL。

流程：nil 接收者→error（"资源服务尚未初始化"）；`len(data)==0`→"资源内容不能为空"；
`validateItemSize(len)`；`normalize ns` 空→"资源命名空间不能为空"；mallocgc+memmove 复制 data；
`TrimSpace(contentType)` 空则 `http.DetectContentType(data)` 空则 "application/octet-stream" →
`register(ns, ct, buf, "", len, ttl)`。

### 3.5 register [S 0x14086f680, 309L]（签名校正）

`(namespace, contentType string, data []byte, filePath string, size int64, ttl time.Duration)`。
旧还原 `(namespace, id, data, version)` 错误：id 恒自动生成（newLauncherAssetID + mapaccess2 查重），
version 恒 `s.next`，无 id/version 入参。

流程：lock → currentTime → pruneExpiredLocked → id 恒生成查重 → `ensureCapacityLocked(namespace, size)`
→ next++ → entry{id, namespace, version=next, contentType, data, filePath, size, createdAt=accessedAt=now,
expiresAt=now+ttl（ttl==0→600s，ttl<0→零值永不过期）} → addEntryLocked →
`buildLauncherAssetURL(namespace, id, next)` → ref{ID, URL, Version=next, ContentType=contentType}。

### 3.6 截图域调用点订正

`attachScreenshotAssetURL`（0x140797380）/`attachScreenshotCaptureAssetURL`（0x140797ba0）：
旧 `RegisterFile("screenshot", filepath.Base(trimmed), trimmed, 0)` → 汇编实证（0x140797535/0x140797d52
`lea rip+0x4c272a`/`+0x4c1f0d` → 0x140c59c66）命名空间 18 字符 `"screenshot/current"`，
第2参 = `trimmed`（全路径，非 basename），第3参 = `screenshotContentTypeForPath(trimmed)`，
第4参 = 600s TTL。修正为
`RegisterFile("screenshot/current", trimmed, screenshotContentTypeForPath(trimmed), 600*time.Second)`。
`attachScreenshotThumbnailAssetURL` 同法订正（其汇编实走
buildScreenshotThumbnailPNGFromData/FromPath，非 RegisterFile，属截图域后续专项，本轮仅保形一致）。

## G4 独立复核

`launcherasset.go`：RegisterBytes/RegisterFile/register 三签名+体全订正；新增 `mime`/`path/filepath`
导入。`bootstrapservice_state_deps.go`：attach 骨架升 [S]（新增 `path/filepath`/`strings` 导入）。
`bootstrapservice_migration.go`：新增 `launcherBackgroundContentTypeForPath` [S]。
`screenshot_services.go`/`screenshot_stubs.go`：三处 RegisterFile 调用点订正。
`launcherasset_test.go`：TestRegisterBytes/TestRegisterFile/TestReadBytes/TestServeAssetRequest/
TestRegisterBody/TestRegisterBodyCapacityError 六处适配新签名（id 恒随机、version 恒 next、
contentType 直传、filePath 化 entry）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

launcher asset 背景域闭环：`attachLauncherBackgroundURL` + `launcherBackgroundContentTypeForPath`
全 [S]；其依赖链 `RegisterFile`/`RegisterBytes`/`register` 签名+体按汇编订正（id 恒随机生成、
version 恒 next、contentType 直传、filePath 化 RegisterFile、TTL 600s 默认），截图域三调用点同步订正。
`attachScreenshotThumbnailAssetURL` 的 buildScreenshotThumbnailPNGFromData/FromPath 链（截图域）仍待专项。
P=40 持平。FUNCS 2874/4754 = 60.45%。
