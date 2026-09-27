import re
src = open('main_types_reconstructed.go', encoding='utf-8').read()
# 1) strip the banner and package
src = src.replace('package main\n\n', '')
# 2) 'type main.X' -> 'type X'
src = re.sub(r'type main\.', 'type ', src)
# 3) field type 'main.X' -> 'X'
src = re.sub(r'\bmain\.([A-Z]\w*)', r'\1', src)
# 4) remove go1.25.12 stray token if any
src = src.replace('go1.25.12', '')
# assemble final
hdr = '''// AUTO-RECONSTRUCTED TYPES (研究用途) — 来自目标 UsbEAm_Launcher.exe 的 go1.25.12 pclntab 类型还原
package main

'''
imports = '''import (
	"context"
	"database/sql"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
	"archive/zip"
	"encoding/xml"
	"golang.org/x/sys/windows"
	"golang.org/x/text/collate"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
	"github.com/wailsapp/go-webview2/pkg/combridge"
	"modernc.org/sqlite"
	"time"
	"fmt"
)

'''
open('types_normalized.go','w',encoding='utf-8').write(import_type + imports + '\n' + src)
