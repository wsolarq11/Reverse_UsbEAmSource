#!/usr/bin/env python
# 研究用途：把 batch_sig_evidence.py 的证据进一步压成「每函数一行摘要」，
# 供批量签名判定（[S-sig]/[P]）。摘要字段：
#   name | recv | params(寄存器槽) | rets(寄存器槽) | first-call | len
import csv
import os
import re
import struct
import sys

import capstone

SYMBOLS = {}
for line in open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "docs", "goresym", "symbols.txt"), encoding="utf-8", errors="replace"):
    line = line.strip()
    if not line:
        continue
    try:
        va, name = line.split(" ", 1)
    except ValueError:
        continue
    SYMBOLS[int(va, 16)] = name
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


PARAM_REGS = ["rax", "rbx", "rcx", "rdi", "rsi", "r8", "r9", "r10", "r11", "r12", "r13", "r14"]
RET_REGS = ["rax", "rbx", "rcx", "rdi", "rsi", "r8", "r9", "r10", "r11"]


def summarize(data, base, length):
    off = va_to_off(data, base)
    if off is None:
        return None
    code = data[off:off + length]
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    ins = list(md.disasm(code, base))
    if not ins:
        return None
    # 参数寄存器：序言（morestack 检查之后、第一个 call 之前）保存/写入 [rsp+X] 的寄存器
    params = []
    first_call = None
    for i in ins:
        if i.mnemonic == "call":
            m = re.search(r"0x([0-9a-f]+)", i.op_str)
            first_call = annotate(int(m.group(1), 16)) if m else None
            break
        if i.mnemonic.startswith("mov") and "rsp" in i.op_str and "[" in i.op_str:
            # 提取源寄存器
            m = re.search(r",\s*([a-z][a-z0-9]+)$", i.op_str)
            if m and m.group(1) in PARAM_REGS and m.group(1) not in params:
                params.append(m.group(1))
    # 返回寄存器：最后一个 ret 之前的 mov 目标寄存器
    rets = []
    ret_idx = None
    for i in range(len(ins) - 1, -1, -1):
        if " ret" in "%s %s" % (ins[i].mnemonic, ins[i].op_str) or ins[i].mnemonic == "ret":
            ret_idx = i
            break
    if ret_idx:
        for i in range(max(0, ret_idx - 6), ret_idx):
            m = re.match(r"([a-z][a-z0-9]*)", ins[i].op_str)
            tgt = m.group(1) if m else None
            if tgt in RET_REGS and tgt not in rets:
                rets.append(tgt)
    return ",".join(params), ",".join(rets), first_call


def main():
    if len(sys.argv) < 4:
        print(__doc__)
        return 2
    exe, csv_path, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
    data = open(exe, "rb").read()
    rows = list(csv.DictReader(open(csv_path, encoding="utf-8-sig")))
    out = []
    for r in rows:
        base = int(r["VA"], 0)
        length = int(r["Len"])
        if length <= 0:
            length = 512
        s = summarize(data, base, length)
        if s is None:
            out.append("%s | <bad> | len=%d" % (r["Func"], length))
            continue
        params, rets, fcall = s
        out.append("%s | params=[%s] | rets=[%s] | call=%s | len=%d" % (r["Func"], params, rets, fcall, length))
    with open(out_path, "w", encoding="utf-8") as fh:
        fh.write("\n".join(out))
    print("写 %s（%d 行）" % (out_path, len(out)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
