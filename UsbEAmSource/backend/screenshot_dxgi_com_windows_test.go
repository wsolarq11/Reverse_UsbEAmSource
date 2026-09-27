package main

import "testing"

func TestDXGIFactoryIIDs(t *testing.T) {
	if iidIDXGIFactory1.Data1 != 0x770aae78 || iidIDXGIFactory1.Data2 != 0xf26f || iidIDXGIFactory1.Data3 != 0x4dba {
		t.Errorf("iidIDXGIFactory1 prefix mismatch: %x-%x-%x", iidIDXGIFactory1.Data1, iidIDXGIFactory1.Data2, iidIDXGIFactory1.Data3)
	}
	want1 := [8]byte{0xa8, 0x29, 0x25, 0x3c, 0x83, 0xd1, 0xb3, 0x87}
	if iidIDXGIFactory1.Data4 != want1 {
		t.Errorf("iidIDXGIFactory1.Data4 = %x, want %x", iidIDXGIFactory1.Data4, want1)
	}

	if iidIDXGIFactory6.Data1 != 0xc1b6694f || iidIDXGIFactory6.Data2 != 0xff09 || iidIDXGIFactory6.Data3 != 0x44a9 {
		t.Errorf("iidIDXGIFactory6 prefix mismatch: %x-%x-%x", iidIDXGIFactory6.Data1, iidIDXGIFactory6.Data2, iidIDXGIFactory6.Data3)
	}
	want6 := [8]byte{0xb0, 0x3c, 0x77, 0x90, 0x0a, 0x0a, 0x1d, 0x17}
	if iidIDXGIFactory6.Data4 != want6 {
		t.Errorf("iidIDXGIFactory6.Data4 = %x, want %x", iidIDXGIFactory6.Data4, want6)
	}
}

func TestReleaseDXGIUnknownNilSafe(t *testing.T) {
	releaseDXGIUnknown(0)
}
