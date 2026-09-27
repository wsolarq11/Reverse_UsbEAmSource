// AUTO-RECONSTRUCTED TYPES — DOMAIN: network
// 研究用途
package main

import (
	"net/http"
)

type directLauncherNetworkAccess struct{}

type LauncherNetworkAccess interface {
	Do(*http.Request, int64) (*http.Response, error)
	Get(string, int64) (*http.Response, error)
}

type LauncherNetworkRedirectPolicyAccess interface {
	DoWithRedirectPolicy(*http.Request, int64, func(*http.Request, []*http.Request) error) (*http.Response, error)
}

type configBackedLauncherNetworkAccess struct {
	configPath string
}
