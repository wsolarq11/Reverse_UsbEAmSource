#!/usr/bin/env python
# 研究用途：批量 dump qrcode_windows/qrcode/qrexternal 未落地函数的反汇编。
# 每个函数输出到 tmp/qrw/<shortname>.asm.txt，并在 stdout 打印序言摘要。
import os, re, struct, sys

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
import capstone

BASE = os.path.dirname(os.path.abspath(__file__))
EXE = r"D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe"
SYM_PATH = os.path.join(BASE, "..", "..", "symbols.txt")
OUT = os.path.join(BASE, "qrw")
os.makedirs(OUT, exist_ok=True)

SYMS = {}
with open(SYM_PATH, encoding="utf-8", errors="replace") as fh:
    for line in fh:
        line = line.strip()
        if not line:
            continue
        try:
            va, name = line.split(" ", 1)
        except ValueError:
            continue
        SYMS[int(va, 16)] = name
SORTED = sorted(SYMS.items())

# (shortname, symbol-name)
TARGETS = [
 ("readQRCodeTextFromClipboard", "main.readQRCodeTextFromClipboard"),
 ("writeQRCodeTextToClipboard", "main.writeQRCodeTextToClipboard"),
 ("runResult", "main.qrCodeScreenSelectionSession.runResult"),
 ("shouldFinalizeAfterOverlayClosed", "main.qrCodeScreenSelectionSession.shouldFinalizeAfterOverlayClosed"),
 ("startHDRPreviewUpgrade", "main.qrCodeScreenSelectionSession.startHDRPreviewUpgrade"),
 ("finishHDRPreviewUpgrade", "main.qrCodeScreenSelectionSession.finishHDRPreviewUpgrade"),
 ("setPendingHDRPreview", "main.qrCodeScreenSelectionSession.setPendingHDRPreview"),
 ("takePendingHDRPreview", "main.qrCodeScreenSelectionSession.takePendingHDRPreview"),
 ("createWindow", "main.qrCodeScreenSelectionSession.createWindow"),
 ("close", "main.qrCodeScreenSelectionSession.close"),
 ("setWindowHandle", "main.qrCodeScreenSelectionSession.setWindowHandle"),
 ("windowHandle", "main.qrCodeScreenSelectionSession.windowHandle"),
 ("setHitTestTransparent", "main.qrCodeScreenSelectionSession.setHitTestTransparent"),
 ("isHitTestTransparent", "main.qrCodeScreenSelectionSession.isHitTestTransparent"),
 ("markTimedOut", "main.qrCodeScreenSelectionSession.markTimedOut"),
 ("didTimeOut", "main.qrCodeScreenSelectionSession.didTimeOut"),
 ("showSelectionWindow", "main.qrCodeScreenSelectionSession.showSelectionWindow"),
 ("handleMessage", "main.qrCodeScreenSelectionSession.handleMessage"),
 ("beginRightButtonAction", "main.qrCodeScreenSelectionSession.beginRightButtonAction"),
 ("applyRightButtonAction", "main.qrCodeScreenSelectionSession.applyRightButtonAction"),
 ("pollRightButtonState", "main.qrCodeScreenSelectionSession.pollRightButtonState"),
 ("updateRightButtonPhysicalState", "main.qrCodeScreenSelectionSession.updateRightButtonPhysicalState"),
 ("cancelSelection", "main.qrCodeScreenSelectionSession.cancelSelection"),
 ("cancelAndClose", "main.qrCodeScreenSelectionSession.cancelAndClose"),
 ("handleHDRPreviewUpgrade", "main.qrCodeScreenSelectionSession.handleHDRPreviewUpgrade"),
 ("applyPendingHDRPreviewForFinalImage", "main.qrCodeScreenSelectionSession.applyPendingHDRPreviewForFinalImage"),
 ("ensureHDRPreviewForFinalImage", "main.qrCodeScreenSelectionSession.ensureHDRPreviewForFinalImage"),
 ("isHDRPreviewUpgradeInProgress", "main.qrCodeScreenSelectionSession.isHDRPreviewUpgradeInProgress"),
 ("replaceSnapshotBitmaps", "main.qrCodeScreenSelectionSession.replaceSnapshotBitmaps"),
 ("releaseSnapshotBitmaps", "main.qrCodeScreenSelectionSession.releaseSnapshotBitmaps"),
 ("paint", "main.qrCodeScreenSelectionSession.paint"),
 ("activeSelectionRect", "main.qrCodeScreenSelectionSession.activeSelectionRect"),
 ("selectionResizeHandleAt", "main.qrCodeScreenSelectionSession.selectionResizeHandleAt"),
 ("beginSelectionResize", "main.qrCodeScreenSelectionSession.beginSelectionResize"),
 ("updateSelectionResize", "main.qrCodeScreenSelectionSession.updateSelectionResize"),
 ("finishSelectionResize", "main.qrCodeScreenSelectionSession.finishSelectionResize"),
 ("updateSelectionResizeHover", "main.qrCodeScreenSelectionSession.updateSelectionResizeHover"),
 ("beginCornerRadiusDrag", "main.qrCodeScreenSelectionSession.beginCornerRadiusDrag"),
 ("updateCornerRadiusDrag", "main.qrCodeScreenSelectionSession.updateCornerRadiusDrag"),
 ("finishCornerRadiusDrag", "main.qrCodeScreenSelectionSession.finishCornerRadiusDrag"),
 ("cornerRadiusSliderHitRect", "main.qrCodeScreenSelectionSession.cornerRadiusSliderHitRect"),
 ("invalidateCornerRadiusControl", "main.qrCodeScreenSelectionSession.invalidateCornerRadiusControl"),
 ("invalidateCornerRadiusChange", "main.qrCodeScreenSelectionSession.invalidateCornerRadiusChange"),
 ("invalidateCornerRadiusSelectionChange", "main.qrCodeScreenSelectionSession.invalidateCornerRadiusSelectionChange"),
 ("drawSelectionResizeHandles", "main.qrCodeScreenSelectionSession.drawSelectionResizeHandles"),
 ("drawControlHover", "main.qrCodeScreenSelectionSession.drawControlHover"),
 ("drawWindowHover", "main.qrCodeScreenSelectionSession.drawWindowHover"),
 ("drawSelectionSizeLabel", "main.qrCodeScreenSelectionSession.drawSelectionSizeLabel"),
 ("resolveControlHoverForWindow", "main.qrCodeScreenSelectionSession.resolveControlHoverForWindow"),
 ("resolveWindowHoverForWindow", "main.qrCodeScreenSelectionSession.resolveWindowHoverForWindow"),
 ("shouldReuseControlHoverForWindow", "main.qrCodeScreenSelectionSession.shouldReuseControlHoverForWindow"),
 ("cachedControlHoverAllowedForPoint", "main.qrCodeScreenSelectionSession.cachedControlHoverAllowedForPoint"),
 ("resolveControlSelectionAtPoint", "main.qrCodeScreenSelectionSession.resolveControlSelectionAtPoint"),
 ("resolveControlSelectionForWindowAtPoint", "main.qrCodeScreenSelectionSession.resolveControlSelectionForWindowAtPoint"),
 ("resolveControlClickSelectionAtPoint", "main.qrCodeScreenSelectionSession.resolveControlClickSelectionAtPoint"),
 ("resolveWindowAtPoint", "main.qrCodeScreenSelectionSession.resolveWindowAtPoint"),
 ("invalidateControlHoverChange", "main.qrCodeScreenSelectionSession.invalidateControlHoverChange"),
 ("invalidateWindowHoverChange", "main.qrCodeScreenSelectionSession.invalidateWindowHoverChange"),
 ("invalidateSelectionChange", "main.qrCodeScreenSelectionSession.invalidateSelectionChange"),
 ("invalidateSelectionSizeLabelChange", "main.qrCodeScreenSelectionSession.invalidateSelectionSizeLabelChange"),
 ("prepareSelectionResult", "main.qrCodeScreenSelectionSession.prepareSelectionResult"),
 ("resolveClickLikeControlSelection", "main.qrCodeScreenSelectionSession.resolveClickLikeControlSelection"),
 ("prepareAreaSelectionResult", "main.qrCodeScreenSelectionSession.prepareAreaSelectionResult"),
 ("finalizePreparedRequest", "main.qrCodeScreenSelectionSession.finalizePreparedRequest"),
 ("beginAnnotationEditing", "main.qrCodeScreenSelectionSession.beginAnnotationEditing"),
 ("confirmAnnotationEditing", "main.qrCodeScreenSelectionSession.confirmAnnotationEditing"),
 ("handleFinalizeRequest", "main.qrCodeScreenSelectionSession.handleFinalizeRequest"),
 ("showAnnotationToolbar", "main.qrCodeScreenSelectionSession.showAnnotationToolbar"),
 ("syncAnnotationToolbar", "main.qrCodeScreenSelectionSession.syncAnnotationToolbar"),
 ("hideAnnotationToolbar", "main.qrCodeScreenSelectionSession.hideAnnotationToolbar"),
 ("annotationToolbarBounds", "main.qrCodeScreenSelectionSession.annotationToolbarBounds"),
 ("annotationToolbarState", "main.qrCodeScreenSelectionSession.annotationToolbarState"),
 ("handleAnnotationToolbarActions", "main.qrCodeScreenSelectionSession.handleAnnotationToolbarActions"),
 ("applyAnnotationToolbarAction", "main.qrCodeScreenSelectionSession.applyAnnotationToolbarAction"),
 ("handleAnnotationPointerDown", "main.qrCodeScreenSelectionSession.handleAnnotationPointerDown"),
 ("handleAnnotationPointerMove", "main.qrCodeScreenSelectionSession.handleAnnotationPointerMove"),
 ("handleAnnotationPointerUp", "main.qrCodeScreenSelectionSession.handleAnnotationPointerUp"),
 ("handleAnnotationKeyDown", "main.qrCodeScreenSelectionSession.handleAnnotationKeyDown"),
 ("handleAnnotationChar", "main.qrCodeScreenSelectionSession.handleAnnotationChar"),
 ("annotationTextSelectionRange", "main.qrCodeScreenSelectionSession.annotationTextSelectionRange"),
 ("replaceAnnotationTextSelection", "main.qrCodeScreenSelectionSession.replaceAnnotationTextSelection"),
 ("deleteAnnotationTextBackward", "main.qrCodeScreenSelectionSession.deleteAnnotationTextBackward"),
 ("deleteAnnotationTextForward", "main.qrCodeScreenSelectionSession.deleteAnnotationTextForward"),
 ("moveAnnotationTextCaretLeft", "main.qrCodeScreenSelectionSession.moveAnnotationTextCaretLeft"),
 ("moveAnnotationTextCaretRight", "main.qrCodeScreenSelectionSession.moveAnnotationTextCaretRight"),
 ("selectedAnnotationText", "main.qrCodeScreenSelectionSession.selectedAnnotationText"),
 ("copySelectedAnnotationText", "main.qrCodeScreenSelectionSession.copySelectedAnnotationText"),
 ("cutSelectedAnnotationText", "main.qrCodeScreenSelectionSession.cutSelectedAnnotationText"),
 ("pasteAnnotationText", "main.qrCodeScreenSelectionSession.pasteAnnotationText"),
 ("annotationTextIndexAtPoint", "main.qrCodeScreenSelectionSession.annotationTextIndexAtPoint"),
 ("updateAnnotationTextFontSizeFromPoint", "main.qrCodeScreenSelectionSession.updateAnnotationTextFontSizeFromPoint"),
 ("cycleAnnotationTextFontFamily", "main.qrCodeScreenSelectionSession.cycleAnnotationTextFontFamily"),
 ("newAnnotationStroke", "main.qrCodeScreenSelectionSession.newAnnotationStroke"),
 ("undoAnnotationAndInvalidate", "main.qrCodeScreenSelectionSession.undoAnnotationAndInvalidate"),
 ("finalizeAnnotationText", "main.qrCodeScreenSelectionSession.finalizeAnnotationText"),
 ("drawAnnotationOverlay", "main.qrCodeScreenSelectionSession.drawAnnotationOverlay"),
 ("applyRoundedSelectionPreview", "main.qrCodeScreenSelectionSession.applyRoundedSelectionPreview"),
 ("roundedCornerMask", "main.qrCodeScreenSelectionSession.roundedCornerMask"),
 ("roundedCornerInnerMask", "main.qrCodeScreenSelectionSession.roundedCornerInnerMask"),
 ("drawAnnotationStrokesOnPaintBuffer", "main.qrCodeScreenSelectionSession.drawAnnotationStrokesOnPaintBuffer"),
 ("hasAnnotationStrokeInPaintRect", "main.qrCodeScreenSelectionSession.hasAnnotationStrokeInPaintRect"),
 ("drawAnnotationTextEditor", "main.qrCodeScreenSelectionSession.drawAnnotationTextEditor"),
 ("invalidateAnnotationTextAreaWith", "main.qrCodeScreenSelectionSession.invalidateAnnotationTextAreaWith"),
 ("invalidateAnnotationDirtyRect", "main.qrCodeScreenSelectionSession.invalidateAnnotationDirtyRect"),
 ("annotationTextInvalidationRect", "main.qrCodeScreenSelectionSession.annotationTextInvalidationRect"),
 ("annotationContentInvalidationRect", "main.qrCodeScreenSelectionSession.annotationContentInvalidationRect"),
 ("annotationTransientInvalidationRect", "main.qrCodeScreenSelectionSession.annotationTransientInvalidationRect"),
 ("invalidateAnnotationStrokeChange", "main.qrCodeScreenSelectionSession.invalidateAnnotationStrokeChange"),
 ("annotationTextEditorRect", "main.qrCodeScreenSelectionSession.annotationTextEditorRect"),
 ("annotationTextEditorBounds", "main.qrCodeScreenSelectionSession.annotationTextEditorBounds"),
 ("annotationTextEditorPosition", "main.qrCodeScreenSelectionSession.annotationTextEditorPosition"),
 ("annotationTextDragHandleRect", "main.qrCodeScreenSelectionSession.annotationTextDragHandleRect"),
 ("annotationTextConfirmRect", "main.qrCodeScreenSelectionSession.annotationTextConfirmRect"),
 ("annotationTextInputRect", "main.qrCodeScreenSelectionSession.annotationTextInputRect"),
 ("annotationTextSwatchRect", "main.qrCodeScreenSelectionSession.annotationTextSwatchRect"),
 ("annotationTextSliderRect", "main.qrCodeScreenSelectionSession.annotationTextSliderRect"),
 ("annotationTextSliderHitRect", "main.qrCodeScreenSelectionSession.annotationTextSliderHitRect"),
 ("annotationTextFontRect", "main.qrCodeScreenSelectionSession.annotationTextFontRect"),
 ("annotationStrokeAt", "main.qrCodeScreenSelectionSession.annotationStrokeAt"),
 ("encodeAnnotatedSelectionPNG", "main.qrCodeScreenSelectionSession.encodeAnnotatedSelectionPNG"),
 ("prepareWindowResult", "main.qrCodeScreenSelectionSession.prepareWindowResult"),
 ("Emit", "main.qrCodeCaptureAcceptedNotifier.Emit"),
 ("OpenQRCodeExternalURL", "main.BootstrapService.OpenQRCodeExternalURL"),
 ("normalizeQRCodeExternalURL", "main.normalizeQRCodeExternalURL"),
 ("containsUnsafeQRCodeURLText", "main.containsUnsafeQRCodeURLText"),
 ("qrExternalURLInvalidError", "main.qrExternalURLInvalidError"),
]


def va_to_off(data, va):
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


def next_va(va):
    for v, _ in SORTED:
        if v > va:
            return v
    return va + 0x400


data = open(EXE, "rb").read()
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_64)

for short, sym in TARGETS:
    va = None
    for v, n in SORTED:
        if n == sym:
            va = v
            break
    if va is None:
        print("MISSING SYMBOL: %s" % sym)
        continue
    nxt = next_va(va)
    length = min(nxt - va, 0x800)
    off = va_to_off(data, va)
    if off is None:
        print("NO SECTION: %s" % sym)
        continue
    code = data[off:off + length]
    lines = []
    for insn in md.disasm(code, va):
        lines.append("0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str))
    with open(os.path.join(OUT, short + ".asm.txt"), "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))
    print("=== %s  %s  +0x%x (%d insns) ===" % (short, sym, length, len(lines)))
    for l in lines[:18]:
        print("   " + l)
    print("   ... (tail)")
    for l in lines[-8:]:
        print("   " + l)
