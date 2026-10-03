package main

import (
	"mime"
	"strings"
)

// remoteIconContentTypeAllowed 判断远程图标的内容类型是否与允许值匹配（忽略大小写）。
// [S] ASM 0x140961fe0: TrimSpace → ParseMediaType，err!=nil→false，否则 EqualFold(mediatype, allowed)。
func remoteIconContentTypeAllowed(contentType, allowedType string) bool {
	mediatype, _, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil {
		return false
	}
	return strings.EqualFold(mediatype, allowedType)
}

// resolveRemoteIconCollectionName 解析远程图标集合显示名：从 collections 查 name，
// TrimSpace(集合 Name) 非空则返回之，否则回退返回 name。
// [S 汇编 0x1409661a0, 160B] 实证：mapaccess2_faststr 查 name；未命中→返回 name；
// 命中→TrimSpace(value) 非空→返回 TrimSpace(value)，否则返回 name。
func resolveRemoteIconCollectionName(name string, collections map[string]remoteIconifyCollectionRef) string {
	c, ok := collections[name]
	if !ok {
		return name
	}
	if trimmed := strings.TrimSpace(c.Name); trimmed != "" {
		return trimmed
	}
	return name
}
