// AUTO-RECONSTRUCTED FUNCTIONS — DOMAIN: launcher update transaction lifecycle
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak（各函数地址见各自 [S] 标注）
//   - 反汇编语义分析：tmp_txn_semantics.md（1286 行，15 函数全解）
//   - 行号蓝图：source_funcs.txt launcherupdate_transaction.go L61-541
//
// 档位：[S] 反汇编实证（已全部语义分析对位）
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ---- transaction lifetime ----

// prepareLauncherUpdateTransaction 从更新计划创建事务：验证 plan → 创建暂存/备份目录 →
// 复制 plan.Files/plan.Deletes → 排序 → 持久化初始状态 "prepared"。
// [S 汇编 0x1408cb500]
func prepareLauncherUpdateTransaction(plan *launcherUpdatePlan) (*launcherUpdateTransaction, error) {
	if err := validateLauncherUpdatePlan(plan); err != nil {
		return nil, err
	}
	installDir := strings.TrimSpace(plan.InstallDir)
	installDir = filepath.Clean(installDir)
	if _, err := validateLauncherUpdateAbsolutePath(installDir); err != nil {
		return nil, err
	}
	// updateRoot must be inside installDir
	if !isPathInsideDirectory(plan.UpdateRoot, installDir) {
		return nil, errors.New("launcherupdate: updateRoot not inside installDir")
	}
	stagingDir := filepath.Join(installDir, ".update-staging")
	backupDir := filepath.Join(installDir, ".update-backup")

	now := time.Now().UTC().Format(time.RFC3339)

	txn := &launcherUpdateTransaction{
		SchemaVersion: 1,
		Nonce:         plan.Nonce,
		State:         "prepared",
		InstallDir:    installDir,
		StagingDir:    stagingDir,
		BackupDir:     backupDir,
		JournalFile:   filepath.Join(installDir, ".update-journal.json"),
		Executable:    plan.ExecutableName,
		HealthNonce:   plan.HealthNonce,
		UpdatedAt:     now,
	}

	// Copy Deletes ([]string → []launcherUpdateTransactionFile with Delete=true)
	for _, d := range plan.Deletes {
		path := strings.ToLower(strings.TrimSpace(d))
		txn.Files = append(txn.Files, launcherUpdateTransactionFile{
			Path:   path,
			Delete: true,
			Status: "prepared",
		})
	}
	// Copy plan.Files
	for _, f := range plan.Files {
		txn.Files = append(txn.Files, launcherUpdateTransactionFile{
			Path:         f.Path,
			Delete:       false,
			Status:       "prepared",
			ExpectedSize: f.Size,
			ExpectedHash: f.SHA256,
		})
	}
	// Sort by Path (case-insensitive)
	sort.SliceStable(txn.Files, func(i, j int) bool {
		return strings.ToLower(txn.Files[i].Path) < strings.ToLower(txn.Files[j].Path)
	})
	// Create staging dir
	if err := os.Mkdir(stagingDir, 0o700); err != nil {
		return nil, err
	}
	if err := txn.persist(); err != nil {
		os.RemoveAll(stagingDir)
		return nil, err
	}
	return txn, nil
}

