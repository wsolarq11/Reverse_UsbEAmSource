// AUTO-RECONSTRUCTED — DOMAIN: Windows automatic path classification
// 研究用途. 实证依据：
//
//	classifyAutomaticWindowsPath          0x14075b180, 1888B
//	classifyAutomaticWindowsPathLexically 0x14075b8e0, 1504B
//	normalizeAutomaticWindowsReparseDestination 0x14075bae0, 480B
//	  — 纯字符串规范化（无 syscall）：NT substitute 名 → Win32 路径
//	automaticWindowsDriveType  0x14075bcc0, 128B — GetDriveTypeW 薄封装
//	automaticWindowsPathAttributes 0x14075bd40, 96B — GetFileAttributesW 薄封装
//	.rdata strings decoded at VAs listed per label.
//
// 汇编确定性语义：
//
//	classifyLexically: TrimSpace → Replace("/→\") → ToLower → 前缀线性匹配
//	classify(path):
//	  1) TrimSpace → empty → ("", "empty")
//	  2) Lex → non-local → (trimmed, lex)
//	  3) "local" → 主循环(最多32轮):
//	     a) Clean→ToLower→dedup(cache)→re-lex
//	     b) callback 返回:
//	        ≤1 → "missing-drive"
//	        =4 → "remote-drive"(i==0) / "reparse-remote"(i>0)
//	        2/3 → VolName→Rel→Split→逐元素 cb.Probe→reparse
//	           成功解析 → jmp 主循环顶（下一轮）; 否则 → "local"
//	     c) 内层循环流程: vol + part 拼接 → callback Probe → 错误处理
//	        → 非 local lex → normalizeReparse → device 检测 → remote-d 前缀检测
package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Classification label constants decoded from .rdata.
const (
	classEmpty         = "empty"          // 0x140C35BCD len=5
	classLocal         = "local"          // 0x140C35BD2 len=5
	classDevice        = "device"         // 0x140C375C6 len=6
	classUNC           = "unc"            // 0x140C33B51 len=3
	classRelative      = "relative"       // 0x140C3C0C4 len=8
	classMissingDrive  = "missing-drive"  // 0x140C4DF8B len=13
	classRemoteDrive   = "remote-drive"   // 0x140C4B900 len=12
	classReparseUnsafe = "reparse-unsafe" // 0x140C506B0 len=14
	classReparseRemote = "reparse-remote" // 0x140C506BE len=14
	classUnavailable   = "unavailable"    // 0x140C4745D len=11
)

// PathAttrsCallback returns an attribute value for a path.
// Return value semantics (inferred from callers):
//
//	0: S_NotFound  → "missing-drive"
//	1: S_Unknown   → "missing-drive"
//	2: S_Drive     → enter parts resolution
//	3: S_Volume    → enter parts resolution
//	4: S_Net       → "remote-drive" / "reparse-remote"
type PathAttrsCallback func(path string) int

// PathProbeCallback returns attributes + error for reparse-resolution.
type PathProbeCallback func(path string) (int, error)

// AutomaticPathCallbacks bundles the two closure structs used in the binary.
type AutomaticPathCallbacks struct {
	Attrs PathAttrsCallback
	Probe PathProbeCallback
}

