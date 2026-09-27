package main

import (
	"path/filepath"
	"strings"
)

// normalizeFileSearchIgnoredDirectoryRule 归一化单条忽略目录规则。
// [S] 0x1407e0760：Trim `"` `'` → 空判 → 绝对路径规则走 path 归一化 →
// TrimRight `/` `\` → 含分隔符的裸名丢弃 → TrimSpace → `.` 丢弃。
func normalizeFileSearchIgnoredDirectoryRule(rule string) string {
	s := strings.TrimSpace(strings.Trim(rule, "\"'"))
	if s == "" {
		return ""
	}
	if isFileSearchIgnoredDirectoryPathRule(s) {
		return normalizeFileSearchIgnoredDirectoryPath(s)
	}
	s = strings.TrimRight(s, "/\\")
	if s == "" {
		return ""
	}
	if strings.IndexAny(s, "/\\") >= 0 {
		return ""
	}
	s = strings.TrimSpace(s)
	if s == "." {
		return ""
	}
	return s
}

// normalizeFileSearchIgnoredDirectoryPath 归一化绝对路径形式的忽略目录。
// [S] 0x1407e0880：Trim 引号 → 空/非绝对路径丢弃 → Clean → `.` 丢弃 →
// 卷根（normalizeVolumeRoot）直接返回根 → TrimRight 分隔符。
func normalizeFileSearchIgnoredDirectoryPath(path string) string {
	s := strings.TrimSpace(strings.Trim(path, "\"'"))
	if s == "" || !filepath.IsAbs(s) {
		return ""
	}
	s = filepath.Clean(s)
	if s == "." {
		return ""
	}
	if root := normalizeVolumeRoot(s); root != "" && strings.EqualFold(s, root) {
		return root
	}
	return strings.TrimRight(s, "/\\")
}

// normalizeFileSearchIgnoredDirectoryMatchPath 生成忽略目录的匹配键（小写、反斜杠分隔）。
// [S] 0x1407e09a0：path 归一化 → 空判 → 分隔符 `/` 换 `\` → 小写。
func normalizeFileSearchIgnoredDirectoryMatchPath(path string) string {
	s := normalizeFileSearchIgnoredDirectoryPath(path)
	if s == "" {
		return ""
	}
	return strings.ToLower(strings.Replace(s, "/", "\\", -1))
}

// isFileSearchIgnoredDirectoryPathRule 判断规则是否为绝对路径形式。
// [S] 0x1407e0a20：Trim 引号 + TrimSpace 后是否绝对路径。
func isFileSearchIgnoredDirectoryPathRule(rule string) bool {
	return filepath.IsAbs(strings.TrimSpace(strings.Trim(rule, "\"'")))
}

// fileSearchPathWithinIgnoredDirectory 判断 path 是否位于 dir 内。
// [S] 0x1407e0a80：空判 → 相等判 → dir 末尾补 `\` → 前缀判。
func fileSearchPathWithinIgnoredDirectory(path, dir string) bool {
	if path == "" || dir == "" {
		return false
	}
	if path == dir {
		return true
	}
	if !strings.HasSuffix(dir, "\\") {
		dir += "\\"
	}
	return strings.HasPrefix(path, dir)
}

// newFileSearchIgnoreMatcher 由规则列表构建忽略目录匹配器。
// [S] 0x1407e0be0：归一化规则 → 空则返回空 nameRules → 路径规则入 pathRules，
// 目录名规则 ToLower 入 nameRules。
func newFileSearchIgnoreMatcher(rules []string) fileSearchIgnoreMatcher {
	rules = normalizeFileSearchIgnoredDirectoryRules(rules)
	if len(rules) == 0 {
		return fileSearchIgnoreMatcher{nameRules: map[string]struct{}{}}
	}
	nameRules := make(map[string]struct{})
	pathRules := make([]string, 0, len(rules))
	for _, r := range rules {
		if isFileSearchIgnoredDirectoryPathRule(r) {
			p := normalizeFileSearchIgnoredDirectoryMatchPath(r)
			if p != "" {
				pathRules = append(pathRules, p)
			}
		} else {
			nameRules[strings.ToLower(r)] = struct{}{}
		}
	}
	return fileSearchIgnoreMatcher{nameRules: nameRules, pathRules: pathRules}
}

// HasRules 是否配置了任意忽略规则。
// [S] 0x1407e0e40：nameRules 或 pathRules 任一非空。
func (m fileSearchIgnoreMatcher) HasRules() bool {
	return len(m.nameRules) > 0 || len(m.pathRules) > 0
}

// IsIgnoredDirectory 判断目录（name/path）是否被忽略。
// [S] 0x1407e0e80：目录名（TrimSpace+ToLower）命中 nameRules → true；
// 路径命中 pathRules 前缀 → true。
func (m fileSearchIgnoreMatcher) IsIgnoredDirectory(name, path string) bool {
	if len(m.nameRules) > 0 {
		if _, ok := m.nameRules[strings.ToLower(strings.TrimSpace(name))]; ok {
			return true
		}
	}
	if len(m.pathRules) > 0 {
		p := normalizeFileSearchIgnoredDirectoryMatchPath(path)
		if p != "" {
			for _, r := range m.pathRules {
				if fileSearchPathWithinIgnoredDirectory(p, r) {
					return true
				}
			}
		}
	}
	return false
}
