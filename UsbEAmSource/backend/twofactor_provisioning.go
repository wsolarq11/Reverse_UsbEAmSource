// AUTO-RECONSTRUCTED — DOMAIN: twofactor provisioning secret decode
// 研究用途
package main

import (
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// decodeTwoFactorHexSecret 解码十六进制二因素密钥。
// [S 汇编实证 0x1409d58a0, 0x40]：hex.DecodeString(stripTwoFactorSecretGrouping(s))。
func decodeTwoFactorHexSecret(s string) ([]byte, error) {
	return hex.DecodeString(stripTwoFactorSecretGrouping(s))
}

// looksLikeTwoFactorBase32Secret 判断文本是否仅由 base32 字符组成。
// [S 汇编实证 0x1409d5020, 0xc0]：ToUpper(stripTwoFactorSecretGrouping(s)) 空 → false；
// 逐 rune 须落在 [A-Z] 或 [2-7]，任一越界 → false；全合法 → true。
func looksLikeTwoFactorBase32Secret(s string) bool {
	s = strings.ToUpper(stripTwoFactorSecretGrouping(s))
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '2' && r <= '7') {
			continue
		}
		return false
	}
	return true
}

// looksLikeTwoFactorHexSecret 判断文本是否仅由十六进制字符组成。
// [S 汇编实证 0x1409d50e0, 0xc0]：stripTwoFactorSecretGrouping(s) 空或长度奇数 → false；
// 逐 rune 须落在 [0-9]/[a-f]/[A-F]，任一越界 → false；全合法 → true。
func looksLikeTwoFactorHexSecret(s string) bool {
	s = stripTwoFactorSecretGrouping(s)
	if s == "" || len(s)%2 != 0 {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

// parseTwoFactorProvisioningURL 解析二因素 provision URI，容错 steam 裸密钥格式。
// [S 汇编实证 0x1409d6ba0, 0x220]：TrimSpace→ToLower；若以 "steam://"(8B @0x140c3c30c)
// 或 "steamguard://"(13B @0x140c4e1c7) 开头，去掉前缀后的 rest 截断到首个 "/?#"
// (@0x140c33d3a)；若 rest 含 "%"，拼 prefix+"/"+rest（把裸密钥当 path），否则原样；
// 最后 net/url.Parse。
func parseTwoFactorProvisioningURL(text string) (*url.URL, error) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	for _, prefix := range []string{"steam://", "steamguard://"} {
		if strings.HasPrefix(lower, prefix) {
			rest := trimmed[len(prefix):]
			if i := strings.IndexAny(rest, "/?#"); i >= 0 {
				rest = rest[:i]
			}
			if strings.IndexByte(rest, '%') >= 0 {
				return url.Parse(trimmed[:len(prefix)] + "/" + rest)
			}
			return url.Parse(trimmed)
		}
	}
	return url.Parse(trimmed)
}

// parseTwoFactorRawQuery 解析二因素 provision query（保留字面 '+'）。
// [S 汇编实证 0x1409d7480, 0xc0]：TrimSpace 空 → (空 map,nil)；否则
// Replace(q, "+"(0x1411caca0), "%2B"(0x140c33d3d), -1) 后 net/url parseQuery。
func parseTwoFactorRawQuery(rawQuery string) (url.Values, error) {
	q := strings.TrimSpace(rawQuery)
	if q == "" {
		return make(url.Values), nil
	}
	q = strings.Replace(q, "+", "%2B", -1)
	return url.ParseQuery(q)
}

// firstNonEmptyQueryValue 返回 names 中第一个 TrimSpace 非空的 query 值。
// [S 汇编实证 0x1409d83e0, 0xe0]：遍历 names，mapaccess1_faststr 取 values[name]，
// 切片空跳过，否则 TrimSpace(values[name][0]) 非空即返回；全部空 → ""。
func firstNonEmptyQueryValue(values url.Values, names []string) string {
	for _, name := range names {
		vs := values[name]
		if len(vs) == 0 {
			continue
		}
		if v := strings.TrimSpace(vs[0]); v != "" {
			return v
		}
	}
	return ""
}

// resolveSteamProvisioningSecretText 解析 steam provision 的密钥文本。
// [S 汇编实证 0x1409d55a0, 0x120]：先 firstNonEmptyQueryValue 取
// ["secret","shared_secret","sharedSecret","key"]；空则 TrimSpace(u.Host) 非空返回；
// 再 TrimSpace(u.Path)，若以 '/' 开头去首个 '/'；否则返回 path。
func resolveSteamProvisioningSecretText(u *url.URL, query url.Values) string {
	if v := firstNonEmptyQueryValue(query, []string{"secret", "shared_secret", "sharedSecret", "key"}); v != "" {
		return v
	}
	if host := strings.TrimSpace(u.Host); host != "" {
		return host
	}
	path := strings.TrimSpace(u.Path)
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	return path
}

// provisioningUsesSteamSharedSecret 判断 provision 文本是否走 steam shared_secret。
// [S 汇编实证 0x1409d4ea0, 0x180]：TrimSpace 空 → false；ToLower 后非
// "otpauth://"(10B @0x140c44289)/"steam://"(8B)/"steamguard://"(13B @0x140c4e1c7)
// 前缀 → false；parseTwoFactorProvisioningURL + parseTwoFactorRawQuery 任一失败 →
// false；否则 firstNonEmptyQueryValue(["shared_secret","sharedSecret"]) != ""。
func provisioningUsesSteamSharedSecret(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	lower := strings.ToLower(t)
	if !strings.HasPrefix(lower, "otpauth://") &&
		!strings.HasPrefix(lower, "steam://") &&
		!strings.HasPrefix(lower, "steamguard://") {
		return false
	}
	u, err := parseTwoFactorProvisioningURL(t)
	if err != nil {
		return false
	}
	query, err := parseTwoFactorRawQuery(u.RawQuery)
	if err != nil {
		return false
	}
	return firstNonEmptyQueryValue(query, []string{"shared_secret", "sharedSecret"}) != ""
}

// shouldPreferSteamBase64Secret 判断 steam 密钥是否优先按 base64 解码。
// [S 汇编实证 0x1409d4e00, 0xa0]：provisioningUsesSteamSharedSecret(provisioningText)
// 真 → true；否则 strings.IndexAny(secret, "+/=_-"(0x140c35d99)) >= 0。
func shouldPreferSteamBase64Secret(provisioningText string, secret string) bool {
	if provisioningUsesSteamSharedSecret(provisioningText) {
		return true
	}
	return strings.IndexAny(secret, "+/=_-") >= 0
}

// resolveTwoFactorSecretDecoders 按 kind/文本特征选择密钥解码器优先级序列。
// [S 汇编实证 0x1409d4c20, 0x1e0]：sanitizeTwoFactorKind(kind) != "steam" →
// [base32,base64,hex]；steam 且 shouldPreferSteamBase64Secret →
// [base64,base32,hex]；steam 且 looksLikeTwoFactorHexSecret 且
// !looksLikeTwoFactorBase32Secret → [hex,base32,base64]；否则 [base32,base64,hex]。
// 解码器函数值 @0x141096a08=base32/0x141096a10=base64/0x141096a18=hex。
func resolveTwoFactorSecretDecoders(provisioningText string, secret string, kind string) []func(string) ([]byte, error) {
	base32Decoder := decodeTwoFactorBase32Secret
	base64Decoder := decodeTwoFactorBase64Secret
	hexDecoder := decodeTwoFactorHexSecret
	if sanitizeTwoFactorKind(kind) != "steam" {
		return []func(string) ([]byte, error){base32Decoder, base64Decoder, hexDecoder}
	}
	if shouldPreferSteamBase64Secret(provisioningText, secret) {
		return []func(string) ([]byte, error){base64Decoder, base32Decoder, hexDecoder}
	}
	if looksLikeTwoFactorHexSecret(secret) && !looksLikeTwoFactorBase32Secret(secret) {
		return []func(string) ([]byte, error){hexDecoder, base32Decoder, base64Decoder}
	}
	return []func(string) ([]byte, error){base32Decoder, base64Decoder, hexDecoder}
}

// extractTwoFactorProvisioningSecretText 从 provision 文本提取密钥字符串。
// [S 汇编实证 0x1409d5200, 0x3a0]：TrimSpace 空 → ("",false)；ToLower 后非
// otpauth:///steam:///steamguard:// 前缀 → ("",false)；parse URL + rawQuery 失败 →
// ("",false)；scheme=ToLower(TrimSpace(u.Scheme))，"steam"/"steamguard" →
// resolveSteamProvisioningSecretText(真)；"otpauth" → host=ToLower(TrimSpace(u.Host))
// 须 "totp"/"steam"，否则 ("",false)；host=="steam" 或 sanitizeTwoFactorKind(kind)
// =="steam" → resolveSteam(真)，否则 firstNonEmptyQueryValue(4 名)(真)。
func extractTwoFactorProvisioningSecretText(text string, kind string) (string, bool) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", false
	}
	lower := strings.ToLower(t)
	if !strings.HasPrefix(lower, "otpauth://") &&
		!strings.HasPrefix(lower, "steam://") &&
		!strings.HasPrefix(lower, "steamguard://") {
		return "", false
	}
	u, err := parseTwoFactorProvisioningURL(t)
	if err != nil {
		return "", false
	}
	query, err := parseTwoFactorRawQuery(u.RawQuery)
	if err != nil {
		return "", false
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	switch scheme {
	case "steam", "steamguard":
		return resolveSteamProvisioningSecretText(u, query), true
	case "otpauth":
		host := strings.ToLower(strings.TrimSpace(u.Host))
		if host != "totp" && host != "steam" {
			return "", false
		}
		if host == "steam" || sanitizeTwoFactorKind(kind) == "steam" {
			return resolveSteamProvisioningSecretText(u, query), true
		}
		return firstNonEmptyQueryValue(query, []string{"secret", "shared_secret", "sharedSecret", "key"}), true
	}
	return "", false
}

