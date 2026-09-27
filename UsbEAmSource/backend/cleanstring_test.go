package main

import (
	"reflect"
	"testing"
)

func TestCleanStringList(t *testing.T) {
	got := cleanStringList([]string{"  A ", "Beta", "beta", "", " a ", "c"})
	want := []string{"A", "Beta", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
	if out := cleanStringList(nil); out == nil {
		t.Error("nil input must yield non-nil slice")
	}
}
