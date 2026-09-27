#!/usr/bin/env python
# 研究用途：从 PE 解码 .rodata 字符串常量
import struct, sys

exe = sys.argv[1]
data = open(exe, 'rb').read()
pe0 = struct.unpack_from('<I', data, 0x3c)[0]
opt = pe0 + 24
nsect = struct.unpack_from('<H', data, pe0 + 6)[0]
optsz = struct.unpack_from('<H', data, pe0 + 20)[0]
ibase = struct.unpack_from('<Q', data, pe0 + 24 + 24)[0]
sec = opt + optsz

def va_to_off(va):
    rva = va - ibase
    for i in range(nsect):
        o = sec + i * 40
        vsize = struct.unpack_from('<I', data, o + 8)[0]
        va0 = struct.unpack_from('<I', data, o + 12)[0]
        rawsz = struct.unpack_from('<I', data, o + 16)[0]
        raw = struct.unpack_from('<I', data, o + 20)[0]
        if va0 <= rva < va0 + max(vsize, rawsz) and raw:
            return raw + (rva - va0)
    return None

# normalizeLinkIconMode targets
targets = {
    0x140C376B6: "6B return (upload, both paths)",
    0x140C39449: "7B return (favicon)",
    0x140C47539: "11B memequal constant (IconData prefix)",
}

for va, desc in sorted(targets.items()):
    off = va_to_off(va)
    if off:
        raw = data[off:off+30]
        null = raw.find(b'\x00')
        s = raw[:null].decode('utf-8', errors='replace') if null >= 0 else repr(raw)
        print(f"VA 0x{va:x} ({desc}):")
        print(f"  file off 0x{off:x}  raw={raw.hex()}")
        print(f"  string={s!r}")
    else:
        print(f"VA 0x{va:x} ({desc}): NOT FOUND")

# Also decode few more targets from normalizeAppEntries
# 0x140889780 is normalizeAppEntries
# Let me also check the call at 0x14088b199: lea rbx,[rip+0x3bc399]
# target = 0x14088b199 + 7 + 0x3bc399 = 0x140C47539
print()
print("=== Additional: normalizeAppEntries constants ===")
# From normalizeLinkEntries asm, at 0x14088ab44 calls normalizeLinkIconMode
# The 11B constant at 0x140C47539 - let me also check for the "data:image/" prefix
# and the .rodata strings around it
off = va_to_off(0x140C47539)
if off:
    for d in range(-20, 40):
        o = off + d
        if o >= 0 and o < len(data):
            b = data[o]
            if b >= 0x20 and b < 0x7f:
                # print char context
                pass
    # Check if there's a pattern like "data:image/" around this offset
    # Read a wider range
    wider = data[off-8:off+40]
    print(f"Context around 0x140C47539: {wider.hex()}")
    print(f"  repr: {wider!r}")
    # Try to find any null-terminated string containing this
    for start in range(max(0, off-20), off+20):
        if data[start] == 0:
            continue
        end = data.find(b'\x00', start)
        if end > start and end - start < 40:
            s = data[start:end].decode('utf-8', errors='replace')
            if len(s) >= 4:
                print(f"  string at off 0x{start:x}: {s!r}")