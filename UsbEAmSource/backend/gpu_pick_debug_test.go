package main

import (
	"os"
	"testing"
)

// TestResolveGPUPickDebugConfiguration 黄金用例：锁定 env + os.Args 解析语义（汇编实证）。
func TestResolveGPUPickDebugConfiguration(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	cases := []struct {
		name     string
		fileEnv  string
		debugEnv string
		args     []string
		wantOn   bool
		wantPath string
	}{
		{"空环境无参数", "", "", []string{"app"}, false, ""},
		{"仅文件环境变量", "  C:\\logs\\gpu.log  ", "", []string{"app"}, true, "C:\\logs\\gpu.log"},
		{"仅调试开关 true", "", "yes", []string{"app"}, true, ""},
		{"仅调试开关 false", "C:\\x.log", "no", []string{"app"}, false, "C:\\x.log"},
		{"参数 --gpu-pick-debug", "", "", []string{"app", "--gpu-pick-debug"}, true, ""},
		{"参数 --gpu-pick-debug=off 覆盖文件 env", "C:\\x.log", "", []string{"app", "--gpu-pick-debug=off"}, false, "C:\\x.log"},
		{"参数 --gpu-pick-debug-file 覆盖路径", "C:\\default.log", "", []string{"app", "--gpu-pick-debug-file=C:\\override.log"}, true, "C:\\override.log"},
		{"参数 --gpu-pick-debug-file 空值不置位", "", "", []string{"app", "--gpu-pick-debug-file="}, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("USBEAM_GPU_PICK_DEBUG_FILE", c.fileEnv)
			t.Setenv("USBEAM_GPU_PICK_DEBUG", c.debugEnv)
			os.Args = c.args
			gotOn, gotPath := resolveGPUPickDebugConfiguration()
			if gotOn != c.wantOn || gotPath != c.wantPath {
				t.Fatalf("got (%v, %q), want (%v, %q)", gotOn, gotPath, c.wantOn, c.wantPath)
			}
		})
	}
}
