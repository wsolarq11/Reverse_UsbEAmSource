package main

import (
	"sort"
	"strings"
)

// webView2ProcessKindSortWeight 返回进程 kind 的排序权重。
// [S] 反汇编实证 0x1409deac0, 256B：
//
//	TrimSpace → ToLower，按 kind 分派：browser=0、renderer=1、gpu-process=2、
//	utility=3、其余=4。常量逐字节实证：browser(0x776f7262/0x6573/0x72)、
//	utility(0x6c697475/0x7469/0x79)、renderer(0x72657265646e6572)、
//	gpu-process(0x636f72702d757067/0x7365/0x73)。
func webView2ProcessKindSortWeight(kind string) int {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "browser":
		return 0
	case "renderer":
		return 1
	case "gpu-process":
		return 2
	case "utility":
		return 3
	default:
		return 4
	}
}

// extractWebView2ProcessType 从命令行提取 "--type=" 之后的进程类型。
// [S] 反汇编实证 0x1409de960, 352B：
//
//	ToLower → Index("--type=")（7B needle @0x140C395A0）。未命中返回 "browser"
//	（7B 常量 @0x140C39599）。命中则取 needle 后到首个空白（tab/LF/CR/space）之间的
//	token，TrimSpace 后再 Trim 掉双引号（1B cutset "\"" @0x1411CAC88）。
func extractWebView2ProcessType(commandLine string) string {
	s := strings.ToLower(commandLine)
	i := strings.Index(s, "--type=")
	if i < 0 {
		return "browser"
	}
	start := i + len("--type=")
	end := start
	for end < len(s) && s[end] != ' ' && s[end] != '\t' && s[end] != '\r' && s[end] != '\n' {
		end++
	}
	return strings.Trim(strings.TrimSpace(s[start:end]), "\"")
}

// normalizeWebView2ProcessKind 归一化进程 kind。
// [S] 反汇编实证 0x1409de840, 288B：
//
//	TrimSpace → ToLower；若为空则 extractWebView2ProcessType(commandLine)。
//	之后 ""/"browser" 归一为 "browser"（7B 常量 @0x140C39599），其余原样返回。
func normalizeWebView2ProcessKind(kind, commandLine string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" {
		kind = extractWebView2ProcessType(commandLine)
	}
	if kind == "" || kind == "browser" {
		return "browser"
	}
	return kind
}

// normalizeWebView2ProcessInfo 归一化单个 WebView2 进程信息。
// [S] 反汇编实证 0x1409de500, 832B：
//
//	kind 归一化（normalizeWebView2ProcessKind(kind, commandLine)）；
//	WorkingSetMB = WorkingSetBytes/1024/1024、PrivateMB = PrivateBytes/1024/1024
//	（float64 常量 0.0009765625 连乘两次，@0x1411CD608）；
//	三个 bool 采用「已置位则保持，否则检查 ToLower(commandLine) 是否含子串」：
//	"--js-flags="(11B)/"--disable-features="(19B)/"--renderer-process-limit="(25B)。
func normalizeWebView2ProcessInfo(info WebView2ProcessInfo) WebView2ProcessInfo {
	info.Kind = normalizeWebView2ProcessKind(info.Kind, info.CommandLine)
	info.WorkingSetMB = float64(info.WorkingSetBytes) / (1024 * 1024)
	info.PrivateMB = float64(info.PrivateBytes) / (1024 * 1024)

	cmd := strings.ToLower(info.CommandLine)
	info.HasJSFlags = info.HasJSFlags || strings.Contains(cmd, "--js-flags=")
	info.HasDisableFeatures = info.HasDisableFeatures || strings.Contains(cmd, "--disable-features=")
	info.HasRendererProcessLimit = info.HasRendererProcessLimit || strings.Contains(cmd, "--renderer-process-limit=")
	return info
}

// normalizeWebView2ProcessSnapshot 归一化进程快照：逐进程归一化、累计总量、按 kind 权重稳定排序。
// [S] 反汇编实证 0x1409dde00, 1472B：
//
//	Processes 为 nil 时置空切片（zerobase）；逐元素 normalizeWebView2ProcessInfo 并累计
//	WorkingSetBytes/PrivateBytes；sort.SliceStable 按 webView2ProcessKindSortWeight 升序，
//	权重相等时按 ProcessID 升序；TotalWorkingSetMB/TotalPrivateMB 由累计量 /1024/1024 得到。
func normalizeWebView2ProcessSnapshot(snapshot WebView2ProcessSnapshot) WebView2ProcessSnapshot {
	if snapshot.Processes == nil {
		snapshot.Processes = []WebView2ProcessInfo{}
	}
	var totalWorking, totalPrivate uint64
	for i := range snapshot.Processes {
		snapshot.Processes[i] = normalizeWebView2ProcessInfo(snapshot.Processes[i])
		totalWorking += snapshot.Processes[i].WorkingSetBytes
		totalPrivate += snapshot.Processes[i].PrivateBytes
	}
	sort.SliceStable(snapshot.Processes, func(i, j int) bool {
		wi := webView2ProcessKindSortWeight(snapshot.Processes[i].Kind)
		wj := webView2ProcessKindSortWeight(snapshot.Processes[j].Kind)
		if wi != wj {
			return wi < wj
		}
		return snapshot.Processes[i].ProcessID < snapshot.Processes[j].ProcessID
	})
	snapshot.TotalWorkingSetBytes = totalWorking
	snapshot.TotalPrivateBytes = totalPrivate
	snapshot.TotalWorkingSetMB = float64(totalWorking) / (1024 * 1024)
	snapshot.TotalPrivateMB = float64(totalPrivate) / (1024 * 1024)
	return snapshot
}

// InspectWebView2Processes 检视 WebView2 进程并返回归一化快照。
// [S] 反汇编实证 0x1409ddc60, 416B：
//
//	bs == nil → userDataDir = ""；否则 bs.workspaceSnapshot().WebView2Dir（偏移 +0x80）。
//	随后 inspectWebView2Processes(userDataDir)（依赖项 [P]，见 webview2_process_windows.go）。
func (bs *BootstrapService) InspectWebView2Processes() WebView2ProcessSnapshot {
	var userDataDir string
	if bs != nil {
		userDataDir = bs.workspaceSnapshot().WebView2Dir
	}
	return inspectWebView2Processes(userDataDir)
}
