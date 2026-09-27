# 研究用途：批量反汇编 oledblackout 域函数到独立 asm.txt。
import os
import sys

TOOLS = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\tools\go-introspect"
sys.path.insert(0, TOOLS)
import va_dump  # noqa: E402  (load_symbols / va_to_off / annotate)

EXE = r"D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe"
OUT = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\pipeline\tmp"

import capstone  # noqa: E402

TARGETS = [
    # overlay.go
    ("overlay_boundsForScreen", 0x140910ae0),
    ("overlay_SetBounds", 0x140910ba0),
    ("overlay_Show", 0x140910ce0),
    ("overlay_Hide", 0x140910e00),
    ("overlay_Focus", 0x140910e80),
    ("overlay_Close", 0x140910f00),
    ("overlay_NativeWindowHandle", 0x140910f80),
    ("overlay_releaseFromService", 0x140911000),
    # overlay_windows.go
    ("owin_UseNativeOverlayWindow", 0x1409111c0),
    ("owin_createNative", 0x1409111e0),
    ("owin_ensureClass", 0x140911440),
    ("owin_run", 0x1409114a0),
    ("owin_createWindow", 0x1409117a0),
    ("owin_windowProc", 0x140911a40),
    ("owin_handleMessage", 0x140911c00),
    ("owin_Show", 0x140911f80),
    ("owin_Hide", 0x140912000),
    ("owin_Close", 0x140912080),
    ("owin_Focus", 0x140912180),
    ("owin_SetBounds", 0x140912200),
    ("owin_NativeWindowHandle", 0x140912340),
    ("owin_applyWindowChrome", 0x140912380),
    ("owin_showOnThread", 0x140912520),
    ("owin_hideOnThread", 0x140912660),
    ("owin_focusOnThread", 0x1409126e0),
    ("owin_setBoundsOnThread", 0x140912860),
    ("owin_paint", 0x140912a00),
    ("owin_setHandle", 0x140912c00),
    ("owin_handle", 0x140912ca0),
    ("owin_setThreadID", 0x140912dc0),
    ("owin_Bounds", 0x140912e60),
    # hotkey_windows.go
    ("hotkey_applyBindings", 0x14090fbc0),
    ("hotkey_buildRegistrations", 0x140910140),
    ("hotkey_failedResult", 0x140910980),
    # windows.go free funcs
    ("win_supported", 0x1409134c0),
    ("win_newCursorController", 0x1409134e0),
    ("win_newWindowsCursorController", 0x140913520),
    ("win_currentIdleSeconds", 0x140914420),
    ("win_applyWindowChrome", 0x140914580),
    ("win_suppressBorder", 0x1409147c0),
    ("win_focusWindow", 0x140914860),
    ("win_focusWindowNow", 0x1409148c0),
    ("win_cursorPhysicalPoint", 0x140914aa0),
    ("win_pressedKeys", 0x140914ac0),
    ("win_keyboardVirtualKeys", 0x140914ba0),
    ("win_idlePauseActive", 0x140914d20),
    ("win_observeBrowserContinuity", 0x140914f40),
    ("win_standaloneBrowserAtCursor", 0x1409152e0),
    ("win_windowContainsPoint", 0x1409157a0),
    ("win_candidateIsStandalone", 0x1409158c0),
    ("win_walkActiveRenderSessions", 0x140915ce0),
    ("win_walkRenderEndpointSessions", 0x140916840),
    ("win_rememberInteractedBrowser", 0x1409178e0),
    ("win_foregroundFullscreen", 0x140917cc0),
    ("win_foregroundPlaybackActive", 0x140917ec0),
    ("win_targetMonitorsPlaying", 0x1409181c0),
    ("win_collectTargetCandidates", 0x140918520),
    ("win_collectVisibleMediaCandidates", 0x140918600),
    ("win_collectScreenCandidates", 0x1409186e0),
    ("win_candidatesContainCoarseApp", 0x1409191c0),
    ("win_candidatesContainPath", 0x140919480),
    ("win_countVisiblePlayingCoarse", 0x1409195e0),
    ("win_rememberBrowserContinuity", 0x140919c00),
    ("win_playingMediaCandidatesCtx", 0x14091a0a0),
    ("win_browserAudioMatchesVisible", 0x14091a8c0),
    ("win_browserContinuityMatches", 0x14091ad00),
    ("win_enumTopLevelWindows", 0x14091b260),
    ("win_visibleRectForScreenPause", 0x14091b560),
    ("win_windowTitle", 0x14091b780),
    ("win_windowTargetVisibleRects", 0x14091b8c0),
    ("win_rectHasVisibleArea", 0x14091bb40),
    ("win_rectSubtract", 0x14091bdc0),
    ("win_windowMatchesMonitors", 0x14091c020),
    ("win_windowMonitorRect", 0x14091c100),
    ("win_monitorRectsForScreens", 0x14091c240),
    ("win_monitorRectsMatch", 0x14091c6c0),
    ("win_mediaPauseProcessExcluded", 0x14091c720),
    ("win_audioSessionMatchesForeground", 0x14091c9a0),
    ("win_audioSessionMatchesCandidate", 0x14091cac0),
    ("win_mediaSessionSourceContainsPID", 0x14091cde0),
    ("win_mediaTitleMatchesWindow", 0x14091cea0),
    ("win_normalizeComparableText", 0x14091cf80),
    ("win_activeAudioCountsByPath", 0x14091d100),
    ("win_foregroundMediaPauseDecision", 0x14091d200),
    ("win_foregroundAudioActive", 0x14091da20),
    ("win_anyCandidateActiveAudio", 0x14091dd20),
    ("win_mediaPauseDecisionForCandidates", 0x14091e160),
    ("win_mediaSourceMatchesCandidate", 0x14091eb60),
    ("win_processAppUserModelID", 0x14091ef40),
    ("win_withWinRT", 0x14091f220),
    ("win_requestMediaSessionManager", 0x14091f560),
    ("win_roGetActivationFactory", 0x14091f9c0),
    ("win_newHString", 0x14091fc80),
    ("win_deleteHString", 0x14091fde0),
    ("win_hstringToString", 0x14091fe40),
    ("win_queryInterface", 0x14091ff40),
]

SORTED = sorted(va_dump.SYMBOLS.items())


def next_len(va):
    for nva, _name in SORTED:
        if nva > va:
            l = nva - va
            if l <= 0x1200:
                return l
            return min(l, 0x1200)
    return 0x200


data = open(EXE, "rb").read()
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

for prefix, va in TARGETS:
    length = next_len(va)
    off = va_dump.va_to_off(data, va)
    if off is None:
        print("MISS VA %x %s" % (va, prefix))
        continue
    code = data[off:off + length]
    with open(os.path.join(OUT, prefix + ".bin"), "wb") as fh:
        fh.write(code)
    lines = []
    for insn in md.disasm(code, va):
        line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
        if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
            m = __import__("re").search(r"0x([0-9a-f]+)", insn.op_str)
            if m:
                name = va_dump.annotate(int(m.group(1), 16))
                if name:
                    line += "   ; %s" % name
        lines.append(line)
    with open(os.path.join(OUT, prefix + ".asm.txt"), "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))
    print("0x%x +0x%x -> %s (%d insn)" % (va, length, prefix, len(lines)))

print("DONE", len(TARGETS))
