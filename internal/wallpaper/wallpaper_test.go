package wallpaper

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
)

// 测试用的画布尺寸：16:9，与素材（bg 3840×2160 / flow-mask 960×540）同比例，
// 这样 Cover 不做裁切，屏幕坐标与遮罩坐标一一对应。
const (
	testW, testH = 320, 180
)

// render 用指定时钟与强度渲染一小块壁纸（尺寸小是为了让 CPU 解释器跑得快）。
func render(t *testing.T, bg *ui.Bitmap, flow *Flow, time float64, play float32) *image.RGBA {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			Draw(p, bg, r, time, play, ui.Hex("#7FD7FF"), 0, flow)
		})
	}, testW, testH).Image()
}

// delta 数一个采样点上两张图的颜色差（三通道绝对值之和，0..765）。
func delta(x, y *image.RGBA, px, py int) int {
	r1, g1, b1, _ := x.At(px, py).RGBA()
	r2, g2, b2, _ := y.At(px, py).RGBA()
	d := 0
	for _, v := range [3]int{
		int(r1>>8) - int(r2>>8),
		int(g1>>8) - int(g2>>8),
		int(b1>>8) - int(b2>>8),
	} {
		if v < 0 {
			v = -v
		}
		d += v
	}
	return d
}

// diff 数一数两张图差异明显的采样点。
func diff(x, y *image.RGBA) int {
	n := 0
	b := x.Bounds()
	for py := b.Min.Y; py < b.Max.Y; py += 2 {
		for px := b.Min.X; px < b.Max.X; px += 2 {
			if delta(x, y, px, py) > 6 {
				n++
			}
		}
	}
	return n
}

// TestAnimation 钉住三条性质：着色器能编译并出图、时间推进会改变画面、play=0 时静止。
func TestAnimation(t *testing.T) {
	bg := assets.Bitmap("ganyu/bg.jpg")
	mask := assets.Bitmap("ganyu/ganyu-flow-mask.png")
	if bg == nil || mask == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}
	flow := &Flow{Mask: mask, Dir: ui.Point{X: 0.94, Y: -0.34}, Strength: 0.012}

	early := render(t, bg, flow, 0, 1)
	late := render(t, bg, flow, 8, 1)
	if moving := diff(early, late); moving < 30 {
		t.Errorf("时间推进 8 秒后画面几乎没变（%d 点差异），动效可能失效", moving)
	}

	still1 := render(t, bg, flow, 0, 0)
	still2 := render(t, bg, flow, 8, 0)
	if still := diff(still1, still2); still != 0 {
		t.Errorf("play=0 时画面仍在变化（%d 点），说明差异不是来自动画参数", still)
	}
}

// TestFlowStaysInsideMask 是防回归的关键一条：形变必须被遮罩圈住。
//
// 曾经的做法是叠第二张人物图并让它相对背景漂移，于是整幅画面都在动、
// 看得出"两个人"。这里用像素证明：只有上半部发区在动，下半部一个点都没动。
//
// flow-mask 实测覆盖 x 31%~82% / y 0%~42%（峰值在 y 17%~33%），下半张全 0。
func TestFlowStaysInsideMask(t *testing.T) {
	bg := assets.Bitmap("ganyu/bg.jpg")
	mask := assets.Bitmap("ganyu/ganyu-flow-mask.png")
	if bg == nil || mask == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}
	flow := &Flow{Mask: mask, Dir: ui.Point{X: 0.94, Y: -0.34}, Strength: 0.012}

	a := render(t, bg, flow, 0, 1)
	b := render(t, bg, flow, 3.7, 1)

	minX, maxX, minY, maxY := testW, -1, testH, -1
	moved := 0
	for y := 0; y < testH; y++ {
		for x := 0; x < testW; x++ {
			if delta(a, b, x, y) <= 6 {
				continue
			}
			moved++
			minX, maxX = min(minX, x), max(maxX, x)
			minY, maxY = min(minY, y), max(maxY, y)
		}
	}

	if moved < 40 {
		t.Fatalf("只有 %d 个像素在动，形变几乎不可见（幅度或遮罩可能没生效）", moved)
	}
	// 下半张必须纹丝不动——这是"人物主体不动"的硬指标。
	if maxY >= testH/2 {
		t.Errorf("形变越过了画面中线：最靠下的动点 y=%d（中线 %d），主体被带着动了", maxY, testH/2)
	}
	// 左右边界也要落在遮罩 bbox（31%~82%）的附近，留 5% 容差给遮罩软化和位移。
	lo, hi := testW*26/100, testW*87/100
	if minX < lo || maxX > hi {
		t.Errorf("形变横向越界：x %d..%d，遮罩区应在 %d..%d（含 5%% 容差）", minX, maxX, lo, hi)
	}
	t.Logf("动点 %d 个，范围 x %d..%d / y %d..%d（画布 %dx%d）", moved, minX, maxX, minY, maxY, testW, testH)
}
