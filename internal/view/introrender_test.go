package view

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/intro"
)

// brightPixels 数一个区域里明显亮于背景的像素，用来判断"标识画出来了没有"。
func brightPixels(img *image.RGBA, cx, cy, half int) int {
	n := 0
	for y := cy - half; y < cy+half; y++ {
		for x := cx - half; x < cx+half; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 > 120 || g>>8 > 120 || b>>8 > 120 {
				n++
			}
		}
	}
	return n
}

// TestIntroRenders 冒烟测试：开场在几个时间点都能画出来，且标识随生长变多变亮。
//
// 这里只做像素级断言，不写文件——渲染细节靠人看截图，回归靠这个。
func TestIntroRenders(t *testing.T) {
	render := func(at float32) *image.RGBA {
		a := NewApp()
		a.intro = intro.New(1280, 800)
		a.intro.T = at
		return ui.NewTester(func(c *ui.Context) {
			// 铺不透明底：离屏画布默认白色，而开场底色带透明度。
			ui.Column(c).Fill().Background(ui.Hex("#101A24")).Draw(func(p *ui.Painter, r ui.Rect) {
				a.drawIntro(p, r)
			})
		}, 1280, 800).Image()
	}

	early := brightPixels(render(0.30), 640, 520, 200)  // 刚开始抽枝
	formed := brightPixels(render(3.75), 640, 360, 200) // 完全成形
	late := brightPixels(render(5.50), 640, 360, 200)   // 已淡出

	if formed < 300 {
		t.Errorf("成形时标识太暗/太小：亮像素只有 %d", formed)
	}
	if early >= formed/2 {
		t.Errorf("缺少生长过程：0.3s 时亮像素 %d，成形时 %d", early, formed)
	}
	if late > formed/4 {
		t.Errorf("5.5s 应已基本淡出，亮像素仍有 %d（成形时 %d）", late, formed)
	}
}
