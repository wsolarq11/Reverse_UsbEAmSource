package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// resolveRegisteredLinkBrowserPath 从注册表 App Paths 解析浏览器可执行路径。
// [S 汇编 0x1408d1740]：key = "Software\Microsoft\Windows\CurrentVersion\App Paths\" + TrimSpace(name)
// （前缀 @0x140c899b2,52B）；依次试 registry.CURRENT_USER(0x80000001)、registry.LOCAL_MACHINE(0x80000002)，
// readRegistryString(hive, key, "") → resolveExistingExecutablePath([]string{value})；命中返回非空则返回，否则 ""。
func resolveRegisteredLinkBrowserPath(name string) string {
	key := `Software\Microsoft\Windows\CurrentVersion\App Paths\` + strings.TrimSpace(name)
	for _, hive := range []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE} {
		value := readRegistryString(hive, key, "")
		if p := resolveExistingExecutablePath([]string{value}); p != "" {
			return p
		}
	}
	return ""
}

// readRegistryString 读注册表字符串值并清理引号与空白。
// [S 汇编 0x1408d1820]：registry.OpenKey(hkey, key, QUERY_VALUE=1) 失败→""；defer k.Close()（deferwrap1 → RegCloseKey）；
// k.GetStringValue(valueName) 失败→""；否则 TrimSpace(Trim(v, "\""))（cutset 单字节 0x22 @0x1411cac88）。
func readRegistryString(hkey registry.Key, key, valueName string) string {
	k, err := registry.OpenKey(hkey, key, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, err := k.GetStringValue(valueName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(v, "\""))
}

// resolveExistingExecutablePath 返回首个存在且非目录的可执行路径。
// [S 汇编 0x1408d19e0]：遍历 paths（rax=ptr/rbx=len/rcx=cap），逐项 TrimSpace(Trim(p,"\""));
// 空则跳过；os.Stat 失败或 info.IsDir()（itab[0x18]）则跳过；命中返回该路径，否则 ""。
func resolveExistingExecutablePath(paths []string) string {
	for _, p := range paths {
		p = strings.TrimSpace(strings.Trim(p, "\""))
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			continue
		}
		return p
	}
	return ""
}

// buildEnvLinkBrowserPath 从环境变量取值并拼接到路径列表末尾。
// [S 汇编 0x1408d1ac0]：v = TrimSpace(os.Getenv(name))；空→""；否则 filepath.Join(append(paths, v)...)
// （growslice 扩容 len+1 → typedslicecopy → filepath.Join）。
func buildEnvLinkBrowserPath(name string, paths []string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return ""
	}
	return filepath.Join(append(paths, v)...)
}
