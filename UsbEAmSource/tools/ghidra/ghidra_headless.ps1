<#
.SYNOPSIS
  Ghidra headless 唯一入口：设置进程级 JAVA_HOME 后转发参数给 analyzeHeadless.bat。

.DESCRIPTION
  背景（实测，见 UsbEAmSource/docs/ENV.md §3）：
  external/ghidra_12.1.3_PUBLIC 自身不带 JDK；而 Windows 上 support/launch.bat
  先尝试 PATH 的 java.exe、再尝试 %JAVA_HOME%\bin\java.exe，两者皆失败即退出，
  之后才轮到 LaunchSupport 读 launch.properties 的 JAVA_HOME_OVERRIDE。
  因此只配 JAVA_HOME_OVERRIDE 不足以启动，必须有可用的 JAVA_HOME。

  本脚本只在当前进程设 JAVA_HOME，不改系统环境变量，不改 scoop 状态。
  JDK 路径可用环境变量 USBREAM_JDK 覆盖。

.EXAMPLE
  pwsh UsbEAmSource/tools/ghidra/ghidra_headless.ps1 -help

.EXAMPLE
  pwsh UsbEAmSource/tools/ghidra/ghidra_headless.ps1 `
    work/ghidra/project usbeam -process -noanalysis -readOnly `
    -scriptPath work/ghidra/scripts -postScript InspectAssemble.java
#>
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Rest)

$ErrorActionPreference = 'Stop'

$Jdk = if ($env:USBREAM_JDK) { $env:USBREAM_JDK } else { 'D:\_tools_\jdk\jdk-21.0.12.1+1' }
if (-not (Test-Path (Join-Path $Jdk 'bin\java.exe'))) {
    throw "JDK not found: $Jdk (override with USBREAM_JDK)"
}

$Repo = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$Ghidra = Join-Path $Repo 'external\ghidra_12.1.3_PUBLIC'
if (-not (Test-Path (Join-Path $Ghidra 'support\analyzeHeadless.bat'))) {
    throw "Ghidra not found: $Ghidra"
}

$env:JAVA_HOME = $Jdk
& (Join-Path $Ghidra 'support\analyzeHeadless.bat') @Rest
exit $LASTEXITCODE
