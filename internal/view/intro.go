package view

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 开场画面覆盖在主界面之上：主界面照常构建，这一层用绝对定位铺满，
// 随着 visual.Opacity() 淡出，主界面自然显现——不需要两套视图切换。
//
// 绘制顺序与原项目一致：底色渐变 → 左右暗边 → 孔明灯 → 雪 → 元素标识 → 品牌字。
// 灯笼原本是「6×14×5 网格 + 旋转投影」的伪 3D（每盏约 840 个三角形），
// 这里改成 2D 表示：1080p 下这些灯只有 3~32px，3D 细节本来就看不出来。

// introDockX/introDockY 是标识停靠点：侧栏品牌图标的中心
// （侧栏 padding 左 20 / 上 64，图标 40×40）——开场结束后品牌图标接着显示同一个 glyph。
// ⚠️ 改品牌区的位置/尺寸时必须同步这里，否则标识会落在品牌位上方的空处
// （曾因这组常量停留在旧值 (40,36)，标识落点比品牌位高 48px）。
const (
	introDockX = 40
	introDockY = 84
)

const tau = 2 * math.Pi

func sinf(v float64) float32 { return float32(math.Sin(v)) }

// introLayer 是开场层每帧的入口：推进时间、绘制、请求下一帧。
func (a *App) introLayer(p *ui.Painter, r ui.Rect) {
	now := p.Now()
	if a.intro == nil {
		a.intro = visual.New(r.W, r.H)
		a.introAt = now
	} else {
		// 开场播放期间切全屏/最大化：把当前绘制区域同步给 Intro，
		// 否则 Pose 仍按旧窗口尺寸算中心，标识会偏在一侧。
		a.intro.Resize(r.W, r.H)
	}
	dt := float32(now.Sub(a.introAt).Seconds())
	a.introAt = now
	a.intro.Update(dt)

	a.drawIntro(p, r)

	if a.intro.Finished() {
		// 结束：让视图重建一次，把这一层摘掉。
		a.introDone = true
		if a.notify != nil {
			a.notify()
		}
		return
	}
	// 开场要流畅，用满帧重绘（只重绘这一层，不重建元素树）。
	p.AnimationFrameRate(60)
}

// drawIntro 绘制开场画面。
func (a *App) drawIntro(p *ui.Painter, r ui.Rect) {
	in := a.intro
	fade := in.Opacity()
	short := min32(r.W, r.H)

	// 标识不吃图层的淡出：背景与粒子在 5.55s 前淡出、主界面透出，
	// 但标识要一路保持可见——持续缩小飞向品牌位，落定后由静态标识接手
	//（如果它跟着 fade 走，会在半空凭空消失，品牌位再凭空亮起，正是
	//「不是想要的效果」的原因）。落定与消失发生在同一帧（introDone）。
	markFade := float32(1)

	// 图层部分（底色/粒子/品牌字）随 fade 淡出；fade 归零后只剩标识。
	if fade > 0.002 {
		a.drawIntroLayer(p, r, in, fade, short)
	}

	pose := in.Pose(introDockX, introDockY)
	// 专属场景在标识之前、4.10s 后先退场；标识继续飞向品牌位。
	if sceneA := 1 - visual.Smooth(in.T, 4.10, 4.65); sceneA > 0.002 {
		a.drawOpeningScene(p, r, pose, sceneA)
	}
	a.drawMark(p, pose, markFade)
	if fade > 0.002 {
		drawIntroBrand(p, r, in, fade)
	}
}

