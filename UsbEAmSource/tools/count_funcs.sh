#!/bin/bash
cd "$(dirname "$0")/../backend"
awk '
/^\/\// {
    if (last == "") {
        s = $0
        if (index(s, "[S-sig]") > 0 || index(s, "[S-sig ") > 0 || index(s, "[S-sig\t") > 0) last = "S-sig"
        else if (index(s, "[S-eq]") > 0 || index(s, "[S-eq ") > 0 || index(s, "[S-eq\t") > 0) last = "S-eq"
        else if (index(s, "[S-inline]") > 0 || index(s, "[S-inline ") > 0) last = "S-inline"
        else if (index(s, "[S]") > 0) last = "S"
        else if (index(s, "[S ") > 0 || index(s, "[S\t") > 0) last = "S"
        else if (index(s, "[P]") > 0) last = "P"
        else if (index(s, "[P ") > 0 || index(s, "[P\t") > 0) last = "P"
    }
    next
}
/^func / {
    if (last != "") {
        c[last]++
        c["MARKED"]++
    } else {
        c["UNMARKED"]++
    }
    c["FUNCS"]++
    last = ""
    next
}
{ last = "" }
END {
    printf "FAITHFUL=%d\n", c["S"] + c["S-inline"]
    printf "USABLE=%d\n", c["S"] + c["S-inline"] + c["S-eq"]
    for (k in c) printf "%s=%d\n", k, c[k]
}' $(ls *.go | grep -v _test) | sort