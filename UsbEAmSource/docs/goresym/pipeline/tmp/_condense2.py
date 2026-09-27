# 研究用途：打印 windows.go 自由函数 asm 的开头 12 行 + 结尾 8 行。
import os

OUT = r"D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\pipeline\tmp"
names = [
    "win_supported", "win_newCursorController", "win_newWindowsCursorController",
    "win_currentIdleSeconds", "win_applyWindowChrome", "win_suppressBorder",
    "win_focusWindow", "win_focusWindowNow", "win_cursorPhysicalPoint",
    "win_pressedKeys", "win_keyboardVirtualKeys", "win_idlePauseActive",
    "win_observeBrowserContinuity", "win_standaloneBrowserAtCursor", "win_windowContainsPoint",
    "win_candidateIsStandalone", "win_walkActiveRenderSessions", "win_walkRenderEndpointSessions",
    "win_rememberInteractedBrowser", "win_foregroundFullscreen", "win_foregroundPlaybackActive",
    "win_targetMonitorsPlaying", "win_collectTargetCandidates", "win_collectVisibleMediaCandidates",
    "win_collectScreenCandidates", "win_candidatesContainCoarseApp", "win_candidatesContainPath",
    "win_countVisiblePlayingCoarse", "win_rememberBrowserContinuity", "win_playingMediaCandidatesCtx",
    "win_browserAudioMatchesVisible", "win_browserContinuityMatches", "win_enumTopLevelWindows",
    "win_visibleRectForScreenPause", "win_windowTitle", "win_windowTargetVisibleRects",
    "win_rectHasVisibleArea", "win_rectSubtract", "win_windowMatchesMonitors",
    "win_windowMonitorRect", "win_monitorRectsForScreens", "win_monitorRectsMatch",
    "win_mediaPauseProcessExcluded", "win_audioSessionMatchesForeground", "win_audioSessionMatchesCandidate",
    "win_mediaSessionSourceContainsPID", "win_mediaTitleMatchesWindow", "win_normalizeComparableText",
    "win_activeAudioCountsByPath", "win_foregroundMediaPauseDecision", "win_foregroundAudioActive",
    "win_anyCandidateActiveAudio", "win_mediaPauseDecisionForCandidates", "win_mediaSourceMatchesCandidate",
    "win_processAppUserModelID", "win_withWinRT", "win_requestMediaSessionManager",
    "win_roGetActivationFactory", "win_newHString", "win_deleteHString",
    "win_hstringToString", "win_queryInterface",
]

for n in names:
    p = os.path.join(OUT, n + ".asm.txt")
    if not os.path.exists(p):
        print("### %s : MISSING" % n)
        continue
    lines = open(p, encoding="utf-8").read().splitlines()
    head = [l for l in lines if "int3" not in l][:12]
    tail = [l for l in lines if "int3" not in l][-8:]
    print("### %s  (%d insn)" % (n, len(lines)))
    for l in head:
        print("  " + l)
    print("  ...")
    for l in tail:
        print("  " + l)
    print()
