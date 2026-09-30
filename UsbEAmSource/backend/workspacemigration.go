// AUTO-RECONSTRUCTED — DOMAIN: workspace data migration（工作区数据迁移域）
// 研究用途。汇编直译（capstone 符号标注），可编译 + 功能一致（非字节级同哈希）。
//
// 覆盖函数（VA 均来自 docs/goresym/pipeline/tmp/*.asm.txt）：
//   - workspaceDataMaintenanceGate.begin  0x1409ef000（gate_begin_end.asm.txt）
//   - begin.func2                         0x1409ef180 / begin.func2.1 0x1409ef1e0
//   - workspaceDataMaintenanceGate.enter  0x1409ef2a0（gateEnterExit.asm.txt）
//   - enter.func2                         0x1409ef420 / enter.func2.1 0x1409ef480
//   - waitWorkspaceMigrationDrain         0x1409ef500（waitWorkspaceMigrationDrain.asm.txt）
//   - beginWorkspaceDataOperation         0x1409ef7c0
//   - beginWorkspaceMigrationMaintenance  0x1409ef840（beginWorkspaceMigrationMaintenance.asm.txt）
//   - validateWorkspaceMigrationRoots     0x1409efac0（validateWorkspaceMigrationRoots.asm.txt）
//   - workspacePathContains               0x1409f0160（workspacePathContains.asm.txt）
//   - stageWorkspaceDataMigration         0x1409f0240（stageWorkspaceDataMigration.asm.txt）
//   - copyWorkspaceDataWithManifest.func1 0x1409f1400
//   - verifyWorkspaceDataManifestWithPolicy 0x1409f1ba0（verifyWorkspaceDataManifestWithPolicy.asm.txt）
//   - verify.func1                        0x1409f1da0
//   - stagedWorkspaceData.Commit          0x1409f2460（stagedWorkspaceData.Commit.asm.txt）
//   - stagedWorkspaceData.Rollback        0x1409f3100（stagedWorkspaceData.Rollback.asm.txt）
//   - removeOwnedWorkspaceMigrationDirectory 0x1409f34c0（removeOwnedWorkspaceMigrationDirectory.asm.txt）
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---- 包级全局（对应 asm 中 lea/mov [rip+...] 加载的全局 error iface / noop funcval） ----

// workspaceMaintenanceNoop 是 nil receiver 路径返回的空操作退出句柄。
// 对应 asm 0x1409ef13d（begin）与 0x1409ef3f5（enter）的包级 noop funcval。
var workspaceMaintenanceNoop = func() {}

// 迁移域共享的包级 error 接口（asm 从 .data 全局加载 itab+data，非每次 newobject）。
var (
	// 0x141BC3E00 begin/enter 维护冲突："数据目录迁移正在进行"（30B，@0x140c6ec2b）
	errWorkspaceMigrationInProgress = newExtractError("数据目录迁移正在进行")
	// 0x141BC3E10 beginWorkspaceMigrationMaintenance app!=nil：
	// "WebView2 数据目录正在使用中，请退出应用后执行离线迁移"（75B，@0x140c91924）
	errWorkspaceAppActive = newExtractError("WebView2 数据目录正在使用中，请退出应用后执行离线迁移")
	// 0x141BC4050 waitWorkspaceMigrationDrain 超时："等待旧数据目录写入排空超时"（39B，@0x140c7db1a）
	errWorkspaceDrainTimeout = newExtractError("等待旧数据目录写入排空超时")
)

// ---- workspaceDataMaintenanceGate 门闸（begin / enter，均经 once.Do 返回退出闭包） ----

// begin 开启一次数据操作：维护中拒绝，active 自增并（首次）创建 idle 通道。
// [S] 汇编 0x1409ef000：nil→(noop,nil)；maintenance→(nil,全局error)；
// active==0→g.idle=make(chan)；active++；返回 once.Do 闭包（func2 0x1409ef180）。
func (g *workspaceDataMaintenanceGate) begin() (func(), error) {
	if g == nil {
		return workspaceMaintenanceNoop, nil
	}
	g.mu.Lock()
	if g.maintenance {
		g.mu.Unlock()
		return nil, errWorkspaceMigrationInProgress
	}
	if g.active == 0 {
		g.idle = make(chan struct{})
	}
	g.active++
	g.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			// begin.func2.1 0x1409ef1e0：active>0 则递减；归零且 idle 非空则 close 并清空。
			g.mu.Lock()
			if g.active > 0 {
				g.active--
			}
			if g.active == 0 && g.idle != nil {
				close(g.idle)
				g.idle = nil
			}
			g.mu.Unlock()
		})
	}, nil
}

