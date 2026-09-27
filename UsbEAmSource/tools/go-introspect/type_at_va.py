#!/usr/bin/env python3
# type_at_va.py — 解码 Go 1.25 PE 二进制中指定 VA 的 runtime._type 描述符（hash/kind/名称）。
# 研究用途。用法: type_at_va.py <exe> <VA-hex> [更多VA...]
#
# 依据 Go runtime/type.go 的 _type 布局（amd64）：
#   0x00 Size(u64) 0x08 PtrBytes(u64) 0x10 Hash(u32) 0x14 TFlag(u8)
#   0x15 Align(u8) 0x16 FieldAlign(u8) 0x17 Kind(u8) 0x18 Equal(u64)
#   0x20 GCData(u64) 0x28 Str(i32, 相对 types base) 0x2c PtrToThis(i32)
# name 编码（go1.17+）：varint 长度 + 字节；byte0&1 表示 embedded。
import sys, struct
import pefile

KIND_TABLE = {
    1: "Bool", 2: "Int", 3: "Int8", 4: "Int16", 5: "Int32", 6: "Int64",
    7: "Uint", 8: "Uint8", 9: "Uint16", 10: "Uint32", 11: "Uint64",
    12: "Uintptr", 13: "Float32", 14: "Float64", 15: "Complex64", 16: "Complex128",
    17: "Array", 18: "Chan", 19: "Func", 20: "Interface", 21: "Map",
    22: "Ptr", 23: "Slice", 24: "String", 25: "Struct", 26: "UnsafePointer",
}
KIND_MASK = 0x1f


def read_va(pe, data, va, size):
    rva = va - pe.OPTIONAL_HEADER.ImageBase
    for sec in pe.sections:
        if sec.VirtualAddress <= rva < sec.VirtualAddress + max(sec.Misc_VirtualSize, sec.SizeOfRawData):
            off = sec.PointerToRawData + (rva - sec.VirtualAddress)
            return data[off:off + size], off
    return None, None


def read_varint(data, off):
    v = 0
    for i in range(6):
        x = data[off + i]
        v += (x & 0x7F) << (7 * i)
        if x & 0x80 == 0:
            return i + 1, v
    return 6, v


def decode_name(pe, data, types_base, name_off):
    va = types_base + name_off
    raw, _ = read_va(pe, data, va, 256)
    if raw is None:
        return "<unreadable>"
    embedded = (raw[0] & 1) == 1
    n, ln = read_varint(raw, 1)
    s = raw[1 + n:1 + n + ln]
    try:
        out = s.decode("utf-8")
    except UnicodeDecodeError:
        out = s.decode("latin-1")
    if embedded:
        out = "<embedded>" + out
    return out


def find_types_base(pe, data):
    """从 PE 中定位 runtime.types（.rodata 中 moduledata.types 指向的基址）。
    实用近似：go1.25 的 types base 通常紧邻 runtime.types 符号，但无符号表时
    用启发式——扫描 .rdata 首个 kind 合法的 type。这里改为：读取 moduledata 不可得，
    故以调用方传入的 base 为准（见 main 的 --base）。"""
    return None


def main():
    exe = sys.argv[1]
    va_args = [a for a in sys.argv[2:] if not a.startswith("--")]
    base_override = None
    for a in sys.argv[2:]:
        if a.startswith("--base="):
            base_override = int(a.split("=", 1)[1], 16)

    pe = pefile.PE(exe, fast_load=True)
    data = open(exe, "rb").read()

    types_base = base_override
    if types_base is None:
        print("ERROR: 需要 --base=0x... (types base)", file=sys.stderr)
        return 2

    for vas in va_args:
        va = int(vas, 16)
        raw, off = read_va(pe, data, va, 0x30)
        if raw is None:
            print(f"{vas:#x}: <not mapped>")
            continue
        size, ptrbytes = struct.unpack_from("<QQ", raw, 0)
        h, tflag, align, falign, kind = struct.unpack_from("<IBBBB", raw, 0x10)
        str_off = struct.unpack_from("<i", raw, 0x28)[0]
        name = decode_name(pe, data, types_base, str_off)
        print(f"{va:#x}: size=0x{size:x} hash=0x{h:08x} kind={kind & KIND_MASK}"
              f"({KIND_TABLE.get(kind & KIND_MASK, '?')}) name={name!r} file_off=0x{off:x}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
