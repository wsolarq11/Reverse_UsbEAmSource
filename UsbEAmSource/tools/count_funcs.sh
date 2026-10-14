#!/bin/bash
# 标记口径统计 —— 唯一真相源（docs/STRATEGY.md §4）。
#
# 口径：标记只从「紧贴 func 的连续注释块首行」读取；块内出现非注释行即重置。
# 标记优先级：S-sig > S-eq > S-inline > S > P。
# 变体后缀（[S-sig]、[S-sig ]、[S-sig\t] 等）等价，见下方 index 判定。
#
# 输出（单行，双空格分隔，顺序固定）：
#   FUNCS MARKED UNMARKED S S-inline S-eq S-sig P FAITHFUL USABLE TRUE
# 其中 FAITHFUL = S + S-inline；USABLE = S + S-inline + S-eq；TRUE = S + S-inline + S-sig。
#
# 用量：bash tools/count_funcs.sh
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
    if (last != "") { c[last]++; c["MARKED"]++ } else { c["UNMARKED"]++ }
    c["FUNCS"]++
    last = ""
    next
}
{ last = "" }
END {
    s = c["S"] + 0
    inl = c["S-inline"] + 0
    eq = c["S-eq"] + 0
    sig = c["S-sig"] + 0
    printf "FUNCS=%d  MARKED=%d  UNMARKED=%d  S=%d  S-inline=%d  S-eq=%d  S-sig=%d  P=%d  FAITHFUL=%d  USABLE=%d  TRUE=%d\n", \
        c["FUNCS"] + 0, c["MARKED"] + 0, c["UNMARKED"] + 0, s, inl, eq, sig, c["P"] + 0, s + inl, s + inl + eq, s + inl + sig
}
' $(ls *.go | grep -v _test)
