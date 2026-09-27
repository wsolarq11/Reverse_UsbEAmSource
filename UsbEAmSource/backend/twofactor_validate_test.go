// AUTO-RECONSTRUCTED TESTS — DOMAIN: twofactor stored-config validation
// 研究用途
package main

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func b64(n int) string {
	return base64.StdEncoding.EncodeToString(make([]byte, n))
}

func TestValidateTwoFactorStoredConfigEmpty(t *testing.T) {
	if err := validateTwoFactorStoredConfig(TwoFactorPasswordConfig{}, nil); err != nil {
		t.Fatalf("空配置应通过校验, got %v", err)
	}
}

func TestValidateTwoFactorStoredConfigValid(t *testing.T) {
	pw := TwoFactorPasswordConfig{
		KDF:      "argon2id-v1",
		Salt:     b64(16),
		Verifier: b64(32),
	}
	if err := validateTwoFactorStoredConfig(pw, nil); err != nil {
		t.Fatalf("合法配置应通过校验, got %v", err)
	}
}

func TestValidateTwoFactorStoredConfigInvalidKDF(t *testing.T) {
	pw := TwoFactorPasswordConfig{
		KDF:      "sha256-v1",
		Salt:     b64(16),
		Verifier: b64(32),
	}
	err := validateTwoFactorStoredConfig(pw, nil)
	if err == nil || !strings.Contains(err.Error(), "password.kdf") {
		t.Fatalf("无效 KDF 应报 kdf 错误, got %v", err)
	}
	if !errors.Is(err, errTwoFactorDataCorrupted) {
		t.Fatalf("错误应包裹 errTwoFactorDataCorrupted, got %v", err)
	}
}

func TestValidateTwoFactorStoredConfigBadSalt(t *testing.T) {
	pw := TwoFactorPasswordConfig{
		KDF:      "argon2id-v1",
		Salt:     b64(8),
		Verifier: b64(32),
	}
	err := validateTwoFactorStoredConfig(pw, nil)
	if err == nil || !strings.Contains(err.Error(), "password.salt") {
		t.Fatalf("Salt 非 16B 应报 salt 错误, got %v", err)
	}
}

func TestValidateTwoFactorStoredConfigBadVerifier(t *testing.T) {
	pw := TwoFactorPasswordConfig{
		KDF:      "argon2id-v1",
		Salt:     b64(16),
		Verifier: b64(8),
	}
	err := validateTwoFactorStoredConfig(pw, nil)
	if err == nil || !strings.Contains(err.Error(), "password.verifier") {
		t.Fatalf("Verifier 非 32B 应报 verifier 错误, got %v", err)
	}
}

func TestValidateTwoFactorStoredConfigEntriesWithoutPassword(t *testing.T) {
	entries := []TwoFactorEntryConfig{{ID: "x"}}
	err := validateTwoFactorStoredConfig(TwoFactorPasswordConfig{}, entries)
	if err == nil || !strings.Contains(err.Error(), "密码配置缺失") {
		t.Fatalf("有条目但密码配置缺失应报错, got %v", err)
	}
	if !errors.Is(err, errTwoFactorDataCorrupted) {
		t.Fatalf("错误应包裹 errTwoFactorDataCorrupted, got %v", err)
	}
}

func validEntry() TwoFactorEntryConfig {
	return TwoFactorEntryConfig{
		Kind:             "totp",
		Algorithm:        "SHA1",
		Digits:           6,
		Period:           30,
		SecretNonce:      b64(12),
		SecretCiphertext: b64(16),
	}
}

func TestValidateTwoFactorStoredEntryValid(t *testing.T) {
	if err := validateTwoFactorStoredEntry(validEntry()); err != nil {
		t.Fatalf("合法条目应通过校验, got %v", err)
	}
}

func TestValidateTwoFactorStoredEntryBadKind(t *testing.T) {
	e := validEntry()
	e.Kind = "hotp"
	err := validateTwoFactorStoredEntry(e)
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("非法 Kind 应报错, got %v", err)
	}
}

