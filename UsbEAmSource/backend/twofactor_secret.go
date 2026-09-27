// AUTO-RECONSTRUCTED — DOMAIN: twofactor secret decode/normalize
// 研究用途
package main

import (
	"encoding/base64"
	"strings"
)

// looksLikeStrictTwoFactorBase32Secret 判断文本是否仅由严格 base32 字符组成。
// [S 汇编实证 0x1409d4b60, 0xc0]：ToUpper(TrimSpace(s)) 空 → false；逐 rune 须落在
// [A-Z] 或 [2-7]，任一越界 → false；全合法 → true。
func looksLikeStrictTwoFactorBase32Secret(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
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

// normalizeTwoFactorSecretText 去除密钥文本中的分组空白。
// [S 汇编实证 0x1409d4600, 0x1c0]：strings.NewReplacer 删除 ' '(0x1411ca5c8)/
// '\t'(0x1411ca430)/'\r'(0x1411ca988)/'\n'(0x1411ca400) 四种单字符（Replacer 每次
// 内联构造），作用于 TrimSpace 后的文本。
func normalizeTwoFactorSecretText(s string) string {
	r := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "")
	return r.Replace(strings.TrimSpace(s))
}

// repairLegacySteamSecretBytes 修复遗留 Steam 密钥（base64 密文包裹的 base32 明文）。
// [S 汇编实证 0x1409d48a0, 0x2c0]：len==0 → (nil,false)；对 secret 的 RawStd/RawURL
// 两种 base64 编码，TrimRight "=" 后 normalizeTwoFactorSecretText，去重后仅当
// looksLikeStrictTwoFactorBase32Secret 且 decodeTwoFactorBase32Secret 成功且
// len(decoded)<len(secret) 时返回 (decoded,true)；否则 (nil,false)。
func repairLegacySteamSecretBytes(secret []byte) ([]byte, bool) {
	if len(secret) == 0 {
		return nil, false
	}
	candidates := []string{
		base64.RawStdEncoding.EncodeToString(secret),
		base64.RawURLEncoding.EncodeToString(secret),
	}
	seen := make(map[string]struct{}, 2)
	for _, cand := range candidates {
		text := normalizeTwoFactorSecretText(strings.TrimRight(cand, "="))
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		if !looksLikeStrictTwoFactorBase32Secret(text) {
			continue
		}
		decoded, err := decodeTwoFactorBase32Secret(text)
		if err != nil || len(decoded) == 0 {
			continue
		}
		if len(decoded) < len(secret) {
			return decoded, true
		}
	}
	return nil, false
}

// normalizeStoredTwoFactorSecretBytes 归一化存储态密钥字节。
// [S 汇编实证 0x1409d47c0, 0xe0]：sanitizeTwoFactorKind(kind)=="steam" 时尝试
// repairLegacySteamSecretBytes，修复成功返回修复结果，否则原样返回 secret。
func normalizeStoredTwoFactorSecretBytes(kind string, secret []byte) []byte {
	if sanitizeTwoFactorKind(kind) == "steam" {
		if fixed, ok := repairLegacySteamSecretBytes(secret); ok {
			return fixed
		}
	}
	return secret
}
