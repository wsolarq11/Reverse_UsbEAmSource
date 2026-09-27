# 批次 243 — filelocator walkRoot/processFile 签名订正（2 升档）

## 基线 / 收口

| 指标 | 基线（批次 242 收口） | 收口（批次 243） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1447 | **1449** |
| P | 49 | **47** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2739（57.61%） | **2741**（57.65%） |

SHA256 `8e78592bda4dde01bb5d8daacc2d19c9f1b9431d8503a0c881e67a35003b4a11`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go`：`walkRoot`、`processFile` 双双升档 `[S-sig]`，订正为
`fileLocatorPreparedSearch` 值传 + ctx + generation + path + progress + onFileDone + startedAt 全形参。
`backend/types_filelocator.go`：新增 `fileLocatorSearchProgress` 累加器类型（0x78 字节布局经
processFile 汇编实证，计数器名按 FileLocatorState 推断，字段级语义 [P]）。

### 签名证据

**walkRoot（0x1407d0d20）** morestack 序言（第 124-141 行）保存 8 槽（rax..r10）：

| 项 | 类型 | 证据 |
|---|---|---|
| rax | receiver | 全程透传（0x1407d0d3d 存 [0x378]） |
| rbx/rcx | ctx context.Context | 0x1407d0e42 `mov rcx,[rax+0x18]; call rcx`=FileInfo.IsDir 前的接口方法派发链透传 |
| rdi | generation uint64 | 透传 processFile（0x1407d0feb `mov rdi,[0x390]`） |
| rsi/r8 | path string | 0x1407d0d7d `mov rax,rsi / mov rbx,r8` 作 os.Stat 首参 |
| r9 | progress *fileLocatorSearchProgress | 0x1407d0f02 存 [0x1e8] 闭包捕获、processFile 传 r11 |
| r10 | onFileDone func(bool) | 0x1407d0f42 存 [0x208] 闭包捕获、processFile 栈传 |

stack 两栈块：`prepared`（值传 288B @ [rsp+0x240]，duffcopy 0x1407d0f8a 铺入）+
`startedAt`（time.Time 3 word @ [0x360]/[0x368]/[0x370]）。

**processFile（0x1407d15a0）** morestack 序言（第 599-618 行）保存 9 槽（rax..r11）：

| 项 | 类型 | 证据 |
|---|---|---|
| rax | receiver | 全程透传 |
| rbx/rcx | ctx | waitIfPaused 首参透传（0x1407d1638） |
| rdi | generation | 0x1407d1638 waitIfPaused 第二参 |
| rsi/r8 | path | 0x1407d1699/0x1407d16a1 作 pathFilter.Allows 首参 |
| r9/r10 | info os.FileInfo | 0x1407d16b6 读 `[r9+0x30]` 接口方法槽、data=r10 |
| r11 | progress *fileLocatorSearchProgress | 0x1407d15d7 存 [0x458]，各计数器自增（+0x10/+0x18/…） |

stack：`prepared`（值传 288B）+ `onFileDone`（func(bool)，0x1407d20e5 `mov rdx,[0x3f8]; mov rcx,[rdx]; mov eax,1; call rcx`）+ `startedAt`（3 word）。

### 关键纠正

- **旧签名整体错误**：`walkRoot(prepared *..., root string, ctx)` 与 `processFile(prepared *..., path string, ctx) (bool, error)` 均被推翻。
  真实形参为 `(prepared 值传, ctx, generation, path, progress, onFileDone, startedAt)`（processFile 多一个 `info os.FileInfo`）。
- **processFile 返回 `error` 非 `(bool, error)`**：所有尾迹只清 rax/rbx 两字（error 接口），无 bool 位。
- `prepared` 是**值传**（288B 大结构走栈），非指针；旧桩误作指针。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1449 / P=47 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return nil`），无调用方，无行为变更；既有测试全量 PASS。
- **G4 review**：`filelocator_runtime.go`（2 升档）+ `types_filelocator.go`（新增 1 类型）。

## 遗留（下一批）

- P 已降至 47。launcherupdate_runtime.go 剩余 6 个 [P]（runLauncherUpdateTask /
  fetchLauncherRemoteConfigWithWorkspace / beginLauncherUpdateTask /
  prepareLauncherUpdatePackageWithWorkspace / downloadLauncherUpdatePackageWithWorkspace /
  downloadLauncherUpdatePackageConcurrently——该族签名与 downloadLauncherUpdatePackageWithWorkspace
  的栈参互相耦合，需先落地调用方全形参）。
- 截图域、oledblackout_windows（25）、nativedrag createTemporaryDirectoryShortcut。
- §10 差集 54 文件；FUNCS 缺口核心是 697 个未落地顶层函数 + 闭包/方法。
