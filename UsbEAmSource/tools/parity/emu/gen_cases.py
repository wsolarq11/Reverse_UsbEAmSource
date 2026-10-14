#!/usr/bin/env python
# 研究用途：对拍用例生成器。
#
# 用例设计目标（覆盖 asm 中每一条已确证分支）：
#   - rank 不等：a/b 分别落在 rank 1 / rank 2 / -1，三档两两组合
#   - 提前出局：0xc356（Image）与 0xc364（Text）分别单独出现
#   - ok=false 两侧
#   - 矩形退化：Dx<2、Dy<2，含 Dx=0/1、Dy=0/1，以及负宽高
#   - 面积分支：areaB<=0、areaA<=0、areaB<areaA、areaB>areaA、两者相等
#   - 同 rank 同面积（不可达内层分支的落点）
#   - 窗口边界：0xc358-1 / 0xc358 / 0xc358+0x19 / 0xc358+0x1a
#
# 用法: python gen_cases.py > cases.txt
import itertools
import random
import sys

try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass

# 由表项推导的 rank=1 集合（与 ref_puiahi.py 同一推导；此处显式列出以便生成器自洽）。
# 该列表不得手改：改动需回到 extract 产物重新推导。
RANK1 = [50008, 50009, 50010, 50017, 50018, 50021, 50023, 50026, 50028, 50030, 50032, 50033]
# 窗口 50008..50033 内、未被 T1 判为 rank=1 的值。50020 被下方 EARLY 提前出局，
# 故不属于 rank=2；集合间必须两两不交（由 assert 保证）。
RANK2 = [50011, 50012, 50013, 50014, 50015, 50016, 50019, 50022, 50024, 50025, 50027, 50029, 50031]
EARLY = [50006, 50020]

assert not (set(RANK1) & set(RANK2)), "RANK1/RANK2 相交"
assert not (set(RANK1) & set(EARLY)), "RANK1/EARLY 相交"
assert not (set(RANK2) & set(EARLY)), "RANK2/EARLY 相交"
assert set(RANK1) | set(RANK2) | set(EARLY) | {50007} == set(range(50006, 50034)), "窗口覆盖不完整"
OUT_OF_WINDOW = [50007, 50033, 50034, 0, 100, 49999, 50005, 50035]


# 用例格式：四个角坐标直给（aminx aminy amaxx amaxy aCT aOK bminx bminy bmaxx bmaxy bCT bOK）。
# 不用「左上角 + 宽高」：Go 的 image.Rect 会交换 Min/Max 使矩形合法，而原函数的入参
# 已由调用方构造完毕。直给四角可消除该规整语义在两套实现间的分歧，并使 Min>Max 的
# 退化形态可被显式表达。


def fmt(rect, ct, ok):
    mnx, mny, mxx, mxy = rect
    return "%d %d %d %d %d %d" % (mnx, mny, mxx, mxy, ct, 1 if ok else 0)


