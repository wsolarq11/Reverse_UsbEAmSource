// AUTO-RECONSTRUCTED -- DOMAIN: start menu scanning
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak（0x1409c3da0/0x1409c3e00/0x1409c3e60/0x1409c3ec0/
//     0x1409c4a40/0x1409c4d40/0x1409c5180/0x1409c5480）
//   - 行号蓝图：source_funcs.txt File: startmenu.go L26-219
//   - 反汇编：tmp_disasm_shortcut/scanStartMenuApps.asm.txt、
//     attachStartMenuAppIconResources.asm.txt、hydrateStartMenuAppEntryWithResolvers.asm.txt、
//     resolveWindowsStartMenuRoots.asm.txt、resolveWindowsDesktopRoots.asm.txt
//
// 档位说明：
//   - [S] 汇编实证：scanStartMenuApps 主体（0x1409c3ec0：roots 遍历 + WalkDir 收集 +
//     sort.Slice 排序）；resolveWindows*Roots（filepath.Join + os.Getenv 双根）。
//   - [S] 结构实证：遍历元素 0x20B = startMenuRoot{Path,Source}（types_app.go）。
//   - [P] 动态图标/COM 快捷方式明细：resolveShortcutInfoWithIconResolver 为 shortcutinfo.go
//     stub（离线无 go-ole）；IconLocation 级图标解析在真实 COM 落地前由
//     startmenu_icon_windows.go 链做路径兜底。
package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---- 核心入口（RPC 可见的扫描三入口） ----

// ScanStartMenuApps 扫描 Windows 开始菜单快捷方式（静态扫描入口）。
// [S 汇编实证 0x1409c3da0 start area]：简单包装 -> scanStartMenuApps(roots, dynamic=false)。
func (bs *BootstrapService) ScanStartMenuApps() []StartMenuApp {
	return bs.scanStartMenuApps(composeStartMenuScanRoots(resolveWindowsStartMenuRoots(), "startmenu"), false)
}

// ScanDynamicStartMenuApps 动态扫描开始菜单快捷方式（每次调用即时重扫）。
// [S 汇编实证 0x1409c3e00 start area]：包装 -> scanStartMenuApps(roots, dynamic=true)。
func (bs *BootstrapService) ScanDynamicStartMenuApps() []StartMenuApp {
	return bs.scanStartMenuApps(composeStartMenuScanRoots(resolveWindowsStartMenuRoots(), "startmenu"), true)
}

// ScanDynamicDesktopApps 动态扫描桌面快捷方式。
// [S 汇编实证 0x1409c3e60 start area]：包装 -> scanStartMenuApps(desktop roots, dynamic=true)。
func (bs *BootstrapService) ScanDynamicDesktopApps() []StartMenuApp {
	return bs.scanStartMenuApps(composeStartMenuScanRoots(resolveWindowsDesktopRoots(), "desktop"), true)
}

// ---- 扫描主循环 ----

// scanStartMenuApps 遍历 start menu/desktop roots 扫描快捷方式应用。
// [S 汇编实证 0x1409c3ec0, 0x338B]：
//
//	参数（ABI 推断）：receiver(AX) + []startMenuRoot(roots.data=BX, roots.len=CX) + bool(dynamic, SIL)。
//	流程：逐 root -> TrimSpace -> classifyAutomaticWindowsPath=="local" -> os.Stat 存在 ->
//	filepath.WalkDir(闭包 func1=0x1409c4420 收集 + hydrate + attach 图标) ->
//	sort.Slice(比较闭包 func2=0x1409c4240)。
func (bs *BootstrapService) scanStartMenuApps(roots []startMenuRoot, dynamic bool) []StartMenuApp {
	if bs == nil || len(roots) == 0 {
		return nil
	}
	out := make([]StartMenuApp, 0, 64)
	var alloc idAllocator

	for _, root := range roots {
		rootPath := strings.TrimSpace(root.Path)
		if rootPath == "" {
			continue
		}
		class, _ := classifyAutomaticWindowsPath(rootPath, nil)
		if class != classLocal {
			continue
		}
		if fi, err := os.Stat(rootPath); err != nil || !fi.IsDir() {
			continue
		}

		_ = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d == nil || d.IsDir() {
				return nil
			}
			p := strings.TrimSpace(path)
			if p == "" || !isShortcutFilePath(p) {
				return nil
			}
			app := StartMenuApp{
				ID:     alloc.Next([]string{p}),
				Name:   resolveStartMenuScanDisplayName(p),
				Path:   p,
				Source: root.Source,
			}
			// 快捷方式明细填充（COM stub 期间仅 DisplayName 可用, [P]）
			hydrateStartMenuAppEntryWithResolvers(&app, defaultStartMenuAppResolvers())
			// 图标资源关联
			bs.attachStartMenuAppIconResources(&app, dynamic)
			out = append(out, app)
			return nil
		})
	}

	// [S] 尾部 sort.Slice：汇编常量闭包 func2=0x1409c4240；判序取路径大小写不敏感升序
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Path) < strings.ToLower(out[j].Path)
	})
	return out
}

// ---- 条目填充与图标关联 ----

