#!/usr/bin/env python
"""扫描 PE .text 中 call rel32 指令，反查目标函数的调用点。

用法:
  python find_callers.py <exe> <targetVA_hex> [moreVA...]

输出: 每个目标 VA 的调用点 VA + 所属符号（用 symbols.txt 反查）。
"""
import os
import re
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass


def load_symbols():
    here = os.path.dirname(os.path.abspath(__file__))
    for cand in (
        os.path.join(here, "..", "docs", "goresym", "symbols.txt"),
        os.path.join(here, "docs", "goresym", "symbols.txt"),
    ):
        if os.path.exists(cand):
            table = {}
            with open(cand, encoding="utf-8", errors="replace") as fh:
                for line in fh:
                    line = line.strip()
                    if not line:
                        continue
                    parts = line.split(" ", 1)
                    if len(parts) != 2:
                        continue
                    try:
                        table[int(parts[0], 16)] = parts[1]
                    except ValueError:
                        pass
            return table
    return {}


SYMS = load_symbols()
_SORTED = sorted(SYMS.items())


def sym_at(va):
    best_va, best_name = None, None
    for sva, name in _SORTED:
        if sva > va:
            break
        best_va, best_name = sva, name
    if best_name is not None and va - best_va < 0x4000:
        return "%s+0x%x" % (best_name, va - best_va)
    return "(unknown 0x%x)" % va


def sections(data):
    pe = struct.unpack_from("<I", data, 0x3C)[0]
    base = struct.unpack_from("<Q", data, pe + 24 + 24)[0]
    nsec = struct.unpack_from("<H", data, pe + 6)[0]
    optsz = struct.unpack_from("<H", data, pe + 20)[0]
    sec = pe + 24 + optsz
    out = []
    for i in range(nsec):
        o = sec + i * 40
        name = data[o:o + 8].rstrip(b"\0").decode("ascii", "replace")
        vsize = struct.unpack_from("<I", data, o + 8)[0]
        va0 = struct.unpack_from("<I", data, o + 12)[0] + base
        rawsz = struct.unpack_from("<I", data, o + 16)[0]
        raw = struct.unpack_from("<I", data, o + 20)[0]
        out.append((name, va0, vsize, raw, rawsz))
    return out


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    exe = sys.argv[1]
    targets = {int(a, 0) for a in sys.argv[2:]}
    data = open(exe, "rb").read()

    for t in sorted(targets):
        hits = []
        for name, va0, vsize, raw, rawsz in sections(data):
            if not name.startswith(".text"):
                continue
            body = data[raw:raw + rawsz]
            for i in range(len(body) - 5):
                if body[i] != 0xE8:
                    continue
                rel = struct.unpack_from("<i", body, i + 1)[0]
                insn_va = va0 + i
                if insn_va + 5 + rel == t:
                    hits.append(insn_va)
        print("### target 0x%x" % t)
        if not hits:
            print("    (no xref)")
        for h in hits:
            print("    0x%x  in %s" % (h, sym_at(h)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
