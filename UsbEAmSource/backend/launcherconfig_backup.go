// ---- launcherConfig 备份辅助函数 ----
// [S 汇编 0x140877aa0/0x140877b20/0x140877ca0] — 全量实证

package main

import (
	"path/filepath"
	"time"
)

// launcherConfigBackupDir 返回当前配置的备份目录：filepath.Dir(path)/backup。
// [S 汇编 0x140877aa0, 128B]：filepathlite.Dir → filepath.Join("backup")。
func launcherConfigBackupDir(path string) string {
	return filepath.Join(filepath.Dir(path), "backup")
}

// launcherConfigBackupPath 返回备份路径：backupDir/<名>.<日期>.<原扩展>。
// [S 汇编 0x140877b20, 384B]：
// launcherConfigBackupFilenameParts → launcherConfigBackupDir → time.Format →
// concatstring3(name_prefix, date, ext) → filepath.Join(dir, result)。
func launcherConfigBackupPath(path string, now time.Time) string {
	prefix, ext := launcherConfigBackupFilenameParts(path)
	dir := launcherConfigBackupDir(path)
	date := now.Format("2006-01-02")
	return filepath.Join(dir, prefix+date+ext)
}

// launcherConfigBackupFilenameParts 拆分配置文件名：去掉扩展名的基底名+'.' 与扩展名(含点)。
// [S 汇编 0x140877ca0, 320B]：
// filepathlite.Base → 逆向扫描末位 '.'（遇 '/' 或 '\\' 中断）→ 返回 (名+".", 含点扩展)；
// 无扩展时返回 (名+".", "")。
func launcherConfigBackupFilenameParts(path string) (prefix, ext string) {
	base := filepath.Base(path)
	dotIdx := -1
	for i := len(base) - 1; i >= 0; i-- {
		switch base[i] {
		case '.':
			dotIdx = i
			goto found
		case '/', '\\':
			goto found
		}
	}
found:
	if dotIdx >= 0 {
		return base[:dotIdx] + ".", base[dotIdx:]
	}
	return base + ".", ""
}
