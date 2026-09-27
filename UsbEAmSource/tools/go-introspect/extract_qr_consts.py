#!/usr/bin/env python
# 研究用途：从已 dump 的 qr asm 中提取 rip 相对引用的 float 常量与字符串字面量。
import os
import re
import struct
import sys

EXE = r"D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
ASMDIR = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\pipeline\tmp\qr"


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


def read_f64(data, va):
    off = va_to_off(data, va)
    if off is None:
        return None
    return struct.unpack_from("<d", data, off)[0]


def main():
    data = open(EXE, "rb").read()
    float_insns = {"movsd", "mulsd", "addsd", "subsd", "divsd", "ucomisd", "maxsd", "minsd", "comisd", "sqrtsd"}
    floats = {}   # target_va -> set of (func, insn_va)
    strings = {}  # target_va -> (len, set of func)

    for fn in sorted(os.listdir(ASMDIR)):
        if not fn.endswith(".asm.txt"):
            continue
        path = os.path.join(ASMDIR, fn)
        for line in open(path, encoding="utf-8", errors="replace"):
            m = re.match(r"0x([0-9a-f]+):\s+(\S+)\s+(.*)", line)
            if not m:
                continue
            addr = int(m.group(1), 16)
            mnem = m.group(2)
            ops = m.group(3)
            # instruction length = next line addr - this addr; approximate via rip-relative disp
            rm = re.search(r"rip \+ (0x[0-9a-f]+)", ops)
            if not rm:
                continue
            disp = int(rm.group(1), 16)
            # find next instruction length: parse from next line in same read
            # For lea/movsd we use length 7 for lea, 8 for movsd (rip-rel xmm), 6 for lea r64.
            if mnem in float_insns:
                tlen = 8  # movsd xmm, [rip+disp32]
                if mnem in ("ucomisd", "comisd"):
                    tlen = 8
                tgt = addr + tlen + disp
                floats.setdefault(tgt, set()).add((fn, addr, mnem))
            elif mnem == "lea":
                # lea r?, [rip+disp32] -> 7 bytes
                tgt = addr + 7 + disp
                strings.setdefault(tgt, set()).add(fn)

    print("=== FLOAT CONSTANTS ===")
    for tgt in sorted(floats):
        v = read_f64(data, tgt)
        if v is None:
            print("0x%x -> <no section>" % tgt)
            continue
        print("0x%x = %.17g" % (tgt, v))
        for fn, addr, mnem in sorted(floats[tgt]):
            print("    %s @0x%x %s" % (fn, addr, mnem))

    print()
    print("=== STRING-LIKE LEA TARGETS (may be type desc or string) ===")
    for tgt in sorted(strings):
        off = va_to_off(data, tgt)
        if off is None:
            print("0x%x -> <no section>" % tgt)
            continue
        raw = data[off:off + 16]
        # heuristic: try decode as string if mostly printable
        s = ""
        printable = True
        for b in raw:
            if 0x20 <= b < 0x7F:
                s += chr(b)
            else:
                break
        print("0x%x : %r" % (tgt, s if s else raw.hex()))
        for fn in sorted(strings[tgt]):
            print("    %s" % fn)
    return 0


if __name__ == "__main__":
    sys.exit(main())
