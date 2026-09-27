package main

import "testing"

func TestLauncherConfigIconDataError(t *testing.T) {
	var nilErr *launcherConfigIconDataError
	if got := nilErr.Error(); got != "CONFIG_ICON_DATA_INVALID" {
		t.Errorf("nil Error = %q", got)
	}
	e := &launcherConfigIconDataError{Path: "/a.png", Reason: "bad"}
	if got := e.Error(); got != "CONFIG_ICON_DATA_INVALID: /a.png: bad" {
		t.Errorf("Error = %q", got)
	}
	if e.Error() == "" {
		t.Error("must be non-empty")
	}
}
