// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: mousegesture 工厂 (装配依赖叶子)
// 研究用途
//
// 契约来源：
//   - newMouseGestureService(0x1408dfda0, 704B) 汇编
//   - 结构 mouseGestureService / MouseGestureConfig / GestureSettings：types_gesture.go
//
// 档位：[S] 主体装配（service + 默认 GestureSettings + 双 runtime chan）；Global/Apps/HotCorners
//
//	默认片断与精确字段-offset 比对待字节级续作标 [P]。
package main

// newMouseGestureService 构造鼠标手势服务。
// [S 汇编 0x1408dfda0]：构造 mouseGestureService，config.Settings 默认
// （StartDistancePx/StartTimeoutMs/StopTimeoutMs + 各 set 标志），runtimeStop/runtimeDone chan。
func newMouseGestureService() *mouseGestureService {
	settings := GestureSettings{
		StartDistancePx:    0x96,  // 150
		StartTimeoutMs:     0x12c, // 300
		StopTimeoutMs:      0x1f4, // 500
		startDistancePxSet: true,
		startTimeoutMsSet:  true,
		stopTimeoutMsSet:   true,
	}
	return &mouseGestureService{
		config: MouseGestureConfig{
			Settings: settings,
		},
		runtimeStop: make(chan struct{}),
		runtimeDone: make(chan struct{}),
	}
}
