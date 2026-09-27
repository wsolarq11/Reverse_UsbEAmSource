#!/usr/bin/env python
# 研究用途：批次 54 字符串解码（增强版）。
# 模式 A: lea reg,[rip+disp] 后 6 行内 mov eXx,imm / mov [mem],imm
# 模式 B: newobject 后 mov qword ptr [rax+8],0xLEN 再 lea rcx,[rip+disp]
import glob
import os
import re
import struct
import sys

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
TMP = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "docs", "goresym", "pipeline", "tmp"))


def va_to_off(data, va):
    pe = struct.unpack_from("<I", data, 0x3C)[0]
    ibase = struct.unpack_from("<Q", data, pe + 24 + 24)[0]
    rva = va - ibase
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


data = open(EXE, "rb").read()
files = sorted(glob.glob(os.path.join(TMP, "b54_*.asm.txt")))
seen = {}

for f in files:
    base = os.path.basename(f)
    lines = open(f, encoding="utf-8", errors="replace").read().splitlines()
    for i, line in enumerate(lines):
        # 模式 B: newobject 后 len，再 lea rcx 字符串
        mb = re.search(r"mov\s+qword ptr \[[^\]]+\],\s*(0x[0-9a-f]+)", line)
        if mb:
            length = int(mb.group(1), 16)
            if 0 < length <= 512:
                for j in range(i + 1, min(i + 8, len(lines))):
                    mm = re.search(r"0x([0-9a-f]+):\s+lea\s+rcx,\s*\[rip \+ (0x[0-9a-f]+)\]", lines[j])
                    if mm:
                        addr = int(mm.group(1), 16)
                        disp = int(mm.group(2), 16)
                        target = addr + 7 + disp
                        off = va_to_off(data, target)
                        if off is None:
                            break
                        raw = data[off:off + length]
                        s = raw.decode("utf-8", "replace")
                        key = (target, length)
                        if key not in seen and b"\x00" not in raw:
                            seen[key] = True
                            print("%s B 0x%x len=%d -> %r" % (base, target, length, s))
                        break
                    if re.search(r"\b(call|ret|jmp|je|jne)\b", lines[j]):
                        break
