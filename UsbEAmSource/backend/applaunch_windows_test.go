package main

import (
	"errors"
	"testing"

	"golang.org/x/sys/windows"
)

// startApplicationUsingCurrentPrivileges 空 shellTarget 分支为纯逻辑（无系统副作用），实测。
func TestStartApplicationUsingCurrentPrivilegesEmptyAppName(t *testing.T) {
	err := startApplicationUsingCurrentPrivileges(launchContext{entryType: "app", shellTarget: "   "})
	if err == nil {
		t.Fatal("空 shellTarget 应返回错误")
	}
	if err.Error() != "入口路径不能为空" {
		t.Fatalf("错误消息不符: %q", err.Error())
	}
}

// startApplication 空 shellTarget 分支（resolveLaunchableAppEntry 后）为纯逻辑，实测。
func TestStartApplicationEmptyAppName(t *testing.T) {
	err := startApplication("admin", launchContext{entryType: "app", shellTarget: "   "})
	if err == nil {
		t.Fatal("空 shellTarget 应返回错误")
	}
	if err.Error() != "入口路径不能为空" {
		t.Fatalf("错误消息不符: %q", err.Error())
	}
}

// retryLaunchAsAdminIfElevationRequired 非提权错误应原样返回（纯逻辑分支）。
func TestRetryLaunchAsAdminIfElevationRequiredPassthrough(t *testing.T) {
	sentinel := errors.New("普通失败")
	ctx := launchContext{entryType: "app", shellTarget: "x.exe"}
	if got := retryLaunchAsAdminIfElevationRequired(sentinel, ctx); got != sentinel {
		t.Fatalf("非提权错误应原样返回, got %v", got)
	}
}

// retryLaunchAsAdminIfElevationRequired 对 740 判据：ERROR_ELEVATION_REQUIRED 应被识别。
// 真分支会触发提权重试（依赖 isProcessElevated/ShellExecute），此处仅锁定判据不误报。
func TestRetryLaunchAsAdminIfElevationRequiredDetects740(t *testing.T) {
	if !isElevationRequiredError(windows.ERROR_ELEVATION_REQUIRED) {
		t.Fatal("ERROR_ELEVATION_REQUIRED 应被识别为提权错误")
	}
}
