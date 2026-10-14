#!/usr/bin/env python3
# va_disasm.py — 以 VA 为基址反汇编函数 bin，并对 call/jmp 目标做符号标注。
# 研究用途。依赖 capstone；符号表 docs/goresym/symbols.txt。
#
# 用法: va_disasm.py <binfile> <baseVA> [limit] [out]
#   baseVA 为该 bin 首个字节的虚拟地址（须 0x 十六进制）。
#
# 符号批注的唯一实现在 tools/index/common.py（annotate_symbol，窗口 0x2000）。
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "index"))

import common  # noqa: E402
from capstone import Cs, CS_ARCH_X86, CS_MODE_64  # noqa: E402

HEX_IN_OPS = re.compile(r"0x([0-9a-f]+)")


def main():
    if len(sys.argv) < 3:
        print("usage: va_disasm.py <binfile> <baseVA> [limit] [out]", file=sys.stderr)
        sys.exit(2)
    binf = sys.argv[1]
    base = int(sys.argv[2], 0)
    lim = int(sys.argv[3], 0) if len(sys.argv) > 3 else 10 ** 9
    out = sys.argv[4] if len(sys.argv) > 4 else None

    data = open(binf, "rb").read()
    md = Cs(CS_ARCH_X86, CS_MODE_64)
    md.detail = False
    lines = []
    for insn in md.disasm(data, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith(("call", "j", "loop")):
            m = HEX_IN_OPS.search(insn.op_str)
            if m:
                s = common.annotate_symbol(int(m.group(1), 16))
                if s:
                    line += "    ; %s" % s
        lines.append(line)
        if len(lines) >= lim:
            break
    txt = "\n".join(lines)
    if out:
        Path(out).write_text(txt, encoding="utf-8")
        print("wrote %d lines -> %s" % (len(lines), out), file=sys.stderr)
    else:
        print(txt)


if __name__ == "__main__":
    main()
