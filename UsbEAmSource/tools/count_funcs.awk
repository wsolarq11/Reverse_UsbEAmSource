#!/usr/bin/env awk -f
# 精确函数级档位统计 —— 修正版
# 匹配 [S]、[S-sig]、[S-inline]、[P] 子串（含前缀/后缀变体如 "[S 汇编...]"）
# 用法: cd backend && awk -f ../tools/count_funcs.awk *.go | sort

/^\/\// {
    if (last == "") {
        s = $0
        if (index(s, "[S-sig]") > 0 || index(s, "[S-sig ") > 0 || index(s, "[S-sig\t") > 0) last = "S-sig"
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
{
    last = ""
}
END {
    for (k in c) printf "%s=%d\n", k, c[k]
}