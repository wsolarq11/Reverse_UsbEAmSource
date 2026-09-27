#!/usr/bin/env python
# 研究用途：解析批次 54 asm 里所有 RIP 相对 lea 字符串 + 常用 float 常量。
import os, re, struct, sys

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go-introspect"))
import va_dump

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
TMP = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "docs", "goresym", "pipeline", "tmp"))

data = open(EXE, "rb").read()

def read_str(va, maxlen=300):
    off = va_dump.va_to_off(data, va)
    if off is None:
        return None
    raw = data[off:off+maxlen]
    try:
        return raw.split(b"\x00", 1)[0].decode("utf-8")
    except Exception:
        return repr(raw[:40])

def read_f32(va):
    off = va_dump.va_to_off(data, va)
    if off is None:
        return None
    return struct.unpack_from("<f", data, off)[0]

# 每条 lea 指令长度：RIP 相对 lea 通常是 7 字节 (lea r64) 或 8 字节（REX 前缀）。
# 用 capstone 精确反汇编。
import capstone
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

results = {}
for fn in sorted(f for f in os.listdir(TMP) if f.startswith("b54_") and f.endswith(".asm.txt")):
    path = os.path.join(TMP, fn)
    lines = open(path, encoding="utf-8", errors="replace").read().splitlines()
    # 收集每个 lea rip 指令的 VA 和 disp
    instrs = []
    for ln in lines:
        m = re.match(r"0x([0-9a-f]+):\s+(\S+)\s+(.*)", ln)
        if not m:
            continue
        va = int(m.group(1), 16)
        mn = m.group(2)
        ops = m.group(3)
        if "[rip" in ops:
            rm = re.search(r"\[rip\s*\+\s*0x([0-9a-f]+)\]", ops)
            if rm:
                disp = int(rm.group(1), 16)
                instrs.append((va, mn, ops, disp))
    # 用 capstone 反汇编每个 .bin 对应文件以获得指令长度
    binpath = os.path.join(TMP, fn[:-len(".asm.txt")] + ".bin")
    if not os.path.exists(binpath):
        continue
    base = None
    first = lines[0] if lines else ""
    m = re.match(r"0x([0-9a-f]+):", first)
    if m:
        base = int(m.group(1), 16)
    if base is None:
        continue
    code = open(binpath, "rb").read()
    insn_map = {}
    for insn in md.disasm(code, base):
        insn_map[insn.address] = insn
    for va, mn, ops, disp in instrs:
        insn = insn_map.get(va)
        if insn is None:
            continue
        tgt = insn.address + insn.size + disp
        if tgt in results:
            continue
        if mn == "lea":
            s = read_str(tgt)
            if s is not None:
                results[tgt] = ("str", s)
        elif mn in ("movss", "movsd", "mulss", "mulss", "divss", "addss", "subss", "minss", "maxss", "ucomiss"):
            results[tgt] = ("f32", read_f32(tgt))

for tgt in sorted(results):
    kind, val = results[tgt]
    if kind == "str":
        print("0x%x  str(%d)  %r" % (tgt, len(val), val))
    else:
        print("0x%x  f32  %.9g (0x%08X)" % (tgt, val, struct.unpack("<I", struct.pack("<f", val))[0]))
