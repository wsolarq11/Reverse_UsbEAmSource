#!/usr/bin/env python
# 研究用途：按 VA 区间直接从 PE 抽取字节并反汇编（含符号批注）。
#
# 相比 addr_tool.exe 的批量落盘，本工具按 (baseVA, length) 精确定位，
# 因此可以处理 addr_tool 漏掉的闭包符号（其 Name 以 "." 开头，落盘文件名异常），
# 也避免 pipeline/tmp 下同名 bin 互相覆盖。
#
# 用法:
#   python va_dump.py <exe> <baseVA> <length> <outPrefix>
# 产物:
#   <outPrefix>.bin      原始字节
#   <outPrefix>.asm.txt  反汇编（call/jmp 目标自动批注符号）
#
# PE 读取与符号批注的唯一实现在 tools/index/common.py（va_to_off / annotate_symbol）。
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "index"))

import common  # noqa: E402

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

import capstone  # noqa: E402

HEX_IN_OPS = re.compile(r"0x([0-9a-f]+)")


def main():
    if len(sys.argv) < 5:
        print(__doc__)
        return 2
    exe = sys.argv[1]
    base = int(sys.argv[2], 0)
    length = int(sys.argv[3], 0)
    out_prefix = sys.argv[4]

    data = open(exe, "rb").read()
    off = common.va_to_off(data, base)
    if off is None:
        print("未找到 VA 0x%x 所在节" % base)
        return 1
    code = data[off:off + length]
    if len(code) < length:
        print("警告：实际只取到 %d/%d 字节" % (len(code), length))

    Path(out_prefix + ".bin").write_bytes(code)

    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    lines = []
    for insn in md.disasm(code, base):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith(("call", "j")):
            m = HEX_IN_OPS.search(insn.op_str)
            if m:
                name = common.annotate_symbol(int(m.group(1), 16))
                if name:
                    line += "   ; %s" % name
        lines.append(line)

    Path(out_prefix + ".asm.txt").write_text("\n".join(lines), encoding="utf-8")
    print("0x%x +%d -> %s.bin / %s.asm.txt (%d 行)" % (base, length, out_prefix, out_prefix, len(lines)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
