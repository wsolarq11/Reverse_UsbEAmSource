#!/usr/bin/env python3
"""
基于反汇编实证，批量填充 bootstrapservice.go 函数体。
按 asm.txt 模式分类：纯委托 / null-check委托 / Lock+委托 / package调用
"""
import re, os, sys

ROOT = "D:/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource"
GOFILE = os.path.join(ROOT, "backend", "bootstrapservice.go")
ASMDIR = os.path.join(ROOT, "docs", "goresym", "pipeline", "tmp")

# ---- 场偏移映射（实证验证） ----
FIELD_OFFSETS = {
    0x440: ("desktopWidgets", "*desktopWidgetService"),
    0x400: ("screenshotPin", "*screenshotPinWindowService"),
    0x398: ("iconAssetOwner", "*launcherAssetService"),
    0x390: ("mouseGestures", "*mouseGestureService"),
    0x388: ("windowManagement", "*windowManagementService"),
    0x380: ("oledBlackout", "*oledBlackoutService"),
    0x378: ("memoryRelease", "*memoryReleaseService"),
    0x370: ("twoFactor", "*twoFactorService"),
    0x438: ("fileLocator", "*fileLocatorService"),
    0x418: ("pluginWindows", "*pluginWindowService"),
    0x410: ("inputMonitor", "*inputMonitorService"),
    0x408: ("screenshotPreview", "*screenshotPreviewWindowService"),
    0x420: ("assets", "*launcherAssetService"),
    0x428: ("fileIndex", "fileSearchRuntimeWarmer"),
}

CALLEE_MAP = {}  # fn_name -> (field_name, callee_method, params, ret_type, is_null_check)

def parse_asm_pattern(filename):
    """Parse a .asm.txt file and determine Go pattern."""
    path = os.path.join(ASMDIR, filename)
    if not os.path.exists(path):
        return None
    
    with open(path, encoding='utf-8', errors='replace') as f:
        lines = f.readlines()
    
    fn_name = filename.replace('.asm.txt', '')
    
    # Extract: struct access offset, call targets, null checks
    field_load = None
    call_targets = []
    has_null_check = False
    struct_reg = None
    
    for l in lines:
        l = l.strip()
        # Struct field load: mov rax, qword ptr [rax + 0xNNN]
        m = re.search(r'mov\s+(rax|rbx|rcx|rdx|rdi|rsi)\s*,\s*qword ptr\s*\[(rax|rbx|rcx|rdx)\s*\+\s*0x([0-9a-f]+)\]', l)
        if m:
            dst_reg = m.group(1)
            src_reg = m.group(2)
            off = int(m.group(3), 16)
            if off in FIELD_OFFSETS:
                field_load = (dst_reg, src_reg, off)
        
        # Call targets (ignore runtime/internal calls)
        m2 = re.search(r'call\s+0x[0-9a-f]+\s*;\s*(main\.\S+)', l)
        if m2:
            target = m2.group(1)
            if not target.startswith('runtime') and not target.startswith('internal') and not target.startswith('main.BootstrapService'):
                call_targets.append(target)
        
        # null check: test rax, rax; je ...
        m3 = re.search(r'test\s+(rax|rbx|rcx|rdx)\s*,\s*\1', l)
        if m3:
            has_null_check = True
        
        # Package-level function call (no struct receiver)
        m4 = re.search(r'call\s+0x[0-9a-f]+\s*;\s*(main\.(?:generateQRCodeDataURL|detectBookmarkSources|detectLinkBrowsers|resolveWindowsStartMenuRoots|resolveWindowsDesktopRoots|buildDesktopWorldClockSnapshot|buildDesktopCalendarMonth|ensureWorkspaceDirectories|loadLanguageMessages|discoverLanguages|normalizeLauncherHotkeyBindings|isLauncherConfigEmptyInitialConfig|fileExists|resolveProcessWorkingDirectory))', l)
        if m4:
            call_targets.append(m4.group(1))
    
    return {
        'name': fn_name,
        'field_load': field_load,
        'calls': call_targets,
        'null_check': has_null_check,
        'struct_reg': struct_reg,
    }

