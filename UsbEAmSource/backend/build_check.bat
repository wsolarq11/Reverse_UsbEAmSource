@echo off
set GOBIN=%USERPROFILE%\go\bin
set HTTPS_PROXY=http://127.0.0.1:7890
set HTTP_PROXY=http://127.0.0.1:7890
set ALL_PROXY=http://127.0.0.1:7890
cd /d "%~dp0"
"%GOBIN%\go1.25.12.exe" build -tags production -trimpath -buildmode=exe -v ./backend 2>&1
echo EXIT_CODE=%errorlevel%