// drawLantern 画一盏孔明灯：柔光晕 → 灯身 → 顶口 → 灯芯。
func drawLantern(p *ui.Painter, cx, cy, size, alpha float32) {
	// 外光晕：柔和的暖色晕开，灯像悬在夜里。
	glow(p, cx, cy+size*0.06, size*3.0, ui.Hex("#F5A64B"), alpha*0.55)

	w, h := size*0.74, size*1.0
	// 灯身：暖橙色的圆角矩形，中间叠一层更亮的窄条，形成一点体积感。
	p.Fill(ui.Rect{X: cx - w/2, Y: cy - h/2, W: w, H: h}, ui.RGBA(240, 158, 72, alpha*0.46), w*0.42)
	p.Fill(ui.Rect{X: cx - w*0.30, Y: cy - h*0.50, W: w * 0.60, H: h * 1.00},
		ui.RGBA(252, 196, 116, alpha*0.34), w*0.30)
	// 顶口收窄，读起来才是「天灯」而不是胶囊。
	p.Fill(ui.Rect{X: cx - w*0.20, Y: cy - h*0.60, W: w * 0.40, H: h * 0.14},
		ui.RGBA(255, 220, 160, alpha*0.36), w*0.07)
	// 灯芯：底部亮核 + 一小团光。
	glow(p, cx, cy+h*0.20, w*0.62, ui.Hex("#FFE8B8"), alpha*0.85)
	circle(p, cx, cy+h*0.20, w*0.15, ui.RGBA(255, 246, 220, alpha*0.80))
}

// drawMark 按当前角色的 Motif 分发标识画法。
//
// 这是"多角色"这一层的关键：绘制层不知道角色名，只认 Motif。
// 加一个新元素 = 加一个 draw*Mark，主题包里指过去即可。
func (a *App) drawMark(p *ui.Painter, pose visual.MarkPose, fade float32) {
	switch a.char.Motif {
	case visual.MotifFlame:
		drawFlameMark(p, a.intro, pose, fade, a.tok.Accent)
	case visual.MotifBolt:
		drawBoltMark(p, a.intro, pose, fade, a.tok.Accent)
	case visual.MotifGeo:
		drawGeoMark(p, a.intro, pose, fade, a.tok.Accent)
	case visual.MotifLeaf:
		drawLeafMark(p, a.intro, pose, fade, a.tok.Accent)
	case visual.MotifWave:
		drawWaveMark(p, a.intro, pose, fade, a.tok.Accent)
	case visual.MotifWind:
		drawWindMark(p, a.intro, pose, fade, a.tok.Accent)
	default:
		drawCrystalMark(p, a.intro, pose, fade, a.tok.Accent)
	}
}

