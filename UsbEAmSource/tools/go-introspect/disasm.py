import sys, os
from capstone import Cs, CS_ARCH_X86, CS_MODE_64

SYM = {}
_t = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "docs", "goresym", "symbols.txt")
for line in open(_t, encoding="utf-8", errors="replace"):
    line=line.strip()
    if not line: continue
    try:
        va, name = line.split(" ", 1)
    except ValueError:
        continue
    SYM[int(va,16)] = name

def sym(va):
    if va in SYM:
        return SYM[va]
    # 找最近符号（区间内）
    best=None; bv=None
    for v,n in SYM.items():
        if v<=va and (bv is None or v>bv):
            bv=v; best=n
    if best is not None and va-bv < 0x2000:
        return "%s+0x%x" % (best, va-bv)
    return None

def dump(fn, start_va, limit=10**9, out=None):
    data=open(fn,'rb').read()
    md=Cs(CS_ARCH_X86, CS_MODE_64)
    lines=[]
    for insn in md.disasm(data, start_va):
        line="0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        # annotate call targets
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            import re
            m=re.search(r'0x([0-9a-f]+)', insn.op_str)
            if m:
                t=int(m.group(1),16)
                s=sym(t)
                if s:
                    line += "   ; %s"%s
        lines.append(line)
        if len(lines)>=limit: break
    if out: open(out,'w',encoding='utf-8').write("\n".join(lines))
    else:
        print("\n".join(lines))

if __name__=="__main__":
    fn=sys.argv[1]
    va=int(sys.argv[2],0) if len(sys.argv)>2 else 0x14086f680
    lim=int(sys.argv[3],0) if len(sys.argv)>3 else 200
    out=sys.argv[4] if len(sys.argv)>4 else None
    dump(fn, va, lim, out)
