#!/usr/bin/env python
# 研究用途：批次 54 字符串批量解码。
# 对每个 asm.txt 扫描 lea reg,[rip+disp]，在其后 6 行内找长度立即数
# （mov eXx,imm 或 mov [mem],imm 或 mov qword ptr [rax+8],imm），解码 PE 字节。
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
        m = re.search(r"0x([0-9a-f]+):\s+lea\s+(\w+),\s*\[rip \+ (0x[0-9a-f]+)\]", line)
        if not m:
            continue
        addr = int(m.group(1), 16)
        disp = int(m.group(3), 16)
        length = None
        for j in range(i + 1, min(i + 8, len(lines))):
            lm = re.search(r"mov\s+(?:e\w+|qword ptr \[[^\]]+\]|dword ptr \[[^\]]+\]),\s*(0x[0-9a-f]+)", lines[j])
            if lm:
                length = int(lm.group(1), 16)
                break
            if re.search(r"\b(call|ret|jmp|je|jne|jl|jg|jle|jge|ja|jb)\b", lines[j]):
                break
        if length is None or length <= 0 or length > 512:
            continue
        target = addr + 7 + disp
        off = va_to_off(data, target)
        if off is None:
            continue
        raw = data[off:off + length]
        if b"\x00" in raw and len(raw) > 1:
            # 可能是 UTF-16 或二进制，尝试 utf-8 失败则跳过
            try:
                s = raw.decode("utf-8")
            except UnicodeDecodeError:
                continue
        else:
            s = raw.decode("utf-8", "replace")
        key = (target, length)
        if key in seen:
            continue
        seen[key] = True
        print("%s 0x%x len=%d -> %r" % (base, target, length, s))
