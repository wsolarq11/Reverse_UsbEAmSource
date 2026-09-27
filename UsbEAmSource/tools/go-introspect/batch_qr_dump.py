#!/usr/bin/env python
# 研究用途：一次性 dump 59 个 qrcode 函数 asm 到指定目录。
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import va_dump

FUNCS = [
    ("qrCodeAnnotationLinePaintBounds", 0x1409529e0, 0x2e0),
    ("fillQRCodeRect", 0x140952ac0, 0x1e0),
    ("drawQRCodeRoundedRect", 0x140952d20, 0x360),
    ("drawQRCodeSliderThumb", 0x140953220, 0x380),
    ("drawQRCodeAnnotationFinishOverlay", 0x140953740, 0x460),
    ("drawQRCodeRectFrame", 0x140953d20, 0x240),
    ("centerRect", 0x140953fe0, 0x60),
    ("drawQRCodeAnnotationOutlinedText", 0x140954040, 0x440),
    ("qrCodeAnnotationNumberLabelSize", 0x140954480, 0x240),
    ("qrCodeAnnotationNumberOutlineWidth", 0x1409546c0, 0xe0),
    ("qrCodeAnnotationTextOutlineOffsets", 0x1409547a0, 0x140),
    ("drawQRCodeAnnotationStrokeOnImageWithOffset", 0x140954c20, 0x8e0),
    ("drawQRCodeAnnotationMosaicOnImageWithOffset", 0x140955500, 0x2c0),
    ("drawQRCodeAnnotationBlurOnImageWithOffset", 0x1409557c0, 0x2a0),
    ("drawArrowHeadOnRGBA", 0x140955a60, 0x140),
    ("qrCodeAnnotationArrowHeadPoints", 0x140955ba0, 0x360),
    ("drawEllipseOnRGBA", 0x140955f00, 0x460),
    ("drawNumberOnRGBA", 0x140956360, 0x4c0),
    ("drawTextBlockOnRGBAWithOutline", 0x140956820, 0x360),
    ("drawTextMaskOnRGBARegion", 0x140956ce0, 0x4a0),
    ("copyQRCodeTextMaskDIBBitsToRGBARegion", 0x140957b40, 0x3c0),
    ("drawSmoothCapsuleOnRGBA", 0x140957f00, 0xcc0),
    ("qrCodeCapsulePixelCoverage", 0x140958da0, 0x160),
    ("qrCodeEllipseStrokeDistance", 0x140958f00, 0x180),
    ("qrCodeEllipseStrokePixelCoverage", 0x140959080, 0x1e0),
    ("drawSmoothCircleOnRGBA", 0x140959260, 0x480),
    ("qrCodeCirclePixelCoverage", 0x1409596e0, 0xc0),
    ("blendQRCodeRGBAPixel", 0x1409597a0, 0x220),
    ("fillRGBA", 0x1409599c0, 0x1a0),
    ("qrCodeSelectionWindowRect", 0x140959ee0, 0xa0),
    ("captureQRCodeVirtualScreenSnapshotWithCursor", 0x140959f80, 0xe0),
    ("encodeQRCodeSelectionPNG", 0x14095a360, 0x240),
    ("qrCodeSelectionResizeHandlePositions", 0x14095a5a0, 0x100),
    ("qrCodeSelectionResizeHandleAtPoint", 0x14095a6a0, 0x1a0),
    ("resizeQRCodeSelectionRect", 0x14095a840, 0x1e0),
    ("qrCodeSelectionCursorPoint", 0x14095aa20, 0x80),
    ("buildQRCodeSelectionDirtyRects", 0x14095aaa0, 0x4a0),
    ("buildQRCodeSelectionDirtyRectsWithSizeLabels", 0x14095af40, 0x360),
    ("qrCodeSelectionSizeLabelText", 0x14095b2a0, 0xe0),
    ("qrCodeSelectionSizeLabelRect", 0x14095b380, 0x120),
    ("qrCodeSelectionCornerRadiusLayout", 0x14095b4a0, 0x5a0),
    ("qrCodeSelectionInfoPlacementRect", 0x14095ba40, 0x260),
    ("qrCodeCornerRadiusThumbX", 0x14095bca0, 0xe0),
    ("qrCodeCornerRadiusFromPoint", 0x14095bd80, 0x100),
    ("qrCodeRoundedCornerRects", 0x14095be80, 0x1c0),
    ("buildQRCodeRoundedCornerMask", 0x14095c040, 0x380),
    ("qrCodeRoundedCornerMaskCoverage", 0x14095c3c0, 0x160),
    ("drawQRCodeRoundedSelectionBorderOnImage", 0x14095c520, 0x800),
    ("applyQRCodeRoundedCorners", 0x14095cd20, 0x360),
    ("blendQRCodeRGBAByCoverage", 0x14095d080, 0x100),
    ("clampQRCodeSelectionSizeLabelRect", 0x14095d180, 0x140),
    ("subtractQRCodeSelectionRect", 0x14095d2c0, 0x280),
    ("buildQRCodeSelectionBorderRects", 0x14095d540, 0x1c0),
    ("filterQRCodeSelectionRectangles", 0x14095d700, 0x360),
    ("drawQRCodeSelectionBorder", 0x14095da60, 0x160),
    ("drawQRCodeRoundedSelectionBorder", 0x14095dbc0, 0x180),
    ("drawQRCodeDashedSelectionBorder", 0x14095e0e0, 0x580),
    ("createQRCodePaintBuffer", 0x14095e840, 0x260),
    ("flushQRCodePaintBuffer", 0x14095eaa0, 0x140),
]


def main():
    exe = sys.argv[1]
    outdir = sys.argv[2]
    os.makedirs(outdir, exist_ok=True)
    data = open(exe, "rb").read()
    for name, base, length in FUNCS:
        off = va_dump.va_to_off(data, base)
        if off is None:
            print("SKIP %s 0x%x (no section)" % (name, base))
            continue
        code = data[off:off + length]
        prefix = os.path.join(outdir, name)
        with open(prefix + ".bin", "wb") as fh:
            fh.write(code)
        md = va_dump.capstone.Cs(va_dump.capstone.CS_ARCH_X86, va_dump.capstone.CS_MODE_64)
        lines = []
        for insn in md.disasm(code, base):
            line = "0x%x: %-8s %s" % (insn.address, insn.mnemonic, insn.op_str)
            if insn.mnemonic.startswith("call") or insn.mnemonic.startswith("j"):
                m = va_dump.re.search(r"0x([0-9a-f]+)", insn.op_str)
                if m:
                    nm = va_dump.annotate(int(m.group(1), 16))
                    if nm:
                        line += "   ; %s" % nm
            lines.append(line)
        with open(prefix + ".asm.txt", "w", encoding="utf-8") as fh:
            fh.write("\n".join(lines))
        print("%-44s 0x%x +%d (%d lines)" % (name, base, length, len(lines)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
