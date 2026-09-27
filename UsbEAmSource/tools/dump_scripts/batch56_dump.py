#!/usr/bin/env python
# 研究用途：批次 156 缺失 asm 资产抽取（gpu_windows.go registry/DXGI 入口链）。
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go-introspect"))
import va_dump

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
OUT = os.path.abspath(os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "..", "..", "docs", "goresym", "pipeline", "tmp",
))

targets = [
    ("resolveGPUPreferenceAdapterInfo", 0x14085d240, 0x100),
    ("resolveDXGIAdapterNameByPreference", 0x14085d340, 0x6c0),
    ("saveGPUPreferenceEntry", 0x14085f340, 0x400),
]

data = open(EXE, "rb").read()
md = va_dump.capstone.Cs(va_dump.capstone.CS_ARCH_X86, va_dump.capstone.CS_MODE_64)

for prefix, base, length in targets:
    off = va_dump.va_to_off(data, base)
    if off is None:
        print("SKIP %s: VA 0x%x no section" % (prefix, base))
        continue
    code = data[off:off + length]
    out_pfx = os.path.join(OUT, prefix)
    with open(out_pfx + ".bin", "wb") as fh:
        fh.write(code)
    lines = []
    for insn in md.disasm(code, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            m = va_dump.re.search(r"0x([0-9a-f]+)", insn.op_str)
            if m:
                name = va_dump.annotate(int(m.group(1), 16))
                if name:
                    line += "   ; %s" % name
        lines.append(line)
    with open(out_pfx + ".asm.txt", "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))
    print("0x%x +%d -> %s (%d lines)" % (base, length, prefix, len(lines)))

print("done: %d targets" % len(targets))
