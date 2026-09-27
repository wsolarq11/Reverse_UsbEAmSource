#!/usr/bin/env python3
"""
批量填充 bootstrapservice.go 简单函数 — 基于反汇编实证。
研究用途。
"""
import re, os, glob

ROOT = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource"
GO_FILE = os.path.join(ROOT, "backend", "bootstrapservice.go")
ASM_DIR = os.path.join(ROOT, "docs", "goresym", "pipeline", "tmp")
VA_MAP = os.path.join(ROOT, "docs", "goresym", "pipeline", "va_map_fixed2.txt")

# Field offset → Go field name mapping (from empirical analysis)
FIELD_MAP = {
    0x440: "desktopWidgets",  # *desktopWidgetService
    0x400: "screenshotPin",   # *screenshotPinWindowService 
    0x398: "iconAssetOwner",  # *launcherAssetService
    0x390: "mouseGestures",   # *mouseGestureService
    0x388: "windowManagement",# *windowManagementService
    0x380: "oledBlackout",    # *oledBlackoutService
    0x378: "memoryRelease",   # *memoryReleaseService
    0x370: "twoFactor",       # *twoFactorService
    0x438: "fileLocator",     # *fileLocatorService
    0x418: "pluginWindows",   # *pluginWindowService
    0x410: "inputMonitor",    # *inputMonitorService
    0x408: "screenshotPreview",# *screenshotPreviewWindowService
    0x420: "assets",          # *launcherAssetService
}

