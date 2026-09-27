// addr_tool: 从 go1.25 Go PE 二进制提取指定方法/函数的二进制地址区间（Off/End）。
// 研究用途。依赖 github.com/goretk/gore（复用 redress 底层 pclntab 解析）。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goretk/gore"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: addr_tool <binary> <name-substr> [dumpdir]")
		os.Exit(2)
	}
	binPath := os.Args[1]
	sub := os.Args[2]
	dumpDir := ""
	if len(os.Args) >= 4 {
		dumpDir = os.Args[3]
	}

	// 若提供第 5 参（symbols 文件路径），输出全量 main 方法 VA→name。
	symOut := ""
	if len(os.Args) >= 5 {
		symOut = os.Args[4]
	}

	f, err := gore.Open(binPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer f.Close()

	pkgs, err := f.GetPackages()
	if err != nil {
		fmt.Fprintln(os.Stderr, "packages:", err)
		os.Exit(1)
	}

	// 兼容性探测：Go1.25 下标准 gosym 可能返回 0 函数，故优先走 packages 路径。
	count := 0
	for _, p := range pkgs {
		if p.Name != "main" {
			continue
		}
		fmt.Fprintf(os.Stderr, "[pkg main] funcs=%d methods=%d\n", len(p.Functions), len(p.Methods))
		for _, m := range p.Methods {
			// 方法：匹配 receiver OR 名
			if strings.Contains(m.Receiver, sub) || strings.Contains(m.Name, sub) {
				fmt.Printf("method %s%s off=0x%x end=0x%x len=%d\n", m.Receiver, m.Name, m.Offset, m.End, m.End-m.Offset)
				count++
				if dumpDir != "" {
					dumpBytes(f, m.Function, dumpDir)
				}
			}
		}
		for _, fn := range p.Functions {
			if strings.Contains(fn.Name, sub) {
				fmt.Printf("func %s off=0x%x end=0x%x len=%d\n", fn.Name, fn.Offset, fn.End, fn.End-fn.Offset)
				count++
				if dumpDir != "" {
					dumpBytes(f, fn, dumpDir)
				}
			}
		}
	}
	fmt.Printf("total matches: %d\n", count)

	if symOut != "" {
		writeSymbols(f, pkgs, symOut)
	}
}

// writeSymbols 输出全部 main 包函数/方法 VA→name 表（含外部库符号）。
func writeSymbols(f *gore.GoFile, pkgs []*gore.Package, out string) {
	var b strings.Builder
	for _, p := range pkgs {
		prefix := p.Name
		if prefix == "" {
			prefix = "?"
		}
		for _, fn := range p.Functions {
			fmt.Fprintf(&b, "0x%x %s.%s\n", fn.Offset, prefix, fn.Name)
		}
		for _, m := range p.Methods {
			recv := strings.TrimSuffix(strings.TrimPrefix(m.Receiver, "(*"), ")")
			fmt.Fprintf(&b, "0x%x %s.%s.%s\n", m.Offset, prefix, recv, m.Name)
		}
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write symbols:", err)
		return
	}
	fmt.Fprintln(os.Stderr, "symbols ->", out, len(b.String()), "bytes")
}

// receiver 提取方法接收者名。
func receiver(r string) string {
	return strings.TrimSuffix(strings.TrimPrefix(r, "(*"), ")")
}

// dumpBytes 将函数区间字节写到 <dumpDir>/<name>.bin。
func dumpBytes(f *gore.GoFile, fn *gore.Function, dumpDir string) {
	b, err := f.Bytes(fn.Offset, fn.End-fn.Offset)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dump", fn.Name, err)
		return
	}
	name := strings.ReplaceAll(fn.Name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	p := filepath.Join(dumpDir, name+".bin")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write", err)
		return
	}
	fmt.Fprintln(os.Stderr, "dumped", len(b), "bytes ->", p)
}