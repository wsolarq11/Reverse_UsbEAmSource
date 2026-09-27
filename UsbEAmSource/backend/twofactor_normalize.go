// AUTO-RECONSTRUCTED — DOMAIN: twofactor normalize
// 研究用途
package main

import (
	"encoding/base32"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

// normalizeTwoFactorConfig 归一化二因素配置顶层。
// [S 汇编实证 0x1409c65a0, 0x1a0]：Password 经 normalizeTwoFactorPasswordConfig，
// Entries 经 normalizeTwoFactorEntryConfigs，其余原样组装返回。
func normalizeTwoFactorConfig(cfg TwoFactorConfig) TwoFactorConfig {
	cfg.Password = normalizeTwoFactorPasswordConfig(cfg.Password)
	cfg.Entries = normalizeTwoFactorEntryConfigs(cfg.Entries)
	return cfg
}

// normalizeTwoFactorPasswordConfig 归一化二因素密码配置。
// [S 汇编实证 0x1409c7240, 0x1e0]：Salt/Verifier TrimSpace，二者任一为空返回零值；
// 否则 KDF TrimSpace 空则默认 "argon2id-v1"（11B @0x140c47683），Blank 原样透传。
func normalizeTwoFactorPasswordConfig(p TwoFactorPasswordConfig) TwoFactorPasswordConfig {
	salt := strings.TrimSpace(p.Salt)
	verifier := strings.TrimSpace(p.Verifier)
	if salt == "" || verifier == "" {
		return TwoFactorPasswordConfig{}
	}
	kdf := strings.TrimSpace(p.KDF)
	if kdf == "" {
		kdf = "argon2id-v1"
	}
	return TwoFactorPasswordConfig{
		KDF:      kdf,
		Salt:     salt,
		Verifier: verifier,
		Blank:    p.Blank,
	}
}

// normalizeTwoFactorEntryConfigs 归一化条目列表并按 ID（小写）去重。
// [S 汇编实证 0x1409c7420, 0x380]：make 等长 slice + map[string]struct{} 去重；
// 逐条 normalizeTwoFactorEntryConfig，ID 空跳过，ToLower(ID) 已见跳过。
func normalizeTwoFactorEntryConfigs(entries []TwoFactorEntryConfig) []TwoFactorEntryConfig {
	result := make([]TwoFactorEntryConfig, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		e = normalizeTwoFactorEntryConfig(e)
		if e.ID == "" {
			continue
		}
		key := strings.ToLower(e.ID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, e)
	}
	return result
}

// normalizeTwoFactorEntryConfig 归一化单个二因素条目。
// [S 汇编实证 0x1409c77a0, 0x500]：kind 经 sanitizeTwoFactorKind；name 空则
// resolveTwoFactorDisplayName(kind, issuer, accountName)；steam 清空 iconRef/iconURL、
// digits=5、period=30；非 steam digits 限 4..10 默认 6、period 限 5..300 默认 30。
func normalizeTwoFactorEntryConfig(e TwoFactorEntryConfig) TwoFactorEntryConfig {
	kind := sanitizeTwoFactorKind(e.Kind)
	name := strings.TrimSpace(e.Name)
	issuer := strings.TrimSpace(e.Issuer)
	accountName := strings.TrimSpace(e.AccountName)
	if name == "" {
		name = resolveTwoFactorDisplayName(kind, issuer, accountName)
	}
	iconRef := strings.TrimSpace(e.IconRef)
	iconURL := strings.TrimSpace(e.IconURL)
	if kind == "steam" {
		iconRef = ""
		iconURL = ""
	}
	digits := e.Digits
	if kind == "steam" {
		digits = 5
	} else if digits < 4 || digits > 10 {
		digits = 6
	}
	period := e.Period
	if kind == "steam" {
		period = 30
	} else if period < 5 || period > 300 {
		period = 30
	}
	return TwoFactorEntryConfig{
		ID:               strings.TrimSpace(e.ID),
		Kind:             kind,
		Name:             name,
		Issuer:           issuer,
		AccountName:      accountName,
		Icon:             sanitizeTwoFactorEntryIcon(kind, e.Icon),
		IconData:         sanitizeTwoFactorEntryIconData(kind, e.IconData),
		IconRef:          iconRef,
		IconURL:          iconURL,
		Algorithm:        sanitizeTwoFactorAlgorithm(e.Algorithm, kind),
		Digits:           digits,
		Period:           period,
		SecretNonce:      strings.TrimSpace(e.SecretNonce),
		SecretCiphertext: strings.TrimSpace(e.SecretCiphertext),
	}
}

// sanitizeTwoFactorEntryIconData 归一化二因素条目图标数据。
// [S 汇编实证 0x1409c7e00, 0x160]：kind 为 steam 返回空；TrimSpace 空返回空；
// lower 以 "data:image/"（11B @0x140c47539）开头原样返回；以 "base64:"（7B @0x140c39568）
// 开头去前缀后交 normalizeTwoFactorOpaqueIconPayload；其余返回空。
func sanitizeTwoFactorEntryIconData(kind, iconData string) string {
	if sanitizeTwoFactorKind(kind) == "steam" {
		return ""
	}
	d := strings.TrimSpace(iconData)
	if d == "" {
		return ""
	}
	lower := strings.ToLower(d)
	if strings.HasPrefix(lower, "data:image/") {
		return d
	}
	if strings.HasPrefix(lower, "base64:") {
		return normalizeTwoFactorOpaqueIconPayload(d[7:])
	}
	return ""
}

// normalizeTwoFactorOpaqueIconPayload 归一化不透明图标负载。
// [S 汇编实证 0x1409c7f60, 0x1c0]：先以 base64 解码器尝试，成功且结果为 data:image/
// 直接返回；否则负载为 base32 时再以 base32 解码器尝试；两次均未得 data:image 时
// 优先返回成功结果，最终兜底返回 "base64:" + 原负载。
func normalizeTwoFactorOpaqueIconPayload(payload string) string {
	p := strings.TrimSpace(payload)
	if p == "" {
		return ""
	}
	res1, ok1 := normalizeTwoFactorOpaqueIconPayloadWithDecoder(p, decodeTwoFactorBase64Secret)
	if ok1 && strings.HasPrefix(strings.ToLower(res1), "data:image/") {
		return res1
	}
	res2, ok2 := "", false
	if looksLikeTwoFactorBase32Payload(p) {
		res2, ok2 = normalizeTwoFactorOpaqueIconPayloadWithDecoder(p, decodeTwoFactorBase32Secret)
		if ok2 && strings.HasPrefix(strings.ToLower(res2), "data:image/") {
			return res2
		}
	}
	if ok1 {
		return res1
	}
	if ok2 {
		return res2
	}
	return "base64:" + p
}

// normalizeTwoFactorOpaqueIconPayloadWithDecoder 用给定解码器归一化不透明图标负载。
// [S 汇编实证 0x1409c8120, 0x200]：解码失败或空返回 ("",false)；解码文本 lower 以
// "data:image/" 开头则 sanitize 后返回；否则 base64 重编码，经 http.DetectContentType
// 判 image/* 则包装 "data:<mime>;base64,<b64>"，非图像包装 "base64:<b64>"。
func normalizeTwoFactorOpaqueIconPayloadWithDecoder(payload string, decoder func(string) ([]byte, error)) (string, bool) {
	decoded, err := decoder(payload)
	if err != nil || len(decoded) == 0 {
		return "", false
	}
	s := strings.TrimSpace(string(decoded))
	if strings.HasPrefix(strings.ToLower(s), "data:image/") {
		return sanitizeTwoFactorEntryIconData("totp", s), true
	}
	b64 := base64.StdEncoding.EncodeToString(decoded)
	mime := strings.ToLower(strings.TrimSpace(http.DetectContentType(decoded)))
	if strings.HasPrefix(mime, "image/") {
		return sanitizeTwoFactorEntryIconData("totp", "data:"+mime+";base64,"+b64), true
	}
	return "base64:" + b64, true
}

// resolveTwoFactorDisplayName 解析二因素显示名。
// [S 汇编实证 0x1409c84e0, 0x100]：accountName TrimSpace 非空返回；否则 issuer TrimSpace
// 非空返回；否则 kind 归一化为 steam 返回 "steam"，其余返回 "lock"。
func resolveTwoFactorDisplayName(kind, issuer, accountName string) string {
	issuer = strings.TrimSpace(issuer)
	accountName = strings.TrimSpace(accountName)
	if accountName != "" {
		return accountName
	}
	if issuer != "" {
		return issuer
	}
	if sanitizeTwoFactorKind(kind) == "steam" {
		return "steam"
	}
	return "lock"
}

// stripTwoFactorSecretGrouping 去掉密钥字符串中的空白与空格。
// [S 汇编实证 0x1409d51a0, 0x60]：TrimSpace 后 Replace(" ", "", -1)。
func stripTwoFactorSecretGrouping(s string) string {
	return strings.Replace(strings.TrimSpace(s), " ", "", -1)
}

// decodeTwoFactorBase32Secret 解码 base32 二因素密钥。
// [S 汇编实证 0x1409d56c0, 0x60]：strip 分组 + ToUpper 后 base32.StdEncoding 解码。
func decodeTwoFactorBase32Secret(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(strings.ToUpper(stripTwoFactorSecretGrouping(s)))
}

// decodeTwoFactorBase64Secret 容错解码 base64 二因素密钥（4 变体依次尝试）。
// [S 汇编实证 0x1409d5720, 0x180]：TrimSpace 空返回 errors.New("empty secret")；
// 依次 RawStd/RawURL 解码，失败则按 4 倍数补 "=" 后 URL/Std 兜底。
func decodeTwoFactorBase64Secret(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty secret")
	}
	if d, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(d) > 0 {
		return d, nil
	}
	if d, err := base64.RawURLEncoding.DecodeString(s); err == nil && len(d) > 0 {
		return d, nil
	}
	if n := len(s) % 4; n != 0 {
		s += strings.Repeat("=", 4-n)
	}
	if d, err := base64.URLEncoding.DecodeString(s); err == nil && len(d) > 0 {
		return d, nil
	}
	return base64.StdEncoding.DecodeString(s)
}