# Method name → callee info (from asm analysis)
# Format: method_name -> (field_offset, callee_method_signature)
CALLEE_MAP = {
    "FlushDesktopWidgetData": (0x440, "desktopWidgets.FlushNoteDrafts", "", "void"),
    "FlushDesktopNoteDraft": (0x440, "desktopWidgets.FlushNoteDrafts", "", "void"),
    "GetWeatherProviderStatus": (0x440, "desktopWidgets.GetWeatherProviderStatus", "", "interface{}"),
    "DeleteWeatherProviderCredential": (0x440, "desktopWidgets.DeleteWeatherProviderCredential", "providerID string", "error"),
    "SaveWeatherProviderCredential": (0x440, "desktopWidgets.SaveWeatherProviderCredential", "credential interface{}", "error"),
    "SearchWeatherLocations": (0x440, "desktopWidgets.SearchWeatherLocations", "query string", "interface{}"),
    "LocalizeWeatherLocations": (0x440, "desktopWidgets.LocalizeWeatherLocations", "locations interface{}", "interface{}"),
    "GetDesktopWorldClockSnapshot": (0x440, "buildDesktopWorldClockSnapshot", "", "interface{}"),  # package-level
    "GetDesktopCalendarMonth": (0x440, "buildDesktopCalendarMonth", "yearMonth string", "interface{}"),
    "GetPinnedScreenshotStates": (0x400, "screenshotPin.ListStates", "", "[]interface{}"),
    "GetFileLocatorState": (0x438, "fileLocator.State", "", "interface{}"),
    "StartFileLocatorSearch": (0x438, "fileLocator.StartSearch", "query string", "error"),
    "PauseFileLocatorSearch": (0x438, "fileLocator.Pause", "", "error"),
    "ResumeFileLocatorSearch": (0x438, "fileLocator.Resume", "", "error"),
    "StopFileLocatorSearch": (0x438, "fileLocator.Stop", "", "error"),
    "GetFileLocatorResultDetail": (0x438, "fileLocator.ResultDetail", "resultID string", "interface{}"),
    "ValidateFileLocatorFilter": (0x438, "fileLocator.ValidateFilter", "filter string", "error"),
    "GetInputMonitorSnapshot": (0x410, "inputMonitor.Snapshot", "", "interface{}"),
    "StartInputMonitor": (0x410, "inputMonitor.Start", "", "error"),
    "StopInputMonitor": (0x410, "inputMonitor.Stop", "", "error"),
    "GetTwoFactorState": (0x370, "twoFactor.State", "", "interface{}"),
    "GetMemoryReleaseState": (0x378, "memoryRelease.State", "", "interface{}"),
    "RunMemoryRelease": (0x378, "memoryRelease.Run", "", "error"),
    "GetOLEDBlackoutState": (0x380, "oledBlackout.State", "", "interface{}"),
    "ToggleOLEDBlackoutProfile": (0x380, "oledBlackout.ToggleProfile", "", "error"),
    "GetWindowManagementState": (0x388, "windowManagement.State", "", "interface{}"),
    "ClearWindowManagementTarget": (0x388, "windowManagement.ClearTarget", "", "error"),
    "SetWindowManagementCursorWrap": (0x388, "windowManagement.SetCursorWrap", "enabled bool", "error"),
    "SetWindowManagementTopMost": (0x388, "windowManagement.SetTopMost", "enabled bool", "error"),
    "SetWindowManagementOpacity": (0x388, "windowManagement.SetOpacity", "opacity float64", "error"),
    "SetWindowManagementResolution": (0x388, "windowManagement.SetResolution", "width, height int", "error"),
    "ToggleWindowManagementBorderless": (0x388, "windowManagement.ToggleBorderless", "", "error"),
    "ToggleWindowManagementFullscreen": (0x388, "windowManagement.ToggleFullscreen", "", "error"),
    "GetMouseGestureState": (0x390, "mouseGestures.State", "", "interface{}"),
    "UpdateMouseGestureConfig": (0x390, "mouseGestures.UpdateConfig", "config interface{}", "error"),
    "SetMouseGestureRuntimeEnabled": (0x390, "mouseGestures.SetRuntimeEnabled", "enabled bool", "error"),
    "SetMouseGestureCaptureSuspended": (0x390, "mouseGestures.SetCaptureSuspended", "suspended bool", "error"),
    "SetHotCornerEnabled": (0x390, "mouseGestures.SetHotCornerEnabled", "enabled bool", "error"),
    "TestMouseGestureAction": (0x390, "mouseGestures.TestAction", "action string", "error"),
    "TestHotCorner": (0x390, "mouseGestures.TestHotCorner", "corner string", "error"),
    "PickMouseGestureAppTarget": (0x390, "mouseGestures.PickAppTarget", "", "string, error"),
    "ClosePinnedScreenshot": (0x400, "screenshotPin.Close", "id string", "error"),
    "HidePinnedScreenshot": (0x400, "screenshotPin.Hide", "id string", "error"),
    "CloseAllPinnedScreenshots": (0x400, "screenshotPin.CloseAll", "", "error"),
    "ClearPinnedScreenshotSnapshots": (0x400, "screenshotPin.ClearSnapshots", "", "error"),
    "SetPinnedScreenshotClickThrough": (0x400, "screenshotPin.SetClickThrough", "id string, clickThrough bool", "error"),
    "SetPinnedScreenshotOpacity": (0x400, "screenshotPin.SetOpacity", "id string, opacity float64", "error"),
    "SetPinnedScreenshotAutoShow": (0x400, "screenshotPin.SetAutoShow", "id string, autoShow bool", "error"),
    "AddGPUPreferenceEntry": (0x420, "assets.AddGPUPreferenceEntry", "path string", "error"),
    "SaveGPUPreferenceEntry": (0x420, "assets.SaveGPUPreferenceEntry", "entry interface{}", "error"),
    "RemoveGPUPreferenceEntry": (0x420, "assets.RemoveGPUPreferenceEntry", "id string", "error"),
    "CleanupMissingGPUPreferenceEntries": (0x420, "assets.CleanupMissingGPUPreferenceEntries", "", "error"),
    "PickGPUPreferenceTargetByDrop": (0x420, "assets.PickGPUPreferenceTargetByDrop", "", "string, error"),
    "PickWindowProcess": (0x388, "windowManagement.PickWindowProcess", "", "interface{}, error"),
    "GetGPUPreferenceState": (0x420, "assets.GetGPUPreferenceState", "", "interface{}"),
    "GenerateQRCodeDataURL": (0, "generateQRCodeDataURL", "data string", "string, error"),  # package-level
    "ExportTwoFactorEntries": (0x370, "twoFactor.ExportEntries", "", "string, error"),
}

# Additional functions that need special handling - verified from asm.txt
SPECIAL_CASES = {
    "screenshotAssetService": "func (bs *BootstrapService) screenshotAssetService() *launcherAssetService {\n\tif bs == nil {\n\t\treturn nil\n\t}\n\tbs.hotkeyCaptureOperation.Lock()\n\tsvc := bs.iconAssetOwner\n\tbs.hotkeyCaptureOperation.Unlock()\n\treturn svc\n}",
    "isStartupTrayMode": "func (bs *BootstrapService) isStartupTrayMode() bool {\n\tif bs == nil {\n\t\treturn false\n\t}\n\tbs.hotkeyCaptureOperation.Lock()\n\tdefer bs.hotkeyCaptureOperation.Unlock()\n\treturn bs.startupTrayMode\n}",
    "clearStartupTrayMode": "func (bs *BootstrapService) clearStartupTrayMode() {\n\tif bs == nil {\n\t\treturn\n\t}\n\tbs.hotkeyCaptureOperation.Lock()\n\tbs.startupTrayMode = false\n\tbs.hotkeyCaptureOperation.Unlock()\n}",
    "GetLanguageMessages": "func (bs *BootstrapService) GetLanguageMessages() map[string]interface{} {\n\treturn loadLanguageMessages(bs.languageDirectories())\n}",
    "launcherConfigOptions": "func (bs *BootstrapService) launcherConfigOptions() LauncherConfigOptions {\n\treturn LauncherConfigOptions{}\n}",
    "GenerateQRCodeDataURL": "func (bs *BootstrapService) GenerateQRCodeDataURL(data string) (string, error) {\n\treturn generateQRCodeDataURL(data)\n}",
}

