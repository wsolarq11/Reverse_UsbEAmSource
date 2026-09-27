package main

import (
	"image"
	"strings"
	"testing"
	"unsafe"
)

func TestIIDIDXGIOutput6(t *testing.T) {
	if iidIDXGIOutput6.Data1 != 0x068346e8 || iidIDXGIOutput6.Data2 != 0xaaec || iidIDXGIOutput6.Data3 != 0x4b84 {
		t.Errorf("iidIDXGIOutput6 prefix mismatch: %x-%x-%x", iidIDXGIOutput6.Data1, iidIDXGIOutput6.Data2, iidIDXGIOutput6.Data3)
	}
	want := [8]byte{0xad, 0xd7, 0x13, 0x7f, 0x51, 0x3f, 0x77, 0xa1}
	if iidIDXGIOutput6.Data4 != want {
		t.Errorf("iidIDXGIOutput6.Data4 = %x, want %x", iidIDXGIOutput6.Data4, want)
	}
}

func TestDXGIAdapterDescLayout(t *testing.T) {
	if got := unsafe.Sizeof(dxgiAdapterDesc{}); got != 0x130 {
		t.Errorf("dxgiAdapterDesc size = %d, want %d", got, 0x130)
	}
	if got := unsafe.Offsetof(dxgiAdapterDesc{}.Description); got != 0 {
		t.Errorf("Description offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(dxgiAdapterDesc{}.VendorID); got != 0x100 {
		t.Errorf("VendorID offset = %d, want 0x100", got)
	}
	if got := unsafe.Offsetof(dxgiAdapterDesc{}.DedicatedVideoMemory); got != 0x110 {
		t.Errorf("DedicatedVideoMemory offset = %d, want 0x110", got)
	}
	if got := unsafe.Offsetof(dxgiAdapterDesc{}.AdapterLuidLow); got != 0x128 {
		t.Errorf("AdapterLuidLow offset = %d, want 0x128", got)
	}
}

func TestDXGIOutputDesc1Layout(t *testing.T) {
	if got := unsafe.Sizeof(dxgiOutputDesc1{}); got != 0x98 {
		t.Errorf("dxgiOutputDesc1 size = %d, want %d", got, 0x98)
	}
	checks := []struct {
		field string
		got   uintptr
		want  uintptr
	}{
		{"DeviceName", unsafe.Offsetof(dxgiOutputDesc1{}.DeviceName), 0x00},
		{"DesktopLeft", unsafe.Offsetof(dxgiOutputDesc1{}.DesktopLeft), 0x40},
		{"AttachedToDesktop", unsafe.Offsetof(dxgiOutputDesc1{}.AttachedToDesktop), 0x50},
		{"Rotation", unsafe.Offsetof(dxgiOutputDesc1{}.Rotation), 0x54},
		{"Monitor", unsafe.Offsetof(dxgiOutputDesc1{}.Monitor), 0x58},
		{"BitsPerColor", unsafe.Offsetof(dxgiOutputDesc1{}.BitsPerColor), 0x60},
		{"ColorSpace", unsafe.Offsetof(dxgiOutputDesc1{}.ColorSpace), 0x64},
		{"MinLuminance", unsafe.Offsetof(dxgiOutputDesc1{}.MinLuminance), 0x88},
		{"MaxLuminance", unsafe.Offsetof(dxgiOutputDesc1{}.MaxLuminance), 0x8c},
		{"MaxFullFrameLuminance", unsafe.Offsetof(dxgiOutputDesc1{}.MaxFullFrameLuminance), 0x90},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s offset = 0x%x, want 0x%x", c.field, c.got, c.want)
		}
	}
}

func TestScreenshotDisplayCaptureInfoFromDesc(t *testing.T) {
	desc := dxgiOutputDesc1{
		DeviceName:            [32]uint16{'D', 'I', 'S', 'P', '1', ' ', 0},
		DesktopLeft:           1920,
		DesktopTop:            0,
		DesktopRight:          0,
		DesktopBottom:         1080,
		AttachedToDesktop:     1,
		Rotation:              2,
		BitsPerColor:          10,
		ColorSpace:            12,
		MinLuminance:          0.5,
		MaxLuminance:          270,
		MaxFullFrameLuminance: 270,
	}
	info := screenshotDisplayCaptureInfoFromDesc(desc, 3, 5, "  Intel(R) UHD  ")

	if info.AdapterIndex != 3 || info.OutputIndex != 5 {
		t.Errorf("indices = (%d,%d), want (3,5)", info.AdapterIndex, info.OutputIndex)
	}
	if info.AdapterName != "Intel(R) UHD" {
		t.Errorf("AdapterName = %q, want %q", info.AdapterName, "Intel(R) UHD")
	}
	if info.DeviceName != "DISP1" {
		t.Errorf("DeviceName = %q, want %q", info.DeviceName, "DISP1")
	}
	if info.Bounds != (image.Rect(0, 0, 1920, 1080)) {
		t.Errorf("Bounds = %v, want (0,0)-(1920,1080)", info.Bounds)
	}
	if info.Rotation != 2 || !info.AttachedToDesktop {
		t.Errorf("Rotation=%d Attached=%v, want 2/true", info.Rotation, info.AttachedToDesktop)
	}
	if info.BitsPerColor != 10 || info.ColorSpace != 12 {
		t.Errorf("BitsPerColor=%d ColorSpace=%d, want 10/12", info.BitsPerColor, info.ColorSpace)
	}
	if !info.HDR || !info.AdvancedColor {
		t.Errorf("HDR=%v Advanced=%v, want true/true", info.HDR, info.AdvancedColor)
	}
	if info.MinLuminance != 0.5 || info.MaxLuminance != 270 || info.MaxFullFrameLuminance != 270 {
		t.Errorf("luminance = (%v,%v,%v), want (0.5,270,270)", info.MinLuminance, info.MaxLuminance, info.MaxFullFrameLuminance)
	}
}

func TestScreenshotDisplayCaptureInfoFromDescAdvancedColor(t *testing.T) {
	desc := dxgiOutputDesc1{ColorSpace: 0, BitsPerColor: 10}
	if info := screenshotDisplayCaptureInfoFromDesc(desc, 0, 0, ""); !info.AdvancedColor {
		t.Errorf("10bpc SDR advanced color = false, want true")
	}
	desc.BitsPerColor = 8
	if info := screenshotDisplayCaptureInfoFromDesc(desc, 0, 0, ""); info.AdvancedColor {
		t.Errorf("8bpc SDR advanced color = true, want false")
	}
}

func TestScreenshotDXGIAdapterNameNil(t *testing.T) {
	name, err := screenshotDXGIAdapterName(0)
	if err != nil {
		t.Fatalf("nil adapter error: %v", err)
	}
	if name != "" {
		t.Errorf("nil adapter name = %q, want empty", name)
	}
}

func TestQueryScreenshotDXGIOutput6Nil(t *testing.T) {
	_, err := queryScreenshotDXGIOutput6(0)
	if err == nil {
		t.Fatal("nil output error = nil, want error")
	}
	if !strings.Contains(err.Error(), "DXGI 输出为空") {
		t.Errorf("nil output error = %q, want DXGI 输出为空", err.Error())
	}
}
