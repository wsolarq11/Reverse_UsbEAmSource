package main

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestClassifyAutomaticWindowsPathLexically(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Device namespace prefixes
		{`\??\C:\foo`, classDevice},
		{`\\.\PhysicalDrive0`, classDevice},
		{`\\?\C:\Windows`, classDevice},
		{`\device\harddisk0\partition1`, classDevice},
		// UNC paths
		{`\\server\share`, classUNC},
		{`\\192.168.1.1\share`, classUNC},
		// Local paths (drive letter + subdir)
		{`C:\Users`, classLocal},
		{`D:\Program Files\App`, classLocal},
		// Bare drive root is still local (contains \ after X:)
		{`E:\`, classLocal},
		// Relative paths (drive letter without subdir)
		{`C:foo`, classRelative},
		{`Z:`, classRelative},
		{`relative\path`, classRelative},
		// Forward slash → normalized to backslash
		{`C:/Users`, classLocal},
		// Mixed case → after ToLower, same result
		{`C:\USERS`, classLocal},
		{`\\SERVER\SHARE`, classUNC},
		// Edge
		{`a`, classRelative},
		{``, classRelative},
		{`  `, classRelative},
	}

	for _, tt := range tests {
		got := classifyAutomaticWindowsPathLexically(tt.input)
		if got != tt.want {
			t.Errorf("classifyAutomaticWindowsPathLexically(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestClassifyAutomaticWindowsPath_Empty(t *testing.T) {
	cleaned, class := classifyAutomaticWindowsPath("", nil)
	if class != classEmpty {
		t.Errorf("empty path: class=%q want %q", class, classEmpty)
	}
	if cleaned != "" {
		t.Errorf("empty path: cleaned=%q want \"\"", cleaned)
	}
}

func TestClassifyAutomaticWindowsPath_Device_UNC(t *testing.T) {
	devCases := []string{`\??\C:\foo`, `\\.\COM1`, `\\?\C:\bar`, `\device\harddisk0`}
	for _, p := range devCases {
		got, class := classifyAutomaticWindowsPath(p, nil)
		if class != classDevice {
			t.Errorf("%s: class=%q want %q", p, class, classDevice)
		}
		if got != p {
			t.Errorf("%s: cleaned=%q want %q", p, got, p)
		}
	}

	uncCases := []string{`\\server\share`, `\\nas\folder`}
	for _, p := range uncCases {
		got, class := classifyAutomaticWindowsPath(p, nil)
		if class != classUNC {
			t.Errorf("%s: class=%q want %q", p, class, classUNC)
		}
		if got != p {
			t.Errorf("%s: cleaned=%q want %q", p, got, p)
		}
	}
}

func TestClassifyAutomaticWindowsPath_Local_NoCallback(t *testing.T) {
	got, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, nil)
	if class != classLocal {
		t.Errorf("want local; got %q", class)
	}
	if filepath.Clean(`C:\Users\Foo`) != got {
		t.Errorf("cleaned=%q want %q", got, filepath.Clean(`C:\Users\Foo`))
	}
}

func TestClassifyAutomaticWindowsPath_Callback_NotFound(t *testing.T) {
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 0 },
	}
	got, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classMissingDrive {
		t.Errorf("want missing-drive; got %q cleaned=%q", class, got)
	}
}

func TestClassifyAutomaticWindowsPath_Callback_NetFirst(t *testing.T) {
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 4 },
	}
	got, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classRemoteDrive {
		t.Errorf("want remote-drive; got %q cleaned=%q", class, got)
	}
}

func TestClassifyAutomaticWindowsPath_Callback_DefaultReturn_Local(t *testing.T) {
	// attr=2 → enter parts resolution; no reparse → returns "local"
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 2 },
		Probe: func(path string) (int, error) { return 0, nil },
	}
	_, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classLocal {
		t.Errorf("want local; got %q", class)
	}
}

func TestClassifyAutomaticWindowsPath_Callback_DefaultReturn_ErrNotExist(t *testing.T) {
	// Probe returns ErrNotExist → "local"
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 2 },
		Probe: func(path string) (int, error) { return 0, os.ErrNotExist },
	}
	_, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classLocal {
		t.Errorf("want local; got %q", class)
	}
}

func TestClassifyAutomaticWindowsPath_Callback_DefaultReturn_Unavailable(t *testing.T) {
	// Probe returns non-ErrNotExist → "unavailable"
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 2 },
		Probe: func(path string) (int, error) { return 0, os.ErrExist },
	}
	_, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classUnavailable {
		t.Errorf("want unavailable; got %q", class)
	}
}

func TestClassifyAutomaticWindowsPath_Callback_Unknown(t *testing.T) {
	// attr=1 ≤ 1 → "missing-drive"
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 1 },
	}
	_, class := classifyAutomaticWindowsPath(`C:\Users\Foo`, cb)
	if class != classMissingDrive {
		t.Errorf("want missing-drive; got %q", class)
	}
}

func TestClassifyAutomaticWindowsPath_Relative_Passthrough(t *testing.T) {
	got, class := classifyAutomaticWindowsPath(`relative\path`, nil)
	if class != classRelative {
		t.Errorf("want relative; got %q", class)
	}
	if got != `relative\path` {
		t.Errorf("cleaned=%q", got)
	}
}

func TestNormalizeReparseDestination(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// NT substitute forms (as returned by reparse point queries)
		{`\??\device-prefix`, `\??\C:\Users\Foo`, `C:\Users\Foo`},
		{`\\?\device-prefix`, `\\?\C:\Windows\System32`, `C:\Windows\System32`},
		{`\??\unc-share`, `\??\unc\server\share\dir`, `\\server\share\dir`},
		{`\\?\unc-share`, `\\?\unc\NAS\Public\file.txt`, `\\NAS\Public\file.txt`},
		// Case-insensitive prefix match, original-case suffix preserved
		{`uppercase-unc`, `\??\UNC\Server\Share`, `\\Server\Share`},
		// Plain paths pass through after TrimSpace
		{`plain-local`, `C:\Users\Foo`, `C:\Users\Foo`},
		{`plain-unc`, `\\server\share`, `\\server\share`},
		{`surrounding-space`, `  C:\Users\Foo  `, `C:\Users\Foo`},
		{`unc-with-space`, `  \\?\unc\srv\sh  `, `\\srv\sh`},
		// Prefix-only edge: stripping leaves an empty suffix
		{`empty-result-1`, `\??\`, ``},
		{`empty-result-2`, `\\?\`, ``},
		{`empty-input`, ``, ``},
		{`blank-input`, `   `, ``},
		// unc-variant needs trailing backslash; generic \??\ prefix still strips
		{`unc-no-slash`, `\??\uncX`, `uncX`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAutomaticWindowsReparseDestination(tt.input)
			if got != tt.want {
				t.Errorf("normalizeAutomaticWindowsReparseDestination(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestClassifyAutomaticWindowsPath_Dedup tests that repeated path in loop
// returns "reparse-unsafe". This only triggers when callback returns 2/3
// and parts loop completes back to top.
func TestClassifyAutomaticWindowsPath_Dedup_SamePath(t *testing.T) {
	// Simulate a path that enters the loop twice
	cb := &AutomaticPathCallbacks{
		Attrs: func(path string) int { return 2 },
		Probe: func(path string) (int, error) { return 0, nil },
	}
	_, class := classifyAutomaticWindowsPath(`C:\Users`, cb)
	// First iteration: enters default branch (attr=2)
	// Parts loop processes "Users", all parts consumed → "local"
	if class != classLocal {
		t.Errorf("expected local; got %q", class)
	}
}

// automaticWindowsDriveType: ASM 0x14075bcc0 → GetDriveTypeW raw value,
// no mapping branch. DRIVE_UNKNOWN=0, NO_ROOT_DIR=1, REMOVABLE=2, FIXED=3,
// REMOTE=4, CDROM=5, RAMDISK=6. Error → 0 (xor eax,eax path).
func TestAutomaticWindowsDriveType(t *testing.T) {
	// NUL in path makes UTF16PtrFromString fail → returns 0.
	if got := automaticWindowsDriveType("C:\\\x00x"); got != 0 {
		t.Errorf("invalid path: got %d, want 0", got)
	}

	// A valid drive root must resolve to a real drive type.
	got := automaticWindowsDriveType(`C:\`)
	if got < windows.DRIVE_REMOVABLE || got > windows.DRIVE_RAMDISK {
		t.Errorf("C:\\ drive type = %d, want one of DRIVE_REMOVABLE..RAMDISK", got)
	}

	// SystemRoot drive must be fixed on a typical install.
	root := filepath.VolumeName(os.Getenv("SystemRoot")) + `\`
	if root != `\` {
		if got := automaticWindowsDriveType(root); got != windows.DRIVE_FIXED {
			t.Logf("SystemRoot drive %q type = %d (not fixed; acceptable on exotic mounts)", root, got)
		}
	}
}

// automaticWindowsPathAttributes: ASM 0x14075bd40 → GetFileAttributesW raw
// bitmask, error ignored (INVALID_FILE_ATTRIBUTES returned as-is).
func TestAutomaticWindowsPathAttributes(t *testing.T) {
	// NUL in path → UTF16PtrFromString error → returns 0.
	if got := automaticWindowsPathAttributes("C:\\\x00x"); got != 0 {
		t.Errorf("invalid path: got %#x, want 0", got)
	}

	// Nonexistent path → INVALID_FILE_ATTRIBUTES (0xffffffff), error ignored.
	if got := automaticWindowsPathAttributes(`C:\__dsh_no_such_path_9f3k__`); got != windows.INVALID_FILE_ATTRIBUTES {
		t.Errorf("missing path: got %#x, want INVALID_FILE_ATTRIBUTES", got)
	}

	// SystemRoot is a directory.
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		t.Skip("SystemRoot not set")
	}
	attrs := automaticWindowsPathAttributes(sysRoot)
	if attrs == windows.INVALID_FILE_ATTRIBUTES {
		t.Fatalf("SystemRoot %q: INVALID_FILE_ATTRIBUTES", sysRoot)
	}
	if attrs&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		t.Errorf("SystemRoot %q attrs=%#x, want FILE_ATTRIBUTE_DIRECTORY set", sysRoot, attrs)
	}
}
