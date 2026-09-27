package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLauncherNetworkAccessFactory(t *testing.T) {
	// 空路径 → directLauncherNetworkAccess。 [S 汇编实证]
	if na := newLauncherNetworkAccess(""); na == nil {
		t.Fatal("empty path must still return access")
	}
	if _, direct := newLauncherNetworkAccess("").(*directLauncherNetworkAccess); !direct {
		t.Error("empty path must yield directLauncherNetworkAccess")
	}

	// 非空路径 → configBacked（TrimSpace 后存 configPath）。
	cb, ok := newLauncherNetworkAccess("  C:/cfg.json ").(*configBackedLauncherNetworkAccess)
	if !ok {
		t.Fatal("non-empty path must yield configBackedLauncherNetworkAccess")
	}
	if cb.configPath != "C:/cfg.json" {
		t.Errorf("configPath=%q want trimmed %q", cb.configPath, "C:/cfg.json")
	}
}

func TestNetworkAccessDoGet(t *testing.T) {
	// 用 httptest 验证 direct 的 Get 真实请求。
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	na := newLauncherNetworkAccess("")
	resp, err := na.Get(srv.URL, 1024)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Get err=%v status=%d", err, resp.StatusCode)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("Get body=%q want ok", body)
	}
}