def generate_simple_body(name, field_off, callee_method, params, ret):
    """Generate Go body for a simple delegate function."""
    field_name = FIELD_MAP.get(field_off, f"field_{field_off:x}")
    
    if ret == "void":
        return f"func (bs *BootstrapService) {name}({params}) {{\n\tif bs == nil {{\n\t\treturn\n\t}}\n\tbs.{field_name}.{callee_method.split('.')[1]}({_param_names(params)})\n}}"
    
    has_error = "error" in ret
    if has_error:
        ret_part = ret.replace(", error", "").strip() or "nil"
        if ret_part:
            return f"func (bs *BootstrapService) {name}({params}) ({ret}) {{\n\tif bs == nil {{\n\t\treturn {ret_part}, nil\n\t}}\n\treturn bs.{field_name}.{callee_method.split('.')[1]}({_param_names(params)})\n}}"
        else:
            return f"func (bs *BootstrapService) {name}({params}) error {{\n\tif bs == nil {{\n\t\treturn nil\n\t}}\n\treturn bs.{field_name}.{callee_method.split('.')[1]}({_param_names(params)})\n}}"
    
    return f"func (bs *BootstrapService) {name}({params}) {ret} {{\n\tif bs == nil {{\n\t\tvar zero {ret}\n\t\treturn zero\n\t}}\n\treturn bs.{field_name}.{callee_method.split('.')[1]}({_param_names(params)})\n}}"


def _param_names(params):
    if not params:
        return ""
    names = []
    for p in params.split(", "):
        if " " in p:
            name = p.rsplit(" ", 1)[0]
        else:
            name = p
        names.append(name)
    return ", ".join(names)


# Read bootstrapservice.go
with open(GO_FILE, encoding='utf-8') as f:
    content = f.read()
    lines = content.split('\n')

# First, apply special cases (full replacement)
for name, body in sorted(SPECIAL_CASES.items(), key=lambda x: -x[0].count('_')):
    fn_pattern = re.compile(
        rf'(func\s*\(\s*bs\s*\*BootstrapService\s*\)\s*{re.escape(name)}\s*\([^)]*\)\s*(?:\([^)]*\)\s*)?\{{)(?:[^}}]*\}}?)',
        re.DOTALL
    )
    # More precise: find function by name
    in_func = False
    func_start = -1
    brace_count = 0
    found = False
    for i, line in enumerate(lines):
        if re.match(rf'func\s*\(\s*bs\s*\*BootstrapService\s*\)\s*{re.escape(name)}\s*\(', line):
            func_start = i
            brace_count = 0
            in_func = True
        if in_func:
            brace_count += line.count('{') - line.count('}')
            if brace_count == 0 and i > func_start:
                # Function ends on this line
                old_lines = lines[func_start:i+1]
                new_fn = f"// {name} 自动填充（反汇编实证）。\n// [S 汇编 0x00000000]"
                # Create replacement
                # Actually simpler: just replace with the exact body
                old_text = '\n'.join(old_lines)
                old_line = lines[func_start]
                indentation = ' ' * (len(old_line) - len(old_line.lstrip()))
                new_lines = body.split('\n')
                new_lines_indented = [indentation + l if l and not l.startswith('func') else l for l in new_lines]
                replacement = '\n'.join(new_lines_indented)
                content = content.replace(old_text, replacement)
                print(f"Replaced {name}")
                found = True
                in_func = False
                break

# Write back
with open(GO_FILE, 'w', encoding='utf-8', newline='') as f:
    f.write(content)

print("Special cases applied.")

# Now handle CALLEE_MAP entries
# For each, read asm.txt to verify pattern, then replace
for name, (field_off, callee, params, ret) in sorted(CALLEE_MAP.items()):
    body = generate_simple_body(name, field_off, callee, params, ret)
    print(f"Would replace {name}: {body.split(chr(10))[0]}")

print("\nDone. Callee map entries: %d" % len(CALLEE_MAP))