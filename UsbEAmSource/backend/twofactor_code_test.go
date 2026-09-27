// AUTO-RECONSTRUCTED TESTS — DOMAIN: twofactor code generation (TOTP/HOTP/Steam)
// 研究用途
package main

import (
	"strings"
	"testing"
	"time"
)

// RFC 4226 附录 D 测试向量（secret="12345678901234567890"，SHA1）。
func TestBuildHOTPValueRFC4226(t *testing.T) {
	secret := []byte("12345678901234567890")
	vectors := []struct {
		counter uint64
		decimal uint32 // 31-bit 截断值
		digits6 uint32 // 6 位 HOTP
	}{
		{0, 1284755224, 755224},
		{1, 1094287082, 287082},
		{2, 137359152, 359152},
		{3, 1726969429, 969429},
		{4, 1640338314, 338314},
		{5, 868254676, 254676},
		{6, 1918287922, 287922},
		{7, 82162583, 162583},
		{8, 673399871, 399871},
		{9, 645520489, 520489},
	}
	for _, v := range vectors {
		got, err := buildHOTPValue(secret, v.counter, "SHA1")
		if err != nil {
			t.Fatalf("counter=%d: error %v", v.counter, err)
		}
		if got != v.decimal {
			t.Errorf("counter=%d: got %d, want %d", v.counter, got, v.decimal)
		}
		if got%1000000 != v.digits6 {
			t.Errorf("counter=%d: 6-digit = %d, want %d", v.counter, got%1000000, v.digits6)
		}
	}
}

func TestBuildStandardTwoFactorCode(t *testing.T) {
	secret := []byte("12345678901234567890")

	// counter=1 的 HOTP=287082，6 位补零。
	code, err := buildStandardTwoFactorCode(secret, 1, 6, "SHA1")
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" {
		t.Fatalf("6-digit code = %q, want %q", code, "287082")
	}

	// 8 位补零（RFC 6238 T=1 官方值：hotp=1094287082，mod 10^8 = 94287082）。
	code, err = buildStandardTwoFactorCode(secret, 1, 8, "SHA1")
	if err != nil {
		t.Fatal(err)
	}
	if code != "94287082" {
		t.Fatalf("8-digit code = %q, want %q", code, "94287082")
	}
}

func TestBuildSteamTwoFactorCode(t *testing.T) {
	secret := []byte("12345678901234567890")

	// counter=0 的 HOTP=1284755224 → Steam 编码 "GG5F5"。
	code, err := buildSteamTwoFactorCode(secret, 0)
	if err != nil {
		t.Fatal(err)
	}
	if code != "GG5F5" {
		t.Fatalf("steam code = %q, want %q", code, "GG5F5")
	}

	// 格式：5 位，且全部落在 Steam 字母表。
	if len(code) != 5 {
		t.Fatalf("steam code length = %d, want 5", len(code))
	}
	const alpha = "23456789BCDFGHJKMNPQRTVWXY"
	for _, r := range code {
		if !strings.ContainsRune(alpha, r) {
			t.Fatalf("steam code contains %q outside alphabet", r)
		}
	}
}

func TestBuildTwoFactorCodeTotp(t *testing.T) {
	secret := []byte("12345678901234567890")
	entry := TwoFactorEntryConfig{
		Kind:      "totp",
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
	}
	now := time.Unix(59, 0) // T = 59/30 = 1, remaining = 30-29 = 1

	code, remaining, err := buildTwoFactorCode(entry, secret, now)
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" {
		t.Fatalf("totp code = %q, want %q", code, "287082")
	}
	if remaining != 1 {
		t.Fatalf("remaining = %d, want 1", remaining)
	}
}

func TestBuildTwoFactorCodeSteam(t *testing.T) {
	secret := []byte("12345678901234567890")
	entry := TwoFactorEntryConfig{
		Kind:   "SteamGuard",
		Period: 60, // steam 强制 30
	}
	now := time.Unix(0, 0) // T = 0

	code, remaining, err := buildTwoFactorCode(entry, secret, now)
	if err != nil {
		t.Fatal(err)
	}
	if code != "GG5F5" {
		t.Fatalf("steam code = %q, want %q", code, "GG5F5")
	}
	// steam period 强制 30，remaining = 30 - 0 = 30。
	if remaining != 30 {
		t.Fatalf("steam remaining = %d, want 30", remaining)
	}
}

func TestBuildTwoFactorCodeClamp(t *testing.T) {
	secret := []byte("12345678901234567890")

	// period 越界（>300）→ 30；digits 越界（>10）→ 6。
	entry := TwoFactorEntryConfig{
		Kind:      "totp",
		Algorithm: "SHA1",
		Digits:    12,
		Period:    999,
	}
	now := time.Unix(59, 0) // T=1 需 period=30

	code, remaining, err := buildTwoFactorCode(entry, secret, now)
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" { // digits 钳回 6，period 钳回 30 → counter=1
		t.Fatalf("clamped code = %q, want %q", code, "287082")
	}
	if remaining != 1 {
		t.Fatalf("clamped remaining = %d, want 1", remaining)
	}
}
