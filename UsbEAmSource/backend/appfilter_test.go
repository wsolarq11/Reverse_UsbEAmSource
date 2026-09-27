package main

import (
	"reflect"
	"testing"
)

func TestFilterDragLaunchAppIDs(t *testing.T) {
	apps := []AppEntry{
		{ID: "a1", EntryType: "app"},
		{ID: "a2", EntryType: "directory"},
		{ID: "a3", EntryType: "app"},
	}
	got := filterDragLaunchAppIDsByAppEntries([]string{"a1", "a2", "a3", "x9"}, apps)
	want := []string{"a1", "a3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filter got %v want %v", got, want)
	}
	if filterDragLaunchAppIDsByAppEntries(nil, nil) == nil {
		t.Error("must never return nil slice")
	}
}