// decodeTwoFactorSecret 解码二因素密钥文本（provision URI 或裸密钥）。
// [S 汇编实证 0x1409d4440, 0x1c0]：TrimSpace 得 trimmed；extract 成功用提取文本否则
// 原始 text；normalizeTwoFactorSecretText 后空 → errors.New("二步验证密钥不能为空")
// (@0x140c6fa3b)；resolveTwoFactorSecretDecoders(trimmed,secret,kind) 依次解码，
// 首个 err==nil 且 len>0 返回；全失败 → errors.New("无法识别二步验证密钥格式")
// (@0x140c78aa7)。
func decodeTwoFactorSecret(text string, kind string) ([]byte, error) {
	trimmed := strings.TrimSpace(text)
	secretText, ok := extractTwoFactorProvisioningSecretText(text, kind)
	if !ok {
		secretText = text
	}
	secret := normalizeTwoFactorSecretText(secretText)
	if secret == "" {
		return nil, errors.New("二步验证密钥不能为空")
	}
	for _, decoder := range resolveTwoFactorSecretDecoders(trimmed, secret, kind) {
		decoded, err := decoder(secret)
		if err != nil {
			continue
		}
		if len(decoded) == 0 {
			continue
		}
		return decoded, nil
	}
	return nil, errors.New("无法识别二步验证密钥格式")
}

