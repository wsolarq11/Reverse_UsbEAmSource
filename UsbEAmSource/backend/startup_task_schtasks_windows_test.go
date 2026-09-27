package main

import "testing"

func TestCreateLauncherStartupTaskUsingSchtasksEmptyExe(t *testing.T) {
	err := createLauncherStartupTaskUsingSchtasks("  ", "", "user", true, 5)
	if err == nil || err.Error() != "开机启动程序路径不能为空" {
		t.Fatalf("空 exe 应传播入口路径错误, got %v", err)
	}
}

func TestCreateLauncherStartupTaskUsingSchtasksEmptyUserID(t *testing.T) {
	err := createLauncherStartupTaskUsingSchtasks(`C:\tools\app.exe`, "", " ", true, 5)
	if err == nil || err.Error() != "开机启动任务用户标识不能为空" {
		t.Fatalf("空 userID 应传播用户标识错误, got %v", err)
	}
}