// drawCrystalMark 画冰元素标识：六角冰晶 + 分叉 + 中心核 + 柔光。
//
// 原项目这里是 5122 个三角形加自定义顶点着色（冰晶材质）。这里用矢量近似：
// 6.2 秒里它只做一次「展开 + 转正 + 停靠」，观感主体是轮廓与光——
// 规整的分叉结构比堆更多线条更能读出冰晶。
func drawCrystalMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)
	impact := visual.Smooth(in.T, 3.50, 3.70) * (1 - visual.Smooth(in.T, 3.74, 4.16))

	// 柔光：成形瞬间（3.5~4.16s）额外亮一下。
	glow(p, x, y, rad*2.6, accent, al*(0.42+0.55*impact))

	ice := accent.Mix(ui.Hex("#FFFFFF"), 0.72).Alpha(al * 0.95)
	iceDim := accent.Mix(ui.Hex("#FFFFFF"), 0.52).Alpha(al * 0.62)
	armW := max32(1.8, rad*0.055)

	// 整体随成形过程旋转 5τ/6（与原项目同一个转角）。
	rot := float64(5*math.Pi/3) * float64(turn)

	for i := 0; i < 6; i++ {
		ang := rot + float64(i)*math.Pi/3
		ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))

		// 每条臂错开生长（0.35s 起、每条晚 0.16s），开场才有「抽枝」的过程，
		// 而不是一上来就是一整朵雪花。原项目用 IceArm 的 growth 参数做同样的事。
		grow := visual.Smooth(in.T, 0.35+float32(i)*0.16, 0.95+float32(i)*0.16)
		rl := rad * grow
		if rl < 1 {
			continue
		}

		// 主干：比侧枝明显粗，六臂的骨架才立得住。
		seg(p, x, y, x+rl*ca, y+rl*sa, armW, ice)

		// 三对侧枝：与主干成 60°，越靠外越短——冰晶的经典分叉。
		// 起点从 0.40 开始（太靠内会和中心核挤在一起，读起来像碎片）。
		for j := 1; j <= 3; j++ {
			f := 0.20 + 0.21*float64(j) // 0.41 / 0.62 / 0.83
			bx := x + rl*float32(f)*ca
			by := y + rl*float32(f)*sa
			bl := rl * float32(0.26-0.045*float64(j))
			for _, sgn := range []float64{-1, 1} {
				ba := ang + sgn*(math.Pi/3)
				seg(p, bx, by,
					bx+bl*float32(math.Cos(ba)), by+bl*float32(math.Sin(ba)),
					armW*0.72, iceDim)
			}
		}

		// 臂尖：一个小菱形收口（沿半径方向拉长，像水晶的尖端）。
		tx, ty := x+rl*ca, y+rl*sa
		ux, uy := ca, sa  // 径向
		vx, vy := -sa, ca // 切向
		tr := rl * 0.075
		ax, ay := tx+ux*tr, ty+uy*tr
		bx2, by2 := tx+vx*tr*0.5, ty+vy*tr*0.5
		cx2, cy2 := tx-ux*tr*0.30, ty-uy*tr*0.30
		dx2, dy2 := tx-vx*tr*0.5, ty-vy*tr*0.5
		for _, edge := range [][4]float32{
			{ax, ay, bx2, by2}, {bx2, by2, cx2, cy2},
			{cx2, cy2, dx2, dy2}, {dx2, dy2, ax, ay},
		} {
			seg(p, edge[0], edge[1], edge[2], edge[3], armW*0.7, ice)
		}
	}

	// 中心核：小六边形压住六臂汇聚点。
	hexR := rad * 0.16
	for i := 0; i < 6; i++ {
		a0 := rot + float64(i)*math.Pi/3
		a1 := rot + float64(i+1)*math.Pi/3
		seg(p,
			x+hexR*float32(math.Cos(a0)), y+hexR*float32(math.Sin(a0)),
			x+hexR*float32(math.Cos(a1)), y+hexR*float32(math.Sin(a1)),
			armW*0.9, ice,
		)
	}
	// 中心的一点亮，别太强——光晕已经在小范围里叠了两层（外光晕 + 这里）。
	glow(p, x, y, hexR*2.0, accent.Mix(ui.Hex("#FFFFFF"), 0.85), al*0.45)
}

// drawFlameMark 画火元素标识：放射火舌 + 末端火星，成形时向外炸开。
//
// 宵宫是烟花师，所以用「炸开」而不是「抽枝」：3.15~3.95s 火花从中心扩散，
// 之后每根火舌按相位轻微长短跳动，像还在燃烧。
func drawFlameMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	boom := visual.Smooth(in.T, 3.15, 3.95)
	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)

	// 光晕：炸开时更大更亮。
	glow(p, x, y, rad*(1.8+0.9*boom), accent, al*(0.40+0.50*boom))

	hot := accent.Mix(ui.Hex("#FFE9A8"), 0.55) // 靠近中心的亮部
	armW := max32(1.4, rad*0.045)

	for i := 0; i < 12; i++ {
		ang := float64(turn)*math.Pi*0.8 + float64(i)*tau/12
		ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))

		// 长短交错（三根一组），再叠一点燃烧的跳动。
		l := rad * (0.42 + 0.58*boom) * float32(0.72+0.14*float64(i%3))
		l *= 1 + 0.10*sinf(float64(in.T)*3.4+float64(i)*1.7)

		// 火舌不从正中心起笔，留出亮核。
		inner := rad * 0.22
		seg(p, x+ca*inner, y+sa*inner, x+ca*l, y+sa*l, armW, accent)

		// 末端火星。
		tipR := rad * (0.030 + 0.030*boom)
		glow(p, x+ca*l, y+sa*l, tipR*3.2, hot, al*0.9*max32(0.2, boom))
		circle(p, x+ca*l, y+sa*l, tipR, hot.Alpha(al*0.85))
	}

	// 内圈短焰，反向缓慢旋转，让中心看起来在燃。
	for i := 0; i < 6; i++ {
		ang := -float64(turn)*math.Pi*1.1 + float64(i)*tau/6
		l := rad * (0.16 + 0.14*boom)
		ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
		seg(p, x, y, x+ca*l, y+sa*l, armW*0.9, hot)
	}
	glow(p, x, y, rad*0.42, ui.Hex("#FFF3D0"), al*0.85)
}

