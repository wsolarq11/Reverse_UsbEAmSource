// AUTO-RECONSTRUCTED BOOKMARKS — DOMAIN: bookmark 标题解析链
// 研究用途
//
// 档位：
//
//	[S] 反汇编实证体（批次 60）：normalizeBookmarkPageTitle / normalizeBookmarkPageTitleURL /
//	    bookmarkPageTitleCandidates / extractBookmarkPageTitle / fetchBookmarkPageTitle /
//	    resolveBookmarkPageTitleWithNetwork
//
// asm 资产：docs/goresym/pipeline/tmp/{normalizeBookmarkPageTitle,normalizeBookmarkPageTitleURL,
// bookmarkPageTitleCandidates,extractBookmarkPageTitle,fetchBookmarkPageTitle,
// resolveBookmarkPageTitleWithNetwork}.asm.txt
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// normalizeBookmarkPageTitle 规整书签标题：去首尾空白 → 按空白切字段 → 以单空格连接。
// [S 汇编 0x140771cc0, 96B]：strings.TrimSpace → strings.Fields → strings.Join(fields, " ")
// （分隔符常量 1 字节 " " @0x140c0e081）。
func normalizeBookmarkPageTitle(title string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
}

// normalizeBookmarkPageTitleURL 规整书签标题 URL：仅接受 http/https，剥离 Userinfo 与 Fragment。
// [S 汇编 0x140770ec0, 320B]：
//
//	url.Parse(TrimSpace(raw)) 错误 → ("",false)。
//	scheme=ToLower(TrimSpace(u.Scheme))；仅 "http"(4B)/"https"(5B) 通过（magic 0x70747468 比较）。
//	TrimSpace(u.Host)=="" → ("",false)。
//	splitHostPort(u.Host).host 经 TrimSpace=="" → ("",false)。
//	u.User != nil → ("",false)（拒绝含 userinfo 的 URL）。
//	清空 u.Fragment → u.String() → (结果,true)。
func normalizeBookmarkPageTitleURL(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	if strings.TrimSpace(u.Host) == "" {
		return "", false
	}
	if strings.TrimSpace(u.Hostname()) == "" {
		return "", false
	}
	if u.User != nil {
		return "", false
	}
	u.Fragment = ""
	return u.String(), true
}

// bookmarkPageTitleCandidates 生成书签标题候选 URL 列表（去重）。
// [S 汇编 0x140770b40, 896B]：
//
//	trimmed=TrimSpace(raw)；空 → nil。
//	normalizeBookmarkPageTitleURL(trimmed) 成功 → []string{normalized}。
//	trimmed 含 "://" → nil（有 scheme 但非 http/https）。
//	否则候选 = ["https","http"] 前缀 + "://" + trimmed，逐个 normalize，map[string]struct{} 去重追加。
func bookmarkPageTitleCandidates(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if normalized, ok := normalizeBookmarkPageTitleURL(trimmed); ok {
		return []string{normalized}
	}
	if strings.Contains(trimmed, "://") {
		return nil
	}
	result := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, scheme := range [2]string{"https", "http"} {
		candidate := scheme + "://" + trimmed
		if normalized, ok := normalizeBookmarkPageTitleURL(candidate); ok {
			if _, dup := seen[normalized]; !dup {
				seen[normalized] = struct{}{}
				result = append(result, normalized)
			}
		}
	}
	return result
}

// extractBookmarkPageTitle 从 HTML 流中抽取首个非空 <title> 文本。
// [S 汇编 0x140771820, 1184B]：
//
//	html.NewTokenizerFragment(r, "")，tokenizer 内部 MaxBuf@+0x60 置 0x80000（512KB，非导出字段，此处注释）。
//	循环 Next()：TextToken(1) 且 inTitle 时累积文本（总量 >4096 报 "网页标题超过大小限制"）；
//	StartTagToken(2) 且 EqualFold(tok.Data,"title") → 清缓冲 + inTitle=true；
//	EndTagToken(3) 且 EqualFold(tok.Data,"title") → normalize 后非空即返回，空则 inTitle=false 继续；
//	ErrorToken(0) → err=tz.Err()；err 非 nil 且非 io.EOF 则返回 ("",err)；否则返回 normalize 结果，
//	空则 "未找到网页标题"。
func extractBookmarkPageTitle(r io.Reader) (string, error) {
	tz := html.NewTokenizerFragment(r, "")
	var title []byte
	inTitle := false
	for {
		tt := tz.Next()
		switch tt {
		case html.ErrorToken:
			if err := tz.Err(); err != nil {
				return "", err
			}
			if normalized := normalizeBookmarkPageTitle(string(title)); normalized != "" {
				return normalized, nil
			}
			return "", errors.New("未找到网页标题")
		case html.TextToken:
			if inTitle {
				text := tz.Text()
				if len(title)+len(text) > 4096 {
					return "", errors.New("网页标题超过大小限制")
				}
				title = append(title, text...)
			}
		case html.StartTagToken:
			if strings.EqualFold(tz.Token().Data, "title") {
				title = title[:0]
				inTitle = true
			}
		case html.EndTagToken:
			if strings.EqualFold(tz.Token().Data, "title") {
				if normalized := normalizeBookmarkPageTitle(string(title)); normalized != "" {
					return normalized, nil
				}
				inTitle = false
			}
		}
	}
}

