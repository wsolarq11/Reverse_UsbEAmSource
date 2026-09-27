package main

import (
	"path/filepath"
	"strings"
)

// workspacePathsEqual 判断两个工作区路径是否等价（规范化后精确或忽略大小写相等）。
// [S 汇编 0x1409f4560]：Clean(a)==Clean(b)（runtime.memequal）→ true；
// 否则 strings.EqualFold(Clean(a), Clean(b))（大小写不敏感）作为结果。
func workspacePathsEqual(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}
