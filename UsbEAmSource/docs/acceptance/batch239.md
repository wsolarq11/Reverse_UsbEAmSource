# 批次 239 — screenshotCOMQueryWorkerPool.enqueue 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 238 收口） | 收口（批次 239） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1443 | **1444** |
| P | 53 | **52** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2735（57.53%） | **2736**（57.55%） |

SHA256 `cedc2c3fc2adca50cedf21ec3f2ac7e3c1a6a70e81dc2c125d2f6cb784f04987`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_worker_windows.go`：`(*screenshotCOMQueryWorkerPool).enqueue` 升档
`[S-sig]`，签名 `(req *screenshotCOMQueryRequest)` →
`(x, y int32, targetWindow uintptr, priority uint8, busyError error) (*screenshotCOMQueryRequest, error)`。

### 签名证据（0x1409b44a0）

morestack 序言（第 258-273 行）保存 7 槽（rax=接收者 + 6 非接收者参数），逐寄存器实证：

| 项 | 类型 | 证据 |
|---|---|---|
| rax | 接收者 *screenshotCOMQueryWorkerPool | 读 [rax+0x58]/[rax+0x68]/[rax+0xc8] 等 pool 字段 |
| ebx | x int32 | 存 dword，写 req+0x10（payload.point.X） |
| ecx | y int32 | 存 dword，写 req+0x14（payload.point.Y） |
| rdi | targetWindow uintptr | 写 req+0x18（payload.targetWindow） |
| sil | priority uint8 | 写 req+0x08（priority 字段） |
| r8/r9 | busyError error（itab/data） | 0x1409b44e7 `test r8,r8` 判 nil；nil 则 `fmt.Errorf` 构造默认错误；写 req+0x20/+0x28 |

返回三寄存器实证：
- 成功路径（0x1409b48e7）`rax=req 指针 + xor rbx/rcx` = `(req, nil)`；
- busy/shutdown 路径（0x1409b460c / 0x1409b465d / 0x1409b47a1）`rax=0 + rbx/rcx=error` = `(nil, err)`。

故返回 `(*screenshotCOMQueryRequest, error)`。

### 关键纠正

旧注释「参数数量与结构超出单 req 指针，无法唯一确定」已解：入参不是单一 req 指针，而是把
`payload.point.X/Y`、`payload.targetWindow`、`priority`、`busyError` 作为 6 个独立标量传入，
体内 newobject 组装 `screenshotCOMQueryRequest`（id 取自 pool.nextRequestID，state=0，
events=makechan(2)）。返回是「请求对象 + 错误」二元组，非空返回 void。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1444 / P=52 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return nil,nil`），无调用方，无行为变更；
  既有测试全量 PASS。
- **G4 review**：`screenshot_uia_worker_windows.go`（1 升档）。

## 遗留（下一批）

- P 已降至 52。截图域剩余复杂 P：captureScreenshotAreaPNG（7 参数）、
  captureScreenshotWindowSelectionWithOptions（5 参数 + WindowProcessPickResult 大结构返回）、
  newScreenshotWindowSelectionSession、resolveControlHoverForWindow、screenshotAccessibleHitTest。
- launcherupdate_runtime.go 9 个 [P]（beginLauncherUpdateTask 6 寄存器返回结构等）。
- filelocator walkRoot/processFile；oledblackout_windows 25；nativedrag createTemporaryDirectoryShortcut。
