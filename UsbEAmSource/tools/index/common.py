#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""usbeam 分析索引 —— 公共口径（唯一实现点）。

本模块承载三个口径，建库与查询两侧只允许从这里取：

1. `scan_markers()`：标记状态机，镜像 `tools/count_funcs.sh` 的 awk 实现。
   口径唯一真相源是 `count_funcs.sh`（docs/STRATEGY.md §4）；本函数逐条对齐它，
   建库后由 `build_index.py` 打印同格式计数供直接 diff 验证。
2. `func_key()`：函数名归一为「类型.方法」形式，已落地侧的唯一键来源
   （取代 `work/scripts/func_gap.py` 与 `work/scripts/cross_ref.py` 各自的实现）。
3. `va_to_off()` / `load_symbols()` / `annotate_symbol()`：PE 与符号表的底层读取。
   此前散在 12 份脚本里各写一遍，现已收敛到本模块。

研究用途。
"""

import bisect
import re
from functools import lru_cache
from pathlib import Path

# tools/index/common.py -> tools/ -> UsbEAmSource/
ENGINE = Path(__file__).resolve().parents[2]
REPO = ENGINE.parent

SYMBOLS_TXT = ENGINE / "docs" / "goresym" / "symbols.txt"
SOURCE_FUNCS_TXT = ENGINE / "docs" / "goresym" / "source_funcs.txt"
VA_MAP_FIXED2 = ENGINE / "docs" / "goresym" / "pipeline" / "va_map_fixed2.txt"
BACKEND_DIR = ENGINE / "backend"
DB_PATH = ENGINE / "work" / "index" / "usbeam.db"

# 目标二进制：不入库，坐标见 docs/ENV.md §1（SHA256 锁定）
TARGET_EXE = Path(r"D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe")

# 反汇编资产根（产物不入库，磁盘保留；见 docs/ENV.md §4）
ASSET_ROOTS = (
    ENGINE / "docs" / "goresym" / "pipeline" / "tmp",
    ENGINE / "docs" / "goresym" / "disasm",
    ENGINE / "docs" / "goresym" / "disasm_archive",
    ENGINE / "docs" / "goresym" / "disasm_assemble",
    ENGINE / "docs" / "goresym" / "disasm_assemble_image",
    ENGINE / "docs" / "goresym" / "disasm_lu",
    ENGINE / "docs" / "goresym" / "dump_archive",
    REPO / "work" / "disasm",
)

ASSET_SUFFIXES = (".asm.txt", ".asm", ".bin", ".dis.txt")

MARKER_ORDER = ("S-sig", "S-eq", "S-inline", "S", "P")

# 按需反汇编产物表（asmtool.py 写入；定义放这里保持 schema 单源）
ASM_SCHEMA = """
CREATE TABLE asm_line (
    va       INTEGER PRIMARY KEY,
    func_va  INTEGER NOT NULL,
    bytes    TEXT NOT NULL,
    mnemonic TEXT NOT NULL,
    ops      TEXT NOT NULL
);

CREATE TABLE call_edge (
    src_va   INTEGER NOT NULL,
    dst_va   INTEGER NOT NULL,
    dst_name TEXT,
    PRIMARY KEY (src_va, dst_va)
);

