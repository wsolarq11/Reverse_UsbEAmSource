package main

import (
	"errors"
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

// inspectWorkspaceMigrationPath 检查迁移路径，返回现存祖先身份与缺失段拼接后的最终路径。
// [S] ASM 0x1409f3c00: Abs→Clean；循环 Lstat(current)：成功→break 得 info；
// ErrNotExist→Dir(current)==current 则 Errorf "找不到迁移路径的现有父目录: %s"(cleaned)，
// 否则 append(Base(current)) 后 current=Dir(current)；其他 err 直接返回；
// validateWorkspaceMigrationExistingChain(current)；
// workspaceMigrationIdentityForExisting(current, info)；
// 从尾到头 Join 缺失段回 identity.canonical 后 Clean；exists=(缺失段数==0)；
// identity 仅 exists 时填 identity，ancestorIdentity 恒填 identity。
func inspectWorkspaceMigrationPath(path string) (workspacePathInspection, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return workspacePathInspection{}, err
	}
	cleaned := filepath.Clean(abs)

	var missing []string
	current := cleaned
	var info os.FileInfo
	for {
		fi, err := os.Lstat(current)
		if err == nil {
			info = fi
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return workspacePathInspection{}, err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return workspacePathInspection{}, fmt.Errorf("找不到迁移路径的现有父目录: %s", cleaned)
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}

	if err := validateWorkspaceMigrationExistingChain(current); err != nil {
		return workspacePathInspection{}, err
	}

	identity, err := workspaceMigrationIdentityForExisting(current, info)
	if err != nil {
		return workspacePathInspection{}, err
	}

	fullPath := identity.canonical
	for i := len(missing) - 1; i >= 0; i-- {
		fullPath = filepath.Join(fullPath, missing[i])
	}
	fullPath = filepath.Clean(fullPath)

	inspection := workspacePathInspection{
		canonical:        fullPath,
		exists:           len(missing) == 0,
		ancestorIdentity: identity,
	}
	if len(missing) == 0 {
		inspection.identity = identity
	}
	return inspection, nil
}

// sameWorkspacePathInspection 判断两个检查结果是否指向同一目标。
// [S] ASM 0x1409f43e0: exists 不等→false；exists 时比较 identity 的
// valid/volume/file；否则比较 ancestorIdentity 的 valid/volume/file
// 且 workspacePathsEqual(canonical)。
func sameWorkspacePathInspection(a, b workspacePathInspection) bool {
	if a.exists != b.exists {
		return false
	}
	if a.exists {
		if !a.identity.valid || !b.identity.valid {
			return false
		}
		if a.identity.volume != b.identity.volume {
			return false
		}
		return a.identity.file == b.identity.file
	}
	if !a.ancestorIdentity.valid || !b.ancestorIdentity.valid {
		return false
	}
	if a.ancestorIdentity.volume != b.ancestorIdentity.volume {
		return false
	}
	if a.ancestorIdentity.file != b.ancestorIdentity.file {
		return false
	}
	return workspacePathsEqual(a.canonical, b.canonical)
}
