# 批次 50 验收（screenshot DXGI COM 薄包装）

日期：2026-09-20
子批次：dxgi_com（Screenshot 运行时族第九子批次）
目标：DXGI 工厂创建 / IDXGIFactory6 查询 / COM 释放 asm 直译——`createDXGIFactory1` /
`queryDXGIFactory6` / `releaseDXGIUnknown` + 辅助 `dxgiUnknownVtable`，新建
`backend/screenshot_dxgi_com_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1026 / MARKED=837 / S=656 / S-inline=6 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 49（1022/833/653/5）：+4 标记（3 [S] + 1 [S-inline]），`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `createDXGIFactory1`（0x14085da00）：`procCreateDXGIFactory1.Find` 失败 → 返回 err；`SyscallN(CreateDXGIFactory1, &IID_IDXGIFactory1, &factory)`；HRESULT 非零 → `fmt.Errorf("%s failed: HRESULT 0x%08X", "CreateDXGIFactory1", uint32(hresult))`；factory 零 → `errors.New("CreateDXGIFactory1 did not return a factory")`。
- `queryDXGIFactory6`（0x14085db80）：factory 零 → `errors.New("DXGI factory is nil")`；`vtable[0]` QueryInterface(`SyscallN(qi, factory, &IID_IDXGIFactory6, &factory6)`)；HRESULT 非零 → `fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIFactory6 unavailable", uint32(hresult))`；factory6 零 → `errors.New("IDXGIFactory6 query returned nil")`。
- `releaseDXGIUnknown`（0x14085dd20）：p/vtable/release 三层零检查；`SyscallN(release, p)`；无返回。
- `dxgiUnknownVtable` 对应 asm `mov rdx,[rax]; mov rax,[rdx]` 内联模式（[S-inline]）。
- GUID 实证：`IID_IDXGIFactory1 = 770AAE78-F26F-4DBA-A829-253C83D1B387`@0x141962a10；`IID_IDXGIFactory6 = C1B6694F-FF09-44A9-B03C-77900A0A1D17`@0x141962a20。
- 字符串实证：`"%s failed: HRESULT 0x%08X"`@0x140c66b0a；`"CreateDXGIFactory1"`@0x140c59ab6；`"IDXGIFactory6 unavailable"`@0x140c66b23；`"CreateDXGIFactory1 did not return a factory"`@0x140c81f53；`"DXGI factory is nil"`@0x140c5bbd0；`"IDXGIFactory6 query returned nil"`@0x140c724a6。

## G4 测试

新增 `backend/screenshot_dxgi_com_windows_test.go`：

- `TestDXGIFactoryIIDs`：两个 IID 的 Data1/Data2/Data3/Data4 全字段字节序断言。
- `TestReleaseDXGIUnknownNilSafe`：`releaseDXGIUnknown(0)` 安全返回。

全 PASS。

## 判定

四路 PASS，批次 50 闭环。
