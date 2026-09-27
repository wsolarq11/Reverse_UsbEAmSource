#!/usr/bin/env python3
"""Split backend/types.go into domain files."""
import re, os, shutil
from collections import defaultdict

SRC = "../backend/types.go"
OUT = "../backend"

RULES = [
    ('BootstrapService', 'launcher'), ('FileIndexService', 'filesearch'),
    ('PluginHostService', 'plugin'), ('VolumeIndex', 'filesearch'),
    ('WindowManagementConfig', 'windowmgt'), ('LauncherNetworkAccess', 'network'),
    ('LauncherNetworkRedirectPolicyAccess', 'network'), ('GestureSettings', 'gesture'),
    ('screenshot', 'screenshot'), ('Screenshot', 'screenshot'),
    ('audio', 'audio'), ('Audio', 'audio'),
    ('bookmark', 'bookmark'), ('Bookmark', 'bookmark'), ('chromium', 'bookmark'), ('firefox', 'bookmark'),
    ('desktop', 'desktopwidget'), ('Desktop', 'desktopwidget'), ('openMeteo', 'desktopwidget'),
    ('qweather', 'desktopwidget'), ('weather', 'desktopwidget'),
    ('fileLocator', 'filelocator'), ('FileLocator', 'filelocator'),
    ('fileSearch', 'filesearch'), ('FileSearch', 'filesearch'), ('FileIndex', 'filesearch'),
    ('volume', 'filesearch'), ('Volume', 'filesearch'), ('rawIndex', 'filesearch'),
    ('namePrefix', 'filesearch'), ('nameBigram', 'filesearch'), ('nameTrigram', 'filesearch'),
    ('scored', 'filesearch'), ('nodePath', 'filesearch'),
    ('AppEntry', 'app'), ('AppImport', 'app'), ('ConsoleItem', 'app'),
    ('StartMenu', 'app'), ('startMenu', 'app'), ('Shortcut', 'app'), ('shortcut', 'app'),
    ('Drag', 'app'), ('FileEntry', 'app'), ('LinkEntry', 'app'), ('LinkBrowser', 'app'),
    ('DetectedLink', 'app'), ('Gesture', 'gesture'), ('gesture', 'gesture'),
    ('Mouse', 'gesture'), ('mouse', 'gesture'), ('HotCorner', 'gesture'),
    ('gpu', 'gpu'), ('GPU', 'gpu'),
    ('inputMonitor', 'inputmonitor'), ('InputMonitor', 'inputmonitor'),
    ('launcher', 'launcher'), ('Launcher', 'launcher'), ('Bootstrap', 'launcher'),
    ('bootstrap', 'launcher'), ('Language', 'launcher'), ('StartupState', 'launcher'),
    ('memory', 'memoryrelease'), ('Memory', 'memoryrelease'),
    ('oled', 'oled'), ('oledBlackout', 'oled'), ('OLED', 'oled'),
    ('plugin', 'plugin'), ('Plugin', 'plugin'),
    ('qr', 'qrcode'), ('qrCode', 'qrcode'), ('QR', 'qrcode'),
    ('twoFactor', 'twofactor'), ('TwoFactor', 'twofactor'),
    ('WindowManagement', 'windowmgt'), ('windowManagement', 'windowmgt'),
    ('WindowFullscreen', 'windowmgt'), ('Workspace', 'workspace'), ('workspace', 'workspace'),
    ('configBacked', 'network'), ('directLauncher', 'network'), ('LauncherNetwork', 'network'),
    ('remote', 'icon'), ('Remote', 'icon'),
    ('displayconfig', 'windows'), ('dxgi', 'windows'), ('shFile', 'windows'),
    ('memoryStatusEx', 'windows'), ('nativeFile', 'windows'), ('shellExecute', 'windows'),
    ('Config', 'config'), ('config', 'config'), ('Preferences', 'config'),
    ('StorageConfig', 'config'), ('Background', 'config'), ('TagCatalog', 'config'),
    ('Initialization', 'config'), ('Migrate', 'config'), ('staged', 'config'),
]

def classify(name):
    for prefix, domain in RULES:
        if name.startswith(prefix):
            return domain
    return 'misc'

def main():
    lines = open(SRC, 'r', encoding='utf-8').readlines()
    
    imp_start = imp_end = -1
    for i, line in enumerate(lines):
        if line.strip() == 'import (':
            imp_start = i
        if imp_start >= 0 and line.strip() == ')':
            imp_end = i
            break
    
    import_paths = {}
    for line in lines[imp_start+1:imp_end]:
        s = line.strip()
        if s:
            m = re.search(r'"([^"]+)"', s)
            if m:
                path = m.group(1)
                short = path.rstrip('/').split('/')[-1]
                import_paths[short] = path
    
    type_defs = []
    i = imp_end + 1
    while i < len(lines):
        stripped = lines[i].strip()
        m = re.match(r'^type\s+(\w+)', stripped)
        if m:
            name = m.group(1)
            start = i
            if '{' in stripped:
                depth = 0
                j = i
                while j < len(lines):
                    depth += lines[j].count('{') - lines[j].count('}')
                    if depth <= 0:
                        type_defs.append((name, start, j))
                        i = j + 1
                        break
                    j += 1
                else:
                    i += 1
            else:
                j = i + 1
                while j < len(lines) and not lines[j].strip().startswith('type '):
                    j += 1
                type_defs.append((name, start, j - 1))
                i = j
        else:
            i += 1
    
    print(f"Found {len(type_defs)} type definitions")
    
    domain_groups = defaultdict(list)
    for name, start, end in type_defs:
        domain_groups[classify(name)].append((name, start, end))
    
    for domain in sorted(domain_groups.keys()):
        print(f"  {domain}: {len(domain_groups[domain])} types")
    
    domain_imports = defaultdict(set)
    for domain, types in domain_groups.items():
        text = ''
        for name, start, end in types:
            for line in lines[start:end+1]:
                text += line + '\n'
        for short, path in import_paths.items():
            if short in text:
                domain_imports[domain].add(path)
        for pkg in ['sync', 'time', 'io', 'os', 'net/http', 'image', 'image/color',
                    'regexp', 'strings', 'encoding/xml', 'database/sql', 'context',
                    'archive/zip', 'unsafe', 'fmt', 'sort', 'math', 'strconv',
                    'encoding/json', 'encoding/base64', 'net', 'net/url',
                    'path/filepath', 'bytes', 'crypto/sha256', 'crypto/hmac',
                    'crypto/rand', 'hash', 'io/fs', 'log', 'reflect']:
            if pkg.split('/')[-1] in text:
                domain_imports[domain].add(pkg)
    
    for domain in sorted(domain_groups.keys()):
        types = sorted(domain_groups[domain], key=lambda x: x[1])
        out_lines = ['// AUTO-RECONSTRUCTED TYPES — DOMAIN: ' + domain,
                     '// 研究用途', 'package main', '']
        imports = sorted(domain_imports[domain])
        if imports:
            out_lines.append('import (')
            for imp in imports:
                out_lines.append('\t"' + imp + '"')
            out_lines.append(')')
            out_lines.append('')
        for name, start, end in types:
            for line in lines[start:end+1]:
                out_lines.append(line.rstrip('\n'))
            out_lines.append('')
        filename = 'types_' + domain + '.go'
        open(os.path.join(OUT, filename), 'w', encoding='utf-8').write('\n'.join(out_lines))
        print(f"  Written: {filename}")
    
    shutil.copy2(SRC, SRC + '.bak')
    print(f"\nBackup: {SRC}.bak")

if __name__ == '__main__':
    main()