@echo off
SET HTTPS_PROXY=http://127.0.0.1:7890
SET HTTP_PROXY=http://127.0.0.1:7890
SET ALL_PROXY=http://127.0.0.1:7890

FOR /F "tokens=*" %%i IN ('go env GOPATH') DO SET GOPATH_BIN=%%i\bin
ECHO GOPATH_BIN=%GOPATH_BIN%

ECHO === go vet ./backend ===
"%GOPATH_BIN%\go1.25.12.exe" vet ./backend 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO VET FAILED
    EXIT /B 1
)
ECHO VET PASSED