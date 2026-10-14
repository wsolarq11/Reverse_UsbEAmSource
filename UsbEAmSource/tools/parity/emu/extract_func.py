#!/usr/bin/env python
# 研究用途：函数级对拍地基 —— 从原 exe 抽取目标函数的机器码与只读跳转表。
#
# 输出物（tools/parity/bin/）：
#   <func>.code.bin   函数机器码（已验证无 call/无外部重定位依赖）
#   <func>.tables.bin 跳转表字节（顺序拼接，供驱动按固定偏移映射）
#   <func>.meta.json  元数据（函数 VA/长度、各表 VA/偏移/表项数）
#
# 前置校验（硬门禁，不改动直接失败）：
#   1. 函数体扫描到的 call 数必须为 0 —— 否则需要更完整的加载器，本抽取器不适用。
#   2. 所有 [rip+disp] lea 的目标必须落在同一张已声明的表集合内 —— 防止漏表。
#
# 用法:
#   python tools/parity/emu/extract_func.py <funcVA> <length> <outName>
import json
import os
import re
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

import capstone

EXE = r"D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
OUT_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "bin")

# 跳转表表项数：索引经 `cmp <idx>,0x19` / `ja` 限界，有效 0x1a 项。
TABLE_ENTRY_COUNT = 0x1A
TABLE_ENTRY_SIZE = 8


class PE:
    def __init__(self, path):
        self.data = open(path, "rb").read()
        pe = struct.unpack_from("<I", self.data, 0x3C)[0]
        self.image_base = struct.unpack_from("<Q", self.data, pe + 24 + 24)[0]
        nsect = struct.unpack_from("<H", self.data, pe + 6)[0]
        optsz = struct.unpack_from("<H", self.data, pe + 20)[0]
        sec = pe + 24 + optsz
        self.sections = []
        for i in range(nsect):
            o = sec + i * 40
            self.sections.append({
                "vsize": struct.unpack_from("<I", self.data, o + 8)[0],
                "va": struct.unpack_from("<I", self.data, o + 12)[0],
                "rawsz": struct.unpack_from("<I", self.data, o + 16)[0],
                "raw": struct.unpack_from("<I", self.data, o + 20)[0],
            })

    def va_to_off(self, va):
        rva = va - self.image_base
        for s in self.sections:
            if s["va"] <= rva < s["va"] + max(s["vsize"], s["rawsz"]) and s["raw"]:
                return s["raw"] + (rva - s["va"])
        return None

    def read(self, va, length):
        off = self.va_to_off(va)
        if off is None:
            raise ValueError("VA 0x%x 不在任何已映射节" % va)
        return self.data[off:off + length]


def scan_rip_lea(code, base_va):
    """返回 [(指令地址, lea 目标 VA)]，以及 call 指令地址列表。"""
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    leas, calls = [], []
    for insn in md.disasm(code, base_va):
        if insn.mnemonic == "call":
            calls.append(insn.address)
        if insn.mnemonic == "lea" and "[rip" in insn.op_str:
            m = re.search(r"\[rip \+ (0x[0-9a-f]+)\]", insn.op_str)
            if m:
                leas.append((insn.address, insn.address + insn.size + int(m.group(1), 16)))
    return leas, calls


def main():
    if len(sys.argv) < 4:
        print(__doc__)
        return 2
    func_va = int(sys.argv[1], 0)
    length = int(sys.argv[2], 0)
    name = sys.argv[3]

    pe = PE(EXE)
    code = pe.read(func_va, length)
    leas, calls = scan_rip_lea(code, func_va)

    if calls:
        print("门禁失败：函数体含 %d 条 call（%s），本抽取器只支持无调用纯逻辑函数。"
              % (len(calls), ", ".join("0x%x" % c for c in calls)))
        return 1

    # 去重表基址；表按 VA 升序排列，驱动用固定偏移访问。
    table_vas = sorted({target for _, target in leas})
    if not table_vas:
        print("门禁失败：未扫描到任何 [rip+disp] 表引用，与预期不符。")
        return 1

    os.makedirs(OUT_DIR, exist_ok=True)
    tables_blob = bytearray()
    table_meta = []
    for tva in table_vas:
        entries = struct.unpack_from("<%dQ" % TABLE_ENTRY_COUNT, pe.read(tva, TABLE_ENTRY_COUNT * TABLE_ENTRY_SIZE), 0)
        if any(entries[i] == 0 for i in range(TABLE_ENTRY_COUNT)):
            print("门禁失败：表 0x%x 存在零表项，表项数假设（0x1a）不成立。" % tva)
            return 1
        table_meta.append({
            "va": tva,
            "offset": len(tables_blob),
            "count": TABLE_ENTRY_COUNT,
            "entries": list(entries),
        })
        tables_blob.extend(pe.read(tva, TABLE_ENTRY_COUNT * TABLE_ENTRY_SIZE))

    code_path = os.path.join(OUT_DIR, name + ".code.bin")
    tables_path = os.path.join(OUT_DIR, name + ".tables.bin")
    meta_path = os.path.join(OUT_DIR, name + ".meta.json")

    open(code_path, "wb").write(code)
    open(tables_path, "wb").write(bytes(tables_blob))

    meta = {
        "name": name,
        "exe": EXE,
        "funcVA": "0x%x" % func_va,
        "length": length,
        "imageBase": "0x%x" % pe.image_base,
        "ripLeaSites": [{"at": "0x%x" % a, "target": "0x%x" % t} for a, t in leas],
        "tables": table_meta,
        "bytesSHA256": None,
    }
    import hashlib
    meta["bytesSHA256"] = hashlib.sha256(code).hexdigest()
    open(meta_path, "w", encoding="utf-8").write(json.dumps(meta, indent=2))

    print("OK  func=0x%x len=%d  code=%s  tables=%s(%d)  meta=%s"
          % (func_va, length, code_path, tables_path, len(table_vas), meta_path))
    for tm in table_meta:
        print("    表 0x%x offset=0x%x count=%d" % (tm["va"], tm["offset"], tm["count"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())