// drawBoltMark 画雷元素标识：六道折线闪电 + 末端亮点。
func drawBoltMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	charge := visual.Smooth(in.T, 3.10, 3.80)
	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)
	glow(p, x, y, rad*(1.9+0.7*charge), accent, al*(0.38+0.45*charge))

	hot := accent.Mix(ui.Hex("#FFFFFF"), 0.75)
	armW := max32(1.4, rad*0.040)

	for i := 0; i < 6; i++ {
		ang := float64(turn)*math.Pi*1.2 + float64(i)*tau/6
		ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
		tx, ty := -sa, ca // 切向，用来做折线偏移
		length := rad * (0.55 + 0.45*charge)

		// 折线：4 段，交替向两侧偏折（用固定公式而非随机，保证确定性）。
		px, py := x, y
		const n = 4
		for k := 1; k <= n; k++ {
			f := float32(k) / n
			j := sinf(float64(i)*3.7+float64(k)*2.1) * rad * 0.10 * (1 - f)
			nx := x + ca*length*f + tx*j
			ny := y + sa*length*f + ty*j
			seg(p, px, py, nx, ny, armW, hot)
			px, py = nx, ny
		}
		glow(p, px, py, rad*0.10, hot, al*0.8)
	}
	glow(p, x, y, rad*0.30, ui.Hex("#F0E4FF"), al*0.9)
}

// drawGeoMark 画岩元素标识：六边形岩印 + 内嵌菱形。
//
// 岩的符号感来自"规整、厚重"：外轮廓是正六边形，中间一枚缓慢转正的菱形，
// 不像冰晶那样有大量分叉，靠线宽和实心度撑重量。
func drawGeoMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)
	solid := visual.Smooth(in.T, 3.20, 3.90)

	glow(p, x, y, rad*2.4, accent, al*(0.38+0.45*solid))

	gold := accent.Mix(ui.Hex("#FFFFFF"), 0.55).Alpha(al * 0.92)
	armW := max32(1.6, rad*0.05)

	// 正六边形：顶点朝上，六个顶点连线。
	hexR := rad * (0.72 + 0.28*solid)
	for i := 0; i < 6; i++ {
		a0 := -math.Pi/2 + float64(i)*math.Pi/3
		a1 := -math.Pi/2 + float64(i+1)*math.Pi/3
		seg(p,
			x+hexR*float32(math.Cos(a0)), y+hexR*float32(math.Sin(a0)),
			x+hexR*float32(math.Cos(a1)), y+hexR*float32(math.Sin(a1)),
			armW, gold)
	}

	// 内嵌菱形：随成形进度转正 45°。
	dm := rad * (0.26 + 0.26*solid)
	r45 := float64(turn) * math.Pi / 4
	for i := 0; i < 4; i++ {
		a0 := r45 + float64(i)*math.Pi/2
		a1 := r45 + float64(i+1)*math.Pi/2
		seg(p,
			x+dm*float32(math.Cos(a0)), y+dm*float32(math.Sin(a0)),
			x+dm*float32(math.Cos(a1)), y+dm*float32(math.Sin(a1)),
			armW*0.85, gold)
	}

	// 中心亮核。
	glow(p, x, y, rad*0.30, ui.Hex("#FFF3C4"), al*0.85)
	circle(p, x, y, rad*0.07, ui.Hex("#FFFCF0").Alpha(al*0.9))
}

