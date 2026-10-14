#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""按需反汇编：从目标 exe 取指定函数的字节，落 asm_line / call_edge 两张表。

与既有链路的关系：
  - 取字节的 VA→偏移逻辑与 `tools/go-introspect/va_dump.py` 一致（common.va_to_off）。
  - 长度优先取 va_map_fixed2 的实测值；缺失时用「下一个符号 VA - 本 VA」作上界。
  - 符号批注走 `common.annotate_symbol`（唯一实现，窗口 0x2000）。

用法：
  python tools/index/asmtool.py disasm RunNow [--len N] [--print]
  python tools/index/asmtool.py callers parseMemoryReleaseMode
  python tools/index/asmtool.py scan "main.screenshot" [--limit 200]
"""

import argparse
import re
import sqlite3
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import common  # noqa: E402

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

import capstone  # noqa: E402

MD = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
MD.detail = False
HEX_IN_OPS = re.compile(r"0x([0-9a-f]+)")


def load_exe():
    if not common.TARGET_EXE.is_file():
        sys.exit("目标 exe 不存在：%s（见 docs/ENV.md §1）" % common.TARGET_EXE)
    return common.TARGET_EXE.read_bytes()


def resolve(conn, needle):
    """符号名片段或 0xVA → [(va, name)]。"""
    if needle.lower().startswith("0x"):
        va = int(needle, 16)
        row = conn.execute("SELECT name FROM symbol WHERE va = ?", (va,)).fetchone()
        return [(va, row[0] if row else "0x%x" % va)]
    rows = conn.execute(
        "SELECT va, name FROM symbol WHERE name LIKE ? ORDER BY va", ("%" + needle + "%",)).fetchall()
    return [(r[0], r[1]) for r in rows]


def size_of(conn, va, explicit):
    if explicit:
        return explicit
    row = conn.execute("SELECT size FROM method_size WHERE va = ?", (va,)).fetchone()
    if row:
        return row[0]
    nxt = conn.execute("SELECT min(va) FROM symbol WHERE va > ?", (va,)).fetchone()[0]
    return max(16, (nxt - va)) if nxt else 4096


def cmd_disasm(conn, args):
    data = load_exe()
    hits = resolve(conn, args.needle)
    if not hits:
        sys.exit("无命中：%s" % args.needle)
    if len(hits) > 1 and not args.all:
        print("命中 %d 个，只处理第一个（--all 处理全部）：%s" % (len(hits), hits[0][1]))
        hits = hits[:1]
    for va, name in hits:
        _disasm_with_len(conn, data, va, name, size_of(conn, va, args.len), args.print_all)
    conn.commit()


def _disasm_with_len(conn, data, va, name, size, show):
    off = common.va_to_off(data, va)
    if off is None:
        print("VA 不在任何节：0x%x" % va)
        return
    code = data[off:off + size]
    lines, edges = [], []
    for insn in MD.disasm(code, va):
        lines.append((insn.address, va, insn.bytes.hex(), insn.mnemonic, insn.op_str))
        if insn.mnemonic.startswith(("call", "jmp")):
            m = HEX_IN_OPS.search(insn.op_str)
            if m:
                target = int(m.group(1), 16)
                edges.append((insn.address, target, common.annotate_symbol(target)))
    conn.executemany(
        "INSERT OR REPLACE INTO asm_line(va, func_va, bytes, mnemonic, ops) VALUES (?, ?, ?, ?, ?)",
        lines)
    conn.executemany(
        "INSERT OR REPLACE INTO call_edge(src_va, dst_va, dst_name) VALUES (?, ?, ?)", edges)
    print("0x%x  %-58s len=%d  指令 %d  call/jmp %d" % (va, name[:58], size, len(lines), len(edges)))
    if show:
        for va_, _f, b, mn, ops in lines:
            tag = ""
            if mn.startswith(("call", "j")):
                m = HEX_IN_OPS.search(ops)
                if m:
                    n = common.annotate_symbol(int(m.group(1), 16))
                    tag = "   ; %s" % n if n else ""
            print("0x%x: %-8s %s%s" % (va_, mn, ops, tag))


def cmd_callers(conn, args):
    hits = resolve(conn, args.needle)
    if not hits:
        sys.exit("无命中：%s" % args.needle)
    for va, name in hits:
        rows = conn.execute(
            "SELECT src_va, dst_name FROM call_edge WHERE dst_va = ? ORDER BY src_va", (va,)).fetchall()
        print("0x%x %s ← 调用点 %d 个" % (va, name, len(rows)))
        for src, _dn in rows:
            owner = conn.execute("SELECT func_va FROM asm_line WHERE va = ?", (src,)).fetchone()
            owner_name = None
            if owner:
                o = conn.execute("SELECT name FROM symbol WHERE va = ?", (owner[0],)).fetchone()
                owner_name = o[0] if o else "0x%x" % owner[0]
            print("   0x%x  %s" % (src, owner_name or "(未反汇编的调用方)"))


def cmd_scan(conn, args):
    data = load_exe()
    rows = conn.execute(
        "SELECT va, name FROM symbol WHERE name LIKE ? ORDER BY va LIMIT ?",
        ("%" + args.prefix + "%", args.limit)).fetchall()
    total_lines = total_edges = 0
    for va, name in rows:
        size = size_of(conn, va, None)
        off = common.va_to_off(data, va)
        if off is None:
            continue
        lines, edges = [], []
        for insn in MD.disasm(data[off:off + size], va):
            lines.append((insn.address, va, insn.bytes.hex(), insn.mnemonic, insn.op_str))
            if insn.mnemonic.startswith(("call", "jmp")):
                m = HEX_IN_OPS.search(insn.op_str)
                if m:
                    t = int(m.group(1), 16)
                    edges.append((insn.address, t, common.annotate_symbol(t)))
        conn.executemany(
            "INSERT OR REPLACE INTO asm_line(va, func_va, bytes, mnemonic, ops) VALUES (?, ?, ?, ?, ?)", lines)
        conn.executemany(
            "INSERT OR REPLACE INTO call_edge(src_va, dst_va, dst_name) VALUES (?, ?, ?)", edges)
        total_lines += len(lines)
        total_edges += len(edges)
    conn.commit()
    print("扫描 %d 个符号：指令 %d，call/jmp 边 %d" % (len(rows), total_lines, total_edges))


def main():
    ap = argparse.ArgumentParser()
    sub = ap.add_subparsers(dest="cmd", required=True)

    p_d = sub.add_parser("disasm")
    p_d.add_argument("needle")
    p_d.add_argument("--len", type=int, default=0)
    p_d.add_argument("--all", action="store_true")
    p_d.add_argument("--print", dest="print_all", action="store_true")

    p_c = sub.add_parser("callers")
    p_c.add_argument("needle")

    p_s = sub.add_parser("scan")
    p_s.add_argument("prefix")
    p_s.add_argument("--limit", type=int, default=100)

    args = ap.parse_args()
    conn = sqlite3.connect(str(common.DB_PATH))
    if args.cmd == "disasm":
        cmd_disasm(conn, args)
    elif args.cmd == "callers":
        cmd_callers(conn, args)
    else:
        cmd_scan(conn, args)


if __name__ == "__main__":
    main()
