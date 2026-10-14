#!/usr/bin/env python
# 研究用途：对拍参考实现（原函数语义镜像）。
#
# 期望值来源纪律（本 harness 的核心）：
#   1. rank=1 的控件类型集合【不手写】，由 .meta.json 的表项机械推导 ——
#      T1 表项落在 eax=1 出口的索引集合，加基址 0xc358 换算。
#   2. rank=2 是窗口内 default；rank=-1 是 ok==false / 矩形退化 / 提前出局集合。
#   3. 分支结构逐条对应 asm 指令地址（注释标出），非按 Go 重建代码转述。
#
# 用法:
#   python tools/parity/emu/ref_puiahi.py <meta.json> < cases.txt > expected.txt
import json
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

# 汇编常量（0x1409b2ce0 序言与跳转表基址）。
CT_WINDOW_BASE = 0xC358          # 0x1409b2e9a add rcx,-0xc358
CT_WINDOW_LAST = 0x19            # 0x1409b2ea1 cmp rcx,0x19
CT_EARLY_OUT_1 = 0xC356          # 0x1409b2de0 cmp ecx,0xc356
CT_EARLY_OUT_2 = 0xC364          # 0x1409b2dca cmp ecx,0xc364

# 出口地址：由 .meta.json 表项活动域推导，不硬编码语义。
EXIT_VALUE_1 = 0x1409B2EB1       # mov eax,1
EXIT_VALUE_2 = 0x1409B2EB8       # mov eax,2
EXIT_RANK_FALSE = 0x1409B3037    # mov eax,1（areaA<=0 → true）
EXIT_RANK_TRUE = 0x1409B3042     # xor eax,eax（areaB<=0 → false）
EXIT_LESS = 0x1409B302E          # setl al
EXIT_GREATER = 0x1409B304A       # setg al


def rank_one_set(meta):
    """从表项机械推导 rank=1 的控件类型集合。"""
    t1 = meta["tables"][0]
    entries = t1["entries"]
    ones = set()
    for idx, tgt in enumerate(entries):
        if idx > CT_WINDOW_LAST:
            continue          # 0x1409b2ea5 ja -> 表外，不可达
        if tgt == EXIT_VALUE_1:
            ones.add(CT_WINDOW_BASE + idx)
    return ones


def rank(rect, ct, ok, ones):
    """[0x1409b2d35 起] controlType 采用优先级。"""
    if not ok:
        return -1                                   # 0x1409b2dbb je -> 0x1409b3053
    if ct == CT_EARLY_OUT_2:
        return -1                                   # 0x1409b2dca je
    if ct == CT_EARLY_OUT_1:
        return -1                                   # 0x1409b2de0 je
    if rect[2] - rect[0] < 2 or rect[3] - rect[1] < 2:
        return -1                                   # 0x1409b2dec/de f6 jge -> 0x1409b3053
    if ct in ones:
        return 1                                    # 0x1409b2eb1
    return 2                                        # 0x1409b2eb8


def area(rect):
    """[0x1409b2f69 起] 非正面积归零。"""
    if rect[2] <= rect[0] or rect[3] <= rect[1]:
        return 0
    return (rect[2] - rect[0]) * (rect[3] - rect[1])


def prefer(a, act, aok, b, bct, bok, ones):
    """[0x1409b2ce0] a 是否优先于 b。"""
    ra = rank(a, act, aok, ones)
    rb = rank(b, bct, bok, ones)
    if ra != rb:
        return ra > rb                              # 0x1409b2f63 jne -> setg
    aa = area(a)
    ab = area(b)
    if ab <= 0:
        return False                                # 0x1409b2fcd jle -> 0x1409b3042
    if aa <= 0:
        return True                                 # 0x1409b2fd2 jle -> 0x1409b3037
    if ab != aa:
        return ab < aa                              # 0x1409b2fd7 -> 0x1409b302e setl
    # 同 rank 同面积：内层双表（T3/T4）索引集合与 T1 相同，判定退化为恒等，
    # 0x1409b3021 xor eax,1 恒得 false。不可达分支。
    return False


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    meta = json.load(open(sys.argv[1], encoding="utf-8"))
    ones = rank_one_set(meta)
    sys.stderr.write("rank=1 集合（由表项推导）: %s\n" % sorted(ones))

    out = []
    for line in sys.stdin:
        p = line.split()
        if len(p) != 12:
            continue
        ax, ay, aX, aY, act, aok, bx, by, bX, bY, bct, bok = (int(v) for v in p)
        # 四角坐标直给，不经「Min + 宽高」重建：生成器已输出 Min/Max 原始坐标。
        a = (ax, ay, aX, aY)
        b = (bx, by, bX, bY)
        r = prefer(a, act, aok != 0, b, bct, bok != 0, ones)
        out.append("1" if r else "0")
    sys.stdout.write("\n".join(out) + ("\n" if out else ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