// (*launcherUpdateTransaction).persist 更新 UpdatedAt → JSON 序列化 → 原子写入 journal。
// [S 汇编 0x1408cf480]：MkdirAll parentDir → 写 .tmp → Sync → Close → replaceLauncherUpdateMetadataFile。
func (txn *launcherUpdateTransaction) persist() error {
	if txn == nil {
		return errors.New("launcherupdate: nil transaction")
	}
	txn.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(txn, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(txn.JournalFile)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmpPath := txn.JournalFile + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	closeErr := f.Close()
	if closeErr != nil {
		return closeErr
	}
	return replaceLauncherUpdateMetadataFile(tmpPath, txn.JournalFile)
}

// (*launcherUpdateTransaction).validate 全链校验：txn/schema/nonces/路径/目录包含/文件数/文件路径。
// [S 汇编 0x1408cf0c0] 12 步链。
func (txn *launcherUpdateTransaction) validate() error {
	if txn == nil || txn.SchemaVersion != 1 {
		return errors.New("launcherupdate: invalid transaction")
	}
	if err := validateLauncherUpdateNonce(txn.Nonce, 32); err != nil {
		return fmt.Errorf("launcherupdate: nonce: %w", err)
	}
	if err := validateLauncherUpdateNonce(txn.HealthNonce, 32); err != nil {
		return fmt.Errorf("launcherupdate: healthNonce: %w", err)
	}
	for _, p := range []string{txn.InstallDir, txn.StagingDir, txn.BackupDir, txn.JournalFile} {
		if _, err := validateLauncherUpdateAbsolutePath(p); err != nil {
			return err
		}
	}
	if !isPathInsideDirectory(filepath.Dir(txn.StagingDir), txn.InstallDir) {
		return errors.New("launcherupdate: stagingDir not inside installDir")
	}
	if !isPathInsideDirectory(txn.BackupDir, txn.InstallDir) {
		return errors.New("launcherupdate: backupDir not inside installDir")
	}
	expectedJournal := filepath.Join(txn.InstallDir, ".update-journal.json")
	if !samePathFold(txn.JournalFile, expectedJournal) {
		return errors.New("launcherupdate: journalFile path mismatch")
	}
	if len(txn.Files) == 0 || len(txn.Files) > 0x10000 {
		return errors.New("launcherupdate: invalid file count")
	}
	for _, f := range txn.Files {
		if _, err := normalizeLauncherUpdateRelativePath(f.Path); err != nil {
			return err
		}
	}
	return nil
}

// (*launcherUpdateTransaction).Apply 执行文件替换事务。
// [S 汇编 0x1408cbf60]：检查 State == "prepared" → "applying" → persist → 循环 applyFile → "applied"。
func (txn *launcherUpdateTransaction) Apply() error {
	if err := txn.validate(); err != nil {
		return err
	}
	if txn.State != "prepared" {
		return fmt.Errorf("launcherupdate: cannot apply from state %q", txn.State)
	}
	txn.State = "applying"
	if err := txn.persist(); err != nil {
		return err
	}
	for i := range txn.Files {
		if err := txn.applyFile(i); err != nil {
			rollbackErr := txn.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("launcherupdate: apply file %d failed: %v (rollback: %v)", i, err, rollbackErr)
			}
			return err
		}
	}
	txn.State = "applied"
	return txn.persist()
}

// (*launcherUpdateTransaction).applyFile 替换或删除单个文件。
// [S 汇编 0x1408cc320]：检查 Status=="pending" → 安全验证 → Lstat → 备份 → 写入 → "replaced"/"deleted"。
func (txn *launcherUpdateTransaction) applyFile(idx int) error {
	if idx < 0 || idx >= len(txn.Files) {
		panic("launcherupdate: applyFile index out of range")
	}
	f := &txn.Files[idx]
	if f.Status != "pending" {
		return fmt.Errorf("launcherupdate: file %s status not pending", f.Path)
	}
	norm := strings.ReplaceAll(f.Path, "/", "\\")
	dstPath := filepath.Join(txn.InstallDir, norm)
	backupPath := filepath.Join(txn.BackupDir, norm)

	if !isPathInsideDirectory(dstPath, txn.InstallDir) {
		return fmt.Errorf("launcherupdate: dst %s not inside installDir", f.Path)
	}
	if !isPathInsideDirectory(backupPath, txn.BackupDir) {
		return fmt.Errorf("launcherupdate: backup %s not inside backupDir", f.Path)
	}
	if launcherUpdatePathProtected(f.Path) {
		return fmt.Errorf("launcherupdate: %s is protected", f.Path)
	}
	if err := ensureLauncherUpdateParentDirectoriesSafe(txn.InstallDir, filepath.Dir(dstPath)); err != nil {
		return err
	}
	// Check existing target
	fi, err := os.Lstat(dstPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		f.Existed = false
	} else {
		f.Existed = true
		if fi.IsDir() {
			return fmt.Errorf("launcherupdate: %s is a directory", f.Path)
		}
		if hasReparse, rerr := launcherUpdatePathHasReparsePoint(dstPath); rerr != nil || hasReparse {
			if rerr != nil {
				return rerr
			}
			return fmt.Errorf("launcherupdate: %s is a reparse point", f.Path)
		}
	}

	if f.Delete {
		return txn.applyFileDelete(f, dstPath, backupPath)
	}
	return txn.applyFileReplace(f, dstPath, backupPath)
}