// attachStartMenuAppIconResources 关联单条开始菜单应用的图标资源。
// [S 汇编实证 0x1409c4a40]：TrimSpace 条目两个图标源字段 -> buildLauncherIconResource
// （0x1408a4f20, bootstrap 图标资源域）-> 结果写入 IconData/IconURL 槽；refresh=false 时
// 解析到新资源后清空旧内联字段, refresh=true 时保留既有字段。
// 档位：[S] 控制流实证；[P] buildLauncherIconResource 的 asset 注册细节待 bootstrap 图标
// 资源域文件落地（此处直接落字段语义, 避免跨域重复定义）。
func (bs *BootstrapService) attachStartMenuAppIconResources(app *StartMenuApp, refresh bool) {
	if bs == nil || app == nil {
		return
	}
	_ = refresh
	// [P] .lnk 显式 IconLocation 需真实 COM 解析（shortcutinfo.go stub 返回空），
	// 离线桩期间按 目标路径->快捷方式路径 的顺序做路径级图标解析
	if iconData, err := resolveStartMenuIconData(app.Path, app.TargetPath, ""); err == nil && iconData != "" {
		app.IconData = iconData
	}
}

// hydrateStartMenuAppEntryWithResolvers 以解析器填充开始菜单条目。
// [S 汇编实证 0x1409c4d40, 0x3ffB]：
//
//	三组解析函数指针入参（shortcutInfo=0x470、displayName=0x478、fallbackName=0x480）：
//	TrimSpace(path) 空 -> 直接返回空结果；
//	isShortcutFilePath(path) -> shortcutInfo 解析器产出多字段结构, 逐字段 TrimSpace 落槽；
//	否则 displayName(path) 产出显示名；显示名为空再走 fallbackName 兜底。
func hydrateStartMenuAppEntryWithResolvers(entry *StartMenuApp, resolvers startMenuAppHydrationResolvers) {
	if entry == nil || strings.TrimSpace(entry.Path) == "" {
		return
	}
	path := strings.TrimSpace(entry.Path)

	if !isShortcutFilePath(path) {
		if name := invokeStartMenuNameResolver(resolvers.displayName, path); name != "" {
			entry.Name = name
		}
		return
	}
	info, err := resolvers.shortcutInfo(path)
	if err != nil {
		return
	}
	if name := strings.TrimSpace(info.DisplayName); name != "" {
		entry.Name = name
	} else if name := invokeStartMenuNameResolver(resolvers.fallbackName, path); name != "" {
		entry.Name = name
	}
	// [P] ShortcutInfo 尚为 stub（shortcutinfo.go 仅 DisplayName）；
	// TargetPath/Arguments/WorkingDir/IconLocation 字段在 go-ole 完整还原后在此落槽。
	entry.TargetPath = info.TargetPath
	entry.Arguments = info.Arguments
	entry.WorkingDir = info.WorkingDirectory
}

// ---- 内部类型 ----

// startMenuAppHydrationResolvers 定义条目水合的解析器集合（对齐 0x1409c4d40 三个函数指针参数）。
type startMenuAppHydrationResolvers struct {
	shortcutInfo func(string) (ShortcutInfo, error)
	displayName  func(string) string
	fallbackName func(string) string
}

// ---- 内部工具 ----

// defaultStartMenuAppResolvers 提供默认解析器绑定：COM stub + 纯字符串显示名兜底。
// [S-inline] 结构装配（3 函数指针，对齐 0x1409c4d40）；shortcutInfo 目标为 [P] COM stub。
func defaultStartMenuAppResolvers() startMenuAppHydrationResolvers {
	return startMenuAppHydrationResolvers{
		shortcutInfo: resolveShortcutInfoWithIconResolver,
		displayName:  resolveStartMenuScanDisplayName,
		fallbackName: resolveStartMenuScanDisplayName,
	}
}

// invokeStartMenuNameResolver 调用可能为 nil 的名字解析器。
// [S-inline] nil 检查 + TrimSpace 调用。
func invokeStartMenuNameResolver(resolve func(string) string, path string) string {
	if resolve == nil {
		return ""
	}
	return strings.TrimSpace(resolve(path))
}

// resolveStartMenuScanDisplayName 解析开始菜单扫描项的显示名（纯字符串取 basename + 剥扩展名）。
// [S-inline] TrimSpace 空返 ""；fallbackAppDisplayName(0x140748040) 非空则用之，否则原路径。
func resolveStartMenuScanDisplayName(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	if name := fallbackAppDisplayName(p); name != "" {
		return name
	}
	return p
}

// composeStartMenuScanRoots 将根路径列表装配为带来源标签的扫描根。
// [S-inline] 空路径跳过；startMenuRoot{Path,Source} 装配（元素 0x20B）。
func composeStartMenuScanRoots(paths []string, source string) []startMenuRoot {
	if len(paths) == 0 {
		return nil
	}
	roots := make([]startMenuRoot, 0, len(paths))
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		roots = append(roots, startMenuRoot{Path: strings.TrimSpace(p), Source: source})
	}
	return roots
}

// ---- 平台根路径 ----

// resolveWindowsStartMenuRoots 返回 Windows 开始菜单根路径列表。
// [S 汇编实证 0x1409c5180]：filepath.Join(PROGRAMDATA, Microsoft\Windows\Start Menu) +
// filepath.Join(APPDATA, ...)；环境变量缺失时跳过对应根（strings.TrimSpace 空判定）。
func resolveWindowsStartMenuRoots() []string {
	return []string{
		filepath.Join(os.Getenv("PROGRAMDATA"), "Microsoft", "Windows", "Start Menu"),
		filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu"),
	}
}

// resolveWindowsDesktopRoots 返回 Windows 桌面根路径列表。
// [S 汇编实证 0x1409c5480]：filepath.Join(PUBLIC, Desktop) + filepath.Join(USERPROFILE, Desktop)。
func resolveWindowsDesktopRoots() []string {
	return []string{
		filepath.Join(os.Getenv("PUBLIC"), "Desktop"),
		filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
	}
}
