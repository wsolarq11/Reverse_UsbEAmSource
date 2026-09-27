# Go 后端已恢复符号清单（pclntab 提取，供 GoReSym + Ghidra 重建）

## main 包（应用自身）
- 类型/结构：main.AppEntry、main.walEntry、main.IndexNode、main.LinkEntry、main.FileEntry、main.item、main.loadJob、main.walkState、main.AudioState、main.hashWriter
- 切片类型：[]main.AppEntry、[]main.IndexNode、[]main.item
- 方法/函数：func(main.AppEntry ...)、main.loadJob(...)

## 依赖（开源库，可直接在 go.mod 锁定版本还原）
- github.com/wailsapp/wails/v3@v3.0.0-alpha2.117（框架）
- github.com/6tail/lunar-go@v1.4.6（农历）
- github.com/mozillazg/go-pinyin@v0.21.0（拼音）
- github.com/makiuchi-d/gozxing@v0.1.1（二维码）
- modernc.org/sqlite@v1.48.2（内嵌 SQLite）
- golang.design/x/clipboard@v0.7.1（剪贴板）
- github.com/wailsapp/go-webview2@v1.0.23（WebView2）
- 其余见 go.sum

## 前端可见行为（由抽取的前端 JS 推断后端服务）
- 截图（ScreenshotManager）、屏幕录制/区域捕获
- 鼠标手势（MouseGestureManager）
- Oled 防烧屏（OledBlackoutManager）
- GPU 偏好（GpuPreferenceManager）、内存释放（MemoryReleaseManager）
- 桌面小部件/日历（DesktopWidgetsManager）
- 窗口管理（WindowManagementManager）、文件定位（FileLocatorManager）
- 两步验证码（TwoFactorManager）、二维码（QRCodeManager）
- 音频（AudioManager）、插件宿主（PluginHostPanel）

> 说明：pclntab 完整保留函数名/类型名；字段名与布局需 GoReSym(离线安装)进一步恢复。
> 网络恢复后：`go install github.com/mandiant/GoReSym@latest` 提取全量类型与字段。
