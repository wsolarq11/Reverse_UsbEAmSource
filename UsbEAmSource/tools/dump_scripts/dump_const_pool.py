#!/usr/bin/env python
# 研究用途：dump 常量池 0x1411cd470..0x1411cd4c0 与 0x1411cd5f0.. 逐 4 字节解 float。
import os, struct, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "go-introspect"))
import va_dump

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
data = open(EXE, "rb").read()

def dump(lo, hi):
    print("--- 0x%x .. 0x%x ---" % (lo, hi))
    off = va_dump.va_to_off(data, lo)
    for va in range(lo, hi, 4):
        o = va_dump.va_to_off(data, va)
        u = struct.unpack_from("<I", data, o)[0]
        f = struct.unpack_from("<f", data, o)[0]
        print("  0x%x: 0x%08X = %.9g" % (va, u, f))

dump(0x1411cd470, 0x1411cd4c0)
dump(0x1411cd5e8, 0x1411cd608)
dump(0x1411caf10, 0x1411caf20)
