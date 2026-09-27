@echo off
SETLOCAL ENABLEDELAYEDEXPANSION
SET HTTPS_PROXY=http://127.0.0.1:7890
SET HTTP_PROXY=http://127.0.0.1:7890
SET ALL_PROXY=http://127.0.0.1:7890

SET GO_VER=1.25.12
SET GOPATH_BIN=
FOR /F "tokens=*" %%i IN ('go env GOPATH') DO SET GOPATH_BIN=%%i\bin
ECHO GOPATH_BIN=%GOPATH_BIN%

ECHO === go vet ./backend ===
"%GOPATH_BIN%\go%GO_VER%.exe" vet ./backend 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO VET FAILED
    EXIT /B 1
)
ECHO VET PASSED

ECHO === go test -count=1 ./backend ===
"%GOPATH_BIN%\go%GO_VER%.exe" test -count=1 ./backend 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO TEST FAILED
    EXIT /B 1
)
ECHO TEST PASSED

ECHO === go build ===
"%GOPATH_BIN%\go%GO_VER%.exe" build -tags production -trimpath -buildmode=exe -o artifacts\UsbEAm_Launcher_rebuilt.exe ./backend 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO BUILD FAILED
    EXIT /B 1
)
ECHO BUILD PASSED

ECHO === ALL VERIFIED ===