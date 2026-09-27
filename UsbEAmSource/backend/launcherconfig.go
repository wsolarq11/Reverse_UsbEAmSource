// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: launcher config store (装配层/config 域)
// 研究用途
//
// 契约来源：
//   - 目标 VA: launcherConfigStoreForPath(0x140898ac0)、launcherConfigStorePathKey(0x140898c20)
//   - 方法体：capstone 反汇编符号标注（docs/goresym/disasm_assemble/launcherConfigStoreForPath.dis.txt）
//   - 结构字段：backend/types_launcher.go L511 launcherConfigStore
//
// 还原口径（与 launcher 域一致）：可编译 + 功能一致（非字节级同哈希）。
// 档位如实标注：
//
//	[S] 汇编实证（本函数体逐条对位）
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// launcherConfigStorePathKey 归一化配置路径为缓存 key：
// TrimSpace → filepath.Abs（失败则 Clean 兜底）→ ToLower。 [S 汇编实证 0x140898c20]
func launcherConfigStorePathKey(path string) string {
	p := strings.TrimSpace(path)
	if a, err := filepath.Abs(p); err == nil {
		p = a
	} else {
		p = filepath.Clean(p)
	}
	return strings.ToLower(p)
}

// launcherConfigStoreCache 保存 launcherConfigStoreForPath 的全局单例缓存。
// 数据来源：汇编全局 HashTrieMap（internal/sync.HashTrieMap[any,any]，经 sync.Map 语义触达）。
var launcherConfigStoreCache sync.Map

// launcherConfigStoreForPath 为指定路径返回(或创建并缓存)配置存储。
// [S 汇编实证 0x140898ac0]：TrimSpace → PathKey → Load（命中即类型断言返回）→
// 未命中 newobject{path=TrimSpace} → LoadOrStore 回填 → 类型断言返回。
func launcherConfigStoreForPath(path string) *launcherConfigStore {
	p := strings.TrimSpace(path)
	key := launcherConfigStorePathKey(p)
	if v, ok := launcherConfigStoreCache.Load(key); ok {
		st, ok := v.(*launcherConfigStore)
		if !ok {
			panic("launcherconfig: cache holds non-configStore")
		}
		return st
	}
	st := &launcherConfigStore{path: p}
	actual, _ := launcherConfigStoreCache.LoadOrStore(key, st)
	got, ok := actual.(*launcherConfigStore)
	if !ok {
		panic("launcherconfig: load-or-store returned non-configStore")
	}
	return got
}

// matchesPath 判断存储当前路径（经 PathKey 归一化）是否与给定路径一致。
// [S 汇编实证 0x140898ca0, 0x6f]：
//
//	入口 rax=s(接收者), rbx=path.ptr, rcx=path.len。
//	test rax,rax → s==nil 则 xor eax,eax 返回 false（0x140898cb3-0x140898cba）。
//	否则 rcx/rbx = s.path（[rax]/[rax+8]）→ launcherConfigStorePathKey(s.path)
//	→ 结果溢出到 [rsp+0x20]/[rsp+0x18]；再以 rax/rbx=path 调 PathKey(path)
//	→ 先 cmp 长度（[rsp+0x18] vs rbx），不等返 false；相等走 runtime.memequal 逐字节比较。
func (s *launcherConfigStore) matchesPath(path string) bool {
	if s == nil {
		return false
	}
	return launcherConfigStorePathKey(s.path) == launcherConfigStorePathKey(path)
}

