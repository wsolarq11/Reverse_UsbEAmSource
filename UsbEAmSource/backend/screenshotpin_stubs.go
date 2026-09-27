// AUTO-RECONSTRUCTED SERVICE STUBS — screenshotPinWindowService / windowManagementService
// 研究用途
// 档位：[S-sig] 签名实证 + 体骨架。方法签名经符号表 + [S] 调用方（bootstrapservice.go 各
// screenshotPin RPC）的 call target 实证；体骨架（Wails v3 置顶窗口域待整域落地，恒返零值）。
package main

import "github.com/wailsapp/wails/v3/pkg/application"

// ListStates 列出置顶窗口状态。 [S-sig 0x140982ae0]
func (s *screenshotPinWindowService) ListStates() []interface{} { return nil }

// Focus 聚焦指定置顶窗口。 [S-sig 0x140983080]
func (s *screenshotPinWindowService) Focus(id string) error { return nil }

// Close 关闭指定置顶窗口。 [S-sig 0x140984d00]
func (s *screenshotPinWindowService) Close(id string) error { return nil }

// Hide 隐藏指定置顶窗口。 [S-sig 0x140983760（HideWindow）]
func (s *screenshotPinWindowService) Hide(id string) error { return nil }

// CloseAll 关闭全部置顶窗口。 [S-sig 0x140984e80]
func (s *screenshotPinWindowService) CloseAll() error { return nil }

// ClearSnapshots 清除快照。 [S-sig 0x140986200]
func (s *screenshotPinWindowService) ClearSnapshots() error { return nil }

// SetClickThrough 设置点击穿透。 [S-sig 0x140983920]
func (s *screenshotPinWindowService) SetClickThrough(id string, clickThrough bool) error { return nil }

// SetOpacity 设置不透明度。 [S-sig 0x140984380]
func (s *screenshotPinWindowService) SetOpacity(id string, opacity float64) error { return nil }

// SetAutoShow 设置自动显示。 [S-sig 0x140983f20]
func (s *screenshotPinWindowService) SetAutoShow(id string, autoShow bool) error { return nil }

// Show 显示置顶窗口。 [S-sig 0x1409817a0]
func (s *screenshotPinWindowService) Show(id string) error { return nil }

// ShowWithSource 以数据源显示置顶窗口。 [S-sig 0x140981800]
func (s *screenshotPinWindowService) ShowWithSource(data string, source string, screen *application.Screen) error {
	return nil
}

// ShowRef 按引用显示置顶窗口。 [S-sig 0x14098ae00（showFromSnapshot）]
func (s *screenshotPinWindowService) ShowRef(ref string) error { return nil }

// ShowClipboard 显示剪贴板图像。 [S-sig 调用方实证，符号名待复核]
func (s *screenshotPinWindowService) ShowClipboard() error { return nil }

// UpdateScale 更新缩放。 [S-sig 0x140982100]
func (s *screenshotPinWindowService) UpdateScale(scale float64) error { return nil }

// State 返回当前 pin window 状态。 [S-sig 0x1409869a0]
func (s *screenshotPinWindowService) State() {}

// updateScaleForPin 按 id 更新置顶窗口缩放。 [S-sig 0x1409822a0]：xmm0=scale, rcx=id。
func (s *screenshotPinWindowService) updateScaleForPin(id string, scale float64) error { return nil }
