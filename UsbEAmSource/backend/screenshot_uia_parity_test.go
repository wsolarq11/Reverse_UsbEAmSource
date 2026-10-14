// 研究用途：screenshotPreferUIAElementHitInfo 行为对拍测试。
//
// 对拍契约（见 tools/parity/README.md）：
//
//	期望值不得由 Go 实现反推。本测试读取的 puiahi.expected.txt 由
//	tools/parity/emu/ref_puiahi.py 生成；该参考实现的 rank 表由原 exe 跳转表
//	（tools/parity/bin/*.meta.json）逐表项机械推导，非手写。
//
// 用例格式（puiahi.cases.txt 每行）：四个角坐标直给，不经 image.Rect 规范化。
//
//	aminx aminy amaxx amaxy aCT aOK bminx bminy bmaxx bmaxy bCT bOK
//
// 之所以不用「左上角 + 宽高」：image.Rect 会交换 Min/Max 使矩形合法，而原函数的
// 入参已由调用方用 image.Rect 构造完毕。直给四角可消除该规整语义在两套实现间的分歧，
// 也使 Min>Max 的退化形态可被显式表达。
//
// 期望格式（puiahi.expected.txt 每行）：0 或 1。
//
// 用例或期望文件缺失时跳过（对拍资产不在源码树内发布时不应阻断构建）。
package main

import (
	"bufio"
	"image"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	parityCasesPath    = "../tools/parity/cases/screenshot_uia/puiahi.cases.txt"
	parityExpectedPath = "../tools/parity/cases/screenshot_uia/puiahi.expected.txt"
)

func loadLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("对拍资产缺失，跳过: %v", err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	return out
}

func mustInt(t *testing.T, s string) int {
	t.Helper()
	v, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("用例字段非整数 %q: %v", s, err)
	}
	return v
}

func TestParityScreenshotPreferUIAElementHitInfo(t *testing.T) {
	cases := loadLines(t, parityCasesPath)
	expected := loadLines(t, parityExpectedPath)
	if len(cases) != len(expected) {
		t.Fatalf("用例数 %d 与期望数 %d 不一致", len(cases), len(expected))
	}

	seen := make(map[int]int)
	for i, line := range cases {
		f := strings.Fields(line)
		if len(f) != 12 {
			t.Fatalf("第 %d 行字段数 %d != 12", i+1, len(f))
		}
		aminx, aminy := mustInt(t, f[0]), mustInt(t, f[1])
		amaxx, amaxy := mustInt(t, f[2]), mustInt(t, f[3])
		aCT, aOK := mustInt(t, f[4]), mustInt(t, f[5])
		bminx, bminy := mustInt(t, f[6]), mustInt(t, f[7])
		bmaxx, bmaxy := mustInt(t, f[8]), mustInt(t, f[9])
		bCT, bOK := mustInt(t, f[10]), mustInt(t, f[11])

		got := screenshotPreferUIAElementHitInfo(
			image.Rectangle{Min: image.Point{X: aminx, Y: aminy}, Max: image.Point{X: amaxx, Y: amaxy}},
			uint32(aCT), aOK != 0,
			image.Rectangle{Min: image.Point{X: bminx, Y: bminy}, Max: image.Point{X: bmaxx, Y: bmaxy}},
			uint32(bCT), bOK != 0,
		)
		want := expected[i] == "1"
		if got != want {
			t.Errorf("用例 %d 不符: %s\n  重建=%v 参考=%v", i+1, line, got, want)
			if seen[0]++; seen[0] >= 20 {
				t.Fatalf("累计不符已达 20 例，中止")
			}
		}
	}
	t.Logf("对拍通过 %d 例", len(cases))
}
