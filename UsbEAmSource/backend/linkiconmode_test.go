// AUTO-RECONSTRUCTED — DOMAIN: link icon mode normalize
// 研究用途. [S 汇编实证 0x14088b120, 320B]
package main

import "testing"

func TestNormalizeLinkIconMode_IconRefNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		iconMode string
		iconData string
		iconRef  string
		want     string
	}{
		{"iconRef non-empty returns upload", "favicon", "", "https://example.com/icon.ico", "upload"},
		{"iconRef non-empty with iconData still upload", "favicon", "data:image/png;base64,abc", "ref", "upload"},
		{"iconRef empty, iconData matches prefix returns upload", "", "data:image/png;base64,abc", "", "upload"},
		{"iconRef empty, iconData short (<11) returns favicon", "", "short", "", "favicon"},
		{"iconRef empty, iconData non-matching returns favicon", "", "https://example.com/icon.png", "", "favicon"},
		{"iconRef empty, iconData empty returns favicon", "", "", "", "favicon"},
		{"iconRef empty, iconData whitespace returns favicon", "", "  ", "", "favicon"},
		{"iconMode ignored (result redundant), iconRef empty returns favicon", "upload", "", "", "favicon"},
		{"iconRef trimmed whitespace empty returns favicon", "", "", "  ", "favicon"},
		{"iconRef whitespace non-empty after trim returns upload", "", "", "  ref  ", "upload"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLinkIconMode(tt.iconMode, tt.iconData, tt.iconRef)
			if got != tt.want {
				t.Errorf("normalizeLinkIconMode(%q, %q, %q) = %q, want %q",
					tt.iconMode, tt.iconData, tt.iconRef, got, tt.want)
			}
		})
	}
}

func TestNormalizeLinkIconMode_EdgeCases(t *testing.T) {
	// iconData exactly 11 chars matching prefix "data:image/"
	got := normalizeLinkIconMode("", "data:image/", "")
	if got != "upload" {
		t.Errorf("exact 11-char prefix match: got %q, want upload", got)
	}

	// iconData 11 chars with different 11th char: "data:images" != "data:image/"
	got = normalizeLinkIconMode("", "data:images", "")
	if got != "favicon" {
		t.Errorf("11-char non-matching prefix: got %q, want favicon", got)
	}
}
