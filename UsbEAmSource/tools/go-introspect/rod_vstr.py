#!/usr/bin/env python
# 研究用途：对二进制在某 baseVA 反汇编，收集 lea rax,[rip+imm] 的 .rodata 目标并读出 UTF-8 串。
# 用法: rod_vstr.py <exe(文件,VA映射用相对文件夹)> <bins文件夹>/<baseVA>
import sys, struct, gzip
try:
    sys.stdout.reconfigure(encoding='utf-8', errors='replace')
except Exception:
    pass
def va_to_off(data, va):
    pe = struct.unpack_from('<I', data, 0x3c)[0]
    opt = pe + 24
    magic = struct.unpack_from('<H', data, opt)[0]
    nsect = struct.unpack_from('<H', data, pe+6)[0]
    optsz = struct.unpack_from('<H', data, pe+20)[0]
    sec = opt + optsz
    for i in range(nsect):
        o = sec + i*40
        vsize = struct.unpack_from('<I', data, o+8)[0]
        va0 = struct.unpack_from('<I', data, o+12)[0]
        rawsz = struct.unpack_from('<I', data, o+16)[0]
        raw = struct.unpack_from('<I', data, o+20)[0]
        if va0 <= va < va0+max(vsize, rawsz) and raw:
            return raw + (va - va0)
    return None
def read_str(data, off):
    end = data.index(b'\x00', off)
    return data[off:end].decode('utf-8', 'replace')
def main():
    exe = sys.argv[1]; base = int(sys.argv[2], 16); bins = sys.argv[3]
    data = open(exe,'rb').read()
    pe0 = struct.unpack_from('<I', data, 0x3c)[0]
    ibase = struct.unpack_from('<Q', data, pe0 + 24 + 24)[0]  # optional header ImageBase
    # bin 内嵌汇编字节
    code = open(bins,'rb').read()
    import capstone
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)
    md.detail = True
    for insn in md.disasm(code, base):
        # lea rax, [rip + imm]
        op = insn.operands[0] if len(insn.operands) else None
        if insn.mnemonic == 'lea' and insn.operands[0].type == capstone.x86.X86_OP_REG and insn.operands[0].reg in (capstone.x86.X86_REG_RAX, capstone.x86.X86_REG_RCX):
            for o in insn.operands[1:]:
                if o.type == capstone.x86.X86_OP_MEM and o.mem.base == capstone.x86.X86_REG_RIP:
                    target = insn.address + insn.size + o.mem.disp
                    off = va_to_off(data, target - ibase)
                    if off is not None:
                        # Go 常量池无分隔；按最大 40 字节预览以识别 6/7 字符枚举串
                        print(f"0x{target:x} len? -> {data[off:off+40]!r}")
    print('done')
if __name__ == '__main__':
    main()
