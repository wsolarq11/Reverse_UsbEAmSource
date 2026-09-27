#!/usr/bin/env python
# 研究用途：用 capstone 精确解析批次 55 asm 的 RIP 相对 float32/字符串常量。
import os, re, struct, sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go-introspect"))
import va_dump
import capstone

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
TMP = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "docs", "goresym", "pipeline", "tmp"))

data = open(EXE, "rb").read()

def read_f32(va):
    off = va_dump.va_to_off(data, va)
    if off is None:
        return None
    return struct.unpack_from("<f", data, off)[0]

def read_u32(va):
    off = va_dump.va_to_off(data, va)
    if off is None:
        return None
    return struct.unpack_from("<I", data, off)[0]

def read_str(va, maxlen=300):
    off = va_dump.va_to_off(data, va)
    if off is None:
        return None
    raw = data[off:off+maxlen]
    try:
        return raw.split(b"\x00", 1)[0].decode("utf-8")
    except Exception:
        return repr(raw[:40])

FLOAT_MNEM = {"movss", "mulss", "divss", "addss", "subss", "minss", "maxss", "ucomiss", "cvtsd2ss"}

md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

for fn in sorted(f for f in os.listdir(TMP) if f.startswith("b55_") and f.endswith(".bin")):
    prefix = fn[:-4]
    code = open(os.path.join(TMP, fn), "rb").read()
    print("===== %s =====" % prefix)
    for insn in md.disasm(code, 0x0):
        ops = insn.op_str
        if "[rip" not in ops:
            continue
        # 计算 RIP 相对目标：insn.address 是相对 bin 起点的偏移，需加上真实 base
        # 但这里只用 bin 相对地址 + insn.size 得到 disp 偏移，再换算成文件偏移。
        # 直接取 disp 值：capstone 的 op_str 给出 [rip + 0x...]，但更可靠用 bytes。
        m = re.search(r"\[rip\s*\+\s*0x([0-9a-f]+)\]", ops)
        if not m:
            continue
        disp = int(m.group(1), 16)
        # rip = 指令末尾地址。用 insn.address(相对) + insn.size 得到相对末地址，但 disp 是相对真实 VA。
        # 需要真实 base。从 .bin 对应 base 读：文件名映射见下。
        # 简化：直接对 .asm.txt 再解析一次获取真实地址。
    print()

# 直接对 .asm.txt 解析（含真实地址），用 capstone 计算指令长度需要 bin+base。
# 改为：读 asm.txt 行，拿到真实 addr；读对应 bin 反汇编获取长度。
bases = {}
for fn in sorted(f for f in os.listdir(TMP) if f.startswith("b55_") and f.endswith(".bin")):
    # 从 asm.txt 第一行的真实地址推断 base
    atxt = os.path.join(TMP, fn[:-4] + ".asm.txt")
    first = open(atxt, encoding="utf-8", errors="replace").readline()
    m = re.match(r"0x([0-9a-f]+):", first)
    bases[fn] = int(m.group(1), 16) if m else None

for fn in sorted(f for f in os.listdir(TMP) if f.startswith("b55_") and f.endswith(".bin")):
    prefix = fn[:-4]
    base = bases[fn]
    if base is None:
        continue
    code = open(os.path.join(TMP, fn), "rb").read()
    print("===== %s (base 0x%x) =====" % (prefix, base))
    seen = set()
    for insn in md.disasm(code, base):
        ops = insn.op_str
        if "[rip" not in ops:
            continue
        m = re.search(r"\[rip\s*\+\s*0x([0-9a-f]+)\]", ops)
        if not m:
            continue
        disp = int(m.group(1), 16)
        tgt = insn.address + insn.size + disp
        if insn.mnemonic in FLOAT_MNEM:
            f = read_f32(tgt)
            key = ("f", tgt)
            if key in seen:
                continue
            seen.add(key)
            print("  0x%x %-7s %-40s -> VA 0x%x = %.9g (0x%08X)" % (insn.address, insn.mnemonic, ops, tgt, f, read_u32(tgt)))
        elif insn.mnemonic == "lea":
            s = read_str(tgt)
            key = ("s", tgt)
            if key in seen:
                continue
            seen.add(key)
            print("  0x%x %-7s %-40s -> VA 0x%x = %r" % (insn.address, insn.mnemonic, ops, tgt, s))
    print()
