# -*- coding: utf-8 -*-
# 研究用途：紧凑 ABI 摘要（每函数一行）
import os, json, re

BASE = os.path.dirname(os.path.abspath(__file__))
DUMPDIR = os.path.join(BASE, "fs")
MAPF = os.path.join(BASE, "_fs_map.json")
ARG_REGS = ["rax", "rbx", "rcx", "rdi", "rsi", "r8", "r9", "r10", "r11"]

def main():
    results = json.load(open(MAPF, encoding="utf-8"))
    out = []
    for r in results:
        va = r.get("va")
        if not va:
            out.append("%-55s MISSING" % r["full"])
            continue
        path = os.path.join(DUMPDIR, "%x.asm.txt" % va)
        if not os.path.exists(path):
            out.append("%-55s NODUMP" % r["full"])
            continue
        insns = []
        for ln in open(path, encoding="utf-8").read().splitlines():
            m = re.match(r'0x([0-9a-f]+):\s+(\S+)\s+(.*)$', ln)
            if m:
                insns.append((m.group(1), m.group(2), m.group(3)))
        n = len(insns)
        def regs(s):
            return set(re.findall(r'\br(?:ax|bx|cx|dx|di|si|bp|sp|8|9|10|11|12|13|14|15)[bwd]?\b', s))
        head_src = set()
        for (_, mn, op) in insns[:10]:
            if "," in op:
                head_src |= regs(op.split(",", 1)[1])
            elif mn == "push":
                head_src |= regs(op)
        ret_writes = []
        for (_, mn, op) in insns[-10:]:
            if mn in ("mov", "movabs", "lea", "xor", "movq", "movl", "set", "movzx", "movsxd") and op:
                dst = regs(op.split(",")[0].strip())
                for rr in ARG_REGS:
                    if rr in dst:
                        ret_writes.append(rr)
        # 首个 call 目标（用于判断 helper 依赖）
        firstcall = ""
        for (_, mn, op) in insns:
            if mn == "call":
                mm = re.search(r'0x([0-9a-f]+)', op)
                firstcall = mm.group(1) if mm else op
                break
        # morestack check 是否存在
        has_ms = any(mn == "cmp" and "r14" in op for _, mn, op in insns[:4])
        out.append("%-55s va=%-11x n=%-4d arg=[%s] ret=[%s] call=%s ms=%d" % (
            r["full"], va, n,
            ",".join(sorted(rr for rr in ARG_REGS if rr in head_src)),
            ",".join(dict.fromkeys(ret_writes)),
            firstcall, 1 if has_ms else 0))
    with open(os.path.join(BASE, "_fs_abi_compact.txt"), "w", encoding="utf-8") as fh:
        fh.write("\n".join(out))
    print("wrote", len(out), "lines")

if __name__ == "__main__":
    main()
