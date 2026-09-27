import re, struct, pefile

exe = r'D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe'
asm = open(r'docs/goresym/pipeline/tmp/fs_deffilters_full.asm.txt', encoding='utf-8').read().splitlines()

leas = []   # (ip, reg, addr)
lens = []   # (ip, off, imm)
for ln in asm:
    s = ln.strip()
    m = re.match(r'0x([0-9a-f]+):\s+lea\s+(\w+),\s*\[rip\s*\+\s*(0x[0-9a-f]+)\]', s)
    if m:
        ip = int(m.group(1), 16)
        disp = int(m.group(3), 16)
        leas.append((ip, m.group(2), ip + 7 + disp))
        continue
    m = re.match(r'0x([0-9a-f]+):\s+mov\s+qword ptr \[rax\s*\+\s*(0x[0-9a-f]+|\d+)\],\s*(0x[0-9a-f]+|\d+)', s)
    if m:
        ip = int(m.group(1), 16)
        off = int(m.group(2), 0)
        imm = int(m.group(3), 0)
        lens.append((ip, off, imm))

print('total leas', len(leas), 'total lens', len(lens))

bounds = [0x1408125b2, 0x140812ed7, 0x140813848, 0x140813cf9, 0x140814285, 0x140814960, 0x140815073]

pe = pefile.PE(exe)
base = 0x140000000

def readstr(va, n):
    d = pe.get_data(va - base, n)
    return d.decode('utf-8', 'replace')

# 每个区域：字符串 lea（排除 lea rax 旧指针 / lea rsi et）
for bi in range(6):
    lo = bounds[bi]
    hi = bounds[bi + 1]
    # 字符串 lea：寄存器 rdx 或 rcx（growslice 参数是 rax/rsi）
    region_leas = [(ip, a) for ip, reg, a in leas if lo < ip < hi and reg in ('rdx', 'rcx')]
    region_lens = sorted([(off, imm) for ip, off, imm in lens if lo < ip < hi], key=lambda x: x[0])
    strs = [a for ip, a in region_leas]
    lengths = [imm for off, imm in region_lens if off % 0x10 == 8]
    rules = []
    for i in range(min(len(strs), len(lengths))):
        if 0 < lengths[i] < 20:
            rules.append(readstr(strs[i], lengths[i]))
    print(f'=== filter[{bi}] ({len(rules)}) ===')
    print(', '.join(repr(r) for r in rules))
    print()