// applyFileDelete 处理删除模式：备份再删。
// [S 汇编实证]：Status=backed-up→persist→Rename 备份→Status=deleted→persist。
func (txn *launcherUpdateTransaction) applyFileDelete(f *launcherUpdateTransactionFile, dstPath, backupPath string) error {
	f.Status = "backed-up"
	if err := txn.persist(); err != nil {
		return err
	}
	if _, err := os.Stat(dstPath); err == nil {
		if err := os.Rename(dstPath, backupPath); err != nil {
			return err
		}
	}
	f.Status = "deleted"
	return txn.persist()
}

// applyFileReplace 处理替换模式：验证 hash → 备份 → 原子替换。
// [S 汇编实证]：hash 校验→备份→staging temp→Rename 原子替换。
func (txn *launcherUpdateTransaction) applyFileReplace(f *launcherUpdateTransactionFile, dstPath, backupPath string) error {
	srcPath := filepath.Join(txn.StagingDir, strings.ReplaceAll(f.Path, "/", "\\"))
	if !isPathInsideDirectory(srcPath, txn.StagingDir) {
		return fmt.Errorf("launcherupdate: src %s not inside stagingDir", f.Path)
	}
	// Verify hash of source file
	h, err := hashLauncherUpdateFile(srcPath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(h, f.ExpectedHash) {
		return fmt.Errorf("launcherupdate: hash mismatch for %s", f.Path)
	}
	// Ensure target parent exists
	if f.Existed {
		if err := validateLauncherUpdateRollbackBackupFile(backupPath); err != nil {
			return err
		}
		f.Status = "backed-up"
		if err := txn.persist(); err != nil {
			return err
		}
		if err := os.Rename(dstPath, backupPath); err != nil {
			return err
		}
	}
	// Copy to staging temp then rename to target for atomic replace
	tempPath := filepath.Join(txn.StagingDir, "._replace_"+filepath.Base(f.Path))
	_ = os.Remove(tempPath)
	if err := copyLauncherUpdateFile(srcPath, tempPath); err != nil {
		return err
	}
	if err := os.Rename(tempPath, dstPath); err != nil {
		os.Remove(tempPath)
		return err
	}
	f.Status = "replaced"
	return txn.persist()
}

// (*launcherUpdateTransaction).failAfterBackup 文件错误后尝试恢复备份，返回原始 error。
// [S 汇编 0x1408ccec0]。
func (txn *launcherUpdateTransaction) failAfterBackup(f *launcherUpdateTransactionFile, dstPath, backupPath string, cause error) error {
	if f.Existed && f.Status == "backed-up" {
		if err := os.Rename(backupPath, dstPath); err == nil {
			f.Status = "removed"
			_ = txn.persist()
		}
	}
	return cause
}

// (*launcherUpdateTransaction).AwaitHealth 标记事务进入等待健康检查状态。
// [S 汇编 0x1408ccfe0]：检查 "applied" → 设为 "awaiting-health" → persist。
func (txn *launcherUpdateTransaction) AwaitHealth() error {
	if txn.State != "applied" {
		return fmt.Errorf("launcherupdate: cannot await health from %q", txn.State)
	}
	txn.State = "awaiting-health"
	return txn.persist()
}

// (*launcherUpdateTransaction).MarkHealthy 标记事务健康，清理暂存/备份目录。
// [S 汇编 0x1408cd0c0]：检查 "awaiting-health" → "applied" → persist → cleanup staging+backup。
func (txn *launcherUpdateTransaction) MarkHealthy() error {
	if txn.State != "awaiting-health" {
		return fmt.Errorf("launcherupdate: cannot mark healthy from %q", txn.State)
	}
	txn.State = "applied"
	if err := txn.persist(); err != nil {
		return err
	}
	os.RemoveAll(txn.StagingDir)
	os.RemoveAll(txn.BackupDir)
	return nil
}

// (*launcherUpdateTransaction).Rollback 尽力回滚：逆向遍历文件，从备份恢复，累计错误不中断。
// [S 汇编 0x1408cd1c0]：State→"rolling-back"→persist→for i:=len-1; i>=0; i-- → 清理→返回错误列表。
func (txn *launcherUpdateTransaction) Rollback() error {
	if err := txn.validate(); err != nil {
		return err
	}
	txn.State = "rolling-back"
	if err := txn.persist(); err != nil {
		return err
	}
	var rollbackErrs []error
	for i := len(txn.Files) - 1; i >= 0; i-- {
		f := &txn.Files[i]
		norm := strings.ReplaceAll(f.Path, "/", "\\")
		targetPath := filepath.Join(txn.InstallDir, norm)
		backupPath := filepath.Join(txn.BackupDir, norm)

		var restoreErr error
		switch f.Status {
		case "replaced", "backed-up", "backup-intent":
			if f.Existed {
				if err := validateLauncherUpdateRollbackBackupFile(backupPath); err != nil {
					rollbackErrs = append(rollbackErrs, err)
					continue
				}
				if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
					rollbackErrs = append(rollbackErrs, err)
					continue
				}
				if err := os.Rename(backupPath, targetPath); err != nil {
					rollbackErrs = append(rollbackErrs, err)
				}
			} else {
				if err := os.Remove(targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
					rollbackErrs = append(rollbackErrs, err)
				}
			}
		case "removed", "rolling-back":
			// Already processed, skip
		case "pending":
			// Not yet processed, skip
		default:
			f.Status = "rolling-back"
		}
		if restoreErr == nil {
			f.Status = "rolled_back"
		}
	}
	os.RemoveAll(txn.StagingDir)
	os.RemoveAll(txn.BackupDir)
	if len(rollbackErrs) > 0 {
		return fmt.Errorf("launcherupdate: rollback had %d errors: %v", len(rollbackErrs), rollbackErrs[0])
	}
	return txn.persist()
}

