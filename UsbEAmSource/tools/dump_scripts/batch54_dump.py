#!/usr/bin/env python
# 研究用途：批次 54 缺失 asm 资产批量现场抽取。
# 复用 va_dump 的 VA->偏移 与 反汇编逻辑，输出到 docs/goresym/pipeline/tmp。
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go-introspect"))
import va_dump

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
OUT = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "..", "..", "docs", "goresym", "pipeline", "tmp",
)
OUT = os.path.abspath(OUT)

# (前缀, baseVA, length)
targets = [
    ("b54_cache_prepare", 0x1409768E0, 0xA0),
    ("b54_cache_capture", 0x140976980, 0x2E0),
    ("b54_cache_captureRegion", 0x140976C60, 0x3C0),
    ("b54_cache_entryForDisplay", 0x140977020, 0x400),
    ("b54_cache_createEntryLocked", 0x140977480, 0x660),
    ("b54_cache_remove", 0x140977AE0, 0x160),
    ("b54_cache_Reconcile", 0x140977C40, 0xAA0),
    ("b54_cache_Close", 0x140978780, 0x300),
    ("b54_entry_capture", 0x140978A80, 0x3C0),
    ("b54_entry_captureRegion", 0x140978EA0, 0x460),
    ("b54_entry_release", 0x140979360, 0x1A0),
    ("b54_entry_updateLastFrameLocked", 0x140979560, 0x140),
    ("b54_entry_lastUsedSnapshot", 0x1409796A0, 0x120),
    ("b54_createD3D11Device", 0x140979FA0, 0xE0),
    ("b54_createD3D11DeviceWithFeatureLevels", 0x14097A080, 0x3E0),
    ("b54_d3d11Device_release", 0x14097A460, 0xC0),
    ("b54_openTarget_func2", 0x140979F00, 0x60),
    ("b54_openTarget_func1", 0x140979F60, 0x40),
    ("b54_cache_entryForDisplay_deferwrap1", 0x140977420, 0x60),
    ("b54_entry_capture_deferwrap1", 0x140978E40, 0x60),
    ("b54_entry_captureRegion_deferwrap1", 0x140979300, 0x60),
    ("b54_entry_release_deferwrap1", 0x140979500, 0x60),
    ("b54_entry_lastUsedSnapshot_deferwrap1", 0x1409797C0, 0x60),
]

data = open(EXE, "rb").read()
md = va_dump.capstone.Cs(va_dump.capstone.CS_ARCH_X86, va_dump.capstone.CS_MODE_64)

for prefix, base, length in targets:
    off = va_dump.va_to_off(data, base)
    if off is None:
        print("SKIP %s: VA 0x%x 无节" % (prefix, base))
        continue
    code = data[off:off + length]
    out_pfx = os.path.join(OUT, prefix)
    with open(out_pfx + ".bin", "wb") as fh:
        fh.write(code)
    lines = []
    for insn in md.disasm(code, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        import re
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            m = re.search(r"0x([0-9a-f]+)", insn.op_str)
            if m:
                name = va_dump.annotate(int(m.group(1), 16))
                if name:
                    line += "   ; %s" % name
        lines.append(line)
    with open(out_pfx + ".asm.txt", "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))
    print("0x%x +%d -> %s (%d lines)" % (base, length, prefix, len(lines)))

print("done: %d targets" % len(targets))
