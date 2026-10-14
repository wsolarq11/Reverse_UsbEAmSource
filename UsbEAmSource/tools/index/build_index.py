#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""建 usbeam 分析索引库（唯一写入出口，全量重写、可重复）。

数据源全部是仓库内的权威数据，不含任何臆测：
  docs/goresym/symbols.txt            符号 → VA（24181）
  docs/goresym/source_funcs.txt       行号蓝图（4754 函数项）
  docs/goresym/pipeline/va_map_fixed2.txt  方法名 → VA / 长度（修正版）
  backend/*.go（非 _test）            落地标记 + 已实现函数名
  磁盘反汇编资产                       文件名归一化键

产物：UsbEAmSource/work/index/usbeam.db（不入库，见 docs/STRUCTURE.md §9）

用法：
  python tools/index/build_index.py
  python tools/index/build_index.py --quiet        # 只打印计数行

计数行与 `bash tools/count_funcs.sh` 输出同格式，可直接 diff：
  bash tools/count_funcs.sh
  python tools/index/build_index.py --quiet
"""

import argparse
import sqlite3
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import common  # noqa: E402

SCHEMA = """
DROP TABLE IF EXISTS symbol;
DROP TABLE IF EXISTS blueprint;
DROP TABLE IF EXISTS landed;
DROP TABLE IF EXISTS impl_key;
DROP TABLE IF EXISTS asset;
DROP TABLE IF EXISTS method_size;
DROP TABLE IF EXISTS asm_line;
DROP TABLE IF EXISTS call_edge;

CREATE TABLE symbol (
    va   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE blueprint (
    pkg           TEXT,
    file          TEXT NOT NULL,
    func          TEXT NOT NULL,
    func_key      TEXT NOT NULL,
    func_short    TEXT NOT NULL,
    func_short_lc TEXT NOT NULL,
    start_line    INTEGER,
    end_line      INTEGER,
    nlines        INTEGER,
    PRIMARY KEY (file, func)
);

-- 落地行：raw_file 为 backend 下的文件名；marker 为 count_funcs.sh 同源判定结果
CREATE TABLE landed (
    file       TEXT NOT NULL,
    line       INTEGER NOT NULL,
    marker     TEXT,
    func_sig   TEXT NOT NULL,
    impl_key   TEXT,
    impl_short TEXT,
    PRIMARY KEY (file, line)
);

-- 已实现函数名键（用于蓝图缺口判定）：`类型.方法` 或 裸名，唯一实现见 common.func_key
CREATE TABLE impl_key (
    key TEXT PRIMARY KEY
);

-- 反汇编资产：norm 为归一化键（与 blueprint.func_key 可比）
CREATE TABLE asset (
    root     TEXT NOT NULL,
    path     TEXT PRIMARY KEY,
    basename TEXT NOT NULL,
    norm     TEXT NOT NULL,
    bytes    INTEGER NOT NULL
);

CREATE TABLE method_size (
    name TEXT PRIMARY KEY,
    va   INTEGER NOT NULL,
    size INTEGER NOT NULL
);

CREATE INDEX idx_blueprint_file  ON blueprint(file);
CREATE INDEX idx_blueprint_key   ON blueprint(func_key);
CREATE INDEX idx_landed_key      ON landed(impl_key);
CREATE INDEX idx_asset_norm      ON asset(norm);
""" + common.ASM_SCHEMA

def load_symbols(conn):
    rows = sorted(common.load_symbols().items())
    conn.executemany("INSERT INTO symbol(va, name) VALUES (?, ?)", rows)
    return len(rows)


def load_blueprint(conn):
    rows = []
    for pkg, file_, func, start, end, n in common.parse_blueprint():
        key = common.resolve_func_key(func)
        short = common.func_short_of(key)
        rows.append((pkg, file_, func, key, short, short.lower(), start, end, n))
    conn.executemany(
        "INSERT INTO blueprint(pkg, file, func, func_key, func_short, func_short_lc,"
        " start_line, end_line, nlines) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", rows)
    return len(rows)


def load_landed(conn):
    """扫描 backend 非 _test .go：标记行 + 已实现键 + 计数（一次遍历，三产物）。"""
    landed, keys, counts, all_rows = [], set(), {}, []
    for path in common.backend_files():
        text = path.read_text(encoding="utf-8", errors="replace")
        keys |= common.impl_keys_of(text)
        rows = common.scan_markers(path)
        all_rows.extend(rows)
        for lineno, sig, marker in rows:
            m = common.RE_FUNC_DEF.match("func " + sig)
            impl_key = impl_short = None
            if m:
                recv = common.recv_type(m.group(1))
                name = m.group(2)
                impl_key = common.func_key("(%s)%s" % (recv, name)) if recv else name
                impl_short = name
            landed.append((path.name, lineno, marker, sig, impl_key, impl_short))
    conn.executemany(
        "INSERT INTO landed(file, line, marker, func_sig, impl_key, impl_short)"
        " VALUES (?, ?, ?, ?, ?, ?)", landed)
    conn.executemany("INSERT INTO impl_key(key) VALUES (?)", [(k,) for k in sorted(keys)])
    counts.update(common.count_of(all_rows))
    return len(landed), len(keys), counts


def load_assets(conn):
    rows = common.iter_assets()
    conn.executemany(
        "INSERT OR REPLACE INTO asset(root, path, basename, norm, bytes) VALUES (?, ?, ?, ?, ?)",
        rows)
    return len(rows)


def load_method_size(conn):
    if not common.VA_MAP_FIXED2.is_file():
        return 0
    rows = []
    for line in common.VA_MAP_FIXED2.read_text(encoding="utf-8", errors="replace").splitlines():
        parts = line.strip().split("|")
        if len(parts) != 3:
            continue
        try:
            rows.append((parts[0], int(parts[1], 16), int(parts[2])))
        except ValueError:
            continue
    conn.executemany("INSERT OR REPLACE INTO method_size(name, va, size) VALUES (?, ?, ?)", rows)
    return len(rows)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--quiet", action="store_true", help="只打印计数行")
    args = ap.parse_args()

    common.DB_PATH.parent.mkdir(parents=True, exist_ok=True)
    if common.DB_PATH.exists():
        common.DB_PATH.unlink()

    conn = sqlite3.connect(common.DB_PATH)
    conn.executescript(SCHEMA)
    n_sym = load_symbols(conn)
    n_bp = load_blueprint(conn)
    n_landed, n_keys, counts = load_landed(conn)
    n_asset = load_assets(conn)
    n_msize = load_method_size(conn)
    conn.commit()
    conn.close()

    if not args.quiet:
        print("db          %s (%.1f MB)" % (common.DB_PATH, common.DB_PATH.stat().st_size / 1048576))
        print("symbol      %d" % n_sym)
        print("blueprint   %d" % n_bp)
        print("landed      %d 行 / impl_key %d" % (n_landed, n_keys))
        print("asset       %d" % n_asset)
        print("method_size %d" % n_msize)
        print("计数（与 count_funcs.sh 同格式，可直接 diff）：")
    print(common.count_line(counts))


if __name__ == "__main__":
    main()