// drawLeafMark 画草元素标识：三片叶子从同一中心抽枝展开 + 短茎 + 芽心。
func drawLeafMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	grow := visual.Smooth(in.T, 0.25, 1.35)
	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)
	bloom := visual.Smooth(in.T, 3.20, 3.95)

	glow(p, x, y, rad*(1.7+0.5*bloom), accent, al*(0.34+0.42*bloom))
	green := accent.Mix(ui.Hex("#F4FFD8"), 0.55)
	w := max32(1.4, rad*0.045)

	for i, ang := range []float64{-math.Pi / 2, -math.Pi/2 - .88, -math.Pi/2 + .88} {
		g := visual.Smooth(in.T, 0.28+float32(i)*0.13, 1.10+float32(i)*0.13)
		length := rad * 0.96 * g
		width := rad * 0.26 * g
		petal(p, x, y, float32(ang)+float32(turn)*0.5, length, width, w, green.Alpha(al*0.92))
	}

	// 茎与芽心：让三片叶子看起来是长出来的。
	stem := rad * (0.30 + 0.42*grow)
	seg(p, x, y, x, y+stem, w*0.8, green.Alpha(al*0.70))
	glow(p, x, y, rad*0.30, ui.Hex("#F4FFD8"), al*0.80)
	circle(p, x, y, max32(.9, rad*0.055), ui.Hex("#F8FFE8").Alpha(al*0.92))

	// 几粒飘散的种子，给成形一点空气感。
	for i := 0; i < 7; i++ {
		a := float64(turn)*2.0 + float64(i)*tau/7
		rr := rad * (0.45 + 0.45*grow)
		px := x + rr*float32(math.Cos(a))
		py := y + rr*float32(math.Sin(a)) - rad*0.20*bloom
		circle(p, px, py, max32(.6, rad*0.018), green.Alpha(al*0.35*(1-bloom*0.4)))
	}
}

// drawWaveMark 画水元素标识：水滴升起 + 三圈扩散的水纹。
func drawWaveMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	grow := visual.Smooth(in.T, 0.22, 1.25)
	settle := visual.Smooth(in.T, 3.10, 4.00)

	glow(p, x, y, rad*(1.7+0.45*settle), accent, al*(0.34+0.40*settle))
	blue := accent.Mix(ui.Hex("#EAF8FF"), 0.58)
	w := max32(1.4, rad*0.045)

	// 水滴从下方浮起，最后停住。
	dy := (1 - grow) * rad * 0.30
	droplet(p, x, y+dy, rad*0.58*grow, blue.Alpha(al*0.92))

	// 水纹：三圈从水滴下方扩散，越远越淡。
	centerY := y + rad*0.58
	for i := 0; i < 3; i++ {
		k := visual.Smooth(in.T, 0.95+float32(i)*0.42, 3.10+float32(i)*0.42)
		rr := rad * (0.42 + 0.95*k)
		col := blue.Alpha(al * 0.55 * (1 - k) * (0.6 + 0.4*settle))
		arcLine(p, x, centerY, rr, math.Pi*0.12, math.Pi*0.88, w*0.8, col)
	}

	// 成形后仍保留两道水波，避免最终只剩一颗孤立的水滴。
	for i := 0; i < 2; i++ {
		wy := y + rad*(0.54+float32(i)*0.14)
		curveLine(p, x-rad*0.55, wy, x-rad*0.24, wy-rad*0.09,
			x+rad*0.24, wy+rad*0.09, x+rad*0.55, wy, blue.Alpha(al*0.55), w*0.7)
	}

	// 几颗悬浮的小水珠。
	for i := 0; i < 6; i++ {
		a := float64(i)*tau/6 + float64(in.T)*0.35
		rr := rad * (0.60 + 0.28*grow)
		px := x + rr*float32(math.Cos(a))
		py := y - rad*0.08 + rr*0.42*float32(math.Sin(a))
		circle(p, px, py, max32(.6, rad*0.018), blue.Alpha(al*0.45))
	}
}

