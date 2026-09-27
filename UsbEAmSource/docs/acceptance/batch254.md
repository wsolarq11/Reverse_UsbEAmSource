# 批次 254 · oledblackout 短函数 +4（key 生成 / 编号解析 / 排序 / 输入快照）

## 目标

落地 `backend/oledblackout.go` 的 4 个短函数：媒体暂停排除 key 生成、屏幕编号解析、
屏幕排序、输入快照读取。均为 gap_aggregate 长度升序中已存在文件内的短函数。

## 基线 / 收口

| 指标 | 基线（batch 253 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2822 | 2826 |
| MARKED | 2822 | 2826 |
| S | 1290 | 1294 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1326 | 1330 |
| 真函数（S+S-inline+S-sig） | 2782（58.52%） | 2786（58.60%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+4，全部 [S]）

### backend/oledblackout.go（+4 [S]，新增 import sort/strconv）

1. `oledBlackoutMediaPauseExclusionKey(path, processName string) string` `[S 0x1408fdf40]`
   = normalizePath → strings.ToLower → 非空则 `lower + "path:"`；空则
   normalizeProcessName → `name + "url(http"`。字符串常量经 rodata 解码实证：
   `"path:"`（5B）与 `"url(http"`（8B，无右括号）。
2. `parseOLEDBlackoutScreenNumber(name string) int` `[S 0x140906420]`
   = 逐字节收集 '0'-'9' 到 []byte → 空则 0；否则 strconv.Atoi，err!=nil 则 0。
3. `sortOLEDBlackoutScreens(screens []*OLEDBlackoutScreen)` `[S 0x140906280]`
   = len<=1 直接返回；否则 sort.Slice 闭包比较 parseOLEDBlackoutScreenNumber(Name)，
   编号相同则 Name 字符串比较（cmpstring+setl）。
4. `oledBlackoutReadInputSnapshot() (int, int, bool, map[uintptr]struct{})` `[S 0x14090be40]`
   = 经两个全局函数值调用 oledBlackoutCurrentCursorPhysicalPoint /
   oledBlackoutPressedKeyboardKeys，返回 (x, y, ok, keys)。

## 关键知悉

- `sortOLEDBlackoutScreens` 元素为 8 字节指针（asm `[rdx+rax*8]`），即 `[]*OLEDBlackoutScreen`，
  比较字段为 Name（元素 +0x10=ptr/+0x18=len），与 types_oled.go 布局（ID@0x00、Name@0x10）一致。
- `oledBlackoutReadInputSnapshot` 的两个全局函数值位于 .data @0x141BC1B30/@0x141BC1B38，
  分别指向 funcval（fn=oledBlackoutCurrentCursorPhysicalPoint@0x140914aa0 /
  oledBlackoutPressedKeyboardKeys@0x140914ac0）。落地以直接调用等价还原（hook 变量默认即此二函数）。
- `oledBlackoutProfileKey`（0x1408fef00）与 `dismissOverlayForKeyboardInputLocked`（0x14090b2a0）
  本批未落地：前者依赖 `normalizeOLEDBlackoutProfileScreens`（2240B 大函数，未落地），后者依赖
  `visibleOverlayForKeyboardInputLocked`/`dismissOverlayForInputLocked`（未落地方法链），
  按「签名未定型拒绝落体」留待专项批次。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt`（top-level=883）长度升序落地已存在文件短函数；
下一批优先 windowmanagement_windows.go 的 wrap/rect 短函数链（128–192B）。
FUNCS 2826/4754 = 59.44%。
