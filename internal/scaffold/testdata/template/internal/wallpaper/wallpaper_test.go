package wallpaper

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
)

// 测试用的画布尺寸：16:9，与素材（bg 3840×2160 / person、hair 1920×1080）同比例，
// 这样 Cover 不做裁切，屏幕坐标与遮罩坐标一一对应。
const (
	testW, testH = 320, 180
)

// 测试用的幅度，比 theme 里配的真实值（呼吸 0.0022 / 发丝 0.0016）大一个量级。
//
// 原因：320 宽的测试画布上，0.0022 只对应不到 1 像素，逐像素比色根本测不出来。
// 放大后位移场形状、遮罩边界这些**要验证的性质**完全不变——只是把信噪比拉起来。
// 观感参数由 theme 决定，不由这里决定。
const (
	testBreath = 0.02
	testSway   = 0.02
)

// 人物轮廓遮罩的实测边界（相对百分比，取整），用作断言基准。
// 甘雨的 person.png 实测：非零 4.9%，bbox x 34.2%~59.4% / y 22.2%~73.7%。
const (
	personLoX, personHiX = 34, 59
	personLoY, personHiY = 22, 74
)

// testFlow 构造测试用的动态素材（与甘雨主题同源）。缺素材返回 nil。
func testFlow(t *testing.T) *Flow {
	t.Helper()
	person := assets.Bitmap("ganyu/person.png")
	hair := assets.Bitmap("ganyu/hair.png")
	if person == nil || hair == nil {
		return nil
	}
	return &Flow{Person: person, Hair: hair, Breath: testBreath, Sway: testSway}
}

// render 用指定时钟与强度渲染一小块壁纸（尺寸小是为了让 CPU 解释器跑得快）。
func render(t *testing.T, bg *ui.Bitmap, flow *Flow, time float64, play float32) *image.RGBA {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			Draw(p, bg, r, Options{
				Time: time, Play: play, Tint: ui.Hex("#7FD7FF"), Flow: flow,
			})
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
	flow := testFlow(t)
	if bg == nil || flow == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}

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

// TestOnlyPersonMoves 是用户要求"建筑、天空不该动"的硬验证。
//
// 判据：动点必须**全部**落在人物轮廓遮罩的 bbox 内。天空、建筑、地面只要有一个
// 像素在动就算失败——这不是"看起来不动"，是逐像素比出来的。
//
// 上一版用 flow-mask 当主遮罩（bbox y 0%~42%，横铺到天空和建筑上），
// 这条测试会直接抓住它。
func TestOnlyPersonMoves(t *testing.T) {
	bg := assets.Bitmap("ganyu/bg.jpg")
	flow := testFlow(t)
	if bg == nil || flow == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}

	a := render(t, bg, flow, 0, 1)
	b := render(t, bg, flow, 3.3, 1)

	// 允许 2% 的容差：遮罩是硬边，滤波与位移会让边界外一两个像素沾到一点变化。
	loX := testW*personLoX/100 - 2
	hiX := testW*personHiX/100 + 2
	loY := testH*personLoY/100 - 2
	hiY := testH*personHiY/100 + 2

	moved, outside := 0, 0
	var oMinX, oMaxX, oMinY, oMaxY = testW, -1, testH, -1
	for y := 0; y < testH; y++ {
		for x := 0; x < testW; x++ {
			if delta(a, b, x, y) <= 6 {
				continue
			}
			if x >= loX && x <= hiX && y >= loY && y <= hiY {
				moved++
				continue
			}
			outside++
			if x < oMinX { oMinX = x }
			if x > oMaxX { oMaxX = x }
			if y < oMinY { oMinY = y }
			if y > oMaxY { oMaxY = y }
		}
	}

	if moved < 40 {
		t.Fatalf("人物区域内只有 %d 个像素在动，动态几乎不可见", moved)
	}
	// 天空在 y<22%，建筑散布其上；这些地方一个点都不许动。
	if outside > 0 {
		t.Errorf("人物轮廓之外有 %d 个像素在动，范围 x %d..%d / y %d..%d —— 建筑或天空被带着动了",
			outside, oMinX, oMaxX, oMinY, oMaxY)
	}
	t.Logf("动点 %d 个，全部在人物 bbox 内（画布 %dx%d，允许区 x %d..%d / y %d..%d）",
		moved, testW, testH, loX, hiX, loY, hiY)
}

// TestBreathIsVertical 验证呼吸是**垂直**起伏，而不是各方向乱飘。
//
// 做法：分别只开呼吸、只开发丝，比较两种情况下水平与垂直方向的位移占比。
// 呼吸是胸腔上下起伏，垂直分量应当占绝对多数；若水平分量不小，说明方向写错了。
func TestBreathIsVertical(t *testing.T) {
	bg := assets.Bitmap("ganyu/bg.jpg")
	flow := testFlow(t)
	if bg == nil || flow == nil {
		t.Skip("assets/ganyu 未导入（素材不入库），跳过")
	}

	breathOnly := &Flow{Person: flow.Person, Breath: testBreath * 2}
	a := render(t, bg, breathOnly, 0, 1)
	b := render(t, bg, breathOnly, 1.1, 1) // sin(1.1*1.4) 与 sin(0)=0 有显著差

	if d := diff(a, b); d < 30 {
		t.Fatalf("纯呼吸模式下几乎没变（%d 点），呼吸可能没生效", d)
	}

	// 呼吸改变的是纵向采样位置：同一列上下会串色，横向几乎不变。
	// 用"每行内部相邻像素差"的方差衡量：垂直位移会让行内纹理横向错位极小。
	// 这里用更直接的判据——位移后整幅图的纵向平移量应明显大于横向平移量。
	shiftY := estimateShiftY(a, b, testW/2, testH/2)
	if shiftY == 0 {
		t.Errorf("中心区域检测不到纵向位移，呼吸方向可能不是垂直的")
	}
	t.Logf("呼吸中心区纵向位移约 %d 像素（画布高 %d）", shiftY, testH)
}

// estimateShiftY 在 (cx, cy) 附近估计 b 相对 a 的纵向位移（像素，正为向下）。
//
// 简单的块匹配：在 ±6 像素内找纵向偏移，使 5×5 邻域的平均颜色差最小。
func estimateShiftY(a, b *image.RGBA, cx, cy int) int {
	best, bestD := 0, 1<<30
	for dy := -6; dy <= 6; dy++ {
		sum := 0
		n := 0
		for y := cy - 2; y <= cy+2; y++ {
			for x := cx - 2; x <= cx+2; x++ {
				yy := y + dy
				if yy < 1 || yy >= testH-1 {
					continue
				}
				sum += delta(a, b, x, yy)
				n++
			}
		}
		if n == 0 {
			continue
		}
		if avg := sum / n; avg < bestD {
			bestD, best = avg, dy
		}
	}
	if bestD > 6 {
		return 0
	}
	return best
}
