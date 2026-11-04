package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 卡片装饰线稿。
//
// mygo 的 Painter 只有「矩形填充 / 描边矩形（可带圆角）/ 水平垂直线」三类图元，
// 没有弧线和贝塞尔，所以这里用折线采样近似：arcLine / quadLine 都拆成短线段，
// 配 1.1~1.4 的线宽，小尺寸下读起来就是连续的墨线。
//
// 每幅线稿都画在卡片右半区（r），部分图形故意越出右边被卡片圆角裁掉，
// 这是参考图里那种「纹样从边上长进来」的观感。

// ring 画一个线框圆（用描边矩形，半径取半宽即圆）。
func ring(p *ui.Painter, cx, cy, rad, w float32, c ui.Color) {
	if rad <= 0 {
		return
	}
	p.Stroke(ui.Rect{X: cx - rad, Y: cy - rad, W: rad * 2, H: rad * 2}, c, rad, w)
}

// poly 画连续折线：pts 是 x0,y0,x1,y1,… 的扁平序列。
func poly(p *ui.Painter, w float32, c ui.Color, pts ...float32) {
	for i := 0; i+3 < len(pts); i += 2 {
		seg(p, pts[i], pts[i+1], pts[i+2], pts[i+3], w, c)
	}
}

// arcLine 用折线近似圆弧（弧度制），再用 Path 描边：连续曲线不会像
// 逐段胶囊那样在放大后露出接缝。
func arcLine(p *ui.Painter, cx, cy, rad float32, from, to float64, w float32, c ui.Color) {
	const steps = 64
	var path ui.Path
	for i := 0; i <= steps; i++ {
		a := from + (to-from)*float64(i)/steps
		x := cx + rad*float32(math.Cos(a))
		y := cy + rad*float32(math.Sin(a))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	p.StrokePath(&path, w, c)
}

// quadLine 用折线近似二次贝塞尔曲线，用 Path 一次描边。
func quadLine(p *ui.Painter, x0, y0, cx, cy, x1, y1, w float32, c ui.Color) {
	const steps = 64
	var path ui.Path
	for i := 0; i <= steps; i++ {
		t := float32(i) / steps
		mt := 1 - t
		x := mt*mt*x0 + 2*mt*t*cx + t*t*x1
		y := mt*mt*y0 + 2*mt*t*cy + t*t*y1
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	p.StrokePath(&path, w, c)
}

// quadAt 求二次贝塞尔上 t 处的点与切线角。
func quadAt(x0, y0, cx, cy, x1, y1, t float32) (x, y, ang float64) {
	mt := 1 - t
	x = float64(mt*mt*x0 + 2*mt*t*cx + t*t*x1)
	y = float64(mt*mt*y0 + 2*mt*t*cy + t*t*y1)
	dx := float64(2*mt*(cx-x0) + 2*t*(x1-cx))
	dy := float64(2*mt*(cy-y0) + 2*t*(y1-cy))
	return x, y, math.Atan2(dy, dx)
}

// petal 画一片花瓣：从 (x,y) 朝 ang 方向伸出，两侧各一段对称的曲线。
func petal(p *ui.Painter, x, y, ang, length, width, w float32, c ui.Color) {
	tipX := x + length*float32(math.Cos(float64(ang)))
	tipY := y + length*float32(math.Sin(float64(ang)))
	nx := -float32(math.Sin(float64(ang))) * width
	ny := float32(math.Cos(float64(ang))) * width
	midX := x + (tipX-x)*0.55
	midY := y + (tipY-y)*0.55
	quadLine(p, x, y, midX+nx, midY+ny, tipX, tipY, w, c)
	quadLine(p, tipX, tipY, midX-nx, midY-ny, x, y, w, c)
}

// blossom 画一朵 n 瓣小花：五瓣是樱/霜花的通用形。
func blossom(p *ui.Painter, cx, cy, rad float64, n int, w float32, c ui.Color) {
	for k := 0; k < n; k++ {
		ang := -math.Pi/2 + float64(k)*2*math.Pi/float64(n)
		petal(p, float32(cx), float32(cy), float32(ang), float32(rad), float32(rad*0.36), w, c)
	}
}

// cardArt 按 Art 分发，画一面卡片线稿。
func cardArt(p *ui.Painter, art visual.Art, r ui.Rect, accent ui.Color) {
	ink := ui.RGBA(255, 255, 255, 0.30)
	faint := ui.RGBA(255, 255, 255, 0.15)
	hot := accent.Mix(ui.Hex("#FFFFFF"), 0.30).Alpha(0.62)
	const w = float32(1.15)

	switch art {
	case visual.ArtNone:
		// 由 Ornament 的专属纹样接管（清心纹 / 风纹 / 中国结）。
		return
	case visual.ArtSakura:
		artSakura(p, r, w, ink, faint, hot)
	case visual.ArtFirework:
		artFirework(p, r, w, ink, faint, hot)
	case visual.ArtLeaf:
		artLeaf(p, r, w, ink, faint, hot)
	case visual.ArtSail:
		artSail(p, r, w, ink, faint, hot)
	case visual.ArtCrystal:
		artCrystal(p, r, w, ink, faint, hot)
	case visual.ArtRing:
		artRing(p, r, w, ink, faint, hot)
	default:
		artBlossom(p, r, w, ink, faint, hot)
	}
}

// artBlossom 霜花 + 风铃（甘雨）：右侧一朵大霜花，右缘挂一只铃。
func artBlossom(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	cx, cy := r.X+r.W*0.34, r.Y+r.H*0.46

	blossom(p, float64(cx), float64(cy), float64(s*0.24), 5, w, ink)
	ring(p, cx, cy, s*0.05, w, hot)

	// 花茎：从花心垂到下缘。
	quadLine(p, cx, cy+s*0.16, cx-s*0.04, cy+s*0.24, cx+s*0.08, r.Y+r.H+6, w, faint)

	// 风铃：吊线 → 铃盖（弧）→ 铃舌 → 流苏。
	lx := r.X + r.W*0.80
	hangY := r.Y + s*0.30
	seg(p, lx, r.Y-4, lx, hangY, w*0.9, faint)
	arcLine(p, lx, hangY, s*0.085, math.Pi, 2*math.Pi, w, hot) // 铃盖：上半圆
	seg(p, lx-s*0.085, hangY, lx+s*0.085, hangY, w*0.8, faint)
	circle(p, lx, hangY+s*0.115, s*0.018, hot) // 铃舌
	for i := -1; i <= 1; i++ {
		seg(p, lx, hangY+s*0.14, lx+float32(i)*s*0.05, hangY+s*0.26, w*0.7, faint)
	}
}

// artSakura 樱枝 + 御守（八重神子）：一根斜枝带三朵樱花，右缘挂御守。
func artSakura(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	x0, y0 := r.X+r.W*0.06, r.Y+r.H+4
	x1, y1 := r.X+r.W*1.05, r.Y-s*0.08
	cx, cy := r.X+r.W*0.40, r.Y+r.H*0.70

	// 主枝与一根分杈。
	quadLine(p, x0, y0, cx, cy, x1, y1, w*1.15, faint)
	bx, by, _ := quadAt(x0, y0, cx, cy, x1, y1, 0.62)
	quadLine(p, float32(bx), float32(by), float32(bx)+s*0.02, float32(by)-s*0.10,
		float32(bx)+s*0.16, float32(by)-s*0.20, w*0.9, faint)

	// 三朵樱花：枝上两朵 + 分杈一朵，花心一点红。
	for _, t := range []float64{0.30, 0.55, 0.82} {
		px, py, _ := quadAt(x0, y0, cx, cy, x1, y1, float32(t))
		blossom(p, px, py, float64(s*0.11), 5, w*0.9, ink)
		circle(p, float32(px), float32(py), s*0.014, hot)
	}
	// 两枚花苞。
	for _, t := range []float64{0.44, 0.70} {
		px, py, _ := quadAt(x0, y0, cx, cy, x1, y1, float32(t))
		circle(p, float32(px)+s*0.05, float32(py)-s*0.05, s*0.02, ink)
	}

	// 御守：右侧一挂，绳结 + 牌 + 流苏。
	ox := r.X + r.W*0.88
	seg(p, ox, r.Y-4, ox, r.Y+s*0.28, w*0.9, faint)
	p.Stroke(ui.Rect{X: ox - s*0.05, Y: r.Y + s*0.28, W: s * 0.10, H: s * 0.14},
		hot, 2, w)
	seg(p, ox-s*0.02, r.Y+s*0.42, ox, r.Y+s*0.47, w*0.7, faint)
	seg(p, ox+s*0.02, r.Y+s*0.42, ox, r.Y+s*0.47, w*0.7, faint)
	seg(p, ox, r.Y+s*0.42, ox, r.Y+s*0.50, w*0.8, hot)
	for i := -1; i <= 1; i++ {
		seg(p, ox, r.Y+s*0.50, ox+float32(i)*s*0.035, r.Y+s*0.61, w*0.7, faint)
	}
}

// artFirework 烟花 + 灯笼（宵宫）：中心炸开的放射线，右下挂一盏灯。
func artFirework(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	cx, cy := r.X+r.W*0.40, r.Y+r.H*0.40

	for i := 0; i < 14; i++ {
		ang := float64(i) * 2 * math.Pi / 14
		ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
		l := s * (0.20 + 0.10*float32(i%3))
		seg(p, cx+ca*s*0.06, cy+sa*s*0.06, cx+ca*l, cy+sa*l, w, ink)
		circle(p, cx+ca*l, cy+sa*l, w*1.1, hot)
	}
	ring(p, cx, cy, s*0.13, w*0.9, faint)
	ring(p, cx, cy, s*0.05, w*0.9, hot)

	// 远处的第二朵小烟花
	quadLine(p, r.X+r.W*0.86, r.Y+s*0.06, r.X+r.W*0.80, r.Y+s*0.14, r.X+r.W*0.90, r.Y+s*0.20, w*0.8, faint)

	// 灯笼：吊线 + 灯身 + 上下口 + 流苏。
	lx := r.X + r.W*0.80
	seg(p, lx, r.Y-4, lx, r.Y+s*0.56, w*0.9, faint)
	p.Stroke(ui.Rect{X: lx - s*0.10, Y: r.Y + s*0.56, W: s * 0.20, H: s * 0.22},
		hot, s*0.09, w*1.1)
	seg(p, lx-s*0.05, r.Y+s*0.55, lx+s*0.05, r.Y+s*0.55, w*0.8, faint)
	seg(p, lx-s*0.05, r.Y+s*0.79, lx+s*0.05, r.Y+s*0.79, w*0.8, faint)
	seg(p, lx, r.Y+s*0.79, lx, r.Y+s*0.88, w*0.8, hot)
}

// artLeaf 叶枝 + 新芽（纳西妲）：一根上扬的枝，三片带叶脉的叶子。
func artLeaf(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	x0, y0 := r.X+r.W*0.16, r.Y+r.H*0.98
	x1, y1 := r.X+r.W*0.82, r.Y-s*0.08
	cx, cy := r.X+r.W*0.46, r.Y+r.H*0.62

	quadLine(p, x0, y0, cx, cy, x1, y1, w*1.15, faint)

	// 叶片：沿枝取点，朝枝的切线外侧伸出，两段曲线合成尖叶，中间补一条叶脉。
	for i, t := range []float64{0.24, 0.52, 0.78} {
		px, py, ang := quadAt(x0, y0, cx, cy, x1, y1, float32(t))
		side := float64(1 - 2*(i%2)) // 交替朝两侧
		la := ang - side*math.Pi/2.6
		long := float64(s) * (0.20 + 0.04*float64(i%2))
		tipX := px + long*math.Cos(la)
		tipY := py + long*math.Sin(la)
		nx := -long * 0.30 * math.Sin(la) * side
		ny := long * 0.30 * math.Cos(la) * side
		mx := (px + tipX) / 2
		my := (py + tipY) / 2
		quadLine(p, float32(px), float32(py), float32(mx+nx), float32(my+ny), float32(tipX), float32(tipY), w, ink)
		quadLine(p, float32(tipX), float32(tipY), float32(mx-nx), float32(my-ny), float32(px), float32(py), w, ink)
		seg(p, float32(px), float32(py), float32(tipX), float32(tipY), w*0.6, faint)
	}

	// 芽尖：一枚小叶 + 亮点。
	quadLine(p, x1, y1, x1+s*0.06, y1-s*0.05, x1+s*0.14, y1-s*0.02, w*0.9, faint)
	circle(p, x1+s*0.15, y1-s*0.02, s*0.020, hot)
}

// artSail 帆船 + 水波（芙宁娜）：桅杆、主帆、前帆、船身与水线。
func artSail(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	mx := r.X + r.W*0.48
	water := r.Y + s*0.70
	top := r.Y + s*0.12

	// 桅杆与旗
	seg(p, mx, top, mx, water-s*0.06, w*1.2, faint)
	poly(p, w*0.9, hot,
		mx, top,
		mx+s*0.10, top+s*0.03,
		mx, top+s*0.06)

	// 主帆：前缘贴桅杆，后缘鼓出。
	quadLine(p, mx, top+s*0.05, mx+s*0.30, top+s*0.22, mx+s*0.22, water-s*0.16, w, ink)
	seg(p, mx+s*0.22, water-s*0.16, mx, water-s*0.14, w, ink)
	seg(p, mx, top+s*0.05, mx, water-s*0.14, w*0.9, faint)

	// 前帆：桅杆另一侧。
	quadLine(p, mx-s*0.01, top+s*0.12, mx-s*0.30, top+s*0.30, mx-s*0.26, water-s*0.18, w, ink)
	seg(p, mx-s*0.26, water-s*0.18, mx-s*0.01, water-s*0.16, w, ink)

	// 船身：下半圆弧 + 甲板线。
	arcLine(p, mx, water-s*0.14, s*0.30, 0.12, math.Pi-0.12, w*1.3, ink)
	seg(p, mx-s*0.295, water-s*0.135, mx+s*0.295, water-s*0.135, w, faint)

	// 水波
	for i := 0; i < 3; i++ {
		y := water + s*0.10 + float32(i)*s*0.055
		x := r.X + r.W*0.16 + float32(i)*s*0.06
		quadLine(p, x, y, x+s*0.10, y-s*0.035, x+s*0.20, y, w*0.8, faint)
		quadLine(p, x+s*0.20, y, x+s*0.30, y+s*0.035, x+s*0.40, y, w*0.8, faint)
	}
}

// artCrystal 棱晶 + 书页（珐露珊）：一枚宝石棱晶，下方摊开的书。
func artCrystal(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	cx, cy := r.X+r.W*0.50, r.Y+s*0.32

	// 宝石：台面 + 亭部（经典切割轮廓）。
	poly(p, w, ink,
		cx-s*0.19, cy-s*0.06, // 左腰
		cx-s*0.10, cy-s*0.20, // 左上台面
		cx+s*0.10, cy-s*0.20,
		cx+s*0.19, cy-s*0.06,
		cx, cy+s*0.18, // 底尖
		cx-s*0.19, cy-s*0.06)
	// 台面横线与刻面
	seg(p, cx-s*0.19, cy-s*0.06, cx+s*0.19, cy-s*0.06, w*0.85, faint)
	seg(p, cx-s*0.10, cy-s*0.20, cx-s*0.05, cy-s*0.06, w*0.8, faint)
	seg(p, cx+s*0.10, cy-s*0.20, cx+s*0.05, cy-s*0.06, w*0.8, faint)
	seg(p, cx-s*0.05, cy-s*0.06, cx, cy+s*0.18, w*0.8, faint)
	seg(p, cx+s*0.05, cy-s*0.06, cx, cy+s*0.18, w*0.8, faint)
	circle(p, cx-s*0.02, cy-s*0.115, s*0.016, hot)

	// 书：摊开的两页（页缘上拱）+ 书脊。
	by := r.Y + s*0.80
	quadLine(p, cx-s*0.26, by+s*0.05, cx-s*0.15, by-s*0.05, cx-s*0.01, by-s*0.02, w, ink)
	quadLine(p, cx-s*0.26, by+s*0.05, cx-s*0.15, by+s*0.06, cx-s*0.01, by+s*0.09, w, ink)
	quadLine(p, cx+s*0.26, by+s*0.05, cx+s*0.15, by-s*0.05, cx+s*0.01, by-s*0.02, w, ink)
	quadLine(p, cx+s*0.26, by+s*0.05, cx+s*0.15, by+s*0.06, cx+s*0.01, by+s*0.09, w, ink)
	seg(p, cx, by-s*0.02, cx, by+s*0.10, w*0.9, hot)
}

// artRing 岩环 + 六方（钟离）：同心圆环，中央一枚六边形，上下有吊线。
func artRing(p *ui.Painter, r ui.Rect, w float32, ink, faint, hot ui.Color) {
	s := r.H
	cx, cy := r.X+r.W*0.50, r.Y+r.H*0.46

	ring(p, cx, cy, s*0.30, w*1.2, ink)
	ring(p, cx, cy, s*0.23, w*0.9, faint)

	// 六边形
	hexR := s * 0.14
	for i := 0; i < 6; i++ {
		a0 := -math.Pi/2 + float64(i)*math.Pi/3
		a1 := -math.Pi/2 + float64(i+1)*math.Pi/3
		seg(p,
			cx+hexR*float32(math.Cos(a0)), cy+hexR*float32(math.Sin(a0)),
			cx+hexR*float32(math.Cos(a1)), cy+hexR*float32(math.Sin(a1)), w, hot)
	}

	// 吊线与坠子
	seg(p, cx, r.Y-4, cx, cy-s*0.30, w*0.9, faint)
	seg(p, cx, cy+s*0.30, cx, r.Y+r.H+4, w*0.9, faint)
	for i := -1; i <= 1; i++ {
		if i == 0 {
			continue
		}
		seg(p, cx, cy, cx+float32(i)*s*0.08, cy-s*0.06, w*0.8, faint)
	}
	circle(p, cx, cy, s*0.02, hot)
}
