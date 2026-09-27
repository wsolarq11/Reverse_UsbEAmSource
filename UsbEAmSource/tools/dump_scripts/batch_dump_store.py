#!/usr/bin/env python
"""批量 dump launcherConfigIconStore 域函数，计算 VA 区间长度并调用 va_dump.py"""
import os
import subprocess
import sys

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
OUTDIR = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource/docs/goresym/pipeline/tmp"
VA_DUMP = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource/tools/go-introspect/va_dump.py"

# 按 VA 排序的函数列表：(name, hexVA)
funcs = [
    ("collectLauncherConfigIconRefCounts", 0x1408914c0),
    # .func1 紧随其后自动覆盖
    ("launcherConfigIconStorePath", 0x1408925e0),
    ("launcherConfigIconStoreForConfigPath", 0x140892740),
    ("launcherConfigIconStoreForPath", 0x140892780),
    ("launcherConfigIconStore.Put", 0x1408928e0),
    ("launcherConfigIconStore.Delete", 0x1408934c0),
    ("launcherConfigIconStore.ExtractLauncherConfigIcons", 0x140893700),
    ("launcherConfigIconStore.ExternalizeLauncherConfigIcons", 0x140893740),
    ("launcherConfigIconStore.externalizeLauncherConfigIcons", 0x1408937a0),
    ("launcherConfigIconStore.Prune", 0x1408948e0),
    ("collectLauncherConfigIconRefs", 0x140894ca0),
    ("normalizeLauncherConfigIconRefSet", 0x140894e80),
    ("launcherConfigIconStore.loadUnlocked", 0x140894fa0),
    ("launcherConfigIconStore.writeUnlocked", 0x140895a80),
    ("validateLauncherConfigIconLibrary", 0x140895e80),
    ("launcherConfigIconStoreBudget.add", 0x1408965e0),
    ("readLauncherConfigIconStoreBytes", 0x1408968a0),
    ("writeLauncherConfigIconLibraryFile", 0x140896de0),
    ("decodeLauncherConfigIconDataURL", 0x1408973a0),
    ("newLauncherConfigIconRef", 0x1408977c0),
    ("ensureLauncherConfigIconJSONEOF", 0x140897a00),
    ("launcherConfigIconStoreError", 0x140897b00),
]

# 已知已存在的 asm
already_have = {
    "launcherConfigIconStore.Resolve",
    "launcherConfigIconStore.ensureLoadedUnlocked",
}

# 自动补上已知但未在 funcs 列表中的（用于计算区间）
all_vaddrs = {name: va for name, va in funcs}
all_vaddrs["launcherConfigIconStore.Resolve"] = 0x140892e40
all_vaddrs["launcherConfigIconStore.ensureLoadedUnlocked"] = 0x1408951a0

sorted_items = sorted(all_vaddrs.items(), key=lambda x: x[1])

# 计算每个函数的长度（下一个函数 VA - 当前 VA + 余量 0x100）
print("=== VA 区间表 ===")
for i, (name, va) in enumerate(sorted_items):
    next_va = sorted_items[i+1][1] if i+1 < len(sorted_items) else va + 0x2000
    length = next_va - va
    # 对于跨越 .func1 的函数，自动扩展
    print(f"  {name:50s}  0x{va:x}  →  0x{next_va:x}  ({length} bytes)")

print("\n=== 批量 dump ===")
os.makedirs(OUTDIR, exist_ok=True)

for name, va in funcs:
    outprefix = os.path.join(OUTDIR, name)
    binpath = outprefix + ".bin"
    asmpath = outprefix + ".asm.txt"
    if os.path.exists(asmpath):
        print(f"  [SKIP] {name} — already exists")
        continue
    if name in already_have:
        print(f"  [SKIP] {name} — already have disasm")
        continue

    # 计算长度：找到排序列表中下一个函数 VA
    idx = next(i for i, (n, v) in enumerate(sorted_items) if n == name)
    next_va = sorted_items[idx+1][1] if idx+1 < len(sorted_items) else va + 0x2000
    length = next_va - va
    # 加 buffer
    length = max(length + 0x100, 0x200)

    cmd = [sys.executable, VA_DUMP, EXE, hex(va), str(length), outprefix]
    print(f"  [DUMP] {name:45s}  VA=0x{va:x}  len={length:#x}")
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"         FAIL: {result.stderr.strip()}")
    else:
        print(f"         OK")

print("\nDone.")