func TestValidateTwoFactorStoredEntryBadAlgorithm(t *testing.T) {
	e := validEntry()
	e.Algorithm = "MD5"
	err := validateTwoFactorStoredEntry(e)
	if err == nil || !strings.Contains(err.Error(), "algorithm") {
		t.Fatalf("非法 Algorithm 应报错, got %v", err)
	}
}

func TestValidateTwoFactorStoredEntrySteamDigits(t *testing.T) {
	e := validEntry()
	e.Kind = "steam"
	e.Digits = 6
	err := validateTwoFactorStoredEntry(e)
	if err == nil || !strings.Contains(err.Error(), "Steam") {
		t.Fatalf("Steam Digits=6 应报错, got %v", err)
	}
	e.Digits = 5
	e.Period = 30
	if err := validateTwoFactorStoredEntry(e); err != nil {
		t.Fatalf("Steam 合法参数应通过, got %v", err)
	}
}

func TestValidateTwoFactorStoredEntryEmptySecret(t *testing.T) {
	e := validEntry()
	e.SecretNonce = ""
	e.SecretCiphertext = ""
	err := validateTwoFactorStoredEntry(e)
	if err == nil || !strings.Contains(err.Error(), "加密数据") {
		t.Fatalf("缺少加密数据应报错, got %v", err)
	}
}

func TestValidateTwoFactorStoredEntryBadNonce(t *testing.T) {
	e := validEntry()
	e.SecretNonce = b64(8)
	err := validateTwoFactorStoredEntry(e)
	if err == nil || !strings.Contains(err.Error(), "secretNonce") {
		t.Fatalf("Nonce 非 12B 应报错, got %v", err)
	}
}

func TestSanitizeTwoFactorKind(t *testing.T) {
	cases := map[string]string{
		"steam":      "steam",
		"SteamGuard": "steam",
		"STEAM":      "steam",
		"totp":       "totp",
		"TOTP":       "totp",
		"hotp":       "totp",
		"":           "totp",
	}
	for in, want := range cases {
		if got := sanitizeTwoFactorKind(in); got != want {
			t.Errorf("sanitizeTwoFactorKind(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeTwoFactorAlgorithm(t *testing.T) {
	cases := []struct {
		algo, kind, want string
	}{
		{"sha256", "totp", "SHA256"},
		{"sha512", "totp", "SHA512"},
		{"sha1", "totp", "SHA1"},
		{"md5", "totp", "SHA1"},
		{"", "totp", "SHA1"},
		{"sha256", "steam", "SHA1"},
		{"", "steam", "SHA1"},
	}
	for _, c := range cases {
		if got := sanitizeTwoFactorAlgorithm(c.algo, c.kind); got != c.want {
			t.Errorf("sanitizeTwoFactorAlgorithm(%q, %q) = %q, want %q", c.algo, c.kind, got, c.want)
		}
	}
}

func TestLooksLikeTwoFactorBase32Payload(t *testing.T) {
	cases := map[string]bool{
		"JBSWY3DP":                 true,
		"jbswy3dp":                 true,
		"JBSWY3DPEB3W64TMMQQQ====": true,
		"":                         false,
		"   ":                      false,
		"JBSW Y3DP":                false,
		"ABC018":                   false,
		"abc8":                     false,
	}
	for in, want := range cases {
		if got := looksLikeTwoFactorBase32Payload(in); got != want {
			t.Errorf("looksLikeTwoFactorBase32Payload(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSanitizeTwoFactorEntryIcon(t *testing.T) {
	cases := []struct {
		kind, icon, want string
	}{
		{"steam", "", "steam"},
		{"steamguard", "any", "steam"},
		{"totp", "", "lock"},
		{"totp", "custom-icon", "custom-icon"},
		{"totp", "  spaced  ", "spaced"},
	}
	for _, c := range cases {
		if got := sanitizeTwoFactorEntryIcon(c.kind, c.icon); got != c.want {
			t.Errorf("sanitizeTwoFactorEntryIcon(%q, %q) = %q, want %q", c.kind, c.icon, got, c.want)
		}
	}
}
