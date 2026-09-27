#!/usr/bin/env python
# 研究用途：生成 gap.csv（蓝图函数 - 已落地函数），供 batch_sig_summary.py 使用。
# 用法：python tools/gap_csv.py <target.go> <out.csv>
# 符号映射规则（source_funcs.txt -> symbols.txt）：
#   name                -> main.name
#   (*Type)method       -> main.Type.method
#   (*Type)methodSUFFIX -> main.Type).method.SUFFIX   （SUFFIX = funcN/deferwrapN/gowrapN）
import csv
import os
import re
import sys

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")


def load_symbols():
    syms = {}
    for line in open(os.path.join(ROOT, "docs", "goresym", "symbols.txt"), encoding="utf-8", errors="replace"):
        line = line.strip()
        if not line:
            continue
        try:
            va, name = line.split(" ", 1)
        except ValueError:
            continue
        syms[name] = int(va, 16)
    return syms


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
    m2 = re.match(r"(.+?)(func\d+|deferwrap\d+|gowrap\d+)$", rest)
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
            m = re.match(r"func\s+(\(\s*\*?([A-Za-z_]\w*)\s*\)\s*)?([A-Za-z_]\w*)\s*\(", line)
            if not m:
                continue
            recv = m.group(2)
            name = m.group(3)
            if recv:
                landed.add("main.%s.%s" % (recv, name))
            else:
                landed.add("main." + name)
    return landed


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    target, out = sys.argv[1], sys.argv[2]
    syms = load_symbols()
    sf = load_source_funcs()
    landed = load_landed()
    funcs = sf.get(target, [])
    if not funcs:
        print("source_funcs.txt 中未找到 File: %s" % target)
        return 1
    sorted_syms = sorted(syms.items(), key=lambda x: x[1])
    rows = []
    missing = 0
    for (func, a, b) in funcs:
        sym = norm_sym(func)
        if sym in landed:
            continue
        if sym not in syms:
            rows.append({"Func": func, "Sym": sym, "VA": "0", "Len": "0"})
            missing += 1
            continue
        va = syms[sym]
        length = 512
        for i, (n, v) in enumerate(sorted_syms):
            if v == va and i + 1 < len(sorted_syms):
                length = sorted_syms[i + 1][1] - va
                break
        rows.append({"Func": func, "Sym": sym, "VA": hex(va), "Len": str(length)})
    with open(out, "w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fieldnames=["Func", "Sym", "VA", "Len"])
        w.writeheader()
        w.writerows(rows)
    print("gap %s: %d funcs（符号缺失 %d）-> %s" % (target, len(rows), missing, out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
