// AUTO-RECONSTRUCTED — DOMAIN: twofactor provisioning secret decode
// 研究用途
package main

import (
	"net/url"
	"testing"
)

func TestDecodeTwoFactorHexSecret(t *testing.T) {
	got, err := decodeTwoFactorHexSecret("48656C6C6F")
	if err != nil {
		t.Fatalf("decode hex err: %v", err)
	}
	if string(got) != "Hello" {
		t.Fatalf("decode hex = %q, want Hello", string(got))
	}
	// 分组空格被 strip
	got, err = decodeTwoFactorHexSecret("48 65 6c 6c 6f")
	if err != nil {
		t.Fatalf("decode grouped hex err: %v", err)
	}
	if string(got) != "Hello" {
		t.Fatalf("decode grouped hex = %q, want Hello", string(got))
	}
}

func TestLooksLikeTwoFactorBase32Secret(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"JBSWY3DP", true},
		{"JBSW Y3DP", true},
		{"jbswy3dp", true}, // ToUpper 容错小写
		{"JBSWY3DP=", false},
		{"JBSWY3DP1", false},
		{"", false},
		{"4865", false},
	}
	for _, c := range cases {
		if got := looksLikeTwoFactorBase32Secret(c.in); got != c.want {
			t.Errorf("looksLikeTwoFactorBase32Secret(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLooksLikeTwoFactorHexSecret(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"48656c6c6f", true},
		{"48 65 6c 6c 6f", true},
		{"48656C6C6F", true},
		{"48656", false}, // 奇数长度
		{"", false},
		{"48zz", false},
		{"JBSWY3DP", false}, // G 非 hex
	}
	for _, c := range cases {
		if got := looksLikeTwoFactorHexSecret(c.in); got != c.want {
			t.Errorf("looksLikeTwoFactorHexSecret(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseTwoFactorRawQuery(t *testing.T) {
	vals, err := parseTwoFactorRawQuery("a=1&b=2")
	if err != nil {
		t.Fatalf("parse query err: %v", err)
	}
	if vals.Get("a") != "1" || vals.Get("b") != "2" {
		t.Fatalf("parse query = %v", vals)
	}
	// '+' 被保留为字面加号（先转 %2B 再 parseQuery）
	vals, err = parseTwoFactorRawQuery("s=x+y")
	if err != nil {
		t.Fatalf("parse plus query err: %v", err)
	}
	if vals.Get("s") != "x+y" {
		t.Fatalf("parse plus = %q, want x+y", vals.Get("s"))
	}
	// 空 query
	vals, err = parseTwoFactorRawQuery("   ")
	if err != nil {
		t.Fatalf("parse empty err: %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("empty query len = %d", len(vals))
	}
}

func TestFirstNonEmptyQueryValue(t *testing.T) {
	vals := url.Values{"a": {"  "}, "b": {" x "}}
	if got := firstNonEmptyQueryValue(vals, []string{"a", "b"}); got != "x" {
		t.Fatalf("firstNonEmpty = %q, want x", got)
	}
	if got := firstNonEmptyQueryValue(vals, []string{"missing"}); got != "" {
		t.Fatalf("firstNonEmpty missing = %q, want empty", got)
	}
}

func TestResolveSteamProvisioningSecretText(t *testing.T) {
	// query secret 优先
	u := &url.URL{}
	q := url.Values{"secret": {"ABC"}}
	if got := resolveSteamProvisioningSecretText(u, q); got != "ABC" {
		t.Fatalf("resolve secret = %q, want ABC", got)
	}
	// host 兜底
	q = url.Values{}
	u = &url.URL{Host: "DEF"}
	if got := resolveSteamProvisioningSecretText(u, q); got != "DEF" {
		t.Fatalf("resolve host = %q, want DEF", got)
	}
	// path 去前导 '/'
	u = &url.URL{Path: "/GHI"}
	if got := resolveSteamProvisioningSecretText(u, q); got != "GHI" {
		t.Fatalf("resolve path = %q, want GHI", got)
	}
}

func TestProvisioningUsesSteamSharedSecret(t *testing.T) {
	if !provisioningUsesSteamSharedSecret("otpauth://totp/x?shared_secret=ABC") {
		t.Fatal("otpauth shared_secret should be true")
	}
	if !provisioningUsesSteamSharedSecret("steam://ABC?shared_secret=DEF") {
		t.Fatal("steam shared_secret query should be true")
	}
	if provisioningUsesSteamSharedSecret("steam://ABC") {
		t.Fatal("steam bare host should be false (no shared_secret)")
	}
	if provisioningUsesSteamSharedSecret("not-a-url") {
		t.Fatal("not-a-url should be false")
	}
	if provisioningUsesSteamSharedSecret("") {
		t.Fatal("empty should be false")
	}
}

func TestShouldPreferSteamBase64Secret(t *testing.T) {
	if !shouldPreferSteamBase64Secret("", "SGVsbG8+") {
		t.Fatal("secret with + should prefer base64")
	}
	if shouldPreferSteamBase64Secret("", "JBSWY3DP") {
		t.Fatal("base32 secret should not prefer base64")
	}
	// provision 文本判定优先
	if !shouldPreferSteamBase64Secret("otpauth://totp/x?shared_secret=ABC", "whatever") {
		t.Fatal("shared_secret provision should prefer base64")
	}
}

func TestResolveTwoFactorSecretDecoders(t *testing.T) {
	// 非 steam：base32 优先
	ds := resolveTwoFactorSecretDecoders("", "JBSWY3DP", "totp")
	if len(ds) != 3 {
		t.Fatalf("decoders len = %d", len(ds))
	}
	if got, err := ds[0]("JBSWY3DP"); err != nil || string(got) != "Hello" {
		t.Fatalf("non-steam first decoder = %q/%v, want Hello/nil", got, err)
	}
	// steam hex 且非 base32：hex 优先
	ds = resolveTwoFactorSecretDecoders("", "AB12CD", "steam")
	if got, err := ds[0]("AB12CD"); err != nil || len(got) != 3 {
		t.Fatalf("steam hex first decoder = %x/%v, want 3 bytes", got, err)
	}
	// steam base32：base32 优先
	ds = resolveTwoFactorSecretDecoders("", "JBSWY3DP", "steam")
	if got, err := ds[0]("JBSWY3DP"); err != nil || string(got) != "Hello" {
		t.Fatalf("steam base32 first decoder = %q/%v, want Hello/nil", got, err)
	}
}

func TestExtractTwoFactorProvisioningSecretText(t *testing.T) {
	got, ok := extractTwoFactorProvisioningSecretText("otpauth://totp/Test?secret=JBSWY3DP", "totp")
	if !ok || got != "JBSWY3DP" {
		t.Fatalf("extract otpauth = %q/%v, want JBSWY3DP/true", got, ok)
	}
	got, ok = extractTwoFactorProvisioningSecretText("steam://JBSWY3DP", "steam")
	if !ok || got != "JBSWY3DP" {
		t.Fatalf("extract steam = %q/%v, want JBSWY3DP/true", got, ok)
	}
	_, ok = extractTwoFactorProvisioningSecretText("", "totp")
	if ok {
		t.Fatal("extract empty should be false")
	}
}

func TestDecodeTwoFactorSecret(t *testing.T) {
	// base32
	if got, err := decodeTwoFactorSecret("JBSWY3DP", "totp"); err != nil || string(got) != "Hello" {
		t.Fatalf("decode base32 = %q/%v, want Hello/nil", got, err)
	}
	// base64
	if got, err := decodeTwoFactorSecret("SGVsbG8=", "totp"); err != nil || string(got) != "Hello" {
		t.Fatalf("decode base64 = %q/%v, want Hello/nil", got, err)
	}
	// otpauth URI
	if got, err := decodeTwoFactorSecret("otpauth://totp/Test?secret=JBSWY3DP", "totp"); err != nil || string(got) != "Hello" {
		t.Fatalf("decode otpauth = %q/%v, want Hello/nil", got, err)
	}
	// steam URI
	if got, err := decodeTwoFactorSecret("steam://JBSWY3DP", "steam"); err != nil || string(got) != "Hello" {
		t.Fatalf("decode steam = %q/%v, want Hello/nil", got, err)
	}
	// 空文本
	if _, err := decodeTwoFactorSecret("", "totp"); err == nil {
		t.Fatal("decode empty should error")
	}
	// 无法识别的格式
	if _, err := decodeTwoFactorSecret("!!!!", "totp"); err == nil {
		t.Fatal("decode garbage should error")
	}
}

func TestDecodeProvisioningLabel(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"/Test", "Test"},
		{"Test", "Test"},
		{"", ""},
		{"/", ""},
		{"/Test%20Label", "Test Label"},
		{"/%ZZ", "%ZZ"}, // PathUnescape 失败回退去 '/' 原文
		{" /Test ", "Test"},
	}
	for _, c := range cases {
		if got := decodeProvisioningLabel(c.in); got != c.want {
			t.Errorf("decodeProvisioningLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSplitProvisioningLabel(t *testing.T) {
	cases := []struct {
		in         string
		wantIssuer string
		wantAcct   string
	}{
		{"Issuer:Account", "Issuer", "Account"},
		{" Issuer : Account ", "Issuer", "Account"},
		{"NoColon", "", "NoColon"},
		{"", "", ""},
		{"A:B:C", "A", "B:C"},
	}
	for _, c := range cases {
		issuer, acct := splitProvisioningLabel(c.in)
		if issuer != c.wantIssuer || acct != c.wantAcct {
			t.Errorf("splitProvisioningLabel(%q) = (%q,%q), want (%q,%q)", c.in, issuer, acct, c.wantIssuer, c.wantAcct)
		}
	}
}

func TestParseTwoFactorOptionalInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"30", 30},
		{" 30 ", 30},
		{"abc", 0},
		{"", 0},
		{"-1", -1},
	}
	for _, c := range cases {
		if got := parseTwoFactorOptionalInt(c.in); got != c.want {
			t.Errorf("parseTwoFactorOptionalInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
