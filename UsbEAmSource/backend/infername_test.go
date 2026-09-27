package main

import "testing"

func TestInferName(t *testing.T) {
	cases := map[string]string{
		`C:\a\b\app.exe`: "app",
		`/usr/bin/x`:     "x",
		`tool`:           "tool",
		`./x`:            "x",
		`a.b.c`:          "a.b",
		"":               "",
		`C:\a\`:          "",
	}
	for in, want := range cases {
		if got := inferName(in); got != want {
			t.Errorf("inferName(%q)=%q want %q", in, got, want)
		}
	}
}