def generate_go_body(info):
    """Generate Go function body from parsed asm info."""
    name = info['name']
    calls = info['calls']
    field = info['field_load']
    null_ck = info['null_check']
    
    if not calls:
        return None
    
    callee = calls[0].replace('main.', '')
    
    # Determine Go signature from existing stub
    stub = get_stub_signature(name)
    if not stub:
        return None
    
    ret_type = stub.get('ret', '')
    params = stub.get('params', '')
    has_error = 'error' in ret_type
    has_receiver = stub.get('receiver', False)
    
    # 1) Package-level function calls (no struct access)
    pkg_funcs = ['generateQRCodeDataURL', 'detectBookmarkSources', 'detectLinkBrowsers', 
                 'resolveWindowsStartMenuRoots', 'resolveWindowsDesktopRoots',
                 'buildDesktopWorldClockSnapshot', 'buildDesktopCalendarMonth',
                 'ensureWorkspaceDirectories', 'loadLanguageMessages', 'discoverLanguages',
                 'normalizeLauncherHotkeyBindings', 'isLauncherConfigEmptyInitialConfig',
                 'fileExists', 'resolveProcessWorkingDirectory']
    
    for pf in pkg_funcs:
        if pf in callee:
            if not has_receiver:
                # package-level function wrapper
                body = f"\treturn {callee}({extract_args(params)})\n"
            else:
                body = f"\treturn {callee}({extract_args(params)})\n"
            return format_body(body, name, params, ret_type, has_receiver)
    
    # 2) Simple delegate to sub-service field
    if field:
        dst_reg, src_reg, off = field
        fname = FIELD_OFFSETS[off][0] if off in FIELD_OFFSETS else f"field_{off:x}"
        
        # If callee contains the method name
        meth_name = callee.split('.')[-1] if '.' in callee else callee
        
        if null_ck:
            body = f"\tif bs == nil {{\n"
            body += f"\t\treturn {zero_value(ret_type)}\n"
            body += f"\t}}\n"
            body += f"\treturn bs.{fname}.{meth_name}({extract_args(params)})\n"
        else:
            body = f"\treturn bs.{fname}.{meth_name}({extract_args(params)})\n"
        return format_body(body, name, params, ret_type, has_receiver)
    
    return None

def extract_args(params):
    """Extract argument names from function parameter list."""
    if not params:
        return ''
    args = []
    for p in params.split(', '):
        p = p.strip()
        if ' ' in p:
            name = p.rsplit(' ', 1)[0]
        else:
            name = p
        args.append(name)
    return ', '.join(args)

def zero_value(ret_type):
    """Return Go zero value for a type."""
    if not ret_type:
        return 'nil'
    if ret_type == 'error':
        return 'nil'
    if ret_type.startswith('*') or ret_type.startswith('[]') or ret_type.startswith('map[') or ret_type.startswith('interface{}'):
        return 'nil'
    if ret_type == 'string':
        return '""'
    if ret_type == 'bool':
        return 'false'
    if ret_type == 'int' or ret_type.startswith('int'):
        return '0'
    if ret_type.startswith('float'):
        return '0.0'
    return 'nil'

# -- Read existing go file to get stub signatures --
def get_stub_signature(name):
    """Read the existing function stub signature from the Go file."""
    with open(GOFILE, encoding='utf-8') as f:
        lines = f.readlines()
    
    for i, line in enumerate(lines):
        m = re.match(rf'^func\s*(?:\(\s*(\w+)\s*\*?(\w+)\s*\))?\s*{re.escape(name)}\s*\(([^)]*)\)\s*(?:\(([^)]*)\)\s*)?{{', line)
        if m:
            recv_var = m.group(1)
            recv_type = m.group(2)
            params = m.group(3).strip()
            ret = m.group(4).strip() if m.group(4) else ''
            return {
                'receiver': bool(recv_var),
                'recv_var': recv_var,
                'recv_type': recv_type,
                'params': params,
                'ret': ret,
            }
    return None

def format_body(body, name, params, ret_type, has_receiver):
    """Wrap body with proper function signature."""
    if has_receiver:
        sig = f"func (bs *BootstrapService) {name}({params})"
    else:
        sig = f"func {name}({params})"
    
    if ret_type:
        sig += f" {ret_type}"
    sig += " {"
    
    return f"{sig}\n{body}}}"

