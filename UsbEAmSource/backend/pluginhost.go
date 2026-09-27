// AUTO-RECONSTRUCTED — DOMAIN: plugin host
// 研究用途
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed plugin_runtime_snippet.html
var pluginRuntimeSnippetTemplate string

// pluginRuntimeSnippetReplacer 把 base href 里的 & < > " ' 转义为 HTML 实体，
// 避免注入 <base href="..."> 时破坏属性边界。
var pluginRuntimeSnippetReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
	"'", "&#39;",
)

// pluginDiscoveryLock 保护插件发现（discoverPlugins / ServeHTTP 共享的读锁）。
var pluginDiscoveryLock sync.RWMutex

// ServeHTTP 实现 http.Handler，把 /plugin-host/ 下的请求路由到插件资源。
// [S 汇编 0x140921b80, 1824B] 实证：全局 RWMutex RLock + defer RUnlock；
// 非 GET/HEAD → 设 Allow 头 + http.Error 405；parsePluginRequestPath 失败 → 400；
// resolvePluginManifestUnlocked 失败 → 404；resolvePluginAssetPath 失败 → 400；
// HTML 资源走 servePluginHTML（失败 500），其余走 Cache-Control/X-Content-Type-Options 头 + ServeFile。
func (h *PluginHostService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pluginDiscoveryLock.RLock()
	defer pluginDiscoveryLock.RUnlock()

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pluginID, assetPath, err := parsePluginRequestPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	manifest, err := h.resolvePluginManifestUnlocked(pluginID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	assetPath, err = resolvePluginAssetPath(assetPath, manifest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if isHTMLPluginAsset(assetPath) {
		if err := servePluginHTML(w, assetPath, manifest); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, assetPath)
}

// resolvePluginManifestUnlocked 无锁按 ID 解析插件清单。
// [S 汇编 0x140922300, 608B] 实证：TrimSpace → validatePluginID 早返 →
// bootstrap.pluginDirectories → discoverPluginsUnlocked → EqualFold(ID,id) 匹配即返回；
// 未匹配返回 fmt.Errorf("未找到插件: %s", id)。
func (h *PluginHostService) resolvePluginManifestUnlocked(id string) (PluginManifest, error) {
	id = strings.TrimSpace(id)
	if err := validatePluginID(id); err != nil {
		return PluginManifest{}, err
	}
	for _, m := range discoverPluginsUnlocked(h.bootstrap.pluginDirectories()) {
		if strings.EqualFold(m.ID, id) {
			return m, nil
		}
	}
	return PluginManifest{}, fmt.Errorf("未找到插件: %s", id)
}

// parsePluginRequestPath 把请求路径拆成插件 ID 与插件内资源路径。
// [S 汇编 0x1409225c0, 528B] 实证：TrimSpace → 去一个前导 '/' 再去一个前导 '\' →
// TrimPrefix "plugin-host/" → 空则 errors.New("插件路径不能为空") → Split "/" →
// pluginID=TrimSpace(parts[0]) → validatePluginID 早返；单段返回 (id,"",nil)，
// 多段 assetPath=Join(parts[1:],"/")。
func parsePluginRequestPath(path string) (pluginID, assetPath string, err error) {
	s := strings.TrimSpace(path)
	if len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	if len(s) > 0 && s[0] == '\\' {
		s = s[1:]
	}
	s = strings.TrimPrefix(s, "plugin-host/")
	if s == "" {
		return "", "", errors.New("插件路径不能为空")
	}
	parts := strings.Split(s, "/")
	pluginID = strings.TrimSpace(parts[0])
	if err = validatePluginID(pluginID); err != nil {
		return "", "", err
	}
	if len(parts) == 1 {
		return pluginID, "", nil
	}
	assetPath = strings.Join(parts[1:], "/")
	return pluginID, assetPath, nil
}

// resolvePluginAssetPath 把资源路径解析为插件源目录内的可信绝对路径。
// [S 汇编 0x140922800, 480B] 实证：TrimSpace → 空或 "/" 用 manifest.Entry 兜底 →
// 反斜杠转正斜杠 → TrimPrefix "/" → resolvePluginPathSecurely(SourceDir, s, true) →
// Stat 非目录直接返回；目录则 Rel 取相对、Join "index.html" 再
// resolvePluginPathSecurely(SourceDir, indexPath, false)。
func resolvePluginAssetPath(assetPath string, manifest PluginManifest) (string, error) {
	s := strings.TrimSpace(assetPath)
	if s == "" || s == "/" {
		s = strings.TrimSpace(manifest.Entry)
	}
	s = strings.Replace(s, "\\", "/", -1)
	s = strings.TrimPrefix(s, "/")
	resolved, err := resolvePluginPathSecurely(manifest.SourceDir, s, true)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return resolved, nil
	}
	rel, err := filepath.Rel(manifest.SourceDir, resolved)
	if err != nil {
		return "", err
	}
	return resolvePluginPathSecurely(manifest.SourceDir, filepath.Join(rel, "index.html"), false)
}

// isHTMLPluginAsset 判断资源扩展名是否为 .htm 或 .html（大小写不敏感）。
// [S 汇编 0x1409229e0, 160B] 实证：TrimSpace → 反扫最后一个 '.'（遇 '\' 或 '/' 停）→
// ToLower 后缀 → 返回 ext=='.htm'||ext=='.html'。
func isHTMLPluginAsset(path string) bool {
	s := strings.TrimSpace(path)
	ext := ""
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			ext = s[i:]
			break
		}
	}
	ext = strings.ToLower(ext)
	return ext == ".htm" || ext == ".html"
}

