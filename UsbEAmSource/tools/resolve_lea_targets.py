#!/usr/bin/env python
# 研究用途：精确解析 bin 反汇编中的 lea [rip+disp] 目标地址并读字符串。
# 手算 RIP 目标反复出错，本脚本用 capstone 确定性计算，杜绝进位错误。
# 用法: python tools/resolve_lea_targets.py <exe> <bin> <baseVA>
import struct
import sys

import capstone


def va_to_off(data, va):
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


def decode(raw):
    out = []
    for b in raw:
        if 0x20 <= b < 0x7F:
            out.append(chr(b))
        elif b == 0x0A:
            out.append("\\n")
        elif b == 0x00:
            out.append("\\0")
        else:
            out.append("\\x%02x" % b)
    return "".join(out)


def main():
    exe = sys.argv[1]
    binpath = sys.argv[2]
    base = int(sys.argv[3], 0)
    data = open(exe, "rb").read()
    code = open(binpath, "rb").read()

    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    md.detail = True

    for insn in md.disasm(code, base):
        mnemonic = insn.mnemonic
        if mnemonic.startswith("lea"):
            disp = None
            for op in insn.operands:
                if op.type == capstone.x86.X86_OP_MEM:
                    disp = op.mem.disp
            if disp is not None:
                target = insn.address + insn.size + disp
                off = va_to_off(data, target)
                s = ""
                if off is not None:
                    s = repr(decode(data[off:off + 32]))
                print("0x%x lea %-32s -> 0x%x  %s" % (insn.address, insn.op_str, target, s))


if __name__ == "__main__":
    sys.exit(main())
