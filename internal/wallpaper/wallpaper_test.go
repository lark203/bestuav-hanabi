package wallpaper

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
)

// render 用指定时钟与强度渲染一小块壁纸（尺寸小是为了让 CPU 解释器跑得快）。
func render(t *testing.T, bg *ui.Bitmap, time float64, play float32) *image.RGBA {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			Draw(p, bg, r, time, play, ui.Hex("#7FD7FF"), 0)
		})
	}, 160, 90).Image()
}

// diff 数一数两张图差异明显的采样点。
func diff(x, y *image.RGBA) int {
	n := 0
	b := x.Bounds()
	for py := b.Min.Y; py < b.Max.Y; py += 2 {
		for px := b.Min.X; px < b.Max.X; px += 2 {
			r1, g1, b1, _ := x.At(px, py).RGBA()
			r2, g2, b2, _ := y.At(px, py).RGBA()
			d := int(r1>>8) - int(r2>>8) + int(g1>>8) - int(g2>>8) + int(b1>>8) - int(b2>>8)
			if d < 0 {
				d = -d
			}
			if d > 6 {
				n++
			}
		}
	}
	return n
}

// TestAnimation 钉住三条性质：着色器能编译并出图、时间推进会改变画面、play=0 时静止。
func TestAnimation(t *testing.T) {
	bg := assets.Bitmap("ganyu/ganyu-bg-4k.jpg")
	if bg == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}

	early := render(t, bg, 0, 1)
	late := render(t, bg, 8, 1)
	if moving := diff(early, late); moving < 30 {
		t.Errorf("时间推进 8 秒后画面几乎没变（%d 点差异），动效可能失效", moving)
	}

	still1 := render(t, bg, 0, 0)
	still2 := render(t, bg, 8, 0)
	if still := diff(still1, still2); still != 0 {
		t.Errorf("play=0 时画面仍在变化（%d 点），说明差异不是来自动画参数", still)
	}
}