// enter 进入维护模式：返回排空通道与退出句柄。
// [S] 汇编 0x1409ef2a0：nil→(closed chan,noop,nil)；maintenance→(nil,nil,全局error)；
// maintenance=true；idleChan = active==0 ? make+close : g.idle（不回写 g.idle）；
// 返回 once.Do 闭包（func2 0x1409ef420 → func2.1 0x1409ef480 置 maintenance=false）。
func (g *workspaceDataMaintenanceGate) enter() (chan struct{}, func(), error) {
	if g == nil {
		ch := make(chan struct{})
		close(ch)
		return ch, workspaceMaintenanceNoop, nil
	}
	g.mu.Lock()
	if g.maintenance {
		g.mu.Unlock()
		return nil, nil, errWorkspaceMigrationInProgress
	}
	g.maintenance = true
	var idleChan chan struct{}
	if g.active == 0 {
		idleChan = make(chan struct{})
		close(idleChan)
	} else {
		idleChan = g.idle
	}
	g.mu.Unlock()

	var once sync.Once
	return idleChan, func() {
		once.Do(func() {
			g.mu.Lock()
			g.maintenance = false
			g.mu.Unlock()
		})
	}, nil
}

// waitWorkspaceMigrationDrain 等待旧数据目录写入排空或 ctx 超时。
// [S] 汇编 0x1409ef500：idle==nil→nil；select { ctx.Done→errors.Join(超时error, ctx.Err())；idle→nil }。
func waitWorkspaceMigrationDrain(ctx context.Context, idle <-chan struct{}) error {
	if idle == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return errors.Join(errWorkspaceDrainTimeout, ctx.Err())
	case <-idle:
		return nil
	}
}

// beginWorkspaceDataOperation 开启一次工作区数据操作（透明透传 begin）。
// [S] 汇编 0x1409ef7c0：bs==nil→(nil,newExtractError("工作区服务不可用"))；
// 否则 (&bs.workspaceDataMaintenance).begin() 原样透传 (func(), error)。
func (bs *BootstrapService) beginWorkspaceDataOperation() (func(), error) {
	if bs == nil {
		return nil, newExtractError("工作区服务不可用")
	}
	return (&bs.workspaceDataMaintenance).begin()
}

// beginWorkspaceMigrationMaintenance 进入离线迁移维护模式（抑制并发数据操作）。
// [S] 汇编 0x1409ef840（0 参）：bs==nil→(nil,newExtractError("工作区服务不可用"))；
// lock(+0x540)→bs.app(+0x188)!=nil→(nil,errWorkspaceAppActive)；
// (&bs.workspaceDataMaintenance).enter()→err 透传；
// context.WithTimeout(Background,15s)→defer cancel()→waitWorkspaceMigrationDrain→
// 超时则 exitFn() 后返回 err；成功返回 exitFn。
func (bs *BootstrapService) beginWorkspaceMigrationMaintenance() (func(), error) {
	if bs == nil {
		return nil, newExtractError("工作区服务不可用")
	}
	bs.lock.Lock()
	app := bs.app
	bs.lock.Unlock()
	if app != nil {
		return nil, errWorkspaceAppActive
	}
	idleChan, exitFn, err := (&bs.workspaceDataMaintenance).enter()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := waitWorkspaceMigrationDrain(ctx, idleChan); err != nil {
		exitFn()
		return nil, err
	}
	return exitFn, nil
}

