#!/bin/bash
# 列出 backend/*.go 中所有 [P] 存根（文件:行: 声明）。
# 标记判定与 tools/count_funcs.sh（唯一真相源，docs/STRATEGY.md §4）逐条一致，
# 优先级：S-sig > S-eq > S-inline > S > P。
# 需要 VA 与邻接长度时用 tools/p_funcs.py。
#
# 用量：bash tools/list_p.sh
cd "$(dirname "$0")/../backend" || exit 1

awk '
/^\/\// {
    if (last == "") {
        s = $0
        if (index(s, "[S-sig]") > 0 || index(s, "[S-sig ") > 0 || index(s, "[S-sig\t") > 0) last = "S-sig"
        else if (index(s, "[S-eq]") > 0 || index(s, "[S-eq ") > 0 || index(s, "[S-eq\t") > 0) last = "S-eq"
        else if (index(s, "[S-inline]") > 0 || index(s, "[S-inline ") > 0 || index(s, "[S-inline\t") > 0) last = "S-inline"
        else if (index(s, "[S]") > 0) last = "S"
        else if (index(s, "[S ") > 0 || index(s, "[S\t") > 0) last = "S"
        else if (index(s, "[P]") > 0) last = "P"
        else if (index(s, "[P ") > 0 || index(s, "[P\t") > 0) last = "P"
    }
    next
}
/^func / {
    if (last == "P") printf "%s:%d: %s\n", FILENAME, FNR, $0
    last = ""
    next
}
{ last = "" }
' $(ls *.go | grep -v _test)
