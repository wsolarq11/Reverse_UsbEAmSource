// AUTO-RECONSTRUCTED — DOMAIN: twofactor crypto/password tests
// 研究用途
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestRandomBytes(t *testing.T) {
	b, err := randomBytes(32)
	if err != nil {
		t.Fatalf("randomBytes(32) error: %v", err)
	}
	if len(b) != 32 {
		t.Fatalf("randomBytes(32) len = %d, want 32", len(b))
	}
	if allZero(b) {
		t.Fatal("randomBytes(32) returned all-zero bytes")
	}
	empty, err := randomBytes(0)
	if err != nil || len(empty) != 0 {
		t.Fatalf("randomBytes(0) = %d bytes, err=%v; want empty", len(empty), err)
	}
}

func TestZeroTwoFactorBytes(t *testing.T) {
	b := []byte{1, 2, 3, 4}
	got := zeroTwoFactorBytes(b)
	if !bytes.Equal(got, []byte{0, 0, 0, 0}) {
		t.Fatalf("zeroTwoFactorBytes = %v, want all zero", got)
	}
	// 返回值与入参同一底层数组
	if &got[0] != &b[0] {
		t.Fatal("zeroTwoFactorBytes must return the same backing array")
	}
	// 空切片不 panic
	if zeroTwoFactorBytes(nil) != nil {
		t.Fatal("zeroTwoFactorBytes(nil) != nil")
	}
}

func TestIsTwoFactorPasswordConfigured(t *testing.T) {
	cases := []struct {
		name string
		p    TwoFactorPasswordConfig
		want bool
	}{
		{"empty", TwoFactorPasswordConfig{}, false},
		{"salt only", TwoFactorPasswordConfig{Salt: "AA=="}, false},
		{"verifier only", TwoFactorPasswordConfig{Verifier: "BB=="}, false},
		{"both", TwoFactorPasswordConfig{Salt: "AA==", Verifier: "BB=="}, true},
		{"both padded", TwoFactorPasswordConfig{Salt: "  AA==  ", Verifier: "\tBB==\n"}, true},
	}
	for _, c := range cases {
		if got := isTwoFactorPasswordConfigured(c.p); got != c.want {
			t.Errorf("%s: isTwoFactorPasswordConfigured = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestDeriveTwoFactorKey(t *testing.T) {
	salt := []byte{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	got := deriveTwoFactorKey("password", salt)
	if len(got) != 32 {
		t.Fatalf("deriveTwoFactorKey len = %d, want 32", len(got))
	}
	// 等价性：与直接 argon2.IDKey 调用一致（锁定 time/memory/threads/keyLen）。
	want := argon2.IDKey([]byte("password"), salt, 3, 32768, 4, 32)
	if !bytes.Equal(got, want) {
		t.Fatal("deriveTwoFactorKey != argon2.IDKey(...,3,32768,4,32)")
	}
	// TrimSpace 生效：前后空白密码与原始密码派生结果一致。
	trimmed := deriveTwoFactorKey("  password  ", salt)
	if !bytes.Equal(trimmed, got) {
		t.Fatal("deriveTwoFactorKey should TrimSpace the password")
	}
	// 确定性
	again := deriveTwoFactorKey("password", salt)
	if !bytes.Equal(again, got) {
		t.Fatal("deriveTwoFactorKey not deterministic")
	}
}

func TestBuildTwoFactorPasswordVerifier(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32B
	got := buildTwoFactorPasswordVerifier(key)

	// 独立构造：盐常量 + key，sha256，base64 Std。
	const salt = "usbeam-two-factor-password:"
	data := make([]byte, 0, len(salt)+len(key))
	data = append(data, salt...)
	data = append(data, key...)
	sum := sha256.Sum256(data)
	want := base64.StdEncoding.EncodeToString(sum[:])

	if got != want {
		t.Fatalf("buildTwoFactorPasswordVerifier = %q, want %q", got, want)
	}
}

func TestVerifyTwoFactorPasswordRoundTrip(t *testing.T) {
	const password = "s3cret-password"
	salt := []byte("0123456789abcdef") // 16B
	key := deriveTwoFactorKey(password, salt)
	cfg := TwoFactorPasswordConfig{
		KDF:      "argon2id-v1",
		Salt:     base64.StdEncoding.EncodeToString(salt),
		Verifier: buildTwoFactorPasswordVerifier(key),
	}

	got, err := verifyTwoFactorPassword(cfg, password)
	if err != nil {
		t.Fatalf("verifyTwoFactorPassword ok case error: %v", err)
	}
	if !bytes.Equal(got, key) {
		t.Fatal("verifyTwoFactorPassword returned wrong key")
	}

	if _, err := verifyTwoFactorPassword(cfg, "wrong-password"); err == nil {
		t.Fatal("verifyTwoFactorPassword accepted wrong password")
	}
}

func TestVerifyTwoFactorPasswordErrors(t *testing.T) {
	if _, err := verifyTwoFactorPassword(TwoFactorPasswordConfig{}, "x"); err == nil {
		t.Fatal("unconfigured password should error")
	}
	badKDF := TwoFactorPasswordConfig{
		KDF:      "pbkdf2-v1",
		Salt:     base64.StdEncoding.EncodeToString([]byte("0123456789abcdef")),
		Verifier: buildTwoFactorPasswordVerifier(deriveTwoFactorKey("x", []byte("0123456789abcdef"))),
	}
	if _, err := verifyTwoFactorPassword(badKDF, "x"); err == nil {
		t.Fatal("unsupported KDF should error")
	}
	if !strings.Contains(errText(badKDF, "x"), "TWO_FACTOR_DATA_CORRUPTED") {
		t.Fatal("KDF error should wrap errTwoFactorDataCorrupted")
	}
	badSalt := TwoFactorPasswordConfig{
		KDF:      "argon2id-v1",
		Salt:     "not-base64",
		Verifier: buildTwoFactorPasswordVerifier(deriveTwoFactorKey("x", []byte("0123456789abcdef"))),
	}
	if _, err := verifyTwoFactorPassword(badSalt, "x"); err == nil {
		t.Fatal("invalid salt should error")
	}
}

func TestEncryptTwoFactorSecretRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32B AES-256
	plaintext := []byte("hello two-factor secret")

	nonceB64, ciphertextB64, err := encryptTwoFactorSecret(plaintext, key)
	if err != nil {
		t.Fatalf("encryptTwoFactorSecret error: %v", err)
	}

	// 独立解密验证 round-trip。
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		t.Fatal(err)
	}
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		t.Fatal(err)
	}
	got, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		t.Fatalf("gcm.Open error: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatal("encryptTwoFactorSecret round-trip mismatch")
	}
}

func errText(p TwoFactorPasswordConfig, password string) string {
	_, err := verifyTwoFactorPassword(p, password)
	if err == nil {
		return ""
	}
	return err.Error()
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
