// desktopwidgets_configio.go — 桌面小部件配置 I/O 提取/判断/可移植转换（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// extractDesktopWidgetDocument 从完整启动器配置字节中提取桌面小部件文档。
// [S] 反汇编实证 0x1407ab8c0, 1344B：
//
//	json.Unmarshal(data, &map[string]json.RawMessage) 失败 → (零值,false,err)；
//	无 "desktopWidgets" 键 → (零值,false,nil)；TrimSpace 后空或 == "null" → (零值,false,nil)；
//	否则 json.NewDecoder(bytes.NewReader(raw)).Decode(&doc) 失败 →
//	desktopWidgetStoreError("desktopWidgets", fmt.Sprintf("解析失败: %v", err))；
//	ensureDesktopWidgetJSONEOF 失败 → desktopWidgetStoreError("desktopWidgets", err.Error())；
//	normalize → validate 失败 → (零值,false,err)；成功 → (doc, true, nil)。
func extractDesktopWidgetDocument(data []byte) (DesktopWidgetDocument, bool, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return DesktopWidgetDocument{}, false, err
	}
	raw, ok := m["desktopWidgets"]
	if !ok {
		return DesktopWidgetDocument{}, false, nil
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return DesktopWidgetDocument{}, false, nil
	}
	var doc DesktopWidgetDocument
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&doc); err != nil {
		return DesktopWidgetDocument{}, false, desktopWidgetStoreError("desktopWidgets", fmt.Sprintf("解析失败: %v", err))
	}
	if err := ensureDesktopWidgetJSONEOF(dec); err != nil {
		return DesktopWidgetDocument{}, false, desktopWidgetStoreError("desktopWidgets", err.Error())
	}
	doc = normalizeDesktopWidgetDocument(doc)
	if err := validateDesktopWidgetDocument(doc); err != nil {
		return DesktopWidgetDocument{}, false, err
	}
	return doc, true, nil
}

// launcherConfigDocumentIsWidgetLibrary 判断配置 JSON 是否为纯小部件库文档。
// [S] 反汇编实证 0x1407abe00, 768B：
//
//	json.Unmarshal(data, &map[string]json.RawMessage) 失败 → false；无 "widgets" 键 → false；
//	若含任一白名单键 {initialized,storage,preferences,apps,speedDial,bookmarks,fileSearch,
//	fileLocator,files,memoryRelease,oledBlackout,windowManagement,mouseGestures,twoFactor,
//	desktopWidgets} → false（说明是完整配置而非纯小部件库）；否则 true。
func launcherConfigDocumentIsWidgetLibrary(data []byte) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}
	if _, ok := m["widgets"]; !ok {
		return false
	}
	for _, key := range []string{
		"initialized", "storage", "preferences", "apps", "speedDial",
		"bookmarks", "fileSearch", "fileLocator", "files", "memoryRelease",
		"oledBlackout", "windowManagement", "mouseGestures", "twoFactor",
		"desktopWidgets",
	} {
		if _, ok := m[key]; ok {
			return false
		}
	}
	return true
}

// portableDesktopWidgetDocument 生成可移植文档：深拷贝后清空设备相关字段
// （WeatherCache/NotificationDeliveries/ProtectedSecrets 三个 map）。
// [S] 反汇编实证 0x1407ac4a0, 384B：cloneDesktopWidgetDocument 后 3×makemap_small
// 覆盖偏移 +0x50/+0x58/+0x60 三字段（天气缓存/通知投递/受保护密钥）。
func portableDesktopWidgetDocument(doc DesktopWidgetDocument) DesktopWidgetDocument {
	out := cloneDesktopWidgetDocument(doc)
	out.WeatherCache = make(map[string]DesktopWeatherSnapshot)
	out.NotificationDeliveries = make(map[string]DesktopNotificationRecord)
	out.ProtectedSecrets = make(map[string]DesktopProtectedSecret)
	return out
}

// desktopWidgetDocumentForImport 生成导入用文档：可移植转换后保留受保护密钥
// （ProtectedSecrets 从深拷贝恢复，DPAPI 密钥随设备上下文保留）。
// [S] 反汇编实证 0x1407ac620, 448B：portableDesktopWidgetDocument 后取
// cloneDesktopWidgetDocument 结果的 +0x60（ProtectedSecrets）覆盖回可移植副本。
func desktopWidgetDocumentForImport(doc DesktopWidgetDocument) DesktopWidgetDocument {
	out := portableDesktopWidgetDocument(doc)
	out.ProtectedSecrets = cloneDesktopWidgetDocument(doc).ProtectedSecrets
	return out
}
