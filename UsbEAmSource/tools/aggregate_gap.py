#!/usr/bin/env python
# 研究用途：聚合所有蓝图文件的缺失函数（含闭包），按 VA 长度升序输出，供选批。
# 用法：python tools/aggregate_gap.py [out.txt]
import os
import re
import sys

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")


def load_symbols():
    syms = {}
    order = []
    for line in open(os.path.join(ROOT, "docs", "goresym", "symbols.txt"), encoding="utf-8", errors="replace"):
        line = line.strip()
        if not line:
            continue
        try:
            va_s, name = line.split(" ", 1)
            va = int(va_s, 16)
        except ValueError:
            continue
        syms[name] = va
        order.append((va, name))
    order.sort(key=lambda x: x[0])
    return syms, order


def load_source_funcs():
    files = {}
    cur = None
    for line in open(os.path.join(ROOT, "docs", "goresym", "source_funcs.txt"), encoding="utf-8", errors="replace"):
        line = line.rstrip("\n").rstrip()
        m = re.match(r"File:\s*(.+)", line)
        if m:
            cur = m.group(1).strip()
            files[cur] = []
            continue
        m = re.match(r"\s*(.+?)\s+Lines:\s*(\d+)\s+to\s*(\d+)", line)
        if m and cur:
            files[cur].append((m.group(1).strip(), int(m.group(2)), int(m.group(3))))
    return files


def norm_sym(func):
    m = re.match(r"\(\*?(.+?)\)(.+)$", func)
    if not m:
        return "main." + func
    recv = m.group(1)
    rest = m.group(2)
    m2 = re.match(r"(.+?)(func\d+|deferwrap\d+|gowrap\d+|printf\d+)$", rest, re.I)
    if m2:
        return "main.%s).%s.%s" % (recv, m2.group(1), m2.group(2))
    return "main.%s.%s" % (recv, rest)


def load_landed():
    landed = set()
    bdir = os.path.join(ROOT, "backend")
    for fn in os.listdir(bdir):
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        for line in open(os.path.join(bdir, fn), encoding="utf-8", errors="replace"):
            # 匹配 func name( 或 func (任意 receiver) name(
            m = re.match(r"func\s+(\(\s*([^)]*)\)\s*)?([A-Za-z_]\w*)\s*\(", line)
            if not m:
                continue
            recv = m.group(2)
            name = m.group(3)
            if recv is not None:
                # 从 "(d *Type)" 或 "(*Type)" 提取类型名（去变量名、去 *，保留最后一段）
                tm = re.search(r"([A-Za-z_]\w*)\s*$", recv.strip().rstrip("*").strip())
                if tm:
                    landed.add("main.%s.%s" % (tm.group(1), name))
            else:
                landed.add("main." + name)
    return landed


def main():
    out = sys.argv[1] if len(sys.argv) > 1 else os.path.join(ROOT, "docs", "goresym", "pipeline", "tmp", "gap_aggregate.txt")
    syms, order = load_symbols()
    sf = load_source_funcs()
    landed = load_landed()
    rows = []
    for fname, funcs in sf.items():
        for (func, a, b) in funcs:
            sym = norm_sym(func)
            if sym in landed:
                continue
            if sym not in syms:
                rows.append((10**9, fname, func, "0", 0))
                continue
            va = syms[sym]
            length = 512
            for i, (v, n) in enumerate(order):
                if v == va and i + 1 < len(order):
                    length = order[i + 1][0] - va
                    break
            rows.append((length, fname, func, hex(va), a))
    rows.sort(key=lambda x: (x[0], x[1]))
    with open(out, "w", encoding="utf-8") as fh:
        fh.write("# 缺失函数（按 asm 长度升序；len=512 为未知/末位；含闭包但闭包不计为独立落地目标）\n")
        for length, fname, func, va, a in rows:
            fh.write("%6d  %-42s  %-60s  %s\n" % (length, fname, func, va))
    top = sum(1 for r in rows if not re.search(r"(func\d+|deferwrap\d+|gowrap\d+|printf\d+)", r[2]))
    print("total missing=%d, top-level(non-closure)=%d -> %s" % (len(rows), top, out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
