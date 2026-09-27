package main

import (
	"sync"
	"testing"
)

// TestLauncherConfigStoreForPathConcurrent：并发访问同一归一化路径须返回同一单例（sync.Map 缓存），
// 不同路径返回不同实例；并发下不 panic。（AGENTS-22 并发覆盖）
func TestLauncherConfigStoreForPathConcurrent(t *testing.T) {
	const p = "C:/Concurrent/cfg.json"
	const n = 64
	res := make([]*launcherConfigStore, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res[i] = launcherConfigStoreForPath(p)
		}(i)
	}
	wg.Wait()
	first := res[0]
	if first == nil {
		t.Fatal("first result nil")
	}
	for i := 1; i < n; i++ {
		if res[i] != first {
			t.Fatalf("instance %d differs (concurrent singleton violated)", i)
		}
	}
	// 不同归一化路径 → 不同实例。
	if launcherConfigStoreForPath("C:/Other/cfg.json") == first {
		t.Error("different path must yield different instance")
	}
	// 同一路径不同大小写/空白 → 仍同一实例（PathKey 归一化）。
	if launcherConfigStoreForPath("c:/concurrent/cfg.json ") != first {
		t.Error("case/trim normalization must unify to same instance")
	}
}
