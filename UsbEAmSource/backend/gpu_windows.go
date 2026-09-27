// AUTO-RECONSTRUCTED — GPU 偏好 Windows 侧纯逻辑助手链
// 研究用途
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// iidIDXGIAdapter 是 IDXGIAdapter 接口标识（EnumAdapterByGpuPreference 的 riid）。
var iidIDXGIAdapter = windows.GUID{Data1: 0x2411e7e1, Data2: 0x12ac, Data3: 0x4ccf, Data4: [8]byte{0xbd, 0x14, 0x97, 0x98, 0xe8, 0x53, 0x4d, 0xc0}}

// resolveGPUPreferenceAdapterInfo 解析节能/高性能两块显卡名。
// [S 汇编 0x14085d240]（gpu_windows.go 蓝图 108-125）：
//
//	resolveDXGIAdapterNameByPreference(1)=节能、(2)=高性能；错误仅 gpuPickDebugLog；
//	返回体两个名字均 TrimSpace。
func resolveGPUPreferenceAdapterInfo() GPUPreferenceAdapterInfo {
	powerSavingName, err := resolveDXGIAdapterNameByPreference(1)
	if err != nil {
		gpuPickDebugLog("读取节能 GPU 名称失败: err=%v", err)
	}
	highPerformanceName, err := resolveDXGIAdapterNameByPreference(2)
	if err != nil {
		gpuPickDebugLog("读取高性能 GPU 名称失败: err=%v", err)
	}
	return GPUPreferenceAdapterInfo{
		PowerSavingName:     strings.TrimSpace(powerSavingName),
		HighPerformanceName: strings.TrimSpace(highPerformanceName),
	}
}

// resolveDXGIAdapterNameByPreference 按 GPU 偏好枚举适配器名。
// [S 汇编 0x14085d340]（gpu_windows.go 蓝图 125-171）：createDXGIFactory1 →
// queryDXGIFactory6 → vtable[0xe8] EnumAdapterByGpuPreference（IID_IDXGIAdapter）→
// NOT_FOUND 返 ("",nil) → adapter 零返 ("",nil) → vtable[0x40] GetDesc →
// TrimSpace(UTF16ToString(Description))。三处 defer releaseDXGIUnknown。
func resolveDXGIAdapterNameByPreference(preference int) (string, error) {
	factory, err := createDXGIFactory1()
	if err != nil {
		return "", err
	}
	defer releaseDXGIUnknown(factory)

	factory6, err := queryDXGIFactory6(factory)
	if err != nil {
		return "", err
	}
	defer releaseDXGIUnknown(factory6)

	var adapter uintptr
	vtbl := *(**dxgiFactory6Vtbl)(unsafe.Pointer(&factory6))
	hresult, _, _ := syscall.SyscallN(
		vtbl.EnumAdapterByGpuPreference,
		factory6,
		0,
		uintptr(preference),
		uintptr(unsafe.Pointer(&iidIDXGIAdapter)),
		uintptr(unsafe.Pointer(&adapter)),
	)
	if hresult == 0x887a0002 { // DXGI_ERROR_NOT_FOUND
		return "", nil
	}
	if hresult != 0 {
		return "", fmt.Errorf("%s failed: HRESULT 0x%08X", "EnumAdapterByGpuPreference", uint32(hresult))
	}
	if adapter == 0 {
		return "", nil
	}
	defer releaseDXGIUnknown(adapter)

	var desc dxgiAdapterDesc
	hresult, _, _ = syscall.SyscallN(
		dxgiAdapterVtable(adapter).GetDesc,
		adapter,
		uintptr(unsafe.Pointer(&desc)),
	)
	if hresult != 0 {
		return "", fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIAdapter.GetDesc", uint32(hresult))
	}
	return strings.TrimSpace(syscall.UTF16ToString(desc.Description[:])), nil
}

