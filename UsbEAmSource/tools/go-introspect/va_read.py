#!/usr/bin/env python3
"""va_read.py - 按 VA 直读 PE 文件字节，自动解析节表。

用法:
    python va_read.py <exe> <va_hex> <len> [<va_hex> <len> ...]

输出每个 (VA, len) 的: 命中节名 / 文件偏移 / raw hex / UTF-8 与 GBK 解码。
VA 不在任何节内时明确报告（通常意味着指向运行时区，静态不可读）。
"""
import struct
import sys


def parse_sections(data):
    """解析 PE 节表，返回 [(name, va, vsize, rawptr, rawsize), ...]"""
    if data[:2] != b"MZ":
        raise SystemExit("not a PE file (no MZ)")
    pe_off = struct.unpack_from("<I", data, 0x3C)[0]
    if data[pe_off : pe_off + 4] != b"PE\0\0":
        raise SystemExit("not a PE file (no PE signature)")
    nsec = struct.unpack_from("<H", data, pe_off + 6)[0]
    opt_size = struct.unpack_from("<H", data, pe_off + 20)[0]
    base = struct.unpack_from("<Q", data, pe_off + 24 + 24)[0]
    sec_off = pe_off + 24 + opt_size
    out = []
    for i in range(nsec):
        off = sec_off + i * 40
        name = data[off : off + 8].rstrip(b"\0").decode("ascii", "replace")
        vsize = struct.unpack_from("<I", data, off + 8)[0]
        va = struct.unpack_from("<I", data, off + 12)[0] + base
        rawsize = struct.unpack_from("<I", data, off + 16)[0]
        rawptr = struct.unpack_from("<I", data, off + 20)[0]
        out.append((name, va, vsize, rawptr, rawsize))
    return out, base


def main():
    if len(sys.argv) < 4:
        raise SystemExit(__doc__)
    exe = sys.argv[1]
    with open(exe, "rb") as f:
        data = f.read()
    sections, base = parse_sections(data)
    print(f"ImageBase = 0x{base:x}")
    for name, va, vsize, rawptr, rawsize in sections:
        print(f"  {name:<10} VA=0x{va:x} VSize=0x{vsize:x} RAW=0x{rawptr:x} RAWSize=0x{rawsize:x}")

    pairs = sys.argv[2:]
    for i in range(0, len(pairs) - 1, 2):
        target = int(pairs[i], 16)
        length = int(pairs[i + 1], 0)
        hit = None
        for name, va, vsize, rawptr, rawsize in sections:
            span = max(vsize, rawsize)
            if va <= target < va + span:
                hit = (name, va, rawptr)
                break
        print(f"\n--- VA=0x{target:x} len={length} ---")
        if hit is None:
            print("  MISS: VA 不在任何节内")
            continue
        name, sec_va, rawptr = hit
        foff = rawptr + (target - sec_va)
        raw = data[foff : foff + length]
        if len(raw) < length:
            print(f"  TRUNCATED: 只读到 {len(raw)} 字节（文件尾?）")
        print(f"  section={name} file_off=0x{foff:x}")
        print(f"  raw_hex={raw.hex()}")
        for enc in ("utf-8", "gbk"):
            try:
                print(f"  {enc}: {raw.decode(enc)}")
                break
            except UnicodeDecodeError as e:
                print(f"  {enc}: FAIL ({e})")


if __name__ == "__main__":
    main()
