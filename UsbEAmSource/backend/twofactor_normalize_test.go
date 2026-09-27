// AUTO-RECONSTRUCTED — DOMAIN: twofactor normalize tests
// 研究用途
package main

import "testing"

func TestNormalizeTwoFactorPasswordConfigEmptyComponent(t *testing.T) {
	if got := normalizeTwoFactorPasswordConfig(TwoFactorPasswordConfig{Salt: "", Verifier: "v", KDF: "k"}); got != (TwoFactorPasswordConfig{}) {
		t.Fatalf("salt empty expect zero, got %+v", got)
	}
	if got := normalizeTwoFactorPasswordConfig(TwoFactorPasswordConfig{Salt: "s", Verifier: ""}); got != (TwoFactorPasswordConfig{}) {
		t.Fatalf("verifier empty expect zero, got %+v", got)
	}
}

func TestNormalizeTwoFactorPasswordConfigDefaultKDF(t *testing.T) {
	got := normalizeTwoFactorPasswordConfig(TwoFactorPasswordConfig{Salt: " s ", Verifier: " v ", Blank: true})
	want := TwoFactorPasswordConfig{KDF: "argon2id-v1", Salt: "s", Verifier: "v", Blank: true}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestNormalizeTwoFactorEntryConfigSteam(t *testing.T) {
	got := normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{
		Kind: "steamguard", Name: "n", Issuer: "i", AccountName: "a",
		IconRef: "r", IconURL: "u", Digits: 9, Period: 60,
	})
	if got.Kind != "steam" {
		t.Fatalf("kind expect steam, got %q", got.Kind)
	}
	if got.Digits != 5 || got.Period != 30 {
		t.Fatalf("steam digits/period expect 5/30, got %d/%d", got.Digits, got.Period)
	}
	if got.IconRef != "" || got.IconURL != "" {
		t.Fatalf("steam iconRef/iconURL expect empty, got %q/%q", got.IconRef, got.IconURL)
	}
	if got.Algorithm != "SHA1" {
		t.Fatalf("steam algorithm expect SHA1, got %q", got.Algorithm)
	}
}

func TestNormalizeTwoFactorEntryConfigDigitsPeriodClamp(t *testing.T) {
	for digits, want := range map[int]int{3: 6, 4: 4, 10: 10, 11: 6} {
		got := normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "totp", Digits: digits, Period: 30}).Digits
		if got != want {
			t.Fatalf("digits %d expect %d, got %d", digits, want, got)
		}
	}
	for period, want := range map[int]int{4: 30, 5: 5, 300: 300, 301: 30} {
		got := normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "totp", Digits: 6, Period: period}).Period
		if got != want {
			t.Fatalf("period %d expect %d, got %d", period, want, got)
		}
	}
}

func TestNormalizeTwoFactorEntryConfigNameResolution(t *testing.T) {
	got := normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "totp", Issuer: "iss", AccountName: "acct"})
	if got.Name != "acct" {
		t.Fatalf("name expect acct, got %q", got.Name)
	}
	got = normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "totp", Issuer: "iss"})
	if got.Name != "iss" {
		t.Fatalf("name expect iss, got %q", got.Name)
	}
	got = normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "totp"})
	if got.Name != "lock" {
		t.Fatalf("name expect lock, got %q", got.Name)
	}
	got = normalizeTwoFactorEntryConfig(TwoFactorEntryConfig{Kind: "steam"})
	if got.Name != "steam" {
		t.Fatalf("name expect steam, got %q", got.Name)
	}
}

func TestNormalizeTwoFactorEntryConfigsDedup(t *testing.T) {
	in := []TwoFactorEntryConfig{
		{ID: "abc", Kind: "totp", Name: "a"},
		{ID: "ABC", Kind: "totp", Name: "b"},
		{ID: "", Kind: "totp", Name: "skip"},
		{ID: "def", Kind: "totp", Name: "c"},
	}
	got := normalizeTwoFactorEntryConfigs(in)
	if len(got) != 2 {
		t.Fatalf("expect 2 entries after dedup, got %d", len(got))
	}
	if got[0].ID != "abc" || got[1].ID != "def" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestSanitizeTwoFactorEntryIconData(t *testing.T) {
	if got := sanitizeTwoFactorEntryIconData("steam", "whatever"); got != "" {
		t.Fatalf("steam expect empty, got %q", got)
	}
	if got := sanitizeTwoFactorEntryIconData("totp", "Data:Image/png;base64,AAAA"); got != "Data:Image/png;base64,AAAA" {
		t.Fatalf("data image expect passthrough, got %q", got)
	}
	if got := sanitizeTwoFactorEntryIconData("totp", "garbage"); got != "" {
		t.Fatalf("garbage expect empty, got %q", got)
	}
}

func TestResolveTwoFactorDisplayName(t *testing.T) {
	if got := resolveTwoFactorDisplayName("totp", "iss", "acct"); got != "acct" {
		t.Fatalf("expect acct, got %q", got)
	}
	if got := resolveTwoFactorDisplayName("totp", "iss", ""); got != "iss" {
		t.Fatalf("expect iss, got %q", got)
	}
	if got := resolveTwoFactorDisplayName("totp", "", ""); got != "lock" {
		t.Fatalf("expect lock, got %q", got)
	}
	if got := resolveTwoFactorDisplayName("steam", "", ""); got != "steam" {
		t.Fatalf("expect steam, got %q", got)
	}
}

func TestDecodeTwoFactorBase32Secret(t *testing.T) {
	got, err := decodeTwoFactorBase32Secret("JBSWY3DP")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if string(got) != "Hello" {
		t.Fatalf("expect Hello, got %q", string(got))
	}
	// 带空格分组
	got, err = decodeTwoFactorBase32Secret("JBSW Y3DP")
	if err != nil || string(got) != "Hello" {
		t.Fatalf("grouped expect Hello, got %q err %v", string(got), err)
	}
}

func TestDecodeTwoFactorBase64Secret(t *testing.T) {
	for _, s := range []string{"SGVsbG8=", "SGVsbG8"} {
		got, err := decodeTwoFactorBase64Secret(s)
		if err != nil || string(got) != "Hello" {
			t.Fatalf("%q expect Hello, got %q err %v", s, string(got), err)
		}
	}
	// URL-safe 变体（- _ 替换 + /）
	got, err := decodeTwoFactorBase64Secret("SGVsbG8")
	if err != nil || string(got) != "Hello" {
		t.Fatalf("expect Hello, got %q err %v", string(got), err)
	}
}

func TestDecodeTwoFactorBase64SecretEmpty(t *testing.T) {
	if _, err := decodeTwoFactorBase64Secret("  "); err == nil {
		t.Fatal("expect empty secret error")
	}
}