// validateWorkspaceMigrationRoots 校验源/目标数据目录根并返回规范化绝对路径。
// [S] 汇编 0x1409efac0：TrimSpace 空检查→Abs→Clean→workspacePathsEqual→
// workspacePathContains 双向→inspect 源/目标→最终路径重叠三向→同一文件对象。
func validateWorkspaceMigrationRoots(source, target string) (string, string, error) {
	if strings.TrimSpace(source) == "" {
		return "", "", newExtractError("源数据目录不能为空")
	}
	if strings.TrimSpace(target) == "" {
		return "", "", newExtractError("目标数据目录不能为空")
	}
	absSource, err := filepath.Abs(strings.TrimSpace(source))
	if err != nil {
		return "", "", err
	}
	absTarget, err := filepath.Abs(strings.TrimSpace(target))
	if err != nil {
		return "", "", err
	}
	cleanSource := filepath.Clean(absSource)
	cleanTarget := filepath.Clean(absTarget)

	if workspacePathsEqual(cleanSource, cleanTarget) {
		return "", "", newExtractError("源数据目录与目标数据目录不能相同")
	}
	if workspacePathContains(cleanSource, cleanTarget) || workspacePathContains(cleanTarget, cleanSource) {
		return "", "", newExtractError("源数据目录与目标数据目录不能互为父子目录")
	}

	sourceInspection, err := inspectWorkspaceMigrationPath(cleanSource)
	if err != nil {
		return "", "", err
	}
	if !sourceInspection.exists {
		return "", "", newExtractError("源数据目录不存在")
	}
	targetInspection, err := inspectWorkspaceMigrationPath(cleanTarget)
	if err != nil {
		return "", "", err
	}
	if workspacePathsEqual(sourceInspection.canonical, targetInspection.canonical) ||
		workspacePathContains(sourceInspection.canonical, targetInspection.canonical) ||
		workspacePathContains(targetInspection.canonical, sourceInspection.canonical) {
		return "", "", newExtractError("源数据目录与目标数据目录的最终路径重叠")
	}
	if targetInspection.exists &&
		sourceInspection.identity.valid && targetInspection.identity.valid &&
		sourceInspection.identity.volume == targetInspection.identity.volume &&
		sourceInspection.identity.file == targetInspection.identity.file {
		return "", "", newExtractError("源数据目录与目标数据目录指向同一文件对象")
	}
	return cleanSource, cleanTarget, nil
}

// workspacePathContains 判断 b 是否在 a 目录树内部（严格子路径）。
// [S] 汇编 0x1409f0160：Rel err→false；"."→false；".."→false；
// 前缀 "..\\"→false；否则 !IsAbs(rel)。
func workspacePathContains(a, b string) bool {
	rel, err := filepath.Rel(a, b)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}
	if rel == ".." {
		return false
	}
	if len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return false
	}
	return !filepath.IsAbs(rel)
}

