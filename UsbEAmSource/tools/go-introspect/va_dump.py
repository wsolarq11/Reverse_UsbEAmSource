#!/usr/bin/env python
# 研究用途：按 VA 区间直接从 PE 抽取字节并反汇编（含符号批注）。
#
# 相比 addr_tool.exe 的批量落盘，本工具按 (baseVA, length) 精确定位，
# 因此可以处理 addr_tool 漏掉的闭包符号（其 Name 以 "." 开头，落盘文件名异常），
# 也避免 pipeline/tmp 下同名 bin 互相覆盖。
#
# 用法:
#   python va_dump.py <exe> <baseVA> <length> <outPrefix>
# 产物:
#   <outPrefix>.bin      原始字节
#   <outPrefix>.asm.txt  反汇编（call/jmp 目标自动批注符号）
import os
import re
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

import capstone


def load_symbols():
    sym_path = os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        "..", "..", "docs", "goresym", "symbols.txt",
    )
    table = {}
    if not os.path.exists(sym_path):
        return table
    with open(sym_path, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            try:
                va, name = line.split(" ", 1)
            except ValueError:
                continue
            table[int(va, 16)] = name
    return table


SYMBOLS = load_symbols()
_SORTED = sorted(SYMBOLS.items())


def annotate(target):
    if target in SYMBOLS:
        return SYMBOLS[target]
    best_va, best_name = None, None
    for va, name in _SORTED:
        if va > target:
            break
        best_va, best_name = va, name
    if best_name is not None and target - best_va < 0x2000:
        return "%s+0x%x" % (best_name, target - best_va)
    return None


def va_to_off(data, va):
    """VA(绝对地址) -> 文件偏移；遍历全部节，不硬编码 .rdata。"""
    pe = struct.unpack_from("<I", data, 0x3C)[0]
    image_base = struct.unpack_from("<Q", data, pe + 24 + 24)[0]
    rva = va - image_base
    nsect = struct.unpack_from("<H", data, pe + 6)[0]
    optsz = struct.unpack_from("<H", data, pe + 20)[0]
    sec = pe + 24 + optsz
    for i in range(nsect):
        o = sec + i * 40
        vsize = struct.unpack_from("<I", data, o + 8)[0]
        va0 = struct.unpack_from("<I", data, o + 12)[0]
        rawsz = struct.unpack_from("<I", data, o + 16)[0]
        raw = struct.unpack_from("<I", data, o + 20)[0]
        if va0 <= rva < va0 + max(vsize, rawsz) and raw:
            return raw + (rva - va0)
    return None


def main():
    if len(sys.argv) < 5:
        print(__doc__)
        return 2
    exe = sys.argv[1]
    base = int(sys.argv[2], 0)
    length = int(sys.argv[3], 0)
    out_prefix = sys.argv[4]

    data = open(exe, "rb").read()
    off = va_to_off(data, base)
    if off is None:
        print("未找到 VA 0x%x 所在节" % base)
        return 1
    code = data[off:off + length]
    if len(code) < length:
        print("警告：实际只取到 %d/%d 字节" % (len(code), length))

    with open(out_prefix + ".bin", "wb") as fh:
        fh.write(code)

    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    lines = []
    for insn in md.disasm(code, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            m = re.search(r"0x([0-9a-f]+)", insn.op_str)
            if m:
                name = annotate(int(m.group(1), 16))
                if name:
                    line += "   ; %s" % name
        lines.append(line)

    with open(out_prefix + ".asm.txt", "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))

    print("0x%x +%d -> %s.bin / %s.asm.txt (%d 行)" % (base, length, out_prefix, out_prefix, len(lines)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
