// AUTO-RECONSTRUCTED — DOMAIN: twofactor code generation (TOTP/HOTP/Steam)
// 研究用途
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"hash"
	"time"
)

// newTwoFactorHMAC 依据算法名构造 HMAC 哈希器。
// [S 汇编实证 0x1409d3400, 0x240]：sanitizeTwoFactorAlgorithm(algorithm,"totp")
// 后按 "SHA256"("SHA2"+"56")/"SHA512"("SHA5"+"12")/默认(SHA1) 选择
// hmac.New(sha256/512/1.New, secret)。
func newTwoFactorHMAC(secret []byte, algorithm string) hash.Hash {
	algo := sanitizeTwoFactorAlgorithm(algorithm, "totp")
	switch algo {
	case "SHA256":
		return hmac.New(sha256.New, secret)
	case "SHA512":
		return hmac.New(sha512.New, secret)
	default:
		return hmac.New(sha1.New, secret)
	}
}

// buildHOTPValue 计算 HOTP 动态截断值（RFC 4226）。
// [S 汇编实证 0x1409d3280, 0x180]：newTwoFactorHMAC → make([]byte,8) →
// bswap 写入 big-endian counter → mac.Write → mac.Sum(nil) →
// offset=sum[len-1]&0xf → bswap(sum[offset:offset+4])&0x7fffffff。
func buildHOTPValue(secret []byte, counter uint64, algorithm string) (uint32, error) {
	mac := newTwoFactorHMAC(secret, algorithm)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)
	if _, err := mac.Write(buf); err != nil {
		return 0, err
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	return binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff, nil
}

// buildStandardTwoFactorCode 生成标准 TOTP 验证码（digits 位补零）。
// [S 汇编实证 0x1409d2f80, 0x140]：buildHOTPValue → mod=10^digits →
// code=hotp%uint32(mod) → fmt.Sprintf("%0*d", digits, code)。
func buildStandardTwoFactorCode(secret []byte, counter uint64, digits int, algorithm string) (string, error) {
	hotp, err := buildHOTPValue(secret, counter, algorithm)
	if err != nil {
		return "", err
	}
	mod := 1
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", digits, hotp%uint32(mod)), nil
}

// buildSteamTwoFactorCode 生成 Steam Guard 5 位验证码。
// [S 汇编实证 0x1409d30c0, 0x1c0]：buildHOTPValue(secret, counter, "SHA1") →
// 5 轮 v%26 映射到字母表 "23456789BCDFGHJKMNPQRTVWXY"(26B @0x141bc6980)，v/=26。
func buildSteamTwoFactorCode(secret []byte, counter uint64) (string, error) {
	hotp, err := buildHOTPValue(secret, counter, "SHA1")
	if err != nil {
		return "", err
	}
	const steamAlphabet = "23456789BCDFGHJKMNPQRTVWXY"
	v := int64(hotp)
	code := make([]byte, 0, 5)
	for i := 0; i < 5; i++ {
		code = append(code, steamAlphabet[v%26])
		v /= 26
	}
	return string(code), nil
}

// buildTwoFactorCode 顶层 TOTP 生成：计算 counter 与 remaining，按 kind 分发。
// [S 汇编实证 0x1409d2c80, 0x300]：sanitizeTwoFactorKind → period clamp（steam 或
// 越界 → 30，否则 [5,300]）→ unix=now.Unix() → counter=unix/period →
// remaining=period-(unix%period)（<=0 置 period）→ steam 走 buildSteamTwoFactorCode，
// 否则 digits clamp([4,10]→6) + sanitizeTwoFactorAlgorithm 后 buildStandardTwoFactorCode。
func buildTwoFactorCode(entry TwoFactorEntryConfig, secret []byte, now time.Time) (string, int64, error) {
	kind := sanitizeTwoFactorKind(entry.Kind)
	period := entry.Period
	if kind == "steam" {
		period = 30
	} else if period < 5 || period > 300 {
		period = 30
	}

	unix := now.Unix()
	p := int64(period)
	counter := unix / p
	remaining := p - (unix % p)
	if remaining <= 0 {
		remaining = p
	}

	if kind == "steam" {
		code, err := buildSteamTwoFactorCode(secret, uint64(counter))
		return code, remaining, err
	}
	digits := entry.Digits
	if digits < 4 || digits > 10 {
		digits = 6
	}
	algorithm := sanitizeTwoFactorAlgorithm(entry.Algorithm, kind)
	code, err := buildStandardTwoFactorCode(secret, uint64(counter), digits, algorithm)
	return code, remaining, err
}