// servePluginHTML 读取 HTML 并注入运行时片段后写回响应。
// [S 汇编 0x140922aa0, 1184B] 实证：readPluginFileBounded(path,4<<20) → injectPluginRuntime →
// 设 Content-Type/Cache-Control/Referrer-Policy/X-Content-Type-Options 头 → w.Write(html)。
func servePluginHTML(w http.ResponseWriter, path string, manifest PluginManifest) error {
	data, err := readPluginFileBounded(path, 4<<20)
	if err != nil {
		return err
	}
	html, err := injectPluginRuntime(data, path, manifest)
	if err != nil {
		return err
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Content-Type-Options", "nosniff")
	_, err = w.Write(html)
	return err
}

// injectPluginRuntime 把插件运行时元数据注入 HTML（拼接到 <head>/<body> 标签后）。
// [S 汇编 0x140922f40, 2288B] 实证：APIVersion 空则 "1.0"；深拷贝 Permissions；
// resolvePluginLanguageMessages(manifest,"en-US") → pluginRuntimeMeta JSON 序列化 →
// resolvePluginBaseHref → buildPluginRuntimeSnippet → 小写化 HTML 找 "<head"/"<body" 的 ">"，
// 命中则 bytes.Join 三段（前缀+snippet+后缀），否则 bytes.Join 两段（snippet+html）。
func injectPluginRuntime(html []byte, path string, manifest PluginManifest) ([]byte, error) {
	apiVersion := strings.TrimSpace(manifest.APIVersion)
	if apiVersion == "" {
		apiVersion = "1.0"
	}
	permissions := append([]string(nil), manifest.Permissions...)
	meta := pluginRuntimeMeta{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Description: manifest.Description,
		Version:     manifest.Version,
		Category:    manifest.Category,
		APIVersion:  apiVersion,
		UI:          manifest.UI,
		Permissions: permissions,
		Messages:    resolvePluginLanguageMessages(manifest, "en-US"),
	}
	pluginJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	baseHref, err := resolvePluginBaseHref(path, manifest)
	if err != nil {
		return nil, err
	}
	snippet := []byte(buildPluginRuntimeSnippet(baseHref, string(pluginJSON)))

	lower := strings.ToLower(string(html))
	if idx := strings.Index(lower, "<head"); idx >= 0 {
		if gt := strings.Index(lower[idx:], ">"); gt >= 0 {
			insert := idx + gt + 1
			return bytes.Join([][]byte{html[:insert], snippet, html[insert:]}, nil), nil
		}
	}
	if idx := strings.Index(lower, "<body"); idx >= 0 {
		if gt := strings.Index(lower[idx:], ">"); gt >= 0 {
			insert := idx + gt + 1
			return bytes.Join([][]byte{html[:insert], snippet, html[insert:]}, nil), nil
		}
	}
	return bytes.Join([][]byte{snippet, html}, nil), nil
}

// resolvePluginBaseHref 计算注入到 HTML 的 <base> 相对基准路径。
// [S 汇编 0x140923840, 448B] 实证：filepath.Clean(TrimSpace(SourceDir)) → Rel(srcDir,path) →
// Dir(rel)；目录为 "." 或 "\" 时返回 "/plugin-host/{id}/"，否则
// "/plugin-host/{id}/{dir 反斜杠转正斜杠}/"。
func resolvePluginBaseHref(path string, manifest PluginManifest) (string, error) {
	srcDir := filepath.Clean(strings.TrimSpace(manifest.SourceDir))
	rel, err := filepath.Rel(srcDir, path)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(rel)
	id := urlPathEscape(manifest.ID)
	if dir == "." || dir == `\` {
		return fmt.Sprintf("/plugin-host/%s/", id), nil
	}
	return fmt.Sprintf("/plugin-host/%s/%s/", id, strings.Replace(dir, `\`, "/", -1)), nil
}

// buildPluginRuntimeSnippet 用模板生成注入到插件 HTML 的运行时脚本片段。
// [S 汇编 0x140923a00, 224B] 实证：模板为 40370 字节 rodata 字符串（2 个 %s 占位 + 26 个 %%），
// fmt.Sprintf(template, Replacer.Replace(baseHref), pluginJSON)；模板以 go:embed 内嵌
// plugin_runtime_snippet.html，与原始 rodata 字节一致。
func buildPluginRuntimeSnippet(baseHref, pluginJSON string) string {
	return fmt.Sprintf(pluginRuntimeSnippetTemplate,
		pluginRuntimeSnippetReplacer.Replace(baseHref),
		pluginJSON)
}

// urlPathEscape 去空白后按路径段规则做 URL 转义。
// [S 汇编 0x140923ae0, 48B] 实证：strings.TrimSpace → net/url.escape(mode=2=encodePathSegment)，
// 即 url.PathEscape(strings.TrimSpace(s))。
func urlPathEscape(s string) string {
	return url.PathEscape(strings.TrimSpace(s))
}
