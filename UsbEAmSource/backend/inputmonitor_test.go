package main

import "testing"

func TestNewInputMonitorService(t *testing.T) {
	svc := newInputMonitorService()
	if svc == nil {
		t.Fatal("factory must return non-nil service")
	}
	// 装配：events 大缓冲、map 已建、maxEvents/eventSize=0x1000。
	if svc.events == nil || len(svc.events) != 0x1000 {
		t.Errorf("events cap=%d want 0x1000; nil=%v", len(svc.events), svc.events == nil)
	}
	if svc.pressedKeys == nil || svc.owners == nil {
		t.Error("maps must be initialized")
	}
	if svc.maxEvents != 0x1000 || svc.eventSize != 0x1000 {
		t.Errorf("limits maxEvents=%d eventSize=%d want 0x1000", svc.maxEvents, svc.eventSize)
	}
	// initialize 已置平台健康托管标志。
	if !svc.platformHealthManaged {
		t.Error("platformHealthManaged must be set after init")
	}
	// 幂等：重复 initialize 不 panic。
	inputMonitorInitialize(svc)
}

func TestInputMonitorInitializeNil(t *testing.T) {
	// nil receiver 应安全返回。
	inputMonitorInitialize(nil)
}
