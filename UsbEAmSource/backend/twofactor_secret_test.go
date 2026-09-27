// AUTO-RECONSTRUCTED TESTS — DOMAIN: twofactor secret decode/normalize
// 研究用途
package main

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"testing"
)

func TestLooksLikeStrictTwoFactorBase32Secret(t *testing.T) {
	cases := map[string]bool{
		"JBSWY3DP":     true,
		"jbswy3dp":     true,
		"JBSW Y3DP":    false,
		"":             false,
		"   ":          false,
		"ABC018":       false,
		"JBSWY3DP====": false, // 含 = 非法
		"JBSWY3DPEB3W": true,
	}
	for in, want := range cases {
		if got := looksLikeStrictTwoFactorBase32Secret(in); got != want {
			t.Errorf("looksLikeStrictTwoFactorBase32Secret(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNormalizeTwoFactorSecretText(t *testing.T) {
	cases := map[string]string{
		"JBSW Y3DP":          "JBSWY3DP",
		"JBSW\tY3DP\r\nEB3W": "JBSWY3DPEB3W",
		"  JBSW Y3DP  ":      "JBSWY3DP",
		"":                   "",
		"no-whitespace":      "no-whitespace",
	}
	for in, want := range cases {
		if got := normalizeTwoFactorSecretText(in); got != want {
			t.Errorf("normalizeTwoFactorSecretText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRepairLegacySteamSecretBytes(t *testing.T) {
	// "Hello" 的 base32 = "JBSWY3DP"；legacy 存储 = 该 base32 文本被 RawStd base64 解码成字节。
	base32text := "JBSWY3DP"
	secret, err := base64.RawStdEncoding.DecodeString(base32text)
	if err != nil {
		t.Fatal(err)
	}
	fixed, ok := repairLegacySteamSecretBytes(secret)
	if !ok {
		t.Fatal("repairLegacySteamSecretBytes should repair legacy steam secret")
	}
	if !bytes.Equal(fixed, []byte("Hello")) {
		t.Fatalf("fixed = %q, want %q", fixed, "Hello")
	}

	// 空输入 → 不修复。
	if _, ok := repairLegacySteamSecretBytes(nil); ok {
		t.Fatal("repairLegacySteamSecretBytes(nil) should not report ok")
	}

	// 普通随机字节（非 legacy 格式）→ 不修复。
	if _, ok := repairLegacySteamSecretBytes([]byte{0xde, 0xad, 0xbe, 0xef}); ok {
		t.Fatal("repairLegacySteamSecretBytes(random) should not report ok")
	}
}

func TestNormalizeStoredTwoFactorSecretBytes(t *testing.T) {
	base32text := "JBSWY3DP"
	legacy, err := base64.RawStdEncoding.DecodeString(base32text)
	if err != nil {
		t.Fatal(err)
	}

	// steam → 触发修复。
	if got := normalizeStoredTwoFactorSecretBytes("steam", legacy); !bytes.Equal(got, []byte("Hello")) {
		t.Fatalf("steam normalize = %q, want %q", got, "Hello")
	}
	// 非 steam → 原样返回。
	if got := normalizeStoredTwoFactorSecretBytes("totp", legacy); !bytes.Equal(got, legacy) {
		t.Fatal("totp normalize should return original bytes")
	}
	// 普通字节（steam 但非 legacy）→ 原样返回。
	plain := []byte{0x01, 0x02, 0x03}
	if got := normalizeStoredTwoFactorSecretBytes("steam", plain); !bytes.Equal(got, plain) {
		t.Fatal("steam non-legacy normalize should return original bytes")
	}
}

func TestDecryptTwoFactorSecretRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32B AES-256
	plaintext := []byte("Hello")

	nonceB64, ciphertextB64, err := encryptTwoFactorSecret(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	entry := TwoFactorEntryConfig{
		Kind:             "totp",
		Algorithm:        "SHA1",
		Digits:           6,
		Period:           30,
		SecretNonce:      nonceB64,
		SecretCiphertext: ciphertextB64,
	}

	got, err := decryptTwoFactorSecret(entry, key)
	if err != nil {
		t.Fatalf("decryptTwoFactorSecret error: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypt = %q, want %q", got, plaintext)
	}
}

func TestDecryptTwoFactorSecretErrors(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	// 非法条目（Kind 非法）→ validate 失败。
	badEntry := TwoFactorEntryConfig{Kind: "hotp"}
	if _, err := decryptTwoFactorSecret(badEntry, key); err == nil {
		t.Fatal("invalid entry should error")
	}

	// 合法结构但密文被篡改 → 认证失败。
	nonceB64, _, _ := encryptTwoFactorSecret([]byte("x"), key)
	entry := TwoFactorEntryConfig{
		Kind:             "totp",
		Algorithm:        "SHA1",
		Digits:           6,
		Period:           30,
		SecretNonce:      nonceB64,
		SecretCiphertext: base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}
	if _, err := decryptTwoFactorSecret(entry, key); err == nil {
		t.Fatal("tampered ciphertext should error")
	}
}

// 验证 base32 编码约定（测试自检，非重建函数）。
func TestBase32Convention(t *testing.T) {
	if got := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("Hello")); got != "JBSWY3DP" {
		t.Fatalf("unexpected base32 encoding: %q", got)
	}
}
