#!/usr/bin/env python3
# va_disasm.py — 以 VA 为基址反汇编函数 bin，并对 call/jmp 目标做 std 符号标注。
# 研究用途。依赖 capstone；符号表 docs/goresym/symbols.txt（VA -> name）。
#
# 用法: va_disasm.py <binfile> <baseVA> [limit] [out]
#   baseVA 为该 bin 首个字节的虚拟地址（须 0x 十六进制）。
import sys, os, re
from capstone import Cs, CS_ARCH_X86, CS_MODE_64

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..")
SYM = {}
for line in open(os.path.join(ROOT, "docs", "goresym", "symbols.txt"), encoding="utf-8", errors="replace"):
    line = line.strip()
    if not line:
        continue
    try:
        va, name = line.split(" ", 1)
    except ValueError:
        continue
    SYM[int(va, 16)] = name

STD_PREFIX = ("github.com/", "golang.org/x/", "modernc.org/", "internal/", "runtime.",
              "strings.", "strconv.", "os.", "path.", "crypto/", "encoding/", "time.",
              "syscall", "unsafe", "reflect.", "errors.", "fmt.", "sync.", "io.")

def sym(va):
    if va in SYM:
        n = SYM[va]
        if any(n.startswith(p) for p in STD_PREFIX):
            return n
        return n
    # 最近符号 +delta
    best, bv = None, None
    for v, n in SYM.items():
        if v <= va and (bv is None or v > bv):
            bv, best = v, n
    if best is not None and va - bv < 0x4000:
        return "%s+0x%x" % (best, va - bv)
    return None

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
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j") or insn.mnemonic.startswith("loop"):
            m = re.search(r"0x([0-9a-f]+)", insn.op_str)
            if m:
                t = int(m.group(1), 16)
                s = sym(t)
                if s:
                    line += "    ; %s" % s
        lines.append(line)
        if len(lines) >= lim:
            break
    txt = "\n".join(lines)
    if out:
        with open(out, "w", encoding="utf-8") as f:
            f.write(txt)
        print("wrote %d lines -> %s" % (len(lines), out), file=sys.stderr)
    else:
        print(txt)

if __name__ == "__main__":
    main()