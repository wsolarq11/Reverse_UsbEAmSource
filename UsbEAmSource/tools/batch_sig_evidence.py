#!/usr/bin/env python
# 研究用途：批量抽取一个域的缺口函数的「签名证据」（序言 + 返回序列），
# 供签名实证（[S-sig]/[P] 判定）使用。不落地 .bin/.asm.txt（那是 va_dump.py 的职责）。
#
# 用法:
#   python tools/batch_sig_evidence.py <exe> <gap.csv> <out.evidence.txt>
#
# 输入 CSV 列: Func,Sym,VA,Len
# 输出: 每个函数两段紧凑反汇编（序言前 N 条 + 返回序列 ret 前 M 条）。
import csv
import os
import struct
import sys

import capstone

PROLOGUE_N = 22
EPILOGUE_N = 14


def load_symbols():
    sym_path = os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        "..", "docs", "goresym", "symbols.txt",
    )
    table = {}
    with open(sym_path, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            try:
                va, name = line.split(" ", 1)
            except ValueError:
                continue
            table[int(va, 16)] = name
    return table


SYMBOLS = load_symbols()
_SORTED = sorted(SYMBOLS.items())


def annotate(target):
    if target in SYMBOLS:
        return SYMBOLS[target]
    best_va, best_name = None, None
    for va, name in _SORTED:
        if va > target:
            break
        best_va, best_name = va, name
    if best_name is not None and target - best_va < 0x2000:
        return "%s+0x%x" % (best_name, target - best_va)
    return None


def va_to_off(data, va):
    pe = struct.unpack_from("<I", data, 0x3C)[0]
    image_base = struct.unpack_from("<Q", data, pe + 24 + 24)[0]
    rva = va - image_base
    nsect = struct.unpack_from("<H", data, pe + 6)[0]
    optsz = struct.unpack_from("<H", data, pe + 20)[0]
    sec = pe + 24 + optsz
    for i in range(nsect):
        o = sec + i * 40
        vsize = struct.unpack_from("<I", data, o + 8)[0]
        va0 = struct.unpack_from("<I", data, o + 12)[0]
        rawsz = struct.unpack_from("<I", data, o + 16)[0]
        raw = struct.unpack_from("<I", data, o + 20)[0]
        if va0 <= rva < va0 + max(vsize, rawsz) and raw:
            return raw + (rva - va0)
    return None


def disasm(data, base, length):
    off = va_to_off(data, base)
    if off is None:
        return None
    code = data[off:off + length]
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    out = []
    for insn in md.disasm(code, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            m = re_search_hex(insn.op_str)
            if m:
                name = annotate(m)
                if name:
                    line += "   ; %s" % name
        out.append(line)
    return out


def re_search_hex(s):
    import re
    m = re.search(r"0x([0-9a-f]+)", s)
    return int(m.group(1), 16) if m else None


def main():
    if len(sys.argv) < 4:
        print(__doc__)
        return 2
    exe, csv_path, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
    data = open(exe, "rb").read()
    rows = list(csv.DictReader(open(csv_path, encoding="utf-8-sig")))
    lines = []
    for r in rows:
        base = int(r["VA"], 0)
        length = int(r["Len"])
        if length <= 0:
            length = 512
        ins = disasm(data, base, length)
        if ins is None:
            lines.append("### %s [VA=%s len=%d]  <无法定位>\n" % (r["Func"], r["VA"], length))
            continue
        pro = ins[:PROLOGUE_N]
        # 返回序列：最后一个 ret 之前
        ret_idx = None
        for i in range(len(ins) - 1, -1, -1):
            if ins[i].lstrip().startswith("0x") and " ret" in ins[i]:
                ret_idx = i
                break
        epi = ins[max(0, (ret_idx or len(ins)) - EPILOGUE_N):ret_idx] if ret_idx else ins[-EPILOGUE_N:]
        lines.append("### %s  [VA=%s len=%d]" % (r["Func"], r["VA"], length))
        lines.append("  prologue:")
        lines.extend("    " + x for x in pro)
        lines.append("  epilogue:")
        lines.extend("    " + x for x in epi)
        lines.append("")
    with open(out_path, "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))
    print("写 %s（%d 函数，%d 行）" % (out_path, len(rows), len(lines)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
