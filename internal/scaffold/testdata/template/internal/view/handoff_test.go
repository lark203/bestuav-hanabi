package view

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// slotBright 统计品牌图标槽位（20..60, 64..104）内明显亮于背景的像素数。
func slotBright(img *image.RGBA) int {
	n := 0
	for y := 64; y < 104; y++ {
		for x := 20; x < 60; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 > 110 || g>>8 > 110 || b>>8 > 110 {
				n++
			}
		}
	}
	return n
}

// TestBrandHandoff 启动动画期间品牌位必须留空（位置保留给飞入的标识），
// 动画结束后同一位置显示同款元素标识——「开场飞入品牌位」的交接契约。
func TestBrandHandoff(t *testing.T) {
	a := NewApp()
	a.intro = visual.New(1280, 800)
	a.intro.T = 2.5
	empty := slotBright(ui.NewTester(a.View, 1280, 800).Image())

	b := NewApp()
	b.introDone = true
	shown := slotBright(ui.NewTester(b.View, 1280, 800).Image())

	t.Logf("开场中品牌位亮点=%d，结束后=%d", empty, shown)
	if empty > 20 {
		t.Fatalf("开场中品牌位没有留空（亮点 %d）", empty)
	}
	if shown < 40 {
		t.Fatalf("结束后品牌位没有显示标识（亮点 %d）", shown)
	}
}