// ---- transaction persistence ----

// validateLauncherUpdateRollbackBackupFile 验证单个备份文件的完整性（非目录、非 reparse）。
// [S 汇编 0x1408ce6a0]。
func validateLauncherUpdateRollbackBackupFile(backupPath string) error {
	fi, err := os.Lstat(backupPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("launcherupdate: backup %s is a directory", backupPath)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("launcherupdate: backup %s is a symlink", backupPath)
	}
	hasReparse, err := launcherUpdatePathHasReparsePoint(backupPath)
	if err != nil {
		return err
	}
	if hasReparse {
		return fmt.Errorf("launcherupdate: backup %s has reparse point", backupPath)
	}
	return nil
}

// loadLauncherUpdateTransaction 从 journal 加载事务（含大小检查+严格解码+validate）。
// [S 汇编 0x1408ce820]。
func loadLauncherUpdateTransaction(journalFile string) (*launcherUpdateTransaction, error) {
	data, err := os.ReadFile(journalFile)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > 8<<20 {
		return nil, errors.New("launcherupdate: journal file size invalid")
	}
	var txn launcherUpdateTransaction
	if err := json.Unmarshal(data, &txn); err != nil {
		return nil, err
	}
	if err := txn.validate(); err != nil {
		return nil, err
	}
	return &txn, nil
}

