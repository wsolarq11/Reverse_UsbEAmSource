# -*- coding: utf-8 -*-
# 研究用途：解析蓝图 + 映射 VA + 批量 dump + ABI 初判
import os, re, json, bisect
import capstone

BASE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.normpath(os.path.join(BASE, "..", "..", "..", ".."))
BP = os.path.join(BASE, "_filesearch_blueprint.txt")
SYM = os.path.join(ROOT, "docs", "goresym", "symbols.txt")
EXE = r"D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe"
DUMPDIR = os.path.join(BASE, "fs")

def load_symbols():
    table = {}
    order = []
    with open(SYM, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            parts = line.split(" ", 1)
            if len(parts) != 2:
                continue
            try:
                va = int(parts[0], 16)
            except ValueError:
                continue
            table[parts[1]] = va
            order.append((va, parts[1]))
    order.sort()
    return table, order

SKIP_PAT = re.compile(r'(func\d+$|deferwrap\d*$|gowrap\d*$|-fm$|^init$|initfunc\d+$)')

# 蓝图里的 concatenated 名 → 真实方法形式
CONCAT_METHOD = {
    "fileSearchSortedCandidateHeapLen": ("fileSearchSortedCandidateHeap", "Len"),
    "fileSearchSortedCandidateHeapLess": ("fileSearchSortedCandidateHeap", "Less"),
    "fileSearchSortedCandidateHeapSwap": ("fileSearchSortedCandidateHeap", "Swap"),
    "fileSearchCandidateHeapLen": ("fileSearchCandidateHeap", "Len"),
    "fileSearchCandidateHeapLess": ("fileSearchCandidateHeap", "Less"),
    "fileSearchCandidateHeapSwap": ("fileSearchCandidateHeap", "Swap"),
    "fileSearchFieldSetincludes": ("fileSearchFieldSet", "includes"),
}

def parse_blueprint():
    funcs = []
    with open(BP, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            line = line.rstrip("\n").rstrip()
            if not line:
                continue
            stripped = line.strip()
            if stripped.startswith("File:"):
                continue
            if not (line.startswith("\t") or line.startswith(" ")):
                continue
            m = re.match(r'^(\S+)\s+Lines:\s+\d+\s+to\s+\d+\s+\((\d+)\)\s*$', stripped)
            if not m:
                continue
            fname = m.group(1)
            if SKIP_PAT.search(fname):
                continue
            recv = ""
            base = fname
            if fname in CONCAT_METHOD:
                recv, base = CONCAT_METHOD[fname]
            else:
                rm = re.match(r'^\((\*?)([A-Za-z][A-Za-z0-9_]*)\)\.?(.+)$', fname)
                if rm:
                    recv = rm.group(2)
                    base = rm.group(3)
            funcs.append({"full": fname, "recv": recv, "name": base})
    return funcs

def va_to_off(data, va):
    import struct
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

def main():
    table, order = load_symbols()
    funcs = parse_blueprint()
    main_vas = sorted([va for va, nm in order if nm.startswith("main.")])
    data = open(EXE, "rb").read()
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

    os.makedirs(DUMPDIR, exist_ok=True)
    results = []
    for fn in funcs:
        recv, name = fn["recv"], fn["name"]
        sym = "main.%s.%s" % (recv, name) if recv else "main.%s" % name
        va = table.get(sym)
        if va is None and recv:
            va = table.get("main.(*%s).%s" % (recv, name))
        if va is None:
            results.append({**fn, "va": None, "len": 0, "sym": sym, "asm": []})
            continue
        i = bisect.bisect_right(main_vas, va)
        nxt = main_vas[i] if i < len(main_vas) else va + 0x2000
        length = nxt - va
        if length <= 0 or length > 0x20000:
            length = 0x2000
        off = va_to_off(data, va)
        if off is None:
            results.append({**fn, "va": va, "len": length, "sym": sym, "asm": []})
            continue
        code = data[off:off+length]
        insns = list(md.disasm(code, va))
        # 保存 asm
        with open(os.path.join(DUMPDIR, "%x.asm.txt" % va), "w", encoding="utf-8") as fh:
            for ins in insns:
                fh.write("0x%x: %-8s %s\n" % (ins.address, ins.mnemonic, ins.op_str))
        results.append({**fn, "va": va, "len": length, "sym": sym, "asm": insns})

    with open(os.path.join(BASE, "_fs_map.json"), "w", encoding="utf-8") as fh:
        json.dump([{k: v for k, v in r.items() if k != "asm"} for r in results], fh, indent=0, ensure_ascii=False)

    missing = [r for r in results if r["va"] is None]
    print("total funcs:", len(results))
    print("missing:", len(missing))
    for r in missing:
        print("  MISSING:", r["full"], r["sym"])
    print("dumped asm count:", sum(1 for r in results if r.get("asm")))

if __name__ == "__main__":
    main()
