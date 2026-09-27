#!/usr/bin/env python
# 研究用途：批量 dump mousegestures/nativedrag 目标函数的序言+尾部，
# 用于签名实证（参数/返回寄存器判定）。
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import va_dump  # noqa: E402

EXE = sys.argv[1]
OUT = sys.argv[2]

# (name, VA) 主清单
TARGETS = [
    # mousegestures.go
    ("GestureSettings.UnmarshalJSON", 0x1408d5e60),
    ("GestureAppProfile.UnmarshalJSON", 0x1408d5fc0),
    ("normalizeMouseGestureConfig", 0x1408d6300),
    ("normalizeGestureSettings", 0x1408d6620),
    ("normalizeGestureButtons", 0x1408d6a40),
    ("normalizeGestureButton", 0x1408d6cc0),
    ("normalizeGestureAppProfiles", 0x1408d7060),
    ("normalizeGestureGlobalRulePriority", 0x1408d7d80),
    ("normalizeGestureAppMatch", 0x1408d7ec0),
    ("normalizeGestureAppMatches", 0x1408d8080),
    ("mergeGestureAppMatch", 0x1408d8640),
    ("gestureAppMatchKey", 0x1408d8c20),
    ("normalizeGestureRules", 0x1408d8ce0),
    ("normalizeGesturePattern", 0x1408d9580),
    ("normalizeGestureModifier", 0x1408d9700),
    ("normalizeGestureDirections", 0x1408d9b20),
    ("normalizeGestureDirection", 0x1408d9d20),
    ("normalizeGestureActionKind", 0x1408da940),
    ("normalizeGestureWindowCommand", 0x1408db520),
    ("normalizeGestureWindowMove", 0x1408db720),
    ("normalizeGestureURL", 0x1408db9c0),
    ("MouseGestureDirectionsFromPoints", 0x1408dc8e0),
    ("directionFromDelta", 0x1408dd800),
    ("gestureDirectionFromDelta", 0x1408dda00),
    ("mouseGestureClockwiseAngleFromUp", 0x1408dda80),
    ("mouseGestureAngleBetween", 0x1408ddb00),
    ("gestureDirectionDominatesDelta", 0x1408ddc40),
    ("mouseGestureTargetAllowsGestureStart", 0x1408ddde0),
    ("matchMouseGestureRule", 0x1408de040),
    ("findGestureAppProfile", 0x1408ded60),
    ("findGestureRule", 0x1408df380),
    ("gesturePatternHasDiagonal", 0x1408df7c0),
    ("FindHotCornerAtPoint", 0x1408df8c0),
    ("mouseGestureButtonStateBit", 0x1408e2a00),
    ("mouseGestureRuleDisplayName", 0x1408e3f00),
    ("gestureAppProfileFromWindowProcessPick", 0x1408e3fc0),
    ("mouseGesturePatternKey", 0x1408e4520),
    ("mouseGesturePatternPreviewKey", 0x1408e46e0),
    ("samePathString", 0x1408e4880),
    ("svc.Configure", 0x1408e0060),
    ("svc.GetState", 0x1408e0460),
    ("svc.UpdateConfig", 0x1408e0560),
    ("svc.SetRuntimeEnabled", 0x1408e06a0),
    ("svc.SetCaptureSuspended", 0x1408e0840),
    ("svc.SetHotCornerEnabled", 0x1408e09e0),
    ("svc.TestAction", 0x1408e0ce0),
    ("svc.TestHotCorner", 0x1408e12a0),
    ("svc.PickAppTarget", 0x1408e1800),
    ("svc.Shutdown", 0x1408e1c40),
    ("svc.currentModuleEnabled", 0x1408e1d20),
    ("svc.buildState", 0x1408e1e40),
    ("svc.setLastError", 0x1408e2180),
    ("svc.recordActionResult", 0x1408e2360),
    ("svc.recordGestureStatus", 0x1408e25a0),
    ("svc.setSuppressGestureButtonUp", 0x1408e28a0),
    ("svc.consumeSuppressGestureButtonUp", 0x1408e2960),
    ("svc.executeMatchedGestureAsync", 0x1408e2ae0),
    ("svc.beginRuntimeAction", 0x1408e2e20),
    ("svc.executeMatchedGestureForGeneration", 0x1408e2fa0),
    ("svc.recordActionResultForGeneration", 0x1408e3600),
    ("svc.recordGestureStatusForGeneration", 0x1408e3860),
    ("svc.executeGestureAction", 0x1408e3b80),
    # mousegestures_actions.go
    ("executeMouseGestureHotkey", 0x1408e4940),
    ("executeMouseGestureAction", 0x1408e49c0),
    ("resolveMouseGestureWindowMoveScreen", 0x1408e4fe0),
    ("prepareMouseGestureActionTarget", 0x1408e59e0),
    # mousegestures_actions_windows.go
    ("sendMouseGestureButtonClick", 0x1408e5c40),
    ("sendMouseGestureButtonDown", 0x1408e5d40),
    ("sendMouseGestureButtonUp", 0x1408e5e00),
    ("mouseGestureButtonInputFlags", 0x1408e5ec0),
    ("executeMouseGestureHotkeyPlatform", 0x1408e5fe0),
    ("executeMouseGestureTextInputPlatform", 0x1408e6080),
    ("executeMouseGestureTaskSwitchPlatform", 0x1408e6100),
    ("sendMouseGestureInputs", 0x1408e6160),
    ("buildMouseGestureTextInputInputs", 0x1408e6360),
    ("buildMouseGestureHotkeyInputs", 0x1408e65a0),
    ("mouseGestureVirtualKey", 0x1408e6b00),
    ("executeMouseGestureWindowCommandPlatform", 0x1408e6c40),
    ("mouseGestureMonitorHandleFromScreenID", 0x1408e7a00),
    ("mouseGestureMonitorInfo", 0x1408e7aa0),
    ("mouseGestureWindowMovePosition", 0x1408e7c40),
    ("activateMouseGestureActionTargetPlatform", 0x1408e7ca0),
    ("mouseGestureRootWindow", 0x1408e7ea0),
    ("mouseGestureShowWindow", 0x1408e7fa0),
    # mousegestures_overlay_windows.go
    ("sess.updateGestureOverlay", 0x1408e8100),
    ("sess.updateGestureLabelOverlay", 0x1408e8320),
    ("sess.finishGestureOverlay", 0x1408e8520),
    ("sess.hideGestureOverlay", 0x1408e8600),
    ("sess.closeGestureOverlay", 0x1408e8660),
    ("createMouseGestureOverlayWindow", 0x1408e8720),
    ("ensureMouseGestureOverlayWindowClass", 0x1408e88c0),
    ("win.run", 0x1408e8920),
    ("win.createWindow", 0x1408e8c20),
    ("mouseGestureOverlayWindowProc", 0x1408e8e40),
    ("win.handleMessage", 0x1408e9000),
    ("win.Update", 0x1408e9240),
    ("win.UpdateLabel", 0x1408e9460),
    ("win.Fade", 0x1408e9600),
    ("win.Hide", 0x1408e97a0),
    ("win.BringToTop", 0x1408e98c0),
    ("win.Close", 0x1408e9940),
    ("win.queuePayload", 0x1408e9b40),
    ("win.applyPendingPayloadOnThread", 0x1408e9ce0),
    ("win.applyPayloadOnThread", 0x1408e9e40),
    ("win.hideOnThread", 0x1408ea060),
    ("win.stepFadeOnThread", 0x1408ea1e0),
    ("win.redraw", 0x1408ea2c0),
    ("win.finishRedraw", 0x1408eaa80),
    ("win.render", 0x1408eaba0),
    ("win.paint", 0x1408eae00),
    ("win.setThreadID", 0x1408eafc0),
    ("win.setHandle", 0x1408eb060),
    ("win.handle", 0x1408eb100),
    ("mouseGestureOverlayBringToTop", 0x1408eaf00),
    ("buildMouseGestureOverlayPayload", 0x1408eb220),
    ("buildMouseGestureLabelOverlayPayload", 0x1408eb760),
    ("mouseGestureOverlayVisibleTrailPoints", 0x1408eb9e0),
    ("mouseGestureOverlayBounds", 0x1408ebfc0),
    ("mouseGestureOverlayLabelBounds", 0x1408ec1c0),
    ("mouseGestureOverlayLabelBoundsForScreen", 0x1408ec300),
    ("mouseGestureOverlayScreenForPoint", 0x1408ec820),
    ("renderMouseGestureOverlay", 0x1408ec940),
    ("mouseGestureOverlayDrawLabel", 0x1408ecf60),
    ("mouseGestureOverlayDrawText", 0x1408ed660),
    ("mouseGestureOverlayMeasureTextWidth", 0x1408ee060),
    ("mouseGestureOverlayCreateCanvasBitmap", 0x1408ee620),
    ("mouseGestureOverlayBlendTextDIBToRGBA", 0x1408ee8a0),
    ("mouseGestureOverlayDrawPolylineStroke", 0x1408eeb60),
    ("mouseGestureOverlaySimplifyStrokePoints", 0x1408eecc0),
    ("mouseGestureOverlayDrawMaxStroke", 0x1408eef20),
    ("mouseGestureOverlaySetMaxPixel", 0x1408ef4a0),
    ("mouseGestureOverlayFillRoundedRect", 0x1408ef660),
    ("mouseGestureOverlayBlendPixel", 0x1408ef900),
    ("mouseGestureOverlayPremultiplyAlpha", 0x1408efae0),
    ("mouseGestureOverlayEstimateTextWidth", 0x1408efcc0),
    # mousegestures_runtime_windows.go
    ("q.tryPushAndWake", 0x1408f0040),
    ("q.lockForPush", 0x1408f02e0),
    ("q.pushLocked", 0x1408f03c0),
    ("q.pop", 0x1408f09a0),
    ("svc.syncPlatformRuntime", 0x1408f0c80),
    ("svc.stopPlatformRuntime", 0x1408f1860),
    ("svc.stopPlatformRuntimeLocked", 0x1408f1980),
    ("svc.runtimeSnapshot", 0x1408f1b40),
    ("svc.capturePausedSnapshot", 0x1408f1d60),
    ("rt.run", 0x1408f1e80),
    ("rt.startButtonReplayWorker", 0x1408f25e0),
    ("rt.executeButtonReplay", 0x1408f2a20),
    ("rt.stopWatcher", 0x1408f2b20),
    ("rt.wakeRuntime", 0x1408f2bc0),
    ("rt.processPending", 0x1408f2c80),
    ("rt.installHook", 0x1408f2f20),
    ("rt.acceptHookEvent", 0x1408f3260),
    ("rt.uninstallHook", 0x1408f34c0),
    ("rt.shouldSuppressHookEvent", 0x1408f3620),
    ("rt.shouldQueueHookEvent", 0x1408f37e0),
    ("rt.queueAndWakeHookEvent", 0x1408f3880),
    ("rt.noteQueueFailure", 0x1408f3920),
    ("rt.shouldCaptureGestureButtonDown", 0x1408f39e0),
    ("rt.runtimeConfig", 0x1408f3d60),
    ("rt.gestureRuntimeEnabled", 0x1408f3ee0),
    ("rt.gestureTargetHWNDAtPoint", 0x1408f4020),
    ("rt.storeGestureTargetScope", 0x1408f40a0),
    ("mouseGestureHookEventFromWindows", 0x1408f4140),
    ("sess.handleEvent", 0x1408f4400),
    ("sess.handleGesture", 0x1408f49a0),
    ("sess.currentTime", 0x1408f5960),
    ("sess.sendDown", 0x1408f59a0),
    ("sess.sendUp", 0x1408f5a20),
    ("sess.sendClick", 0x1408f5aa0),
    ("sess.gestureTimeoutStatus", 0x1408f5b20),
    ("sess.armGestureTimeout", 0x1408f5c80),
    ("sess.scheduleGestureTimeoutWake", 0x1408f5e60),
    ("sess.stopGestureTimeout", 0x1408f6000),
    ("sess.handleTimeout", 0x1408f6100),
    ("sess.handleQueueFailure", 0x1408f6260),
    ("sess.shutdown", 0x1408f6320),
    ("sess.handleGestureMove", 0x1408f63a0),
    ("sess.shouldUpdateGestureOverlay", 0x1408f6ce0),
    ("sess.resetGestureSession", 0x1408f6da0),
    ("sess.awaitGestureButtonRelease", 0x1408f6f00),
    ("sess.completeAwaitingGestureButtonRelease", 0x1408f70a0),
    ("sess.queueGestureButtonClickReplay", 0x1408f71a0),
    ("sess.cancelGestureWithRestore", 0x1408f73c0),
    ("sess.finishGesture", 0x1408f75a0),
    ("sess.finishGestureTarget", 0x1408f7d40),
    ("sess.currentMatchedGestureLabel", 0x1408f8100),
    ("sess.observeGestureTargetScope", 0x1408f8460),
    ("sess.ensureGestureTargetResolved", 0x1408f86a0),
    ("sess.gestureTargetHWNDAtPoint", 0x1408f87e0),
    ("sess.gestureTargetAtPoint", 0x1408f8860),
    ("sess.handleHotCorner", 0x1408f89e0),
    ("executeHotCornerRule", 0x1408f91e0),
    ("mouseGestureRuntimeTargetHWNDAtPoint", 0x1408f9440),
    ("mouseGestureTargetFromHWND", 0x1408f94c0),
    ("mouseGestureWindowFromPoint", 0x1408f96e0),
    ("mouseGestureForegroundWindowFullscreen", 0x1408f9760),
    ("mouseGestureScreenBoundsFromRects", 0x1408f98e0),
    # nativedrag_windows.go
    ("nativeFileDragGetData", 0x1408f9d60),
    ("nativeFileDragGetDataHere", 0x1408f9de0),
    ("nativeFileDragQueryGetData", 0x1408f9e60),
    ("nativeFileDragGetCanonicalFormatEtc", 0x1408f9ec0),
    ("nativeFileDragSetData", 0x1408f9f40),
    ("nativeFileDragEnumFormatEtc", 0x1408f9fc0),
    ("nativeFileDragDAdvise", 0x1408fa040),
    ("nativeFileDragDUnadvise", 0x1408fa0e0),
    ("nativeFileDragEnumDAdvise", 0x1408fa140),
    ("nativeFileDragEnumFormatEtcNext", 0x1408fa1a0),
    ("nativeFileDragEnumFormatEtcSkip", 0x1408fa220),
    ("nativeFileDragEnumFormatEtcReset", 0x1408fa280),
    ("nativeFileDragEnumFormatEtcClone", 0x1408fa2c0),
    ("dob.GetData", 0x1408fa320),
    ("dob.GetDataHere", 0x1408fa480),
    ("dob.QueryGetData", 0x1408fa4a0),
    ("dob.GetCanonicalFormatEtc", 0x1408fa500),
    ("dob.SetData", 0x1408fa5a0),
    ("dob.EnumFormatEtc", 0x1408fa5c0),
    ("dob.DAdvise", 0x1408fa680),
    ("dob.DUnadvise", 0x1408fa6a0),
    ("dob.EnumDAdvise", 0x1408fa6c0),
    ("enum.Next", 0x1408fa720),
    ("enum.Skip", 0x1408fa8c0),
    ("enum.Reset", 0x1408fa900),
    ("enum.Clone", 0x1408fa920),
    ("bsvc.StartNativeFileDrag", 0x1408fa9a0),
    ("bsvc.StartAppEntryNativeDrag", 0x1408fab80),
    ("normalizeNativeFileDragPaths", 0x1408faea0),
    ("createTemporaryDirectoryShortcut", 0x1408fb1a0),
    ("sanitizeDirectoryShortcutName", 0x1408fb580),
    ("startNativeFileDrag", 0x1408fb700),
    ("ensureNativeFileDragOleInitialised", 0x1408fbac0),
    ("resolveNativeFileDragPreferredEffectFormat", 0x1408fbb00),
    ("registerNativeFileDragClipboardFormat", 0x1408fbb60),
    ("createNativeFileDragEnumFormatEtcAtIndex", 0x1408fbcc0),
    ("buildNativeFileDragHDropHandle", 0x1408fbe60),
    ("buildNativeFileDragDWORDHandle", 0x1408fc1c0),
    ("buildNativeFileDragWidePathList", 0x1408fc440),
]


def next_va(va):
    best = None
    for v in va_dump._SORTED:
        if v[0] > va:
            best = v[0]
            break
    return best


data = open(EXE, "rb").read()

import capstone  # noqa: E402
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

out = []
for name, va in TARGETS:
    nv = next_va(va)
    size = (nv - va) if nv else 0x400
    size = min(size, 0x800)
    off = va_dump.va_to_off(data, va)
    if off is None:
        out.append("=== %s @ 0x%x : NO SECTION ===" % (name, va))
        continue
    code = data[off:off + size]
    out.append("\n\n===== %s @ 0x%x  (size up to %d) =====" % (name, va, size))
    # prologue: first 0x50 bytes
    for insn in md.disasm(code[:0x50], va):
        out.append("0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str))
    # tail: last 0x30 bytes
    if len(code) > 0x50:
        tail = code[len(code) - 0x30:]
        tva = va + (len(code) - len(tail))
        out.append("  --tail--")
        for insn in md.disasm(tail, tva):
            out.append("0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str))

with open(OUT, "w", encoding="utf-8") as fh:
    fh.write("\n".join(out))
print("dumped %d funcs -> %s" % (len(TARGETS), OUT))
