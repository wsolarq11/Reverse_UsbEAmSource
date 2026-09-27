#!/usr/bin/env python
# 研究用途：从反汇编文本批量解析 RIP 相对 lea 字符串常量。
#
# 动机：手算 `lea rax,[rip+disp]` 的目标 VA 极易出错（批次 35 实测两次偏差），
# 本工具按指令地址 + 指令长度确定性换算，并用紧随其后的长度立即数解码，
# 消除人工算术导致的常量误读。
#
# 用法:
#   python resolve_lea_strings.py <exe> <asm.txt> [more.asm.txt ...]
# 输出:
#   TAB 分隔：asm行号 / lea地址 / 目标VA / 长度 / 解码文本
import os
import re
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

LINE_RE = re.compile(r"^(0x[0-9a-fA-F]+):\s+(\S+)\s+(.*)$")
LEA_RE = re.compile(r"^lea\s+([a-z0-9]+),\s*\[rip\s*([+-])\s*(0x[0-9a-fA-F]+)\]")
IMM_RE = re.compile(r"^mov\s+e[a-z0-9]+,\s*(0x[0-9a-fA-F]+)")


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


def parse_asm(path):
    """返回 [(lineno, addr, mnem, ops)]，忽略非指令行。"""
    rows = []
    with open(path, encoding="utf-8", errors="replace") as fh:
        for idx, line in enumerate(fh, 1):
            m = LINE_RE.match(line.strip())
            if not m:
                continue
            rows.append((idx, int(m.group(1), 16), m.group(2), m.group(3)))
    return rows


def length_of(mnem, ops):
    """x86-64 指令长度估算：仅覆盖本流水线用到的指令族。"""
    if mnem == "lea":
        return 7
    if mnem == "call":
        return 5
    if mnem in ("jmp", "jbe", "je", "jne", "jge", "jg", "jl", "jle", "jb", "ja", "js", "jns"):
        # 短跳转（8 位）出现在 +0x7f 以内的目标，此处按近跳转默认 6 字节
        return 6
    if mnem == "mov" and ops.startswith("qword ptr [rsp"):
        return 8
    if mnem == "mov" and ops.startswith("e"):
        return 5
    if mnem.startswith("nop"):
        return 4
    if mnem == "ret":
        return 1
    return 4


def resolve(exe_data, rows, lookahead=6):
    """对每个 lea rip 常量，向后找最近的长度立即数并解码。"""
    out = []
    for i, (lineno, addr, mnem, ops) in enumerate(rows):
        m = LEA_RE.match("%s %s" % (mnem, ops))
        if not m:
            continue
        sign, disp_s = m.group(2), m.group(3)
        disp = int(disp_s, 16)
        if sign == "-":
            disp = -disp
        next_ip = addr + length_of(mnem, ops)
        target = next_ip + disp

        length = None
        length_line = None
        for j in range(i + 1, min(i + 1 + lookahead, len(rows))):
            n_lineno, _n_addr, n_mnem, n_ops = rows[j]
            if n_mnem == "call":
                break
            im = IMM_RE.match("%s %s" % (n_mnem, n_ops))
            if im:
                length = int(im.group(1), 16)
                length_line = n_lineno
                break

        text = ""
        if length is not None:
            off = va_to_off(exe_data, target)
            if off is not None and 0 < length <= 512:
                text = decode(exe_data[off:off + length])
        out.append((lineno, addr, target, length, length_line, text))
    return out


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    exe = sys.argv[1]
    data = open(exe, "rb").read()
    for asm_path in sys.argv[2:]:
        print("### %s" % os.path.basename(asm_path))
        for lineno, addr, target, length, llen, text in resolve(data, parse_asm(asm_path)):
            print("L%-4d lea@0x%x -> 0x%x len=%s (from L%s)  %s"
                  % (lineno, addr, target, length, llen, text))
        print()
    return 0


if __name__ == "__main__":
    sys.exit(main())
