#!/usr/bin/env python3
# 研究用途：从 asm 文件抽取字符串常量。
# 模式：lea rX, [rip + disp] 后紧跟 mov eXx, imm（长度）。
# 用法: extract_str.py <exe> <asm.txt>
import re
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass


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


def main():
    exe, asmf = sys.argv[1], sys.argv[2]
    data = open(exe, "rb").read()
    lines = open(asmf, encoding="utf-8", errors="replace").read().splitlines()
    for i, line in enumerate(lines):
        m = re.search(r"0x([0-9a-f]+):\s+lea\s+(\w+),\s*\[rip \+ (0x[0-9a-f]+)\]", line)
        if not m:
            continue
        addr = int(m.group(1), 16)
        disp = int(m.group(3), 16)
        # 找后续 3 行内的长度立即数
        length = None
        for j in range(i + 1, min(i + 6, len(lines))):
            lm = re.search(r"mov\s+e\w+, (0x[0-9a-f]+)", lines[j])
            if lm:
                length = int(lm.group(1), 16)
                break
            if re.search(r"\b(call|ret|jmp|je|jne|jl|jg|jle|jge)\b", lines[j]):
                break
        if length is None or length <= 0 or length > 4096:
            continue
        target = addr + 7 + disp
        off = va_to_off(data, target)
        if off is None:
            continue
        raw = data[off:off + length]
        try:
            s = raw.decode("utf-8")
        except UnicodeDecodeError:
            s = raw.decode("utf-8", "replace")
        # 只跳过含 NUL 的二进制串；UTF-8 中文允许。
        if b"\x00" in raw:
            continue
        print("0x%x len=%d -> %r" % (target, length, s))


if __name__ == "__main__":
    main()
