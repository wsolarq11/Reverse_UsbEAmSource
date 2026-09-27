// AUTO-RECONSTRUCTED — DOMAIN: internet shortcut file read
// 研究用途. 实证依据 dump 反汇编（928B，0x14086a1a0）。
// 汇编确定性语义（逐条对齐）：
//   - os.OpenFile(path, O_RDONLY, 0)（0x14086a1e5）
//   - os.File.Stat + defer close（0x14086a218/0x14086a306）；Stat.Size() > 0x40000 → 超限错误
//   - io.LimitReader(0x40001) + io.ReadAll（0x14086a2b5）；读回 > 0x40000 → 超限错误
//     返回原始字节（未做编码解码，调用方 readInternetShortcutInfo 负责 decode）。
package main

import (
	"errors"
	"io"
	"os"
)

// errInternetShortcutTooLarge 表示 .url 文件超过 0x40000 字节上限。
var errInternetShortcutTooLarge = errors.New("internet shortcut too large")

// readInternetShortcutFile 读取 .url 快捷方式文件原始字节（带大小上限）。
// [S 汇编 0x14086a1a0, 928B]：OpenFile(O_RDONLY)→Stat（>0x40000 超限）→LimitReader(0x40001)+ReadAll。
func readInternetShortcutFile(path string) ([]byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > 0x40000 {
		return nil, errInternetShortcutTooLarge
	}

	data, err := io.ReadAll(io.LimitReader(f, 0x40001))
	if err != nil {
		return nil, err
	}
	if len(data) > 0x40000 {
		return nil, errInternetShortcutTooLarge
	}
	return data, nil
}
