# 研究用途：打印每个 asm 文件的开头 14 行 + 结尾 12 行，用于签名判读。
import os

OUT = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\pipeline\tmp"
names = [
    "overlay_boundsForScreen", "overlay_SetBounds", "overlay_Show", "overlay_Hide",
    "overlay_Focus", "overlay_Close", "overlay_NativeWindowHandle", "overlay_releaseFromService",
    "owin_UseNativeOverlayWindow", "owin_createNative", "owin_ensureClass", "owin_run",
    "owin_createWindow", "owin_windowProc", "owin_handleMessage", "owin_Show",
    "owin_Hide", "owin_Close", "owin_Focus", "owin_SetBounds",
    "owin_NativeWindowHandle", "owin_applyWindowChrome", "owin_showOnThread", "owin_hideOnThread",
    "owin_focusOnThread", "owin_setBoundsOnThread", "owin_paint", "owin_setHandle",
    "owin_handle", "owin_setThreadID", "owin_Bounds",
    "hotkey_applyBindings", "hotkey_buildRegistrations", "hotkey_failedResult",
]

for n in names:
    p = os.path.join(OUT, n + ".asm.txt")
    if not os.path.exists(p):
        print("### %s : MISSING" % n)
        continue
    lines = open(p, encoding="utf-8").read().splitlines()
    # strip annotation-only trailing int3
    head = [l for l in lines if "int3" not in l][:14]
    tail = [l for l in lines if "int3" not in l][-12:]
    print("### %s  (%d insn)" % (n, len(lines)))
    for l in head:
        print("  " + l)
    print("  ...")
    for l in tail:
        print("  " + l)
    print()