// getGPUPreferenceState 读取注册表 GPU 偏好并规范化去重排序。
// [S 汇编 0x14085dda0]（gpu_windows.go 蓝图 230-280）：resolveGPUPreferenceAdapterInfo →
// OpenKey(CURRENT_USER, "Software\\Microsoft\\DirectX\\UserGpuPreferences", QUERY_VALUE) →
// ErrNotExist 返 (空 Entries + AdapterInfo, nil)；其余错误返 (零, err)；defer Close →
// ReadValueNames(0) → 逐名 GetStringValue → buildGPUPreferenceEntry →
// gpuPreferencePathKey 空跳过 → 已存在且 !shouldReplace 跳过 → 写 map →
// map 转 slice → sort.Slice（Name TrimSpace+ToLower 升序，相等按 PathKey 升序）。
func getGPUPreferenceState() (GPUPreferenceState, error) {
	adapterInfo := resolveGPUPreferenceAdapterInfo()

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\DirectX\UserGpuPreferences`, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return GPUPreferenceState{Entries: []GPUPreferenceEntry{}, AdapterInfo: adapterInfo}, nil
		}
		return GPUPreferenceState{}, err
	}
	defer key.Close()

	valueNames, err := key.ReadValueNames(0)
	if err != nil {
		return GPUPreferenceState{}, err
	}

	m := make(map[string]GPUPreferenceEntry, len(valueNames))
	for _, name := range valueNames {
		rawValue, _, err := key.GetStringValue(name)
		if err != nil {
			continue
		}
		entry, ok := buildGPUPreferenceEntry(name, rawValue)
		if !ok {
			continue
		}
		pathKey := gpuPreferencePathKey(entry.Path)
		if pathKey == "" {
			continue
		}
		if existing, ok := m[pathKey]; ok {
			if !shouldReplaceGPUPreferenceEntry(existing, entry) {
				continue
			}
		}
		m[pathKey] = entry
	}

	sortGPUPreferenceEntries := make([]GPUPreferenceEntry, 0, len(m))
	for _, entry := range m {
		sortGPUPreferenceEntries = append(sortGPUPreferenceEntries, entry)
	}

	sort.Slice(sortGPUPreferenceEntries, func(i, j int) bool {
		a := strings.ToLower(strings.TrimSpace(sortGPUPreferenceEntries[i].Name))
		b := strings.ToLower(strings.TrimSpace(sortGPUPreferenceEntries[j].Name))
		if a == b {
			return gpuPreferencePathKey(sortGPUPreferenceEntries[i].Path) < gpuPreferencePathKey(sortGPUPreferenceEntries[j].Path)
		}
		return a < b
	})

	return GPUPreferenceState{Entries: sortGPUPreferenceEntries, AdapterInfo: adapterInfo}, nil
}

// addGPUPreferenceEntry 新增 GPU 偏好条目。
// [S 汇编 0x14085e920]（gpu_windows.go 蓝图 284-317）：请求日志 → validateGPUPreferenceTargetPath →
// CreateKey(CURRENT_USER, UserGpuPreferences, QUERY_VALUE|SET_VALUE) → defer Close →
// deleteGPUPreferenceValueAliases → serialize 单条 {GpuPreference,"0"} → SetStringValue →
// 失败各出口 gpuPickDebugLog 后返 (零, err)；成功 getGPUPreferenceState 刷新并返 (state, nil)。
func addGPUPreferenceEntry(path string) (GPUPreferenceState, error) {
	gpuPickDebugLog("请求新增 GPU 首选项: rawPath=%q", path)

	validatedPath, err := validateGPUPreferenceTargetPath(path)
	if err != nil {
		gpuPickDebugLog("新增 GPU 首选项校验失败: rawPath=%q err=%v", path, err)
		return GPUPreferenceState{}, err
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\DirectX\UserGpuPreferences`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		gpuPickDebugLog("新增 GPU 首选项打开注册表失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}
	defer key.Close()

	if err := deleteGPUPreferenceValueAliases(key, validatedPath); err != nil {
		gpuPickDebugLog("新增 GPU 首选项清理别名失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}

	raw := serializeGPUPreferenceRegistrySettings([]GPUPreferenceSetting{{Key: "GpuPreference", Value: "0"}})
	if err := key.SetStringValue(validatedPath, raw); err != nil {
		gpuPickDebugLog("新增 GPU 首选项写入注册表失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}

	state, err := getGPUPreferenceState()
	if err != nil {
		gpuPickDebugLog("新增 GPU 首选项后刷新状态失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}
	gpuPickDebugLog("新增 GPU 首选项成功: path=%q entries=%d", validatedPath, len(state.Entries))
	return state, nil
}

// saveGPUPreferenceEntry 保存 GPU 偏好条目。
// [S 汇编 0x14085f340]（gpu_windows.go 蓝图 320-355）：请求日志（path + len(settings)）→
// validateGPUPreferenceTargetPath → ensureGPUPreferenceRegistrySettings → CreateKey → defer Close →
// deleteGPUPreferenceValueAliases → serializeGPUPreferenceRegistrySettings → SetStringValue →
// 失败各出口日志后返 (零, err)；成功 getGPUPreferenceState 刷新并返 (state, nil)。
func saveGPUPreferenceEntry(path string, settings []GPUPreferenceSetting) (GPUPreferenceState, error) {
	gpuPickDebugLog("请求保存 GPU 首选项: rawPath=%q settings=%d", path, len(settings))

	validatedPath, err := validateGPUPreferenceTargetPath(path)
	if err != nil {
		gpuPickDebugLog("保存 GPU 首选项校验失败: rawPath=%q err=%v", path, err)
		return GPUPreferenceState{}, err
	}

	normalized := ensureGPUPreferenceRegistrySettings(settings)

	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\DirectX\UserGpuPreferences`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		gpuPickDebugLog("保存 GPU 首选项打开注册表失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}
	defer key.Close()

	if err := deleteGPUPreferenceValueAliases(key, validatedPath); err != nil {
		gpuPickDebugLog("保存 GPU 首选项清理别名失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}

	raw := serializeGPUPreferenceRegistrySettings(normalized)
	if err := key.SetStringValue(validatedPath, raw); err != nil {
		gpuPickDebugLog("保存 GPU 首选项写入注册表失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}

	state, err := getGPUPreferenceState()
	if err != nil {
		gpuPickDebugLog("保存 GPU 首选项后刷新状态失败: path=%q err=%v", validatedPath, err)
		return GPUPreferenceState{}, err
	}
	gpuPickDebugLog("保存 GPU 首选项成功: path=%q entries=%d", validatedPath, len(state.Entries))
	return state, nil
}

// removeGPUPreferenceEntry 删除 GPU 偏好条目。
// [S 汇编 0x14085fde0]（gpu_windows.go 蓝图 358-394）：normalizeGPUPreferencePath 空 → 造错
// "显卡调度目标路径不能为空" 日志后返 (零, err)；请求日志 → OpenKey →
// errors.Is(err, ErrNotExist) 真 → 日志后返 (空 Entries + AdapterInfo, nil)；其余错误日志后返；
// defer Close → deleteGPUPreferenceValueAliases → 失败日志后返；成功 getGPUPreferenceState 刷新并返 (state, nil)。
func removeGPUPreferenceEntry(path string) (GPUPreferenceState, error) {
	normalized := normalizeGPUPreferencePath(path)
	if normalized == "" {
		err := errors.New("显卡调度目标路径不能为空")
		gpuPickDebugLog("删除 GPU 首选项失败: rawPath=%q err=%v", path, err)
		return GPUPreferenceState{}, err
	}
	gpuPickDebugLog("请求删除 GPU 首选项: rawPath=%q normalizedPath=%q", path, normalized)

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\DirectX\UserGpuPreferences`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			gpuPickDebugLog("删除 GPU 首选项时注册表键不存在，视为已删除: path=%q", normalized)
			adapterInfo := resolveGPUPreferenceAdapterInfo()
			return GPUPreferenceState{Entries: []GPUPreferenceEntry{}, AdapterInfo: adapterInfo}, nil
		}
		gpuPickDebugLog("删除 GPU 首选项打开注册表失败: path=%q err=%v", normalized, err)
		return GPUPreferenceState{}, err
	}
	defer key.Close()

	if err := deleteGPUPreferenceValueAliases(key, normalized); err != nil {
		gpuPickDebugLog("删除 GPU 首选项清理别名失败: path=%q err=%v", normalized, err)
		return GPUPreferenceState{}, err
	}

	state, err := getGPUPreferenceState()
	if err != nil {
		gpuPickDebugLog("删除 GPU 首选项后刷新状态失败: path=%q err=%v", normalized, err)
		return GPUPreferenceState{}, err
	}
	gpuPickDebugLog("删除 GPU 首选项成功: path=%q entries=%d", normalized, len(state.Entries))
	return state, nil
}

// cleanupMissingGPUPreferenceEntries 清理失效的 GPU 偏好条目。
// [S 汇编 0x140860760]（gpu_windows.go 蓝图 397-432）：OpenKey →
// errors.Is(err, ErrNotExist) 真 → 返 (空 Entries + AdapterInfo, nil)；其余错误返 (零, err)；
// ReadValueNames(0) 失败 → key.Close()（忽略）后返 (零, err)；遍历 valueNames，
// shouldCleanupMissingGPUPreferencePath(name, gpuPreferenceFileExists) 命中 →
// DeleteValue 失败且非 ErrNotExist → key.Close()（忽略）后返 (零, err)；
// 循环后 key.Close() 失败返 (零, err)；成功直返 getGPUPreferenceState()。
func cleanupMissingGPUPreferenceEntries() (GPUPreferenceState, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\DirectX\UserGpuPreferences`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			adapterInfo := resolveGPUPreferenceAdapterInfo()
			return GPUPreferenceState{Entries: []GPUPreferenceEntry{}, AdapterInfo: adapterInfo}, nil
		}
		return GPUPreferenceState{}, err
	}

	valueNames, err := key.ReadValueNames(0)
	if err != nil {
		key.Close()
		return GPUPreferenceState{}, err
	}

	for _, name := range valueNames {
		if shouldCleanupMissingGPUPreferencePath(name, gpuPreferenceFileExists) {
			if err := key.DeleteValue(name); err != nil {
				if errors.Is(err, registry.ErrNotExist) {
					continue
				}
				key.Close()
				return GPUPreferenceState{}, err
			}
		}
	}

	if err := key.Close(); err != nil {
		return GPUPreferenceState{}, err
	}
	return getGPUPreferenceState()
}

// shouldCleanupMissingGPUPreferencePath 判断给定路径是否应因目标缺失而被清理。
// [S 汇编 0x140860b40]（gpu_windows.go 蓝图 294-304）：
//
//	normalizeGPUPreferencePath 空 → false；
//	f 为 nil → false；非绝对路径 → false；
//	gpuPreferencePathMayAccessFile 拒绝 → false；
//	否则返回 !f(p)。
func shouldCleanupMissingGPUPreferencePath(path string, f func(string) bool) bool {
	p := normalizeGPUPreferencePath(path)
	if p == "" {
		return false
	}
	if f == nil {
		return false
	}
	if !filepath.IsAbs(p) {
		return false
	}
	if !gpuPreferencePathMayAccessFile(p) {
		return false
	}
	return !f(p)
}

// buildGPUPreferenceEntry 由注册表原始值构造一条 GPU 偏好条目。
// [S 汇编 0x140860c00]（gpu_windows.go 蓝图 304-330）：路径规范化与文件系统
// 判定 → 解析 settings → 模式/取值 → 可访问 + 文件存在 → 显示名/图标 → 序列化。
// 结果 Icon 字段刻意留空（汇编 duffcopy 零值模板，未写 Icon 槽位）。
func buildGPUPreferenceEntry(path, rawValue string) (GPUPreferenceEntry, bool) {
	p := normalizeGPUPreferencePath(path)
	if p == "" || !looksLikeFilesystemPath(p) {
		return GPUPreferenceEntry{}, false
	}
	settings := parseGPUPreferenceRegistryValue(rawValue)
	mode, value := gpuPreferenceModeFromSettings(settings)
	if !gpuPreferencePathMayAccessFile(p) {
		return GPUPreferenceEntry{}, false
	}
	if !gpuPreferenceFileExists(p) {
		return GPUPreferenceEntry{}, false
	}
	name, _ := resolveAppDisplayName(p)
	iconData := strings.TrimSpace(resolveAppIconDataWithOptions(p, AppIconOptions{}))
	name = strings.TrimSpace(name)
	if name == "" {
		name = fallbackAppDisplayName(p)
	}
	raw := serializeGPUPreferenceRegistrySettings(settings)
	return GPUPreferenceEntry{
		Path:            p,
		Name:            name,
		IconData:        iconData,
		PreferenceMode:  mode,
		PreferenceValue: value,
		Settings:        settings,
		RawValue:        raw,
	}, true
}

// gpuPreferencePathMayAccessFile 判断路径是否指向可访问的本地 .exe 文件。
// [S 汇编 0x140860f40]（gpu_windows.go 蓝图 330-344）：
//
//	Replace("/", "\") → TrimSpace → ToLower；空返 false；
//	filepath.Ext（末段扩展名）非 ".exe" → false；
//	前缀为 \\、\\??\\、\\?\\、\\.\\（UNC/设备路径）→ false；
//	否则 true。
func gpuPreferencePathMayAccessFile(path string) bool {
	p := strings.ToLower(strings.TrimSpace(strings.Replace(path, "/", `\`, -1)))
	if p == "" {
		return false
	}
	if !strings.EqualFold(filepath.Ext(p), ".exe") {
		return false
	}
	if strings.HasPrefix(p, `\\`) || strings.HasPrefix(p, `\??\`) ||
		strings.HasPrefix(p, `\\?\`) || strings.HasPrefix(p, `\\.\`) {
		return false
	}
	return true
}

// shouldReplaceGPUPreferenceEntry 判断候选条目是否应替换已有条目。
// [S 汇编 0x1408610c0]（gpu_windows.go 蓝图 344-355）：
//
//	PreferenceValue 不同 → 取更大者；
//	候选有 Icon 而既有无 → 替换；
//	候选 Path 非空而既有 Path 空 → 替换；
//	否则按 Settings 容量（汇编读 0x68=cap 槽位）比较。
func shouldReplaceGPUPreferenceEntry(existing, candidate GPUPreferenceEntry) bool {
	if candidate.PreferenceValue != existing.PreferenceValue {
		return candidate.PreferenceValue > existing.PreferenceValue
	}
	if candidate.Icon != "" && existing.Icon == "" {
		return true
	}
	if strings.TrimSpace(candidate.Path) != "" && strings.TrimSpace(existing.Path) == "" {
		return true
	}
	return cap(existing.Settings) < cap(candidate.Settings)
}

// validateGPUPreferenceTargetPath 校验并规范化 GPU 偏好目标路径。
// [S 汇编 0x140861180]（gpu_windows.go 蓝图 355-371）：委托 WithResolver，
// 分别注入快捷方式解析器、filepath.EvalSymlinks、os.Stat。
func validateGPUPreferenceTargetPath(path string) (string, error) {
	return validateGPUPreferenceTargetPathWithResolver(
		path,
		func(p string) (string, error) {
			info, err := resolveShortcutInfoWithIconResolver(p)
			if err != nil {
				return "", err
			}
			target := strings.TrimSpace(info.TargetPath)
			if target == "" {
				return "", errors.New("快捷方式没有有效目标")
			}
			return target, nil
		},
		filepath.EvalSymlinks,
		os.Stat,
	)
}

// validateGPUPreferenceTargetPathWithResolver 带依赖注入的 GPU 目标路径校验。
// [S 汇编 0x1408611e0]（gpu_windows.go 蓝图 371-429）：规范化 → 本地路径判定 →
// .lnk 解析（含链式 .lnk 拒绝）→ .exe 判定 → Abs → EvalSymlinks → 再判 .exe →
// stat 非目录。stat 回调返回 os.FileInfo，其 IsDir 判定目录拒绝。
func validateGPUPreferenceTargetPathWithResolver(
	path string,
	resolveShortcut func(string) (string, error),
	resolveSymlinks func(string) (string, error),
	stat func(string) (os.FileInfo, error),
) (string, error) {
	p := normalizeGPUPreferencePath(path)
	if p == "" {
		return "", errors.New("显卡调度目标路径不能为空")
	}
	if !looksLikeFilesystemPath(p) {
		return "", fmt.Errorf("显卡调度只支持本地程序路径: %s", strings.TrimSpace(path))
	}

	ext := strings.ToLower(filepath.Ext(p))
	if ext == ".lnk" {
		if resolveShortcut == nil {
			return "", errors.New("无法解析显卡调度快捷方式")
		}
		target, err := resolveShortcut(p)
		if err != nil {
			return "", fmt.Errorf("解析显卡调度快捷方式失败: %w", err)
		}
		p = normalizeGPUPreferencePath(target)
		if strings.EqualFold(filepath.Ext(p), ".lnk") {
			return "", errors.New("显卡调度不支持链式快捷方式")
		}
	} else if ext != ".exe" {
		return "", fmt.Errorf("显卡调度只支持 .exe 或指向 .exe 的 .lnk: %s", p)
	}

	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("规范化显卡调度目标失败: %w", err)
	}
	if resolveSymlinks != nil {
		abs, err = resolveSymlinks(abs)
		if err != nil {
			return "", fmt.Errorf("解析显卡调度目标失败: %w", err)
		}
	}
	p = normalizeGPUPreferencePath(abs)
	if !strings.EqualFold(filepath.Ext(p), ".exe") {
		return "", fmt.Errorf("显卡调度目标必须是实际 .exe 文件: %s", p)
	}
	if stat != nil {
		info, err := stat(p)
		if err != nil {
			return "", fmt.Errorf("目标文件不存在，无法加入显卡调度: %s", p)
		}
		if info.IsDir() {
			return "", fmt.Errorf("显卡调度目标不能是目录: %s", p)
		}
	}
	return p, nil
}

// deleteGPUPreferenceValueAliases 删除注册表键中与给定路径等价的所有值别名。
// [S 汇编 0x1408617e0]（gpu_windows.go 蓝图 429-448）：gpuPreferencePathKey 生成
// 目标键，遍历 ReadValueNames，键匹配则 DeleteValue；ErrNotExist 忽略，其余错误透传。
func deleteGPUPreferenceValueAliases(key registry.Key, path string) error {
	target := gpuPreferencePathKey(path)
	if target == "" {
		return nil
	}
	names, err := key.ReadValueNames(0)
	if err != nil {
		return err
	}
	for _, name := range names {
		if gpuPreferencePathKey(name) == target {
			if err := key.DeleteValue(name); err != nil && !errors.Is(err, registry.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

// gpuPreferenceFileExists 判断路径是否为存在且非目录的文件。
// [S 汇编 0x140861960]（gpu_windows.go 蓝图 448-461）：TrimSpace → 空返 false →
// UTF16PtrFromString + GetFileAttributes（成功则按 FILE_ATTRIBUTE_DIRECTORY 判定）→
// 失败回退 os.Stat，错误返 false，否则 !info.IsDir()。
func gpuPreferenceFileExists(path string) bool {
	p := strings.TrimSpace(path)
	if p == "" {
		return false
	}
	ptr, err := windows.UTF16PtrFromString(p)
	if err == nil {
		if attrs, err := windows.GetFileAttributes(ptr); err == nil {
			return attrs&windows.FILE_ATTRIBUTE_DIRECTORY == 0
		}
	}
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
