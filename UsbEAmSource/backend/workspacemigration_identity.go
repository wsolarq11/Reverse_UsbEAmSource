package main

import (
	"fmt"
	"os"
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

// validateWorkspaceMigrationExistingChain 校验路径向上祖先链不含 reparse point。
// [S] ASM 0x1409f42a0: 循环 Clean(path) → os.Lstat err 直接返回；
// workspaceMigrationPathIsReparse(cleaned) err 直接返回、true 则 Errorf
// "迁移路径不能经过符号链接、目录联接或 reparse point: %s"(cleaned)；
// parent=Dir(cleaned)，parent==cleaned 则 return nil，否则 path=parent 续环。
func validateWorkspaceMigrationExistingChain(path string) error {
	for {
		cleaned := filepath.Clean(path)
		if _, err := os.Lstat(cleaned); err != nil {
			return err
		}
		isReparse, err := workspaceMigrationPathIsReparse(cleaned)
		if err != nil {
			return err
		}
		if isReparse {
			return fmt.Errorf("迁移路径不能经过符号链接、目录联接或 reparse point: %s", cleaned)
		}
		parent := filepath.Dir(cleaned)
		if parent == cleaned {
			return nil
		}
		path = parent
	}
}
