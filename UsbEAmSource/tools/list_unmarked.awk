/^\/\// {
    if (last == "") {
        s = $0
        if (index(s, "[S") > 0 || index(s, "[P") > 0) last = "X"
    }
    next
}
/^func / {
    if (last == "") { print FILENAME ": " $0 }
    last = ""
    next
}
{ last = "" }
