@echo off
setlocal enabledelayedexpansion

set BIN=D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\artifacts\UsbEAm_Launcher_rebuilt.exe
set ADDR_TOOL=D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\tools\go-introspect\addr_tool.exe
set VA_DISASM=D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\tools\go-introspect\va_disasm.py
set DUMPDIR=D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\disasm_assemble
set PIPEDIR=D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource\docs\goresym\pipeline

echo === Step 1: addr_tool listing for BootstrapService ===
"%ADDR_TOOL%" "%BIN%" "BootstrapService"

echo.
echo === Step 2: Dump specific function binary ===
mkdir "%PIPEDIR%\tmp" 2>nul
"%ADDR_TOOL%" "%BIN%" "BootstrapService.setStartupTrayMode" "%PIPEDIR%\tmp"
echo.
echo --- Listing dumped files ---
dir "%PIPEDIR%\tmp\"

echo.
echo === Step 3: Disassemble the dumped binary ===
for %%f in ("%PIPEDIR%\tmp\*.bin") do (
    set binfile=%%f
    echo Disassembling: %%f
    python "%VA_DISASM%" "%%f" 0x14076f7e0 200 "%PIPEDIR%\tmp\%%~nf.asm.txt"
)
echo.
echo --- Done ---