CREATE INDEX idx_asm_line_func ON asm_line(func_va);
CREATE INDEX idx_call_edge_dst ON call_edge(dst_va);
"""


def scan_markers(path):
    """扫描单个 .go 文件，返回 [(行号, func 签名, 标记 or None)]。

    严格镜像 count_funcs.sh 的 awk 状态机：
      - `^//` 行：仅当 last 为空时判定标记（块首行未命中则继续看块内后续行），不重置。
      - `^func ` 行：结算一次，重置 last。
      - 其它任何行（含空行）：重置 last。
    """
    rows = []
    last = ""
    with open(path, encoding="utf-8", errors="replace") as fh:
        for lineno, raw in enumerate(fh, 1):
            line = raw.rstrip("\r\n")
            if line.startswith("//"):
                if last == "":
                    last = _marker_of(line)
                continue
            if line.startswith("func "):
                rows.append((lineno, line[5:], last or None))
                last = ""
                continue
            last = ""
    return rows


def _marker_of(comment):
    """注释行 → 标记档位。判定顺序与 count_funcs.sh 逐条一致。"""
    for name in MARKER_ORDER:
        for tail in ("]", " ", "\t"):
            if "[" + name + tail in comment:
                return name
    return ""


def count_of(rows):
    """[行] → count_funcs.sh 同格式的计数字典。"""
    c = {k: 0 for k in ("S", "S-inline", "S-eq", "S-sig", "P")}
    marked = 0
    for _lineno, _sig, marker in rows:
        if marker:
            c[marker] += 1
            marked += 1
    total = len(rows)
    s, inl, eq, sig = c["S"], c["S-inline"], c["S-eq"], c["S-sig"]
    return {
        "FUNCS": total,
        "MARKED": marked,
        "UNMARKED": total - marked,
        "S": s,
        "S-inline": inl,
        "S-eq": eq,
        "S-sig": sig,
        "P": c["P"],
        "FAITHFUL": s + inl,
        "USABLE": s + inl + eq,
        "TRUE": s + inl + sig,
    }


def count_line(counts):
    """与 count_funcs.sh 完全相同的单行输出格式，便于直接 diff。"""
    order = ("FUNCS", "MARKED", "UNMARKED", "S", "S-inline", "S-eq", "S-sig", "P",
             "FAITHFUL", "USABLE", "TRUE")
    return "  ".join("%s=%d" % (k, counts[k]) for k in order)


def func_key(name):
    """蓝图 / 符号 / 源码函数名 → 规范化比较键（点号分隔，大小写敏感）。

    全仓唯一实现。取代先前并存的两套：索引侧的下划线小写键、以及
    `tools/list_missing.js` 的点号键——两套曾对同一份代码给出不同的缺失数。
    """
    n = name.strip()
    n = re.sub(r"^main\.", "", n)
    n = re.sub(r"^\(\*([^)]+)\)\.", r"\1.", n)                      # (*T).M -> T.M
    n = re.sub(r"^\(\*([^)]+)\)(?=\w)", r"\1.", n)                  # (*T)M  -> T.M
    n = re.sub(r"^\((?:\w+\s+)?\*?([\w\[\]]+)\)(?=\w)", r"\1.", n)  # (T)M / (r T)M -> T.M
    n = re.sub(r"^\(\*([^)]+)\)", r"\1", n)
    return n


def func_short_of(key):
    """比较键 → 短名（最后一段）。"""
    return key.rsplit(".", 1)[-1]


_RE_LOOSE = re.compile(r"[^0-9A-Za-z]")


def loose_key(s):
    """比较键 → 去分隔符形态（只留字母数字），用于跨形态配对。"""
    return _RE_LOOSE.sub("", s)


@lru_cache(maxsize=1)
def symbol_key_map():
    """符号表 → {loose 键: 规范点号键}。

    蓝图 `source_funcs.txt` 原文自身形态不统一：既有 `(*BootstrapService)CheckPluginUpdates`，
    也有 `volumeIndexReadViewfindNodeIndexByFRN`（类型名与方法名粘连，无任何分隔符）。
    符号表是编译级产物，形态规范，用它把粘连名还原成点号名——否则同一个函数在
    蓝图侧与源码侧会是两个键，缺口判定凭空多出假阳性，输出也没法读。
    """
    out = {}
    for name in load_symbols().values():
        if not name.startswith("main."):
            continue
        key = func_key(name)
        lk = loose_key(key)
        if len(lk) >= 4:
            out.setdefault(lk, key)
    return out


def resolve_func_key(name, sym_map=None):
    """蓝图函数名 → 规范比较键（点号形态）。粘连名经符号表字典还原。"""
    key = func_key(name)
    if sym_map is None:
        sym_map = symbol_key_map()
    return sym_map.get(loose_key(key), key)


# 函数声明解析：接收者可能是 (T) 或 (s *T)，两种都要吃下。
RE_FUNC_DEF = re.compile(r"^func\s+(?:\(\s*([^)]*?)\s*\)\s*)?(\w+)\s*\(", re.M)


def recv_type(recv_raw):
    """接收者原文 → 类型名。`s *memoryReleaseService` → `memoryReleaseService`。"""
    if not recv_raw:
        return ""
    words = re.findall(r"[A-Za-z_]\w*", recv_raw)
    return words[-1] if words else ""


def iter_func_defs(text):
    """文本 → [(接收者类型, 函数名)]。"""
    for m in RE_FUNC_DEF.finditer(text):
        yield (recv_type(m.group(1)), m.group(2))


def impl_keys_of(text):
    """文本 → 实现键集合（`类型.方法` 与 裸名 两种形态）。"""
    keys = set()
    for recv, name in iter_func_defs(text):
        if recv:
            keys.add(func_key("(%s)%s" % (recv, name)))
        keys.add(name)
    return keys


# 编译器生成物：闭包 funcN、defer 包装 deferwrapN、接口包装 gowrapN 等。
# 正则与 tools/list_missing.js 的 isGen 逐条一致（收敛前两边不同）。
RE_GENERATED = re.compile(r"(func\d+|deferwrap\d+|gowrap\d+|Printffunc\d+)(\.\d+)*$|\.\d+$|-fm$")

# 永不落地的名字（依据 HANDOFF.md 给新会话的三条铁律）：
#   nativeFileDrag* 的 IUnknown 三件套是 combridge 自动提升，不是活；
#   screenshotLinearFloatToSRGBByte 名字错配，落一次白落一次。
RE_NEVER_LAND = re.compile(
    r"^(nativeFileDragDataObject|nativeFileDragFormatEnumerator)\.(AddRef|Release|QueryInterface)$"
    r"|^screenshotLinearFloatToSRGBByte$")


def is_generated(key):
    """比较键是否编译器生成物。"""
    return bool(RE_GENERATED.search(key))


def is_never_land(key):
    """比较键是否属于「永不落地」名单。"""
    return bool(RE_NEVER_LAND.match(key))


def blueprint_funcs():
    """蓝图 → [(file, func)]，排除 <autogenerated>（它不是文件，是 58 个函数项）。"""
    return [(f, fn) for _pkg, f, fn, _s, _e, _n in parse_blueprint() if f != "<autogenerated>"]


def backend_impl_keys():
    """backend 非 _test 的全部实现键。"""
    keys = set()
    for path in backend_files():
        keys |= impl_keys_of(path.read_text(encoding="utf-8", errors="replace"))
    return keys


def classify_missing(impl=None):
    """唯一缺失判定。

    返回 [(file, func, key, reason)]，reason ∈ {MISSING, GEN, NEVER}：
      MISSING 需人工还原；GEN 编译器生成物；NEVER 走「永不落地」名单。
    分母是蓝图（排除 <autogenerated>），已实现集是 backend 非 _test 的函数定义；
    匹配只做精确键相等，不做短名回退——短名回退会把「别的类型里的同名方法」
    算成本函数的实现（list_missing.js 的 `run` 就是这样漏报的）。
    """
    if impl is None:
        impl = backend_impl_keys()
    sym_map = symbol_key_map()
    out = []
    for file_, func in blueprint_funcs():
        key = resolve_func_key(func, sym_map)
        if key in impl:
            continue
        if is_generated(key):
            reason = "GEN"
        elif is_never_land(key):
            reason = "NEVER"
        else:
            reason = "MISSING"
        out.append((file_, func, key, reason))
    return out


def norm_asset(basename):
    """反汇编资产名 → 归一化匹配键（小写，词首批次前缀去掉）。

    与 `func_short_of(func_key).lower()` 同形态，供资产命中判定直接相等比较。
    注意只做小写，不把 `_` 换成 `.`——Go 标识符里本来就有下划线。
    """
    b = re.sub(r"\.(asm\.txt|asm|bin|dis\.txt)$", "", basename)
    b = re.sub(r"^_?(b\d{1,3})_", "", b)
    b = re.sub(r"^_(prefix\d|sep|const)_", "", b)
    return b.lower().strip("_")


def backend_files():
    """backend 下参与口径统计的 .go（排除 _test.go），与 count_funcs.sh 同集。"""
    return sorted(p for p in BACKEND_DIR.glob("*.go") if not p.name.endswith("_test.go"))


@lru_cache(maxsize=1)
def load_symbols():
    """symbols.txt → {VA: name}（24181 条）。"""
    table = {}
    with open(SYMBOLS_TXT, encoding="utf-8", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            try:
                va, name = line.split(" ", 1)
            except ValueError:
                continue
            table[int(va, 16)] = name
    return table


@lru_cache(maxsize=1)
def _symbols_sorted():
    """→ (升序 VA 元组, {VA: name})，供 annotate_symbol 二分。"""
    syms = load_symbols()
    return tuple(sorted(syms)), syms


def annotate_symbol(target, window=0x2000):
    """VA → 符号名。精确命中给全名，否则给「最近符号+0xdelta」；超出 window 给 None。

    `tools/index/asmtool.py` 与 `tools/go-introspect/` 的反汇编入口都走这里，
    避免同一地址在不同脚本里标注成不同字符串。
    """
    vas, syms = _symbols_sorted()
    if target in syms:
        return syms[target]
    i = bisect.bisect_right(vas, target) - 1
    if i < 0 or target - vas[i] >= window:
        return None
    return "%s+0x%x" % (syms[vas[i]], target - vas[i])


def parse_blueprint():
    """source_funcs.txt → [(pkg, file, func, start_line, end_line, nlines)]。"""
    rows = []
    pkg = file_ = ""
    with open(SOURCE_FUNCS_TXT, encoding="utf-8", errors="replace") as fh:
        for raw in fh:
            line = raw.rstrip("\r\n")
            if line.startswith("Package "):
                pkg = line[8:].strip()
                continue
            if line.startswith("File: "):
                file_ = line[6:].strip()
                continue
            if not line.startswith("\t"):
                continue
            m = re.match(r"(.+?) Lines: (\d+) to (\d+) \((\d+)\)\s*$", line[1:])
            if m:
                rows.append((pkg, file_, m.group(1), int(m.group(2)), int(m.group(3)), int(m.group(4))))
    return rows


def iter_assets():
    """磁盘上的反汇编资产 → [(root, path, basename, norm, size)]。"""
    out = []
    for root in ASSET_ROOTS:
        if not root.is_dir():
            continue
        for path in root.rglob("*"):
            if not path.is_file() or not path.name.endswith(ASSET_SUFFIXES):
                continue
            try:
                size = path.stat().st_size
            except OSError:
                continue
            out.append((str(root), str(path), path.name, norm_asset(path.name), size))
    return out


def va_to_off(data, va):
    """VA(绝对地址) → 文件偏移；遍历全部节，不硬编码节名。"""
    pe = int.from_bytes(data[0x3C:0x40], "little")
    image_base = int.from_bytes(data[pe + 24 + 24:pe + 24 + 32], "little")
    rva = va - image_base
    nsect = int.from_bytes(data[pe + 6:pe + 8], "little")
    optsz = int.from_bytes(data[pe + 20:pe + 22], "little")
    sec = pe + 24 + optsz
    for i in range(nsect):
        o = sec + i * 40
        vsize = int.from_bytes(data[o + 8:o + 12], "little")
        va0 = int.from_bytes(data[o + 12:o + 16], "little")
        rawsz = int.from_bytes(data[o + 16:o + 20], "little")
        raw = int.from_bytes(data[o + 20:o + 24], "little")
        if va0 <= rva < va0 + max(vsize, rawsz) and raw:
            return raw + (rva - va0)
    return None
