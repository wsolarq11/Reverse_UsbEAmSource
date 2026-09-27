#!/usr/bin/env python
# 研究用途：精确复刻 count_funcs.awk 的 last 状态机，提取 backend/*.go 里所有 [P] 存根，
# 反查 symbols.txt 得 VA+邻接长度，输出 CSV（File,Func,Recv,Sym,VA,Len）。
# 用法：python tools/p_funcs.py <out.csv>
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


def has(s, tag):
    return tag in s


def classify(line):
    # 与 count_funcs.awk 相同的 tier 判定（顺序敏感）
    if "[S-sig]" in line or "[S-sig " in line or "[S-sig\t" in line:
        return "S-sig"
    if "[S-inline]" in line or "[S-inline " in line:
        return "S-inline"
    if "[S]" in line or "[S " in line or "[S\t" in line:
        return "S"
    if "[P]" in line or "[P " in line or "[P\t" in line:
        return "P"
    return ""


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    out = sys.argv[1]
    syms = load_symbols()
    order = sorted(syms.items(), key=lambda x: x[1])
    rows = []
    seen = set()
    bdir = os.path.join(ROOT, "backend")
    for fn in sorted(os.listdir(bdir)):
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        lines = open(os.path.join(bdir, fn), encoding="utf-8", errors="replace").read().splitlines()
        last = ""
        for i, line in enumerate(lines):
            if line.startswith("//"):
                if last == "":
                    last = classify(line)
                continue
            if line.startswith("func "):
                if last == "P":
                    m = re.match(r"func\s+(\(\s*([^)]*)\s*\)\s*)?(\w+)\s*\(", line)
                    if m:
                        recv_raw = m.group(2)
                        name = m.group(3)
                        recv = None
                        if recv_raw:
                            # receiver 原始串形如 "s *Type" / "s Type" / "*Type"；取类型名（最后词，去 *）
                            parts = recv_raw.replace("*", " ").split()
                            if parts:
                                recv = parts[-1]
                        sym = ("main.%s.%s" % (recv, name)) if recv else ("main." + name)
                        key = fn + ":" + sym
                        if key in seen:
                            last = ""
                            continue
                        seen.add(key)
                        va = syms.get(sym)
                        length = 512
                        if va is not None:
                            for k, (n, v) in enumerate(order):
                                if v == va and k + 1 < len(order):
                                    length = order[k + 1][1] - va
                                    break
                        rows.append({
                            "File": fn,
                            "Func": ("(*%s).%s" % (recv, name)) if recv else name,
                            "Recv": recv or "",
                            "Sym": sym,
                            "VA": hex(va) if va is not None else "0",
                            "Len": str(length),
                        })
                last = ""
                continue
            last = ""
    with open(out, "w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fieldnames=["File", "Func", "Recv", "Sym", "VA", "Len"])
        w.writeheader()
        w.writerows(rows)
    have = sum(1 for r in rows if r["VA"] != "0")
    print("[P] 提取 %d 个（有 VA %d，无 VA %d）-> %s" % (len(rows), have, len(rows) - have, out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
