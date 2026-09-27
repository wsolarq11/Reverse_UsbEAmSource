#!/usr/bin/env python
"""批量解码 launcherconfigiconstore.go 所需的 .rdata 字符串常量"""
import subprocess as sp
import sys

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"

pairs = [
    (0x140c440c7, 10, "storeError_fmt"),
    (0x140c69c4e, 27, "storePath_baseRef"),
    (0x140c68307, 26, "storePath_iconsName"),
    (0x140c37638, 6, "storePath_defaultLeaf"),
    (0x140c47565, 11, "storePath_prefix"),
    (0x140896812, 24, "storeAdd_nilMsg"),  # newobject + len=0x18
    (0x14089678b, 33, "storeAdd_countLimitMsg"),
    (0x140896721, 48, "storeAdd_bytesLimitMsg"),
    (0x14089668c, 42, "storeAdd_pixelsLimitMsg"),
    (0x140897aab, 9, "ensureJSONEOF_nilMsg"),
    (0x140897ac6, 26, "ensureJSONEOF_trailingMsg"),
    (0x140897a85, 27, "ensureJSONEOF_errFmt"),
    (0x1408976f7, 24, "decode_missingSepMsg"),
    (0x1408976b1, 29, "decode_missingSemiMsg"),
    (0x140897626, 23, "decode_base64ErrFmt"),
    # Delete error: the 0x1b-byte constant


]

# Also read from Launcher_Icons.json (iconstore name) to verify
pairs.append((0x140c440c7, 10, "VERIFY_storeError_fmt"))

for va, length, name in pairs:
    r = sp.run(
        ["python", "tools/go-introspect/va_read.py", EXE, "0x%x" % va, str(length)],
        capture_output=True,
        text=True,
    )
    lines = r.stdout.strip().split("\n")
    utf8_lines = [l for l in lines if l.startswith("  utf-8: ")]
    if utf8_lines:
        val = utf8_lines[0][len("  utf-8: "):]
        print("%-35s VA=0x%x len=%d -> %s" % (name, va, length, val))
    else:
        raw_lines = [l for l in lines if l.startswith("  raw_hex=")]
        if raw_lines:
            print("%-35s VA=0x%x len=%d -> %s" % (name, va, length, raw_lines[0]))
        else:
            print("%-35s VA=0x%x len=%d -> (no decode) %s" % (name, va, length, lines))