// drawWindMark 画风元素标识：三道回旋气流 + 甩出的尾迹 + 中心风眼。
func drawWindMark(p *ui.Painter, in *visual.Intro, pose visual.MarkPose, fade float32, accent ui.Color) {
	al := pose.Opacity * fade
	if al <= 0.004 || pose.Radius <= 0.5 {
		return
	}
	x, y, rad := pose.X, pose.Y, pose.Radius

	grow := visual.Smooth(in.T, 0.20, 1.30)
	turn := visual.Smooth(in.T, visual.TurnStart, visual.TurnEnd)
	gust := visual.Smooth(in.T, 3.10, 4.00)

	glow(p, x, y, rad*(1.8+0.4*gust), accent, al*(0.34+0.38*gust))
	teal := accent.Mix(ui.Hex("#F2FFF8"), 0.55)
	w := max32(1.4, rad*0.042)

	for i := 0; i < 3; i++ {
		r := rad * (0.90 - 0.19*float32(i)) * grow
		a0 := float64(turn)*math.Pi*1.2 + float64(i)*tau/3 + float64(gust)*0.5
		arcLine(p, x, y, r, a0, a0+math.Pi*1.5, w*(1-float32(i)*0.18),
			teal.Alpha(al*(0.92-0.20*float32(i))))
		// 外圈起点甩出一条短尾。
		tx := x + r*float32(math.Cos(a0))
		ty := y + r*float32(math.Sin(a0))
		seg(p, tx, ty, tx-float32(math.Cos(a0))*rad*0.26, ty-float32(math.Sin(a0))*rad*0.26,
			w*0.8, teal.Alpha(al*0.55))
	}

	// 风眼：一个亮核 + 绕行的小气旋点。
	glow(p, x, y, rad*0.30, ui.Hex("#EFFFF8"), al*0.85)
	circle(p, x, y, max32(.9, rad*0.045), ui.Hex("#F4FFFA").Alpha(al*0.9))
	for i := 0; i < 8; i++ {
		a := float64(turn)*math.Pi*2.4 + float64(i)*tau/8
		rr := rad * (0.35 + 0.30*grow)
		px := x + rr*float32(math.Cos(a))
		py := y + rr*float32(math.Sin(a))
		circle(p, px, py, max32(.5, rad*0.014), teal.Alpha(al*0.40))
	}
}

// drawIntroBrand 右下角的 "HANABI" 字标：逐字绘制以做出字距。
//
// fade 已经是整层的淡出（Intro.Opacity），这里不能再乘一次 in.Opacity()：
// 那会让品牌字按不透明度的平方衰减，比其它元素早得多地消失（踩过）。
func drawIntroBrand(p *ui.Painter, r ui.Rect, in *visual.Intro, fade float32) {
	reveal := visual.Smooth(in.T, 0.45, 1.25) * (1 - visual.Smooth(in.T, 4.90, 5.70)) * fade
	if reveal <= 0.01 {
		return
	}
	const title = "HANABI"
	const size, tracking = float32(20), float32(4)

	short := min32(r.W, r.H)
	margin := min32(48, max32(24, short*0.055))

	// 先量总宽（逐字 + 字距），才能右对齐。
	total := float32(0)
	for _, ch := range title {
		w, _ := p.MeasureText(0, ui.Span{Text: string(ch), Size: size})
		total += w + tracking
	}
	total -= tracking

	x := r.X + r.W - margin - total
	y := r.Y + r.H - margin - size

	for _, ch := range title {
		s := string(ch)
		// 一点点投影，让字在亮背景上也立得住。
		p.Text(x, y+1, s, size, ui.RGBA(0, 0, 0, reveal*0.16))
		p.Text(x, y, s, size, ui.RGBA(208, 229, 242, reveal*0.90))
		w, _ := p.MeasureText(0, ui.Span{Text: s, Size: size})
		x += w + tracking
	}
}

