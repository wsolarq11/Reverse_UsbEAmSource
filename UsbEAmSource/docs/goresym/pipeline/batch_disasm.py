#!/usr/bin/env python3
"""
批量反汇编：遍历 pipeline/tmp/*.bin，根据 addr_ref 查真实 VA，运行 va_disasm 生成 .asm.txt。
研究用途。
"""
import os, sys, re, subprocess, glob

ROOT = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource"
BIN = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
ADDR_REF = os.path.join(ROOT, "docs", "goresym", "pipeline", "bootstrap_addr_ref.txt")
TMP = os.path.join(ROOT, "docs", "goresym", "pipeline", "tmp")
VA_DISASM = os.path.join(ROOT, "tools", "go-introspect", "va_disasm.py")

# 解析 addr_ref → { clean_name: (va_hex, len) }
ADDR_MAP = {}
for line in open(ADDR_REF, encoding="utf-8"):
    line = line.strip()
    # method (*BootstrapService)setStartupTrayMode off=0x140772760 end=0x140772820 len=192
    m = re.match(r"method.*?\(?\*?(\w+)\)([.\w]+)\s+off=0x([0-9a-f]+)\s+end=0x[0-9a-f]+\s+len=(\d+)", line)
    if m:
        # Clean name: strip func suffix, keep main method name
        full = m.group(1) + "." + m.group(2)
        # Remove clutter
        va = int(m.group(3), 16)
        length = int(m.group(4))
        ADDR_MAP[full] = (va, length)
        continue
    # func NewBootstrapService off=0x140771d20 end=0x140772700 len=2528
    m2 = re.match(r"func (\w+)\s+off=0x([0-9a-f]+)\s+end=0x[0-9a-f]+\s+len=(\d+)", line)
    if m2 and m2.group(1) == "NewBootstrapService":
        ADDR_MAP["NewBootstrapService"] = (int(m2.group(2), 16), int(m2.group(3)))

print(f"Loaded {len(ADDR_MAP)} addr mappings")

# Build a name-lookup map for bin files: bin basename (without .bin) → (va, len)
# The addr_tool names are like "setStartupTrayMode" for Method 
# But the .bin file names are raw Go names
NAME_MAP = {}
for full_name, (va, length) in ADDR_MAP.items():
    # Extract the short name after the last dot
    parts = full_name.split(".")
    if len(parts) >= 2:
        short = parts[1] if parts[0] in ("BootstrapService", "NewBootstrapService") else full_name
    else:
        short = full_name
    NAME_MAP[short] = (va, length)
    # Also store with the full name pattern used by addr_tool dump
    if "BootstrapService." in full_name:
        NAME_MAP[full_name.split("BootstrapService.")[1]] = (va, length)

# Also handle the -fm variants
for line in open(ADDR_REF, encoding="utf-8"):
    line = line.strip()
    m = re.search(r"\s+off=0x([0-9a-f]+)\s+end=0x[0-9a-f]+\s+len=(\d+)", line)
    if not m:
        continue
    va = int(m.group(1), 16)
    length = int(m.group(2))
    # Extract the first word that looks like a method name
    w = line.split()
    for w in w:
        if w.startswith("(*"):
            name = w.replace("(*BootstrapService)", "").replace("(*BootstrapService).", "")
            NAME_MAP[name] = (va, length)
            break

print(f"Built name map with {len(NAME_MAP)} entries")

# Now disassemble each .bin
count = 0
ok = 0
skip = 0
for bf in sorted(glob.glob(os.path.join(TMP, "*.bin"))):
    basename = os.path.splitext(os.path.basename(bf))[0]
    # Remove numeric prefix if present
    name = basename
    # Skip these
    if name in ("1", "2"):
        skip += 1
        continue
    
    out_asm = os.path.join(TMP, name + ".asm.txt")
    if os.path.exists(out_asm):
        skip += 1
        continue
    
    # Look up VA
    if name in NAME_MAP:
        va, length = NAME_MAP[name]
    else:
        # Try fuzzy match
        found = False
        for k, (v, l) in NAME_MAP.items():
            if k.replace(".", "") == name.replace("_", ""):
                va, length = v, l
                found = True
                break
        if not found:
            print(f"  [SKIP] {name}: no VA found")
            skip += 1
            continue
    
    cmd = ["python", VA_DISASM, bf, hex(va), str(max(length * 4, 1000)), out_asm]
    r = subprocess.run(cmd, capture_output=True, text=True, cwd=ROOT)
    count += 1
    if r.returncode == 0:
        ok += 1
        if count % 20 == 0:
            print(f"  [{count}] {name} -> {out_asm}")
    else:
        print(f"  [FAIL] {name}: {r.stderr.strip()}")

print(f"\nDone: {count} attempted, {ok} OK, {skip} skipped")