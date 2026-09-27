# 签名实证方法论（[P] → [S-sig] 升档唯一依据）

> 目的：把 backend 里 `[P]` 存根（签名未实证、体为零骨架）逐函数做**签名实证**，
> 升为 `[S-sig]`（签名实证、体保持零骨架）。禁止臆造，禁止把 [P] 直接改成 [S]。
> 铁律：签名不确定就保持 [P] 并注明阻断原因。

## 环境

- 工作目录：`D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource`
- 目标 exe：`D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe`
- 符号表：`docs/goresym/symbols.txt`（格式 `0xVA main.符号名`，符号名带 `main.` 前缀）

## 工具

```powershell
# 1. 查 VA（精确匹配行尾）
Select-String -Path docs\goresym\symbols.txt -Pattern 'main\.<函数名>$'

# 2. dump asm（长度 = symbols.txt 下一符号 VA - 当前 VA）
python tools/go-introspect/va_dump.py <exe> 0xVA <len> docs/goresym/pipeline/tmp/<前缀>
# 产物：<前缀>.asm.txt（反汇编，call/jmp 目标自动批注符号名）
```

## Go 1.25 AMD64 内部 ABI（签名判定依据）

- 参数寄存器序：`AX, BX, CX, DI, SI, R8, R9, R10, R11`
- `int/uint/uintptr/pointer/bool` 各占 1 寄存器
- `string` 占 2 寄存器（ptr + len）
- `slice` 占 3 寄存器（ptr + len + cap）
- `interface` 占 2 寄存器（itab + data）
- 结构体按字段序展开为寄存器（每个标量字段占相应宽度）
- 返回序：`AX, BX, CX, DI, SI`（依次）

## 判定方法（读 morestack 序言 + 函数体）

函数开头 `jbe` 后的 `call runtime.morestack` 之前，会有一串 `mov [rsp+..], reg` 保存参数寄存器。
**这些被保存的寄存器就是参数占用的全部寄存器**，据此数出参数宽度。

- `test reg, reg; je/jz` → nil/零值检查（指针 / interface.itab / 数值 / string.len）
- `mov rcx, [rcx+off]; mov <另一个reg>, <data>; call rcx` → **接口方法调用**（itab.fun[off]，data 作 receiver）
- `call runtime.memequal` → string/slice 比较
- `call convT32 / convT64` → 整型转 interface
- `call strings.ToLower / TrimSpace / ...` → string 处理
- `call runtime.newobject` + 写字段 → 结构体/闭包分配

## 升档规则

1. 只做**签名实证**（参数类型 + 返回值），体保持零骨架（`_ = x; return 零值`）。
2. 标记格式（**紧贴 func，中间无空行**，否则 count_funcs.sh 误判 UNMARKED）：
   ```go
   // [S-sig 0xVA] 序言实证：morestack 保存 AX/BX=...，判 nil=...；体语义=...。
   func xxx(...) ... {
       ...
   }
   ```
3. 若签名无法确证 → 保持 `[P]`，更新注释写明阻断原因。
4. 若发现现有签名错误 → 改正签名，但**先 grep 确认无其他调用点**再改；改签名后必须保证 `go build` 通过。
5. 只改分配给自己的文件，不碰其他文件。

## 验收（由 Lead 统一执行）

- `bash tools/count_funcs.sh`：P 减少、S-sig 增加、UNMARKED 保持 0、FUNCS 不变
- `go1.25.12 build ./backend` EXIT=0
