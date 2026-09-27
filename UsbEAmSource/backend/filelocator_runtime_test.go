package main

import (
	"context"
	"testing"
	"time"
)

// TestAcquireFileLocatorContentScanNilReceiver 实证 0x1407d23e0 的 s==nil 分支：
// 直接返回 context.Canceled（.data 0x141bc4520 字符串 "context canceled"）。
func TestAcquireFileLocatorContentScanNilReceiver(t *testing.T) {
	var s *fileLocatorService
	if err := s.acquireFileLocatorContentScan(context.Background()); err != context.Canceled {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

// TestAcquireFileLocatorContentScanCancelled 实证 select 的 <-ctx.Done() 分支：
// 先占满单缓冲信号量，再以已取消的 ctx 调用，select 只剩 recv 就绪 → ctx.Err()。
func TestAcquireFileLocatorContentScanCancelled(t *testing.T) {
	s := &fileLocatorService{}
	if err := s.acquireFileLocatorContentScan(context.Background()); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.acquireFileLocatorContentScan(ctx); err != context.Canceled {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

// TestWaitIfPausedNotRunning 实证 0x1407d09a0 的收尾逻辑：
// state.Running==false 且 generation 未变、ctx 未取消 → context.Canceled。
func TestWaitIfPausedNotRunning(t *testing.T) {
	s := &fileLocatorService{generation: 7}
	if err := s.waitIfPaused(context.Background(), 7, time.Time{}); err != context.Canceled {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

// TestWaitIfPausedGenerationChanged 实证 generation 变 → context.Canceled。
func TestWaitIfPausedGenerationChanged(t *testing.T) {
	s := &fileLocatorService{generation: 7}
	if err := s.waitIfPaused(context.Background(), 9, time.Time{}); err != context.Canceled {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
