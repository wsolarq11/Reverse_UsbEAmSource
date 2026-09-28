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
