// AUTO-RECONSTRUCTED — 二维码外部 URL 处理
// 研究用途
//
// 原始契约：source_funcs.txt qrexternal.go Lines 15-97
// 档位：[S] 反汇编实证（0x14095eda0 / 0x14095ee20 / 0x14095f520 / 0x14095f600）
package main

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

// OpenQRCodeExternalURL 打开二维码外部 URL。
// [S 汇编 0x14095eda0, 128B] 打开二维码外部 URL：normalizeQRCodeExternalURL(url)
// （rax=url.ptr,rbx=url.len）；err 非 nil（rcx=err.type）→ 原样返回；否则
// openWithDefaultHandler(normalized)（@0x1408a8de0）。
func (b *BootstrapService) OpenQRCodeExternalURL(url string) error {
	normalized, err := normalizeQRCodeExternalURL(url)
	if err != nil {
		return err
	}
	return openWithDefaultHandler(normalized)
}

// normalizeQRCodeExternalURL 规范化二维码外部 URL（仅 http/https、无用户信息、
// 主机名 IP/IDNA 规范化并重建 host:port）。
// [S 汇编 0x14095ee20, 1792B] TrimSpace 空或 != raw → 错误1；含不安全文本 → 错误2；
// 至多两次 url.PathUnescape（mode=encodePathSegment）验证双重转义 → 错误3/4；url.Parse
// 失败或 Opaque 非空 → 错误9；Scheme=lower(trim(Scheme)) 非 http/https → 错误5；User
// 非 nil → 错误8；host=trim(Hostname()) 空 → 错误7；net.ParseIP 命中 → ip.String()，
// 否则 idna.Lookup.ToASCII（err 或 trim 后空 → 错误6）；port=Port()；host 含 ":"（IPv6）
// → "[host]" 或 "[host]:port"，否则 ToLower(host) + 可选 ":port"；返回 u.String()。
func normalizeQRCodeExternalURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed != raw {
		return "", qrExternalURLInvalidError("链接不能为空或包含首尾空白")
	}
	if containsUnsafeQRCodeURLText(trimmed) {
		return "", qrExternalURLInvalidError("链接包含不安全字符")
	}

	s := trimmed
	for i := 0; i < 2; i++ {
		decoded, err := url.PathUnescape(s)
		if err != nil {
			return "", qrExternalURLInvalidError("链接转义无效")
		}
		if containsUnsafeQRCodeURLText(decoded) {
			return "", qrExternalURLInvalidError("链接包含不安全的转义字符")
		}
		if decoded == s {
			break
		}
		s = decoded
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed == nil || parsed.Opaque != "" {
		return "", qrExternalURLInvalidError("链接格式无效")
	}

	parsed.Scheme = strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", qrExternalURLInvalidError("只允许 HTTP 或 HTTPS 链接")
	}
	if parsed.User != nil {
		return "", qrExternalURLInvalidError("链接不能包含用户名或密码")
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return "", qrExternalURLInvalidError("链接必须包含有效主机名")
	}

	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	} else {
		var err error
		host, err = idna.Lookup.ToASCII(host)
		if err != nil || strings.TrimSpace(host) == "" {
			return "", qrExternalURLInvalidError("链接主机名无效")
		}
	}

	port := parsed.Port()
	if strings.Contains(host, ":") {
		if port != "" {
			parsed.Host = "[" + host + "]:" + port
		} else {
			parsed.Host = "[" + host + "]"
		}
	} else {
		lower := strings.ToLower(host)
		if port != "" {
			parsed.Host = lower + ":" + port
		} else {
			parsed.Host = lower
		}
	}

	return parsed.String(), nil
}

// containsUnsafeQRCodeURLText 判断文本是否含反斜杠或控制字符（C0 0x00-0x1f、
// DEL/C1 0x7f-0x9f）。
// [S 汇编 0x14095f520, 224B] strings.Contains(s,"\\")（"\\" 1B @0x1411cac40）命中 →
// true；否则按 rune 查表（256B 位图 @0x14196c560，&1 判不安全），rune>0xff 视为安全。
func containsUnsafeQRCodeURLText(s string) bool {
	if strings.Contains(s, "\\") {
		return true
	}
	for _, r := range s {
		if r <= 0x1f || (r >= 0x7f && r <= 0x9f) {
			return true
		}
	}
	return false
}

// qrExternalURLInvalidError 构造带前缀的无效 URL 错误。
// [S 汇编 0x14095f600, 192B] strings.TrimSpace(url) 空 → 默认消息
// "二维码链接无效"（21B @0x140c5f953）；errors.New("QR_EXTERNAL_URL_INVALID: "+s)
// （前缀 25B @0x140c66cfe）。
func qrExternalURLInvalidError(url string) error {
	s := strings.TrimSpace(url)
	if s == "" {
		s = "二维码链接无效"
	}
	return errors.New("QR_EXTERNAL_URL_INVALID: " + s)
}