// classifyAutomaticWindowsPathLexically classifies a Windows path
// by its lexical form alone (no filesystem access).
// [S 汇编 0x14075b8e0, 1504B]：TrimSpace → Replace("/→\") → ToLower → 前缀线性匹配。
//
// ASM 0x14075b8e0: TrimSpace → Replace("/→\") → ToLower → prefix checks:
//
//	\??\, \\.\, \\?\, \device\ → "device"
//	\\ → "unc"
//	X:[\/] → "local" if remaining after X: contains \, else "relative"
func classifyAutomaticWindowsPathLexically(path string) string {
	p := strings.TrimSpace(path)
	p = strings.Replace(p, "/", "\\", -1)
	p = strings.ToLower(p)
	n := len(p)

	// Win32 device namespace prefixes (4 bytes, dword-cmp in asm):
	//   \??\  0x5c3f3f5c  \\.\  0x5c2e5c5c  \\?\  0x5c3f5c5c
	if n >= 4 {
		w := p[:4]
		if w == `\??\` || w == `\\.\` || w == `\\?\` {
			return classDevice
		}
	}

	// NT device path prefix (8 bytes): \device\
	if n >= 8 && p[:8] == `\device\` {
		return classDevice
	}

	// UNC: begins with "\\"
	if n >= 2 && p[0] == '\\' && p[1] == '\\' {
		return classUNC
	}

	// Drive letter: X: followed by \ or /
	// (asm skips 2 bytes "X:" and searches remaining for \)
	if n >= 3 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
		ch := p[0]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			if strings.Contains(p[2:], "\\") {
				return classLocal
			}
			return classRelative
		}
	}

	return classRelative
}

// classifyAutomaticWindowsPath re-evaluates a path with filesystem callbacks,
// walking parent directories and resolving reparse points.
// [S 汇编 0x14075b180, 1888B]：TrimSpace 空→empty；Lex 非 local 直返；主循环(32 轮) Clean→dedup→re-lex→回调。
//
// ASM 0x14075b180:
//  1. TrimSpace, empty → ("", "empty") [0x14075b2c1]
//  2. Lexically classify; non-"local" → (trimmed, lex)
//  3. check cb not nil; construct seen map → main loop
//  4. Main loop (32 max): Clean→ToLower→dedup→re-lex→callback→...
//  5. callback return:
//     ≤1 → "missing-drive"
//     =4 → conditional "remote-drive"(i==0)/"reparse-remote"(i>0)
//     else → parts resolution → reparse cycle
//  6. Inner parts loop: build candidate=base+part, probe, if non-local → normalize reparse
func classifyAutomaticWindowsPath(rawPath string, cb *AutomaticPathCallbacks) (class string, cleaned string) {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return "", classEmpty
	}

	lex := classifyAutomaticWindowsPathLexically(trimmed)
	if lex != classLocal {
		return trimmed, lex
	}

	// Without callbacks or with Attrs nil → return as-is
	if cb == nil || cb.Attrs == nil {
		return trimmed, classLocal
	}

	seen := make(map[string]struct{}, 32)
	workPath := trimmed

	for i := 0; i < 32; i++ {
		cleanedPath := filepath.Clean(workPath)
		lowered := strings.ToLower(cleanedPath)

		// Dedup check (mapaccess2_faststr)
		if _, ok := seen[lowered]; ok {
			return cleanedPath, classReparseUnsafe
		}
		seen[lowered] = struct{}{}

		// Re-lex after Clean
		lex2 := classifyAutomaticWindowsPathLexically(cleanedPath)
		if lex2 != classLocal {
			if i > 0 {
				return cleanedPath, classReparseUnsafe
			}
			return cleanedPath, lex2
		}

		// Callback Attrs call
		attr := cb.Attrs(cleanedPath)

		switch {
		case attr <= 1:
			return cleanedPath, classMissingDrive

		case attr == 4:
			if i > 0 {
				return cleanedPath, classReparseRemote
			}
			return cleanedPath, classRemoteDrive

		default: // attr ∈ {2, 3}
			// Resolve relative path via VolumeName + "\" + Rel
			// ASM: concatstring2(VolumeName(path), "\\") → Filepath.Rel(cleaned, concat)
			vol := filepath.VolumeName(cleanedPath)
			if vol == "" {
				return cleanedPath, classLocal
			}
			basePart := vol + `\`

			rel, err := filepath.Rel(basePart, cleanedPath)
			if err != nil || (len(rel) > 0 && rel[0] == '.') {
				return cleanedPath, classLocal
			}

			parts := strings.Split(cleanedPath, `\`)
			if len(parts) < 2 {
				return cleanedPath, classLocal
			}

			// baseAccum starts as VolumeName + "\" (first part always the volume)
			base := vol + `\`

			// Inner parts loop
			for _, part := range parts[1:] {
				if part == "" || part == "." {
					continue
				}

				candidate := filepath.Join(base, part)

				// Second callback: Probe
				if cb.Probe != nil {
					_, probeErr := cb.Probe(candidate)
					if probeErr != nil {
						if errors.Is(probeErr, fs.ErrNotExist) {
							return cleanedPath, classLocal
						}
						return cleanedPath, classUnavailable
					}
				}

				// Re-evaluate candidate lexically
				cl2 := classifyAutomaticWindowsPathLexically(candidate)
				if cl2 == classLocal {
					// Still local, advance base and continue
					base = candidate
					continue
				}

				// Non-local result → reparse point detected
				target := normalizeAutomaticWindowsReparseDestination(candidate)
				if !filepath.IsAbs(target) {
					dir := filepath.Dir(candidate)
					target = filepath.Join(dir, target)
				}

				lex3 := classifyAutomaticWindowsPathLexically(target)
				if lex3 == classDevice {
					if i > 0 {
						return cleanedPath, classReparseRemote
					}
					return target, classDevice
				}

				// Inline "remote-drive" / "remote-d" detection
				if strings.HasPrefix(target, "remote-drive") ||
					(strings.HasPrefix(target, "remote-d") &&
						len(target) >= 12 && target[:12] == "remote-drive") {
					return target, classReparseUnsafe
				}

				// Non-local, non-device reparse → advance main loop
				workPath = candidate
				goto nextIteration
			}

			// All parts processed without reparse → return local
			return cleanedPath, classLocal
		}

	nextIteration:
	}

	return workPath, classReparseUnsafe
}

// normalizeAutomaticWindowsReparseDestination normalizes an NT reparse
// substitute name back to Win32 path form.
// [S 汇编 0x14075bae0, 480B]：纯字符串规范化（无 syscall），NT substitute 名 → Win32 路径。
//
// ASM 0x14075bae0, 480B: NO syscalls. Pure string transform:
//
//	TrimSpace → ToLower(prefix compare only) → strip namespace prefix:
//	  \??\unc\...  or  \\?\unc\...  →  \\ + s[8:]
//	  \??\...      or  \\?\...      →  s[4:]
//	  otherwise → trimmed as-is
//
// Prefix match is case-insensitive (compared on ToLower copy), while the
// returned suffix keeps the original case of the trimmed input.
func normalizeAutomaticWindowsReparseDestination(path string) string {
	trimmed := strings.TrimSpace(path)
	lower := strings.ToLower(trimmed)

	if strings.HasPrefix(lower, `\??\unc\`) {
		return `\\` + trimmed[8:]
	}
	if strings.HasPrefix(lower, `\\?\unc\`) {
		return `\\` + trimmed[8:]
	}
	if strings.HasPrefix(lower, `\??\`) {
		return trimmed[4:]
	}
	if strings.HasPrefix(lower, `\\?\`) {
		return trimmed[4:]
	}
	return trimmed
}

// automaticWindowsDriveType wraps kernel32.GetDriveTypeW via LazyProc.
// [S 汇编 0x14075bcc0, 128B]：UTF16PtrFromString → GetDriveTypeW 薄封装。
//
// ASM 0x14075bcc0: UTF16PtrFromString → LazyProc.Call(GetDriveTypeW) → uint32
// Returns DRIVE_UNKNOWN(0), DRIVE_NO_ROOT_DIR(1), DRIVE_REMOVABLE(2),
// DRIVE_FIXED(3), DRIVE_REMOTE(4), DRIVE_CDROM(5), or DRIVE_RAMDISK(6).
// On error (invalid path), returns 0.
func automaticWindowsDriveType(path string) uint32 {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}
	return windows.GetDriveType(ptr)
}

// automaticWindowsPathAttributes wraps kernel32.GetFileAttributesW.
// [S 汇编 0x14075bd40, 96B]：UTF16PtrFromString → GetFileAttributesW 薄封装。
//
// ASM 0x14075bd40: UTF16PtrFromString → windows.GetFileAttributes → uint32
// Returns FILE_ATTRIBUTE_* flags (e.g. FILE_ATTRIBUTE_DIRECTORY=0x10,
// FILE_ATTRIBUTE_REPARSE_POINT=0x400). UTF16 conversion failure returns 0;
// GetFileAttributes failure is returned as INVALID_FILE_ATTRIBUTES
// (0xffffffff) because the ASM ignores the error result.
func automaticWindowsPathAttributes(path string) uint32 {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}
	attrs, _ := windows.GetFileAttributes(ptr)
	return attrs
}
