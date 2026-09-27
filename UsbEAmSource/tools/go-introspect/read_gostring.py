#!/usr/bin/env python
# 研究用途：按 VA + 字节长度直接从 PE 抽取 Go 字符串字面量常量。
#
# 场景：反汇编里 `lea rax, [rip + 0xNNN]` 装载的是字符串头指针，
# 长度立即数（mov ebx, 0xNN）紧随其后；RIP 相对地址需自算（lea 末地址 + disp）。
# 本工具只做 VA→字节→文本，不做反汇编，用于批量解码格式串/命名空间/错误消息常量。
#
# 用法:
#   python read_gostring.py <exe> <VA:len> [<VA:len> ...]
#   python read_gostring.py <exe> --off <VA+disp>:<len> ...   # 打印解析后的绝对 VA
import struct
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass


def va_to_off(data, va):
    """VA(绝对地址) -> 文件偏移；遍历全部节，不硬编码 .rdata。"""
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


def decode(raw):
    """Go 字符串常量以 UTF-8 存字节；控制字符按 \\xNN 转义。"""
    out = []
    for b in raw:
        if 0x20 <= b < 0x7F:
            out.append(chr(b))
        elif b == 0x0A:
            out.append("\\n")
        elif b == 0x00:
            out.append("\\0")
        else:
            out.append("\\x%02x" % b)
    return "".join(out)


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        return 2
    exe = sys.argv[1]
    specs = sys.argv[2:]
    data = open(exe, "rb").read()

    for spec in specs:
        if ":" not in spec:
            print("跳过无长度规格: %s" % spec)
            continue
        va_s, len_s = spec.rsplit(":", 1)
        va = int(va_s, 0)
        length = int(len_s, 0)
        off = va_to_off(data, va)
        if off is None:
            print("0x%x len=%d -> 未找到所在节" % (va, length))
            continue
        raw = data[off:off + length]
        print("0x%x len=%d -> %r" % (va, length, decode(raw)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