// recoverLauncherUpdateTransactionBound 启动时恢复边界：加载 journal → 路径一致性 → 检查可执行文件冲突 → 恢复。
// [S 汇编 0x1408ce940]。
func recoverLauncherUpdateTransactionBound(installDir, updateRoot, executableName string) (*launcherUpdateTransaction, error) {
	journalFile := filepath.Join(installDir, ".update-journal.json")
	txn, err := loadLauncherUpdateTransaction(journalFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	// Path consistency checks
	if !samePathFold(txn.InstallDir, installDir) {
		return nil, errors.New("launcherupdate: installDir mismatch")
	}
	if !samePathFold(filepath.Dir(txn.InstallDir), updateRoot) {
		return nil, errors.New("launcherupdate: updateRoot mismatch")
	}
	if !samePathFold(txn.Executable, executableName) {
		return nil, errors.New("launcherupdate: executable mismatch")
	}
	if launcherUpdateTransactionRollbackWouldReplaceExecutable(txn) {
		return nil, errors.New("launcherupdate: rollback would replace running executable")
	}
	if err := recoverLoadedLauncherUpdateTransaction(txn); err != nil {
		return nil, err
	}
	return txn, nil
}

// launcherUpdateTransactionRollbackWouldReplaceExecutable 检查回滚是否会替换当前进程的可执行文件。
// [S 汇编 0x1408cec00]。
func launcherUpdateTransactionRollbackWouldReplaceExecutable(txn *launcherUpdateTransaction) bool {
	if txn == nil {
		return false
	}
	switch txn.State {
	case "applied", "prepared", "backed-up", "rolling-back", "awaiting-health", "backup-intent":
		// active states
	default:
		return false
	}
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	for _, f := range txn.Files {
		backupPath := filepath.Join(txn.BackupDir, strings.ReplaceAll(f.Path, "/", "\\"))
		if samePathFold(backupPath, execPath) {
			switch f.Status {
			case "removed", "replaced", "backed-up", "rolling-back", "awaiting-health", "backup-intent":
				return true
			}
		}
	}
	return false
}

// recoverLoadedLauncherUpdateTransaction 基于当前 State 执行恢复操作（仅 healthy 清理，其余回滚）。
// [S 汇编 0x1408cef40]。
func recoverLoadedLauncherUpdateTransaction(txn *launcherUpdateTransaction) error {
	switch txn.State {
	case "healthy":
		os.RemoveAll(txn.StagingDir)
		os.RemoveAll(txn.BackupDir)
		return nil
	case "prepared", "applied", "backed-up", "rolling-back", "backup-intent", "awaiting-health":
		return txn.Rollback()
	default:
		return fmt.Errorf("launcherupdate: unrecoverable state %q", txn.State)
	}
}

// ensureLauncherUpdateParentDirectoriesSafe 确保父目录链安全（逐段 Lstat→Mkdir→检查 reparse）。
// [S 汇编 0x1408cf720]。
func ensureLauncherUpdateParentDirectoriesSafe(installDir, dirPath string) error {
	if !isPathInsideDirectory(dirPath, installDir) {
		return errors.New("launcherupdate: dirPath not inside installDir")
	}
	rel, err := filepath.Rel(installDir, dirPath)
	if err != nil {
		return err
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	var accumulated string
	for _, part := range parts {
		if part == "." {
			continue
		}
		accumulated = filepath.Join(accumulated, part)
		parentPath := filepath.Join(installDir, accumulated)
		fi, err := os.Lstat(parentPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := os.Mkdir(parentPath, 0o755); err != nil {
					return err
				}
				continue
			}
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("launcherupdate: %s is not a directory", parentPath)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("launcherupdate: %s is a symlink", parentPath)
		}
		hasReparse, err := launcherUpdatePathHasReparsePoint(parentPath)
		if err != nil {
			return err
		}
		if hasReparse {
			return fmt.Errorf("launcherupdate: %s has reparse point", parentPath)
		}
	}
	return nil
}

// ---- helpers ----

// copyLauncherUpdateFile 拷贝文件（含内容与元数据）。
// [S-sig] 标准拷贝辅助（[R]），用于 staging temp 准备。
func copyLauncherUpdateFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()
	if _, err := dstF.ReadFrom(srcF); err != nil {
		return err
	}
	return dstF.Sync()
}
