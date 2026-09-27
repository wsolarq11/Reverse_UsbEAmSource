import subprocess, sys

FUNCS = [
    ("captureQRCodesFromNativeSelection", 0x1409341e0, 0x320),
    ("captureQRCodesFromNativeSelectionResultWithOptions", 0x140934500, 0x580),
    ("captureQRCodeSelectionPreviewSnapshot", 0x140934ac0, 0x2a0),
    ("captureQRCodeFinalSelectionPNG", 0x140934d60, 0x160),
    ("ensureQRCodeSelectionWindowClass", 0x140934ec0, 0x60),
    ("newQRCodeScreenSelectionSession", 0x140934f20, 0x4e0),
    ("createQRCodeBitmapHandle", 0x140935400, 0x120),
    ("createQRCodeBitmapDC", 0x140935520, 0x1a0),
    ("applyQRCodeSelectionCaptureExclusion", 0x140936ac0, 0xa0),
    ("tryReserveQRCodeSelectionSession", 0x1409375a0, 0xe0),
    ("activateReservedQRCodeSelectionSession", 0x1409376e0, 0x120),
    ("finishQRCodeSelectionSession", 0x140937860, 0x120),
    ("activeQRCodeSelectionSession", 0x1409379e0, 0xc0),
    ("qrCodeSelectionWindowProc", 0x140937b00, 0xe0),
    ("shouldRefreshQRCodeControlHover", 0x14093e000, 0x120),
    ("shouldReuseQRCodeControlHover", 0x14093e120, 0x40),
    ("qrCodeControlSelectionFromCachedHover", 0x14093e860, 0xc0),
    ("hideScreenshotOverlayWindow", 0x140940760, 0x80),
    ("normalizeQRCodeAnnotationClipboardText", 0x140946ce0, 0x1c0),
    ("screenshotCurrentControlSelectionPreference", 0x140947040, 0xc0),
    ("normalizedQRCodeAnnotationFontFamily", 0x140947100, 0x100),
    ("nextQRCodeAnnotationFontFamily", 0x140947200, 0x140),
    ("offsetQRCodeAnnotationStroke", 0x140947ec0, 0x2e0),
    ("translateQRCodeAnnotationStroke", 0x1409481a0, 0x4a0),
    ("clampQRCodeAnnotationTextEditorPoint", 0x140948640, 0x100),
    ("parseQRCodeAnnotationHexColor", 0x140948740, 0x140),
    ("unionQRCodeAnnotationRects", 0x14094c700, 0xa0),
    ("qrCodeAnnotationTextEditorRectForPoint", 0x14094cbc0, 0x80),
    ("qrCodeAnnotationTextEditorRawRectForPointWithLimit", 0x14094cc40, 0xe0),
    ("qrCodeAnnotationTextEditorRectAvoidingPreview", 0x14094cd20, 0x400),
    ("clampQRCodeAnnotationRect", 0x14094d120, 0xc0),
    ("drawQRCodeAnnotationSmallText", 0x14094d4e0, 0x200),
    ("withQRCodeAnnotationTextMeasureHDC", 0x14094d800, 0x220),
    ("qrCodeAnnotationTextVisibleRange", 0x14094dbc0, 0x500),
    ("qrCodeAnnotationTextIndexForX", 0x14094e0c0, 0x260),
    ("measureQRCodeAnnotationTextWidth", 0x14094e320, 0x140),
    ("drawQRCodeAnnotationStroke", 0x14094e460, 0x500),
    ("createQRCodeAnnotationPen", 0x14094ea80, 0x120),
    ("drawQRCodeAnnotationLine", 0x14094eba0, 0x120),
    ("drawQRCodeAnnotationRect", 0x14094ecc0, 0x200),
    ("drawQRCodeAnnotationEllipse", 0x14094eec0, 0x560),
    ("drawQRCodeAnnotationArrow", 0x14094f420, 0x100),
    ("drawQRCodeAnnotationPenPath", 0x14094f520, 0x180),
    ("drawQRCodeAnnotationTextWithColor", 0x14094f6a0, 0x4c0),
    ("qrCodeAnnotationTextW32Rect", 0x14094fe00, 0x80),
    ("qrCodeAnnotationTextBoundsForHDC", 0x14094fe80, 0x180),
    ("qrCodeAnnotationTextEstimatedBounds", 0x140950000, 0xe0),
    ("qrCodeAnnotationTextMeasuredWidth", 0x1409500e0, 0x3e0),
    ("measureQRCodeAnnotationTextSizeForHDC", 0x140950660, 0x140),
    ("createQRCodeAnnotationFontWithQuality", 0x1409507a0, 0x160),
    ("createQRCodeAnnotationNumberFont", 0x140950900, 0x120),
    ("drawQRCodeAnnotationNumber", 0x140950a20, 0x900),
    ("drawQRCodeAnnotationMosaic", 0x140951800, 0x320),
    ("drawQRCodeAnnotationBlur", 0x140951b20, 0x2e0),
    ("averageQRCodeAnnotationColor", 0x140951e00, 0x140),
    ("drawQRCodeAnnotationSelection", 0x140951f40, 0xe0),
    ("qrCodeAnnotationStrokeBounds", 0x140952300, 0x4e0),
    ("qrCodeAnnotationPointsBounds", 0x1409527e0, 0xa0),
    ("qrCodeAnnotationStrokePaintBounds", 0x140952880, 0x160),
]

EXE = "D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
DUMP = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource/tools/go-introspect/va_dump.py"

for name, va, ln in FUNCS:
    out = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource/docs/goresym/pipeline/tmp/qr_%s" % name
    r = subprocess.run([sys.executable, DUMP, EXE, hex(va), hex(ln), out],
                       capture_output=True, text=True, encoding="utf-8", errors="replace")
    print((r.stdout or "").strip())
    if r.returncode != 0:
        print("  STDERR:", (r.stderr or "").strip())