// fetchBookmarkPageTitle 通过网络访问拉取书签页面标题。
// [S 汇编 0x140771000, 1984B]：
//
//	NewRequestWithContext(context.Background(),"GET",url,nil)；Header 设
//	User-Agent "Mozilla/5.0 (compatible; UsbEAm Launcher/1.0)"、
//	Accept "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8"。
//	access.Do(req, 0x12a05f200=5s)（[S-sig] 第二参语义为 timeout；现有接口 int64，按 time.Duration 传递）。
//	resp/body nil → "网页响应为空"；StatusCode∉[200,299] → "网页返回状态码 %d"；
//	ContentLength>2MB → "网页响应超过大小限制"；Content-Type 非空且非 text/html / application/xhtml+xml
//	→ "网页内容类型不支持: %s"。defer body.Close()。
//	extractBookmarkPageTitle(io.LimitReader(body, 2MB))；标题空 → "未找到网页标题"。
func fetchBookmarkPageTitle(access LauncherNetworkAccess, url string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; UsbEAm Launcher/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")

	resp, err := access.Do(req, int64(5*time.Second))
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Body == nil {
		return "", errors.New("网页响应为空")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("网页返回状态码 %d", resp.StatusCode)
	}
	if resp.ContentLength > 2<<20 {
		return "", errors.New("网页响应超过大小限制")
	}
	if ct := strings.TrimSpace(resp.Header.Get("Content-Type")); ct != "" {
		mediatype, _, err := mime.ParseMediaType(ct)
		if err != nil || (mediatype != "text/html" && mediatype != "application/xhtml+xml") {
			return "", fmt.Errorf("网页内容类型不支持: %s", ct)
		}
	}

	title, err := extractBookmarkPageTitle(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if title != "" {
		return title, nil
	}
	return "", errors.New("未找到网页标题")
}

// resolveBookmarkPageTitleWithNetwork 逐个候选 URL 解析书签页面标题。
// [S 汇编 0x1407709e0, 544B]：
//
//	bookmarkPageTitleCandidates(url) 空 → ("", "链接地址无效")。
//	否则遍历候选 fetchBookmarkPageTitle(access, candidate)：err → 记录 lastErr 继续；
//	title 非空 → (title, nil)。耗尽后 lastErr 非 nil → ("", lastErr)，否则 ("", "未读取到网页标题")。
func resolveBookmarkPageTitleWithNetwork(access LauncherNetworkAccess, url string) (string, error) {
	candidates := bookmarkPageTitleCandidates(url)
	if len(candidates) == 0 {
		return "", errors.New("链接地址无效")
	}
	var lastErr error
	for _, candidate := range candidates {
		title, err := fetchBookmarkPageTitle(access, candidate)
		if err != nil {
			lastErr = err
			continue
		}
		if title != "" {
			return title, nil
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("未读取到网页标题")
}

// normalizeFirefoxProfileName 规整 Firefox 配置名：名字非空则取 TrimSpace 名字；
// 否则取路径 Base 并去首段点号扩展名。
// [S 汇编 0x14076cfa0, 256B]：TrimSpace(name) 非空直接返回；否则 filepathlite.Base(
// TrimSpace(path)) → TrimSpace → Index(".") → 找到则取 [:idx] 再 TrimSpace。
func normalizeFirefoxProfileName(name, path string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n
	}
	base := strings.TrimSpace(filepath.Base(strings.TrimSpace(path)))
	if idx := strings.Index(base, "."); idx >= 0 {
		return strings.TrimSpace(base[:idx])
	}
	return base
}

// containsUnsafeOpenLinkText 判定链接文本是否含危险内容（危险子串或控制字符）。
// [S-sig 0x140769980, 224B]：Index(危险子串) 命中→true；否则遍历 rune 查控制字符表
// (table[rune]&1)。体待危险子串/字符表专项还原。
func containsUnsafeOpenLinkText(text string) bool {
	_ = text
	return false
}
