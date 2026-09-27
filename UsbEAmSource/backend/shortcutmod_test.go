package main

import (
	"reflect"
	"testing"
)

func TestCanonicalizeShortcutModifiers(t *testing.T) {
	if m, ok := canonicalizeShortcutModifiers("Ctrl+Alt"); !ok || !reflect.DeepEqual(m, []string{"Ctrl", "Alt"}) {
		t.Errorf("ctrl+alt -> %v,%v", m, ok)
	}
	if m, ok := canonicalizeShortcutModifiers("alt+WIN+ctrl"); !ok || !reflect.DeepEqual(m, []string{"Alt", "Win", "Ctrl"}) {
		t.Fatalf("dedup preserve=%v,%v", m, ok)
	}
	if m, ok := canonicalizeShortcutModifiers("Ctrl+Alt+Ctrl"); !ok || len(m) != 2 {
		t.Fatalf("dedup got %v", m)
	}
	if _, ok := canonicalizeShortcutModifiers("Ctrl+X"); ok {
		t.Error("non-modifier must be false")
	}
	if _, ok := canonicalizeShortcutModifiers("   "); ok {
		t.Error("blank must be false")
	}
}
