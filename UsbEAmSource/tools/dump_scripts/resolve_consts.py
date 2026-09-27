#!/usr/bin/env python
"""解析 asm.txt 中的 `lea reg,[rip+disp]`，用「下一条指令地址 + disp」精确算出绝对 VA。
不依赖手工算术，避免进位错误。输出 VA 列表供 read_gostring.py 解码。
"""
import re
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

LEA_RE = re.compile(
    r"^(0x[0-9a-f]+):\s+lea\s+(\w+)\s*,\s*\[rip\s*\+\s*(0x[0-9a-f]+)\]"
)
IMM_RE = re.compile(r"^0x[0-9a-f]+:\s+mov\s+e(\w+)i\s*,\s*(0x[0-9a-f]+)\s*$")
ADDR_RE = re.compile(r"^(0x[0-9a-f]+):")


def parse(path):
    lines = open(path, encoding="utf-8", errors="replace").read().splitlines()
    rows = []
    for i, ln in enumerate(lines):
        m = LEA_RE.match(ln.strip())
        if not m:
            continue
        addr = int(m.group(1), 16)
        disp = int(m.group(3), 16)
        # 指令结束地址 = 下一条指令的地址（跳过 int3 填充等，取最近的下一条带地址行）
        nxt = None
        for j in range(i + 1, min(i + 4, len(lines))):
            a = ADDR_RE.match(lines[j].strip())
            if a:
                nxt = int(a.group(1), 16)
                break
        if nxt is None:
            continue
        absva = nxt + disp
        # 紧随其后的长度立即数（最多看 4 行）
        length = None
        for j in range(i + 1, min(i + 5, len(lines))):
            im = IMM_RE.match(lines[j].strip())
            if im:
                length = int(im.group(2), 16)
                break
        rows.append((addr, m.group(2), nxt, disp, absva, length))
    return rows


def main():
    for path in sys.argv[1:]:
        print("=== %s ===" % path)
        for addr, reg, nxt, disp, absva, length in parse(path):
            tag = "len=0x%x" % length if length is not None else "len=?"
            print(
                "  insn=0x%x reg=%-4s next=0x%x disp=+0x%x -> VA=0x%x  %s"
                % (addr, reg, nxt, disp, absva, tag)
            )
        print()


main()
