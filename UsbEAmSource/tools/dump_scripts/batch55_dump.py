#!/usr/bin/env python
# 研究用途：批次 55 缺失 asm 资产批量现场抽取（tone-map / display-config / 换算）。
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
    ("b55_textureSizeForDesktop", 0x14099E7E0, 0x140),
    ("b55_resolveCopyBounds", 0x14097C180, 0x1E0),
    ("b55_formatToneMapOptionsForDebug", 0x1409AD660, 0x280),
    ("b55_toneMapOptionsForDisplayWithSDRWhiteResolver", 0x1409AD8E0, 0x260),
    ("b55_applyToneMapEnvOverrides", 0x1409ADB40, 0xC0),
    ("b55_parseToneMapEnvFloat", 0x1409ADC00, 0x100),
    ("b55_toneMapRuntimeRGBToSRGBBytes", 0x1409ADD00, 0xE0),
    ("b55_toneMapRGBWithRuntime", 0x1409ADDE0, 0x1A0),
    ("b55_newToneMapRuntime", 0x1409ADF80, 0x160),
    ("b55_querySDRWhiteLevelForDisplay", 0x140975000, 0x620),
    ("b55_queryActivePaths", 0x140975620, 0x360),
    ("b55_findPathForDisplay", 0x140975980, 0x3A0),
    ("b55_querySourceName", 0x140975D20, 0x140),
    ("b55_queryTargetSDRWhiteLevel", 0x140975E60, 0x120),
    ("b55_normalizeDeviceName", 0x140975F80, 0x40),
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