// decodeProvisioningLabel 解码 provision label 的 URL 路径段。
// [S 汇编实证 0x1409d8260, 0xa0]：TrimSpace 后去前导 '/'（asm dec rbx + neg/sar/and
// 无分支实现 s[1:]）；空 → ""；url.PathUnescape（net/url.unescape mode=2，即
// encodePathSegment）成功返回 TrimSpace 结果，失败返回去 '/' 后的原始 s。
func decodeProvisioningLabel(s string) string {
	s = strings.TrimSpace(s)
	if s != "" && s[0] == '/' {
		s = s[1:]
	}
	if s == "" {
		return ""
	}
	if u, err := url.PathUnescape(s); err == nil {
		return strings.TrimSpace(u)
	}
	return s
}

// splitProvisioningLabel 把 label 按首个 ':' 拆成 (issuer, account)。
// [S 汇编实证 0x1409d8300, 0xe0]：TrimSpace 空 → ("","")；strings.Cut(s, ":"(0x1411cac58))
// 未找到 → ("", s)；找到 → (TrimSpace(before), TrimSpace(after))。
func splitProvisioningLabel(s string) (string, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	issuer, account, found := strings.Cut(s, ":")
	if !found {
		return "", s
	}
	return strings.TrimSpace(issuer), strings.TrimSpace(account)
}

// parseTwoFactorOptionalInt 解析可选整数字段，失败返回 0。
// [S 汇编实证 0x1409d84c0, 0x60]：strconv.Atoi(strings.TrimSpace(s))，err != nil → 0，
// 否则返回解析值。
func parseTwoFactorOptionalInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}