// readLauncherConfigBytes 有界读取配置文件字节（<=64MB）。
// [S 汇编实证 0x140897c20]：os.OpenFile(O_RDONLY) → Stat 校验模式 → ReadAll →
// 读取数或文件大小超 64MB 报错；defer close。
func readLauncherConfigBytes(path string) ([]byte, error) {
	f, err := os.OpenFile(path, 0, 0) // O_RDONLY
	if err != nil {
		return nil, err
	}
	defer f.Close()
	const maxCfg = 64 << 20
	data, err := io.ReadAll(io.LimitReader(f, maxCfg+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxCfg {
		return nil, fmt.Errorf("launcherconfig: config %d exceeds 64MB", len(data))
	}
	return data, nil
}

// writeJSONFile 将任意值以缩进 JSON 原子写盘（临时文件 + rename 重试）。
// [S 汇编实证 0x14088bd80]：TrimSpace(空→错) → MkdirAll(dir,755) → MarshalIndent(v,"","  ")
// → 追加 '\n' → CreateTemp(dir,"*") 写入 → close → rename 到目标（遇可重试错误 sleep50ms 重试 ≤5）。
func writeJSONFile(path string, v interface{}) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("launcherconfig: empty config path")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if len(data) > 64<<20 {
		return fmt.Errorf("launcherconfig: config json %d exceeds 64MB", len(data))
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(dir, "*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	// 原子替换目标路径：rename，遇可重试错误先清理临时文件后交由上层的重试窗口。
	for i := 0; i < 5; i++ {
		if err := os.Rename(tmp.Name(), path); err == nil {
			return nil
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	os.Remove(tmp.Name())
	return nil
}

// launcherConfigIconDataFingerprint 对 LauncherConfig 的图标数据计算 SHA-256 指纹（前 32 字节）。
// [S 汇编 0x14088d300]：newobject{config} → sha256.Digest.Reset → visitLauncherConfigIconData 遍历写；
// [P]：visitLauncherConfigIconData 的图标字段遍历范围待 icon 域字节级对齐（此处按配置序列化字节哈希近似）。
func launcherConfigIconDataFingerprint(cfg LauncherConfig) [32]byte {
	h := sha256.New()
	// [S] 走 visitLauncherConfigIconData 遍历图标数据写哈希（比 json.Marshal 近似更贴近汇编；
	// 具体图标字段集合依托 icon slots，[P] 待更精确）→ SHA-256 前 32 字节。
	visitLauncherConfigIconData(&cfg, func(s string) { _, _ = h.Write([]byte(s)) })
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// writeConfigUnlocked 落盘配置并刷新图标指纹标志。
// [S 汇编 0x14089d580]：若 writeConfig 注入则调用之，否则 writeJSONFile；随后刷 iconFingerprint/
// iconFingerprintReady/runtimeReady，返回错误。
func (s *launcherConfigStore) writeConfigUnlocked(cfg LauncherConfig) error {
	if s.writeConfig != nil {
		if err := s.writeConfig(s.path, cfg); err != nil {
			return err
		}
	} else if err := writeJSONFile(s.path, cfg); err != nil {
		return err
	}
	fp := launcherConfigIconDataFingerprint(cfg)
	s.iconFingerprint = fp
	s.iconFingerprintReady = true
	s.runtimeReady = true
	return nil
}

// loadUnlocked 加载配置并刷新运行时/图标指纹状态。
// [S 汇编 0x140898d40]：若 runtimeReady 已置则短路返回；否则 loadLauncherConfigIfExistsUnlockedWithWriter
// 读取 + 反序列化，成功 → launcherConfigIconDataFingerprint 刷 iconFingerprint/@39，置 runtimeReady=1；
// 失败 → 清 iconFingerprint/@39。返回加载的配置。
func (s *launcherConfigStore) loadUnlocked() (LauncherConfig, error) {
	if s.runtimeReady {
		// [S] 短路走 loadLauncherConfigRuntimeUnlocked（读回但不刷指纹，L0x140898f34）。
		return s.loadLauncherConfigRuntimeUnlocked()
	}
	c, err := s.loadLauncherConfigIfExistsUnlocked()
	if err != nil {
		s.iconFingerprint = [32]byte{}
		s.iconFingerprintReady = false
		return LauncherConfig{}, err
	}
	s.runtimeReady = true
	fp := launcherConfigIconDataFingerprint(c)
	s.iconFingerprint = fp
	s.iconFingerprintReady = true
	return c, nil
}

// loadLauncherConfigRuntimeUnlocked runtime-ready 后再次读取（不刷指纹）。
// [S 汇编 0x140876ce0]：read → json 反序列化，返回。
func (s *launcherConfigStore) loadLauncherConfigRuntimeUnlocked() (LauncherConfig, error) {
	var c LauncherConfig
	data, err := readLauncherConfigBytes(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c, nil
		}
		return c, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}

// loadLauncherConfigIfExistsUnlocked 读取并反序列化配置（存在则加载，不存在视为空）。
// [S 汇编 0x140876220]：readLauncherConfigBytes → 不存在(io/fs.ErrNotExist) 返回空配置；
// 否则 json.Unmarshal → [P]：writer/Migration 边栏未展开，此以纯读实现。
func (s *launcherConfigStore) loadLauncherConfigIfExistsUnlocked() (LauncherConfig, error) {
	var cfg LauncherConfig
	data, err := readLauncherConfigBytes(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// deletePreparedLocked 删除配置：先 executed 可选 prepare 回调(loaded config)，随后 os.Remove 配置，
// 成功(含 NotExist 视为成功)后清理 icon 指纹标志。
// [S 汇编 0x14089ca40]。
func (s *launcherConfigStore) deletePreparedLocked(prepare func(LauncherConfig) error) error {
	cfg, _ := s.loadUnlocked()
	if prepare != nil {
		if err := prepare(cfg); err != nil {
			return err
		}
	}
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.iconFingerprint = [32]byte{}
	s.iconFingerprintReady = false
	return nil
}

// ---- initializeWithCheckpoint 骨架依赖（[P] 脚手架：bootstrap 初始化链待字节级续作）----

// 注：workspaceSnapshot 已按汇编实证迁往 bootstrapservice.go。
// 实证签名是 `func (bs *BootstrapService) workspaceSnapshot() WorkspaceLayout`
// （VA 0x1407a10c0, 272B，duffcopy 0xa0 字节 = WorkspaceLayout；三处调用点
// ChooseInitializationDataRoot/ChooseLauncherBackgroundImage/GetScreenshotHistory
// 均以它直接取 workspace 字段）。旧占位版误返 BootstrapSnapshot，已删除。

// 注：ensureWorkspaceDirectories 已按汇编实证迁往 bootstrapservice_config.go。
// 包级版签名为 (WorkspaceLayout) error（VA 0x1407a2520），
// 与 VA 0x1407a26a0 的 BootstrapService 方法版是两个不同符号（va_map_fixed2.txt:337）。

// loadLauncherConfigIfExists 读配置；不存在或解析失败则 ok=false。
// [S 汇编 0x1408760e0]：launcherConfigStoreForPath(path) → store.Read() → (LauncherConfig, error)。
// 现有签名保留(bool)以兼容旧调用点，err!=nil 时 ok=false。
func loadLauncherConfigIfExists(path string) (LauncherConfig, bool, error) {
	store := launcherConfigStoreForPath(path)
	cfg, err := store.Read()
	if err != nil {
		return LauncherConfig{}, false, err
	}
	return cfg, true, nil
}

// backupLauncherConfigForToday 备份今日配置。
// [S 汇编 0x140877620, 1152B] 全量实证：
// 入口 5 寄存器：rax=path.ptr, rbx=path.len, rcx=retention, rdi=time.wall, r8=time.ext
// （若 rdi/r8 都为零则函数内自调 time.Now）。
// 返回：rax/rbx=backupPath(string), rcx/rdi=error(nil 时为零)。
//
// 流程：TrimSpace(path)→空则 error "当前配置路径不能为空"→
// os.Stat→ErrNotExist 短路返回→IsDir 则 error "当前配置路径不能是目录"→
// launcherConfigBackupDir→MkdirAll(0755)→
// launcherConfigBackupPath(path, now)→fileExists(已存在则跳过写入)→
// launcherConfigStoreForPath→ReadSelfContained→writeSelfContainedConfig(false)→
// retention 钳制[10,1000]→pruneLauncherConfigBackups(path, retention)。
func backupLauncherConfigForToday(path string, retention int, now time.Time) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("当前配置路径不能为空")
	}

	if fi, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	} else if fi.IsDir() {
		return "", errors.New("当前配置路径不能是目录")
	}

	backupDir := launcherConfigBackupDir(path)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	backupPath := launcherConfigBackupPath(path, now)
	exists, err := fileExists(backupPath)
	if err != nil || exists {
		if err != nil {
			return "", err
		}
		// 已存在：跳过写入，但仍做清理
		pruneLauncherConfigBackups(path, clampRetention(retention))
		return backupPath, nil
	}

	store := launcherConfigStoreForPath(path)
	widgetDoc, err := store.ReadSelfContained()
	if err != nil {
		return "", err
	}
	if _, err := writeSelfContainedLauncherConfigWithWidgets(backupPath, path, widgetDoc); err != nil {
		return "", err
	}

	pruneLauncherConfigBackups(path, clampRetention(retention))
	return backupPath, nil
}

// clampRetention 将保留数钳制在 [10, 1000] 范围内，缺省 100。
// [S 汇编实证]：<=0→100；<10→10；>1000→1000。
func clampRetention(n int) int {
	if n <= 0 {
		return 100
	}
	if n < 10 {
		return 10
	}
	if n > 1000 {
		return 1000
	}
	return n
}

// ---- saveUnlockedWithCurrentRefs 链（[P] 前置辅助脚手架；写链走既有 writeConfigUnlocked）----

// errTwoFactorDataCorrupted 二因素存储数据损坏的统一 sentinel。
// [S 实证]：validateTwoFactorStoredConfig(0x1409c6740) 所有 %w 均包裹同一全局 errorString
// "TWO_FACTOR_DATA_CORRUPTED" 25B @0x140c66993。
var errTwoFactorDataCorrupted = errors.New("TWO_FACTOR_DATA_CORRUPTED")

// validateTwoFactorStoredConfig 校验二因素密码配置与条目的一致性。
// [S 汇编实证 0x1409c6740, 0x5a0]：
//
//	参数 pw=TwoFactorPasswordConfig（KDF@+0 Salt@+0x10 Verifier@+0x20 Blank@+0x30）、
//	entries []TwoFactorEntryConfig（stride 0xd0）。
//	flag = Blank || TrimSpace(KDF)!="" || TrimSpace(Salt)!="" || TrimSpace(Verifier)!=""
//	flag 真 → KDF 须空或 "argon2id-v1"；Salt base64 解码须 16B；Verifier base64 解码须 32B。
//	flag 假 → Salt/Verifier 非空报 "password 配置不完整"。
//	len(entries)>0 且 !flag → "存在条目但密码配置缺失"。
//	逐条 validateTwoFactorStoredEntry，失败包裹 "entries[%d]: %w"。
func validateTwoFactorStoredConfig(pw TwoFactorPasswordConfig, entries []TwoFactorEntryConfig) error {
	flag := pw.Blank
	if !flag {
		flag = strings.TrimSpace(pw.KDF) != ""
	}
	if !flag {
		flag = strings.TrimSpace(pw.Salt) != ""
	}
	if !flag {
		flag = strings.TrimSpace(pw.Verifier) != ""
	}

	if flag {
		kdf := strings.TrimSpace(pw.KDF)
		if kdf != "" && kdf != "argon2id-v1" {
			return fmt.Errorf("%w: password.kdf 无效", errTwoFactorDataCorrupted)
		}
		salt, err := base64.StdEncoding.DecodeString(strings.TrimSpace(pw.Salt))
		if err != nil || len(salt) != 16 {
			return fmt.Errorf("%w: password.salt 无效", errTwoFactorDataCorrupted)
		}
		verifier, err := base64.StdEncoding.DecodeString(strings.TrimSpace(pw.Verifier))
		if err != nil || len(verifier) != 32 {
			return fmt.Errorf("%w: password.verifier 无效", errTwoFactorDataCorrupted)
		}
	} else if strings.TrimSpace(pw.Salt) != "" || strings.TrimSpace(pw.Verifier) != "" {
		return fmt.Errorf("%w: password 配置不完整", errTwoFactorDataCorrupted)
	}

	if len(entries) > 0 && !flag {
		return fmt.Errorf("%w: 存在条目但密码配置缺失", errTwoFactorDataCorrupted)
	}

	for i := range entries {
		if err := validateTwoFactorStoredEntry(entries[i]); err != nil {
			return fmt.Errorf("entries[%d]: %w", i, err)
		}
	}
	return nil
}

// validateTwoFactorStoredEntry 校验单条二因素条目存储数据。
// [S 汇编实证 0x1409c6ce0, 0x560]：
//
//	Kind ToLower 后须 "totp"/"steam"；Algorithm ToUpper 后须空/"SHA1"/"SHA256"/"SHA512"。
//	steam：Algorithm 空/SHA1，Digits 0/5，Period 0/30。
//	totp：Digits 0/[4,10]，Period 0/[5,300]。
//	SecretNonce/SecretCiphertext 至少一项非空；nonce base64 解码须 12B；
//	ciphertext base64 解码须 [16,65536]B。
func validateTwoFactorStoredEntry(e TwoFactorEntryConfig) error {
	kind := strings.ToLower(strings.TrimSpace(e.Kind))
	if kind != "totp" && kind != "steam" {
		return fmt.Errorf("%w: kind 无效", errTwoFactorDataCorrupted)
	}
	algo := strings.ToUpper(strings.TrimSpace(e.Algorithm))
	if algo != "" && algo != "SHA1" && algo != "SHA256" && algo != "SHA512" {
		return fmt.Errorf("%w: algorithm 无效", errTwoFactorDataCorrupted)
	}
	if kind == "steam" {
		if algo != "" && algo != "SHA1" {
			return fmt.Errorf("%w: Steam 参数无效", errTwoFactorDataCorrupted)
		}
		if e.Digits != 0 && e.Digits != 5 {
			return fmt.Errorf("%w: Steam 参数无效", errTwoFactorDataCorrupted)
		}
		if e.Period != 0 && e.Period != 30 {
			return fmt.Errorf("%w: Steam 参数无效", errTwoFactorDataCorrupted)
		}
	} else {
		if e.Digits != 0 && (e.Digits < 4 || e.Digits > 10) {
			return fmt.Errorf("%w: TOTP 参数无效", errTwoFactorDataCorrupted)
		}
		if e.Period != 0 && (e.Period < 5 || e.Period > 300) {
			return fmt.Errorf("%w: TOTP 参数无效", errTwoFactorDataCorrupted)
		}
	}

	nonce := strings.TrimSpace(e.SecretNonce)
	cipher := strings.TrimSpace(e.SecretCiphertext)
	if nonce == "" && cipher == "" {
		return fmt.Errorf("%w: 缺少加密数据", errTwoFactorDataCorrupted)
	}
	if b, err := base64.StdEncoding.DecodeString(nonce); err != nil || len(b) != 12 {
		return fmt.Errorf("%w: secretNonce 无效", errTwoFactorDataCorrupted)
	}
	if b, err := base64.StdEncoding.DecodeString(cipher); err != nil || len(b) < 16 || len(b) > 65536 {
		return fmt.Errorf("%w: secretCiphertext 无效", errTwoFactorDataCorrupted)
	}
	return nil
}

// normalizeLauncherConfigForSave 保存前规范化配置。
// [S-sig 0x140877420, 448B]：壳 → normalizeLauncherConfigWithOptions(0x140879380)；体骨架（恒等）。
func normalizeLauncherConfigForSave(cfg LauncherConfig) LauncherConfig {
	return cfg
}

// validateChangedIconData 校验变更的图标数据。
// [S-sig 0x140899000, 8L]：签名经 saveUnlockedWithCurrentRefs 实证；体骨架（恒 nil）。
func (s *launcherConfigStore) validateChangedIconData(cfg LauncherConfig) error {
	_ = s
	_ = cfg
	return nil
}

// saveUnlockedWithCurrentRefs 保存配置（仍保留当前图标引用）。
// [S 汇编实证 0x14089cd20 主流程]：validate2FA → normalize → validateChangedIcon → prepareIcon
// → writeConfigUnlocked 落盘。
func (s *launcherConfigStore) saveUnlockedWithCurrentRefs(cfg LauncherConfig) error {
	if err := validateTwoFactorStoredConfig(cfg.TwoFactor.Password, cfg.TwoFactor.Entries); err != nil {
		return err
	}
	n := normalizeLauncherConfigForSave(cfg)
	if err := s.validateChangedIconData(n); err != nil {
		return err
	}
	prepared, err := prepareLauncherConfigIconsForCommitWithCurrentRefs(n)
	if err != nil {
		return err
	}
	return s.writeConfigUnlocked(prepared)
}

// ---- savePreparedUnlocked 链（icon refs 收集 [S] 汇编实证）----

// collectLauncherConfigIconRefs 已迁至 launcherconfigiconstore.go。
// collectLauncherConfigIconRefCounts 已迁至 launcherconfigiconcommit.go。

// savePreparedUnlocked 保存准备态配置（可选先收集当前图标 refs）。
// [S 汇编实证 0x14089d1c0 主流程]：若带图标收集则 collect refs/refcounts → saveUnlockedWithCurrentRefs。
func (s *launcherConfigStore) savePreparedUnlocked(withIcons bool, cfg LauncherConfig) error {
	if withIcons {
		_ = collectLauncherConfigIconRefs(cfg)
		_ = collectLauncherConfigIconRefCounts(cfg)
	}
	return s.saveUnlockedWithCurrentRefs(cfg)
}

// normalizeStorageConfig 规整存储配置段。 [S 汇编实证 0x140879b40, 352B]：
// 对 DataRoot/IconDir/IndexDir/ScreenshotDir/WebView2Dir 各执行 strings.TrimSpace 后组装返回。
func normalizeStorageConfig(dataRoot, iconDir, indexDir, screenshotDir, webView2Dir string) StorageConfig {
	return StorageConfig{
		DataRoot:      strings.TrimSpace(dataRoot),
		IconDir:       strings.TrimSpace(iconDir),
		IndexDir:      strings.TrimSpace(indexDir),
		ScreenshotDir: strings.TrimSpace(screenshotDir),
		WebView2Dir:   strings.TrimSpace(webView2Dir),
	}
}
