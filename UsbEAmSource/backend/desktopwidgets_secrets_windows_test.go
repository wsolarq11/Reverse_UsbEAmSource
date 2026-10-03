package main

import (
	"strings"
	"testing"
)

func TestProtectDesktopWidgetSecretEmpty(t *testing.T) {
	got, err := protectDesktopWidgetSecret("")
	if err == nil {
		t.Fatal("want error for empty secret")
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
	if !strings.Contains(err.Error(), "凭据不能为空") {
		t.Errorf("err = %q, want 凭据不能为空", err.Error())
	}
}

func TestUnprotectDesktopWidgetSecretInvalid(t *testing.T) {
	for _, in := range []string{"", "not-base64!!!", "@@@@", "===="} {
		got, err := unprotectDesktopWidgetSecret(in)
		if err == nil {
			t.Fatalf("input %q: want error", in)
		}
		if got != "" {
			t.Errorf("input %q: got %q, want empty", in, got)
		}
		if !strings.Contains(err.Error(), "受保护凭据格式无效") {
			t.Errorf("input %q: err = %q, want 受保护凭据格式无效", in, err.Error())
		}
	}
}

func TestDesktopWidgetSecretRoundTrip(t *testing.T) {
	const want = "api-key-示例-凭据-xyz123"
	enc, err := protectDesktopWidgetSecret(want)
	if err != nil {
		t.Fatalf("protect: %v", err)
	}
	if enc == "" {
		t.Fatal("protect returned empty ciphertext")
	}
	got, err := unprotectDesktopWidgetSecret(enc)
	if err != nil {
		t.Fatalf("unprotect: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip = %q, want %q", got, want)
	}
}