// circle 用圆角矩形画圆（mygo 没有椭圆图元，正方形 + 半径=半宽即圆）。
func circle(p *ui.Painter, cx, cy, rad float32, c ui.Color) {
	if rad <= 0 {
		return
	}
	p.Fill(ui.Rect{X: cx - rad, Y: cy - rad, W: rad * 2, H: rad * 2}, c, rad)
}

// seg 画一条任意方向的线段。
//
// 不能用 Painter.Line：它的文档写明只画「水平或垂直线」，内部是用两点包围盒
// 再用 width 覆盖某一维——斜线会被压成一个矩形（冰晶六条斜臂因此全变成竖线，
// 雪花也变成小方块，踩过）。这里用 Transform 把坐标系转到线段方向，
// 再填一个长 l、高 width 的胶囊矩形。
func seg(p *ui.Painter, x0, y0, x1, y1, width float32, c ui.Color) {
	if width <= 0 {
		return
	}
	dx, dy := x1-x0, y1-y0
	l := float32(math.Hypot(float64(dx), float64(dy)))
	if l < 0.05 {
		return
	}
	deg := float32(math.Atan2(float64(dy), float64(dx)) * 180 / math.Pi)
	// 短段（l 小于笔宽）不能用 width/2 当圆角：半径超过短边一半时，
	// 渲染器只剩两端的圆帽，整条线看起来是一串点。半径夹到 l/2，
	// 短段就退化成圆/胶囊，相邻段仍然接得上。
	rad := width / 2
	if half := l / 2; half < rad {
		rad = half
	}
	p.Transform(ui.Rotation(deg), x0, y0, func() {
		p.Fill(ui.Rect{X: x0, Y: y0 - width/2, W: l, H: width}, c, rad)
	})
}

// glow 画柔光晕。
//
// mygo 没有径向渐变。用多层同心圆叠不行——每层都是硬边，会出现一圈圈台阶
// （试过，比单层圆更难看）。正确做法是**预生成一张径向渐变贴图**，
// 再用 ImageRegion 拉伸 + Tint 上色：一次生成，之后 CPU/GPU 都只是贴图采样。
func glow(p *ui.Painter, cx, cy, rad float32, c ui.Color, alpha float32) {
	if rad <= 0 || alpha <= 0.002 {
		return
	}
	b := glowTexture()
	if b == nil {
		return
	}
	w, h := b.Size()
	p.ImageRegion(b,
		ui.Rect{W: float32(w), H: float32(h)},
		ui.Rect{X: cx - rad, Y: cy - rad, W: rad * 2, H: rad * 2},
		ui.ImageOptions{Tint: c.Alpha(alpha)},
	)
}

var (
	glowOnce sync.Once
	glowBmp  *ui.Bitmap
)

// glowTexture 生成白色径向渐变：中心不透明、边缘完全透明。
// 衰减用 (1-r)²：中心集中、边缘拖尾长，接近真实光晕。
//
// 注意 Go 的 image.RGBA 是**预乘 alpha**：RGB 必须等于 颜色×alpha，
// 否则写成 (255,255,255,0) 这种 RGB>A 的非法值，渲染器会把边缘也当成不透明
// （第一版就踩了这个坑，光晕变成实心圆盘）。
func glowTexture() *ui.Bitmap {
	glowOnce.Do(func() {
		const n = 256 // 放大到几百像素也不会有明显锯齿
		img := image.NewRGBA(image.Rect(0, 0, n, n))
		c := float64(n-1) / 2
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				dx := (float64(x) - c) / c
				dy := (float64(y) - c) / c
				a := 1 - math.Hypot(dx, dy)
				if a < 0 {
					a = 0
				}
				a *= a
				v := uint8(a*255 + 0.5)
				img.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: v})
			}
		}
		glowBmp = ui.NewBitmap(img)
	})
	return glowBmp
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
