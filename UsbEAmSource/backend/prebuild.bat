@echo off
chcp 65001 >nul
set PATH=D:\go1.25.12\bin;%PATH%
cd /d D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\backend
echo === go vet ===
go vet ./...
echo === go build ===
go build -tags production -trimpath -buildmode=exe -o ..\backend_test_build.exe . 2>&1
if %errorlevel% equ 0 (echo BUILD PASS) else (echo BUILD FAIL)
echo === go test ===
go test -count=1 ./... 2>&1