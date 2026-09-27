@echo off
set GOBIN=%USERPROFILE%\go\bin
set HTTPS_PROXY=http://127.0.0.1:7890
set HTTP_PROXY=http://127.0.0.1:7890
set ALL_PROXY=http://127.0.0.1:7890
"%GOBIN%\go1.25.12.exe" build -tags production -trimpath -buildmode=exe ./backend
if %errorlevel% equ 0 (
    echo BUILD PASS
) else (
    echo BUILD FAIL
)