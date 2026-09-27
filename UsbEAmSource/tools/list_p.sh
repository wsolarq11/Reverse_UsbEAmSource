#!/bin/bash
cd "$(dirname "$0")/../backend"
awk '
/^\/\// {
    if (last == "") {
        s = $0
        if (index(s, "[S-sig]") > 0 || index(s, "[S-sig ") > 0) last = "S-sig"
        else if (index(s, "[S-inline]") > 0) last = "S-inline"
        else if (index(s, "[S]") > 0) last = "S"
        else if (index(s, "[S ") > 0 || index(s, "[S\t") > 0) last = "S"
        else if (index(s, "[P]") > 0) last = "P"
        else if (index(s, "[P ") > 0 || index(s, "[P\t") > 0) last = "P"
    }
    next
}
/^func / {
    if (last == "P") printf "%s: %s\n", FILENAME, $0
    last = ""
    next
}
{ last = "" }
' $(ls *.go | grep -v _test)
