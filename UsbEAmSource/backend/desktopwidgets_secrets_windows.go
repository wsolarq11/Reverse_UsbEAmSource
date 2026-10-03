// desktopwidgets_secrets_windows.go — 桌面小部件凭据 DPAPI 保护（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)

package main

import (
	"encoding/base64"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// protectDesktopWidgetSecret 用 DPAPI 加密凭据并 base64 编码。
// [S] 反汇编实证 0x1407b1660, 960B：
//
//	空串 → errors.New("凭据不能为空")（18B @0x140C59CD2）；
//	[]byte(secret) → CryptProtectData(&DataBlob{Size:len,Data:&data[0]}, nil,nil,0,nil,
//	CRYPTPROTECT_UI_FORBIDDEN=1, &out) → 失败返回 ("", err)；成功 LocalFree 释放 out.Data、
//	复制出加密字节（memmove）后 base64.StdEncoding.EncodeToString 返回；
//	defer 逆序：clear(encrypted) → LocalFree → clear(data)（memclrNoHeapPointers 清零敏感字节）。
func protectDesktopWidgetSecret(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("凭据不能为空")
	}
	data := []byte(secret)
	defer clear(data)
	var out windows.DataBlob
	err := windows.CryptProtectData(
		&windows.DataBlob{Size: uint32(len(data)), Data: &data[0]},
		nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out,
	)
	if err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	encrypted := make([]byte, out.Size)
	copy(encrypted, unsafe.Slice(out.Data, int(out.Size)))
	defer clear(encrypted)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// unprotectDesktopWidgetSecret 用 DPAPI 解密 base64 编码的凭据。
// [S] 反汇编实证 0x1407b1b40, 1024B：
//
//	base64.StdEncoding.DecodeString → err!=nil 或 len==0 时 errors.New("受保护凭据格式无效")
//	（27B @0x140C69C84）；否则 CryptUnprotectData(&DataBlob{Size:len,Data:&data[0]},
//	nil,nil,0,nil,CRYPTPROTECT_UI_FORBIDDEN=1, &out) → 失败返回 ("", err)；成功 LocalFree 释放
//	out.Data、复制出明文字节后 string(bytes) 返回；defer 逆序 clear/LocalFree/clear。
func unprotectDesktopWidgetSecret(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil || len(data) == 0 {
		return "", errors.New("受保护凭据格式无效")
	}
	defer clear(data)
	var out windows.DataBlob
	err = windows.CryptUnprotectData(
		&windows.DataBlob{Size: uint32(len(data)), Data: &data[0]},
		nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out,
	)
	if err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	decrypted := make([]byte, out.Size)
	copy(decrypted, unsafe.Slice(out.Data, int(out.Size)))
	defer clear(decrypted)
	return string(decrypted), nil
}