// stageWorkspaceDataMigration 把源数据目录内容暂存到 staging 路径并校验清单。
// [S] 汇编 0x1409f0240：validate→Stat 源→inspect 源/目标→目标为空校验→
// MkdirAll 目标父级→rand 命名 staging→Mkdir 0700→卷一致性→Walk 拷贝→
// verify(staging,false)→verify(source,true)→身份复核→构造 stagedWorkspaceData。
func stageWorkspaceDataMigration(source, target string) (*stagedWorkspaceData, error) {
	cleanSource, cleanTarget, err := validateWorkspaceMigrationRoots(source, target)
	if err != nil {
		return nil, err
	}

	sourceInfo, err := os.Stat(cleanSource)
	if err != nil {
		return nil, err
	}
	if !sourceInfo.IsDir() {
		return nil, newExtractError("源数据路径不是目录")
	}

	sourceInspection, err := inspectWorkspaceMigrationPath(cleanSource)
	if err != nil {
		return nil, err
	}
	targetInspection, err := inspectWorkspaceMigrationPath(cleanTarget)
	if err != nil {
		return nil, err
	}
	if targetInspection.exists {
		targetInfo, err := os.Stat(cleanTarget)
		if err != nil {
			return nil, err
		}
		if !targetInfo.IsDir() {
			return nil, newExtractError("目标数据路径已存在且不是目录")
		}
		entries, err := os.ReadDir(cleanTarget)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			return nil, newExtractError("目标数据目录必须为空")
		}
	}

	if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
		return nil, err
	}
	targetInspection2, err := inspectWorkspaceMigrationPath(cleanTarget)
	if err != nil {
		return nil, err
	}
	parentInspection, err := inspectWorkspaceMigrationPath(filepath.Dir(cleanTarget))
	if err != nil {
		return nil, err
	}
	if !parentInspection.exists {
		return nil, newExtractError("目标数据目录父级不存在")
	}

	var randBytes [8]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		return nil, err
	}
	stagingName := "." + filepath.Base(cleanTarget) + ".usbeam-staging-" + hex.EncodeToString(randBytes[:])
	stagingPath := filepath.Join(filepath.Dir(cleanTarget), stagingName)
	if err := os.Mkdir(stagingPath, 0o700); err != nil {
		return nil, err
	}
	stagingInspection, err := inspectWorkspaceMigrationPath(stagingPath)
	if err != nil {
		os.Remove(stagingPath)
		return nil, err
	}
	if stagingInspection.identity.volume != parentInspection.identity.volume {
		os.Remove(stagingPath)
		return nil, newExtractError("迁移暂存目录与目标目录不在同一卷")
	}

	manifest := make(map[string]workspaceDataManifestEntry)
	err = filepath.Walk(cleanSource, func(p string, info os.FileInfo, walkErr error) error {
		// copyWorkspaceDataWithManifest.func1 0x1409f1400：Rel→walkErr→skip→dst→
		// isReparse→IsDir→IsRegular→MkdirAll→openSource+SameFile→openTarget→copy→manifest。
		rel, err := filepath.Rel(cleanSource, p)
		if err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}
		if rel != "." && shouldSkipWorkspaceDataMigration(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dst := filepath.Join(stagingPath, rel)
		isReparse, err := workspaceMigrationPathIsReparse(p)
		if err != nil {
			return err
		}
		if isReparse {
			return fmt.Errorf("数据目录包含不支持的符号链接、目录联接或 reparse point: %s", p)
		}
		if info.IsDir() {
			return os.MkdirAll(dst, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("数据目录包含不支持的文件类型: %s", p)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		srcFile, err := openWorkspaceMigrationSourceFileNoFollow(p)
		if err != nil {
			return err
		}
		fi, statErr := srcFile.Stat()
		if statErr != nil || !os.SameFile(info, fi) {
			srcFile.Close()
			if statErr != nil {
				return statErr
			}
			return fmt.Errorf("迁移源文件身份发生变化: %s", p)
		}
		targetFile, err := openWorkspaceMigrationTargetFileNoFollow(dst, info.Mode().Perm())
		if err != nil {
			srcFile.Close()
			return err
		}
		h := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(targetFile, h), srcFile)
		syncErr := targetFile.Sync()
		targetCloseErr := targetFile.Close()
		srcCloseErr := srcFile.Close()
		if copyErr != nil {
			return copyErr
		}
		if syncErr != nil {
			return syncErr
		}
		if targetCloseErr != nil {
			return targetCloseErr
		}
		if srcCloseErr != nil {
			return srcCloseErr
		}
		manifest[filepath.ToSlash(rel)] = workspaceDataManifestEntry{
			Size:   n,
			SHA256: hex.EncodeToString(h.Sum(nil)),
		}
		return nil
	})
	if err != nil {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, err
	}

	if err := verifyWorkspaceDataManifestWithPolicy(stagingPath, manifest, false); err != nil {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, err
	}
	if err := verifyWorkspaceDataManifestWithPolicy(cleanSource, manifest, true); err != nil {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, fmt.Errorf("迁移期间源数据发生变化: %w", err)
	}

	sourceInspection2, err := inspectWorkspaceMigrationPath(cleanSource)
	if err != nil {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, err
	}
	if !sameWorkspacePathInspection(sourceInspection, sourceInspection2) {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, newExtractError("迁移期间源数据目录身份发生变化")
	}

	parentInspection2, err := inspectWorkspaceMigrationPath(filepath.Dir(cleanTarget))
	if err != nil {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, err
	}
	if !parentInspection.identity.valid || !parentInspection2.identity.valid ||
		parentInspection.identity.volume != parentInspection2.identity.volume ||
		parentInspection.identity.file != parentInspection2.identity.file {
		removeOwnedWorkspaceMigrationDirectory(stagingPath, stagingInspection.identity)
		return nil, newExtractError("迁移期间目标父目录身份发生变化")
	}

	return &stagedWorkspaceData{
		sourcePath:           cleanSource,
		targetPath:           cleanTarget,
		stagingPath:          stagingPath,
		manifest:             manifest,
		sourceInspection:     sourceInspection,
		targetInspection:     targetInspection2,
		targetParentIdentity: parentInspection.identity,
		stagingIdentity:      stagingInspection.identity,
	}, nil
}