def apply_replacement(name, new_body):
    """Find and replace function body in go file."""
    with open(GOFILE, encoding='utf-8') as f:
        content = f.read()
        lines = content.split('\n')
    
    # Find function start and end
    fn_start = -1
    in_func = False
    brace_count = 0
    
    for i, line in enumerate(lines):
        if re.match(rf'^func\s*(?:\(\s*\w+\s*\*\w+\s*\))?\s*{re.escape(name)}\s*\(', line):
            fn_start = i
            in_func = True
            brace_count = 0
        
        if in_func:
            brace_count += line.count('{') - line.count('}')
            if brace_count <= 0 and i > fn_start:
                # Function body ends on this line or after }
                if brace_count < 0:
                    # The '}' that closed the function is on this line
                    old_end = i + 1
                else:
                    old_end = i
                
                old_lines = lines[fn_start:old_end]
                old_text = '\n'.join(old_lines)
                
                # Build proper indentation
                indent = ' ' * (len(lines[fn_start]) - len(lines[fn_start].lstrip()))
                new_lines_body = new_body.split('\n')
                formatted = []
                for j, nl in enumerate(new_lines_body):
                    if j == 0:
                        formatted.append(nl)
                    elif nl.startswith('\t'):
                        formatted.append(indent + nl.lstrip('\t'))
                    elif nl.strip():
                        formatted.append(indent + nl)
                    else:
                        formatted.append(indent)
                
                new_text = '\n'.join(formatted)
                
                content = content.replace(old_text, new_text)
                with open(GOFILE, 'w', encoding='utf-8', newline='') as f:
                    f.write(content)
                return True
    
    return False

# -- Main pipeline --
def main():
    asm_files = sorted([f for f in os.listdir(ASMDIR) if f.endswith('.asm.txt') and not any(x in f for x in ['func1', 'func2', 'func11', 'func12', 'func21', 'deferwrap', 'gowrap', 'Printf', '-fm', '.func'])])
    
    # Filter to only those matching existing functions
    existing = set()
    with open(GOFILE, encoding='utf-8') as f:
        lines = f.readlines()
    for i, line in enumerate(lines):
        m = re.match(r'^func\s*(?:\(\s*\w+\s*\*\w+\s*\))?\s+(\w+)\s*\(', line)
        if m:
            existing.add(m.group(1))
    
    completed = 0
    skipped = 0
    errors = []
    
    for asm_file in asm_files:
        fn_name = asm_file.replace('.asm.txt', '')
        if fn_name not in existing:
            skipped += 1
            continue
        
        # Check if already has empirical body (not stub)
        if has_empirical_body(fn_name):
            skipped += 1
            continue
        
        info = parse_asm_pattern(asm_file)
        if not info or not info['calls']:
            errors.append((fn_name, 'no calls'))
            continue
        
        body = generate_go_body(info)
        if not body:
            errors.append((fn_name, 'no body generated'))
            continue
        
        if apply_replacement(fn_name, body):
            completed += 1
            if completed % 10 == 0:
                print(f"[{completed}] {fn_name} -> OK")
        else:
            errors.append((fn_name, 'replace failed'))
    
    print(f"\nDone: {completed} completed, {skipped} skipped, {len(errors)} errors")
    if errors:
        for name, err in errors[:10]:
            print(f"  {name}: {err}")
        if len(errors) > 10:
            print(f"  ... and {len(errors)-10} more")

def has_empirical_body(fn_name):
    """Check if function already has empirical (non-stub) body."""
    stub = get_stub_signature(fn_name)
    if not stub:
        return True
    with open(GOFILE, encoding='utf-8') as f:
        content = f.read()
    # Find function by name and check first line of body
    lines = content.split('\n')
    for i, line in enumerate(lines):
        if re.match(rf'^func\s*(?:\(\s*\w+\s*\*\w+\s*\))?\s*{re.escape(fn_name)}\s*\(', line):
            # Check next non-comment, non-annotation lines
            for j in range(i+1, min(i+5, len(lines))):
                cl = lines[j].strip()
                if cl.startswith('//') or cl.startswith('/*') or cl == '':
                    continue
                # If body has more than just 'return nil' etc, it's empirical
                if cl.startswith('return bs.') or cl.startswith('return ') and '.' in cl:
                    return True
                if cl.startswith('if bs == nil'):
                    return True
                if 'bs.' in cl and ('=' in cl or '(' in cl):
                    return True
                break
    return False

if __name__ == '__main__':
    main()