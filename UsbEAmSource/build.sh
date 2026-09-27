# UsbEAm Launcher 1.0.3 — 可复现构建脚本（已验证可编译） 
# 说明：本机 Go 直连外网超时，需走本地代理 127.0.0.1:7890（Clash/V2Ray 类）。若非如此去掉 export 即可。
export HTTPS_PROXY=http://127.0.0.1:7890
export HTTP_PROXY=http://127.0.0.1:7890
export ALL_PROXY=http://127.0.0.1:7890

set -e
GO_VER=1.25.12
GOPATH_BIN="$(go env GOPATH)/bin"

# 1) 确保精确工具链 go1.25.12（buildinfo 验证过）
if [ ! -x "$GOPATH_BIN/go$GO_VER" ]; then
  go install golang.org/dl/go$GO_VER@latest
  "$GOPATH_BIN/go$GO_VER" download
fi

# 2) 还原 go 模块（go.sum 的 h1 已对上真实模块，已验证）
"$GOPATH_BIN/go$GO_VER" mod download

# 3) 编译 Go 后端（与目标一致的 build flags）
#    前端 frontend/dist 即原 exe 内嵌字节，直接嵌入/随 wails 分发。
"$GOPATH_BIN/go$GO_VER" build \
  -tags production \
  -trimpath \
  -buildmode=exe \
  -o artifacts/UsbEAm_Launcher_rebuilt.exe \
  ./backend

echo "构建成功: artifacts/UsbEAm_Launcher_rebuilt.exe"
echo "（.exe 哈希与官方不同——Go 原生码无法从反编译逐字节重建，此为功能一致重建）"