// verifyWorkspaceDataManifestWithPolicy 校验路径树与清单一致。
// [S] 汇编 0x1409f1ba0：visited=make(map[string]struct{},len(manifest))；Walk(path,func1)→err；
// len(visited)!=len(manifest)→"迁移目标文件清单不完整"。
func verifyWorkspaceDataManifestWithPolicy(path string, manifest map[string]workspaceDataManifestEntry, strict bool) error {
	visited := make(map[string]struct{}, len(manifest))
	err := filepath.Walk(path, func(p string, info os.FileInfo, walkErr error) error {
		// verify.func1 0x1409f1da0：walkErr→Rel→strict skip→isReparse→IsDir→
		// Mode&ModeType→relKey 归一→清单外→sha256 校验→visited。
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if strict && rel != "." && shouldSkipWorkspaceDataMigration(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		isReparse, err := workspaceMigrationPathIsReparse(p)
		if err != nil {
			return err
		}
		if isReparse {
			return fmt.Errorf("迁移清单路径包含 reparse point: %s", p)
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&fs.ModeType != 0 {
			return fmt.Errorf("迁移目标包含非普通文件: %s", p)
		}
		relKey := strings.ReplaceAll(rel, "\\", "/")
		entry, ok := manifest[relKey]
		if !ok {
			return fmt.Errorf("迁移目标出现清单外文件: %s", relKey)
		}
		f, err := openWorkspaceMigrationSourceFileNoFollow(p)
		if err != nil {
			return err
		}
		h := sha256.New()
		n, copyErr := io.Copy(h, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n != entry.Size {
			return fmt.Errorf("迁移文件校验失败: %s", relKey)
		}
		if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), entry.SHA256) {
			return fmt.Errorf("迁移文件校验失败: %s", relKey)
		}
		visited[relKey] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	if len(visited) != len(manifest) {
		return newExtractError("迁移目标文件清单不完整")
	}
	return nil
}

// Commit 提交暂存数据：校验→身份复核→清空目标→rename→复核提交后身份。
// [S] 汇编 0x1409f2460。
func (s *stagedWorkspaceData) Commit() error {
	if s == nil {
		return newExtractError("迁移暂存状态不可用")
	}
	if err := verifyWorkspaceDataManifestWithPolicy(s.stagingPath, s.manifest, false); err != nil {
		return err
	}
	if err := verifyWorkspaceDataManifestWithPolicy(s.sourcePath, s.manifest, true); err != nil {
		return fmt.Errorf("提交前源数据发生变化: %w", err)
	}

	sourceInspection2, err := inspectWorkspaceMigrationPath(s.sourcePath)
	if err != nil {
		return err
	}
	if !sameWorkspacePathInspection(s.sourceInspection, sourceInspection2) {
		return newExtractError("提交前源数据目录身份发生变化")
	}

	stagingInspection, err := inspectWorkspaceMigrationPath(s.stagingPath)
	if err != nil {
		return err
	}
	if !stagingInspection.exists ||
		!s.stagingIdentity.valid || !stagingInspection.identity.valid ||
		s.stagingIdentity.volume != stagingInspection.identity.volume ||
		s.stagingIdentity.file != stagingInspection.identity.file {
		return newExtractError("迁移暂存目录身份发生变化")
	}

	parentInspection, err := inspectWorkspaceMigrationPath(filepath.Dir(s.targetPath))
	if err != nil {
		return err
	}
	if !parentInspection.exists ||
		!s.targetParentIdentity.valid || !parentInspection.identity.valid ||
		s.targetParentIdentity.volume != parentInspection.identity.volume ||
		s.targetParentIdentity.file != parentInspection.identity.file {
		return newExtractError("目标父目录身份发生变化")
	}

	targetInspection, err := inspectWorkspaceMigrationPath(s.targetPath)
	if err != nil {
		return err
	}
	if !sameWorkspacePathInspection(s.targetInspection, targetInspection) {
		return newExtractError("目标数据目录身份发生变化")
	}

	targetInfo, err := os.Stat(s.targetPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil && !targetInfo.IsDir() {
		return newExtractError("目标数据路径已存在且不是目录")
	}
	if err == nil {
		entries, err := os.ReadDir(s.targetPath)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return newExtractError("目标数据目录必须为空")
		}
		if err := os.Remove(s.targetPath); err != nil {
			return err
		}
		s.targetWasEmpty = true
	}

	if err := os.Rename(s.stagingPath, s.targetPath); err != nil {
		if s.targetWasEmpty {
			os.MkdirAll(s.targetPath, 0o755)
		}
		return err
	}

	s.committed = true
	s.committedIdentity = s.stagingIdentity

	committedInspection, err := inspectWorkspaceMigrationPath(s.targetPath)
	if err != nil {
		return err
	}
	if !committedInspection.exists ||
		!s.committedIdentity.valid || !committedInspection.identity.valid ||
		s.committedIdentity.volume != committedInspection.identity.volume ||
		s.committedIdentity.file != committedInspection.identity.file {
		return newExtractError("迁移提交后的目标目录身份不一致")
	}
	s.committedIdentity = committedInspection.identity
	return nil
}

// Rollback 回滚暂存数据（幂等补偿）。
// [S] 汇编 0x1409f3100：nil→nil；!committed→removeOwned(staging)；
// 否则 inspect(target) 身份复核→verify(target,false)→RemoveAll→targetWasEmpty 时 MkdirAll。
func (s *stagedWorkspaceData) Rollback() error {
	if s == nil {
		return nil
	}
	if !s.committed {
		return removeOwnedWorkspaceMigrationDirectory(s.stagingPath, s.stagingIdentity)
	}

	targetInspection, inspectErr := inspectWorkspaceMigrationPath(s.targetPath)
	if inspectErr == nil && targetInspection.exists &&
		s.committedIdentity.valid && targetInspection.identity.valid &&
		s.committedIdentity.volume == targetInspection.identity.volume &&
		s.committedIdentity.file == targetInspection.identity.file {
		if err := verifyWorkspaceDataManifestWithPolicy(s.targetPath, s.manifest, false); err != nil {
			return fmt.Errorf("迁移目标已被外部修改，拒绝删除: %w", err)
		}
		if err := os.RemoveAll(s.targetPath); err != nil {
			return err
		}
		if s.targetWasEmpty {
			os.MkdirAll(s.targetPath, 0o755)
		}
		return nil
	}

	if errors.Is(inspectErr, os.ErrNotExist) {
		return nil
	}
	if inspectErr != nil {
		return inspectErr
	}
	if !targetInspection.exists {
		return nil
	}
	return newExtractError("迁移目标身份已变化，拒绝删除")
}

// removeOwnedWorkspaceMigrationDirectory 仅当路径身份匹配时删除暂存目录。
// [S] 汇编 0x1409f34c0：inspect→ErrNotExist→nil；!exists→nil；
// 身份不匹配→"迁移暂存目录身份已变化，拒绝删除"；否则 RemoveAll。
func removeOwnedWorkspaceMigrationDirectory(path string, identity workspacePathIdentity) error {
	inspection, err := inspectWorkspaceMigrationPath(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !inspection.exists {
		return nil
	}
	if !identity.valid || !inspection.identity.valid ||
		identity.volume != inspection.identity.volume ||
		identity.file != inspection.identity.file {
		return newExtractError("迁移暂存目录身份已变化，拒绝删除")
	}
	return os.RemoveAll(path)
}
