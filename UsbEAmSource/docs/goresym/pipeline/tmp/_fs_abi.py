# -*- coding: utf-8 -*-
# 研究用途：生成 ABI 分析报告（序言/结尾 + 参数寄存器 + 返回寄存器）
import os, json, re

BASE = os.path.dirname(os.path.abspath(__file__))
DUMPDIR = os.path.join(BASE, "fs")
MAPF = os.path.join(BASE, "_fs_map.json")

# 整数/指针 参数寄存器（ABIInternal 顺序）
ARG_REGS = ["rax", "rbx", "rcx", "rdi", "rsi", "r8", "r9", "r10", "r11"]
RET_REGS = ["rax", "rbx", "rcx", "rdi", "rsi", "r8", "r9", "r10", "r11"]

def regs_of(op_str):
    # 提取操作数里的寄存器名（含 qword ptr [rax+0x10] 等）
    return set(re.findall(r'\b(r(?:ax|bx|cx|dx|di|si|bp|sp|8|9|10|11|12|13|14|15)[bwd]?|xmm\d+)\b', op_str))

def analyze():
    results = json.load(open(MAPF, encoding="utf-8"))
    out = []
    for r in results:
        va = r.get("va")
        if not va:
            out.append((r, "MISSING"))
            continue
        path = os.path.join(DUMPDIR, "%x.asm.txt" % va)
        if not os.path.exists(path):
            out.append((r, "NODUMP"))
            continue
        lines = open(path, encoding="utf-8").read().splitlines()
        insns = []
        for ln in lines:
            m = re.match(r'0x([0-9a-f]+):\s+(\S+)\s+(.*)$', ln)
            if m:
                insns.append((m.group(1), m.group(2), m.group(3)))
        n = len(insns)
        # 序言：前 12 条里，参数寄存器第一次被“读”（作为源操作数）
        # 简化：收集前 12 条里出现在源位置的 arg 寄存器
        head = insns[:12]
        tail = insns[-12:]
        # 结尾：ret 前写入的返回寄存器
        ret_writes = []
        for (_, mn, op) in tail:
            if mn in ("mov", "movabs", "lea", "xor", "movq", "movl", "set", "movzx", "movsxd") and op:
                dst = op.split(",")[0].strip()
                dstregs = regs_of(dst)
                for rr in RET_REGS:
                    if rr in dstregs:
                        ret_writes.append(rr)
        # 序言 arg 读取（前 10 条，含 cmp rsp）
        head_src = set()
        for (_, mn, op) in head:
            if "," in op:
                src = op.split(",", 1)[1]
                head_src |= regs_of(src)
            elif mn in ("push",):
                head_src |= regs_of(op)
        arg_used = [rr for rr in ARG_REGS if rr in head_src]
        out.append((r, n, arg_used, ret_writes, head, tail))
    return out

def fmt(r, n, arg_used, ret_writes, head, tail):
    label = r["full"]
    lines = []
    lines.append("=== %s  [%s] va=0x%x len=%d insn=%d" % (label, r["sym"], r["va"], r["len"], n))
    lines.append("  argRegs(序言源): %s" % (", ".join(arg_used) or "-"))
    lines.append("  retWrites(结尾): %s" % (", ".join(dict.fromkeys(ret_writes)) or "-"))
    lines.append("  HEAD:")
    for h in head:
        lines.append("    %s %-8s %s" % h)
    lines.append("  TAIL:")
    for t in tail:
        lines.append("    %s %-8s %s" % t)
    return "\n".join(lines)

def main():
    rows = analyze()
    with open(os.path.join(BASE, "_fs_abi_report.txt"), "w", encoding="utf-8") as fh:
        for row in rows:
            if row[1] in ("MISSING", "NODUMP"):
                fh.write("=== %s %s\n\n" % (row[0]["full"], row[1]))
                continue
            r, n, arg_used, ret_writes, head, tail = row
            fh.write(fmt(r, n, arg_used, ret_writes, head, tail) + "\n\n")
    print("wrote _fs_abi_report.txt, rows:", len(rows))

if __name__ == "__main__":
    main()