def main():
    lines = []

    def add(a, act, aok, b, bct, bok):
        lines.append("%s %s" % (fmt(a, act, aok), fmt(b, bct, bok)))

    R = lambda x, y, w, h: (x, y, x + w, y + h)
    # 退化形态：直接给 Min/Max，Min>Max 表示原始矩形已被 caller 构造为退化。
    D = lambda mnx, mny, mxx, mxy: (mnx, mny, mxx, mxy)
    r1 = R(0, 0, 10, 10)
    r2 = R(100, 100, 20, 20)
    r3 = R(50, 50, 5, 5)
    small = R(0, 0, 1, 1)
    thin_x = R(0, 0, 1, 10)
    thin_y = R(0, 0, 10, 1)
    zero_w = D(0, 0, 0, 5)
    zero_h = D(0, 0, 5, 0)
    inv_x = D(10, 10, 5, 20)      # Max.X < Min.X
    inv_y = D(10, 10, 20, 5)      # Max.Y < Min.Y
    inv_both = D(10, 10, 5, 5)

    # 1) rank 三档两两组合（同 ok、同矩形几何族）
    for act, aok in itertools.product(RANK1[:2] + RANK2[:2] + EARLY, [True]):
        for bct in RANK1[:2] + RANK2[:2] + EARLY:
            add(r1, act, aok, r2, bct, True)
            add(r2, act, aok, r1, bct, True)

    # 2) ok=false 组合
    for bct in RANK1[:1] + RANK2[:1]:
        add(r1, RANK1[0], False, r2, bct, True)
        add(r1, RANK1[0], True, r2, bct, False)
        add(r1, RANK1[0], False, r2, bct, False)

    # 3) 矩形退化（两侧各种退化形态互配）
    degen = [small, thin_x, thin_y, zero_w, zero_h, inv_x, inv_y, inv_both]
    for a in degen:
        for b in [r1, r2, small, zero_w]:
            add(a, RANK1[0], True, b, RANK1[0], True)
            add(b, RANK1[0], True, a, RANK1[0], True)

    # 4) 面积分支（同 rank 同 CT，走面积比较）
    for a, b in [(r1, r2), (r2, r1), (r3, r1), (r1, r3), (r1, r1), (r3, r3)]:
        add(a, RANK1[0], True, b, RANK1[0], True)

    # 5) 同 rank 同面积（内层不可达分支落点）
    add(R(0, 0, 10, 10), RANK1[0], True, R(0, 0, 10, 10), RANK1[1], True)
    add(R(0, 0, 10, 10), RANK2[0], True, R(0, 0, 10, 10), RANK2[1], True)

    # 6) 窗口边界
    for ct in [0xC357, 0xC358, 0xC371, 0xC372, 0xC356, 0xC364, 50005, 50033, 50034]:
        add(r1, ct, True, r2, RANK1[0], True)
        add(r1, RANK1[0], True, r2, ct, True)

    # 7) 全窗口扫描（0xc350..0xc380）
    for ct in range(0xC350, 0xC381):
        add(r1, ct, True, r2, RANK1[0], True)
        add(r1, RANK1[0], True, r2, ct, True)

    # 8) rank=1 全集合与 rank=2 全集合枚举
    for ct in RANK1:
        add(r1, ct, True, r2, ct, True)
        add(r2, ct, True, r1, ct, True)
        add(r2, ct, True, r1, RANK1[0], True)
    for ct in RANK2:
        add(r1, ct, True, r2, ct, True)
        add(r2, ct, True, r1, RANK1[0], True)
        add(r1, RANK1[0], True, r2, ct, True)

    # 9) 越窗类型
    for ct in OUT_OF_WINDOW:
        add(r1, ct, True, r2, RANK1[0], True)
        add(r1, RANK1[0], True, r2, ct, True)

    # 10) 随机补充（固定种子，可复现）
    # 直接生成四角坐标，允许 Min>Max 的退化形态（原函数收到的是已构造完毕的矩形）。
    rnd = random.Random(397)
    for _ in range(1200):
        def mk():
            mnx = rnd.randrange(-50, 300)
            mny = rnd.randrange(-50, 300)
            mxx = mnx + rnd.choice([-10, -1, 0, 1, 2, 3, 5, 10, 40, 200])
            mxy = mny + rnd.choice([-10, -1, 0, 1, 2, 3, 5, 10, 40, 200])
            return (mnx, mny, mxx, mxy)
        ct_pool = RANK1 + RANK2 + EARLY + OUT_OF_WINDOW + list(range(0xC350, 0xC381))
        add(mk(), rnd.choice(ct_pool), rnd.random() < 0.9,
            mk(), rnd.choice(ct_pool), rnd.random() < 0.9)

    sys.stdout.write("\n".join(lines) + "\n")
    sys.stderr.write("用例数 %d\n" % len(lines))
    return 0


if __name__ == "__main__":
    sys.exit(main())
