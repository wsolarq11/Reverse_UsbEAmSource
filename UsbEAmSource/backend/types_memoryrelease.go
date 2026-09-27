// AUTO-RECONSTRUCTED TYPES — DOMAIN: memoryrelease
// 研究用途
package main

import (
	"sync"
	"time"
)

type MemoryReleaseConfig struct {
	Mode            string `json:"mode,omitempty"`
	TimerEnabled    bool   `json:"timerEnabled,omitempty"`
	IntervalMinutes int    `json:"intervalMinutes,omitempty"`
}

type MemoryReleaseState struct {
	Supported       bool   `json:"supported"`
	IsAdmin         bool   `json:"isAdmin"`
	Running         bool   `json:"running"`
	SchedulerActive bool   `json:"schedulerActive"`
	NextRunAt       string `json:"nextRunAt,omitempty"`
	LastRunAt       string `json:"lastRunAt,omitempty"`
	LastMode        string `json:"lastMode,omitempty"`
	LastError       string `json:"lastError,omitempty"`
}

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type memoryReleaseService struct {
	lock              sync.Mutex
	config            MemoryReleaseConfig
	moduleEnabled     bool
	state             MemoryReleaseState
	timer             *time.Timer
	scheduleID        uint64
	generation        uint64
	shuttingDown      bool
	runningDone       chan struct{}
	execute           func(string) error
	supported         func() bool
	processIsElevated func() bool
	now               func() time.Time
}
