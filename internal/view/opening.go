package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 各角色的开场分镜（参考 characters/*/*_visual.cpp）。
//
// 原版的核心标识是几千个三角形的材质网格；这里是纯 2D 的近似：
// 背景建筑/帷幕/书架、围绕标识的专属场景（霜莲珠、烟花金鱼、神樱雷、
// 梦境莲台、水上舞台、机关风带、岩柱流星），以及元素色氛围粒子。
// 时间轴仍是 6.2 秒：0.2 上升 → 1.45 转向 → 3.65 成形 → 4.65 停靠。

type openingPalette struct {
	top, bottom, glow ui.Color
}

func openingPaletteFor(o visual.Opening) openingPalette {
	switch o {
	case visual.OpeningYoimiya:
		return openingPalette{ui.Hex("#150F19"), ui.Hex("#30201D"), ui.Hex("#F8B365")}
	case visual.OpeningYae:
		return openingPalette{ui.Hex("#130A20"), ui.Hex("#2A1531"), ui.Hex("#CA87D7")}
	case visual.OpeningNahida:
		return openingPalette{ui.Hex("#091A1A"), ui.Hex("#14352A"), ui.Hex("#AFD994")}
	case visual.OpeningFurina:
		return openingPalette{ui.Hex("#07111F"), ui.Hex("#10243D"), ui.Hex("#7EC3EC")}
	case visual.OpeningFaruzan:
		return openingPalette{ui.Hex("#061A1B"), ui.Hex("#103433"), ui.Hex("#7ED9B7")}
	case visual.OpeningZhongli:
		return openingPalette{ui.Hex("#18140F"), ui.Hex("#221A11"), ui.Hex("#E3A33B")}
	default:
		return openingPalette{ui.Hex("#071522"), ui.Hex("#102B3A"), ui.Hex("#68BFEA")}
	}
}

// drawIntroLayer 开场的背景与氛围（不含标识、品牌字另画）。
func (a *App) drawIntroLayer(p *ui.Painter, r ui.Rect, in *visual.Intro, fade, short float32) {
	pal := openingPaletteFor(a.char.Opening)
	p.FillGradient(r, ui.LinearGradient{
		From:  pal.top.Alpha(.995 * fade),
		To:    pal.bottom.Alpha(.985 * fade),
		Angle: 180,
	}, 0)

	sideW := r.W * 0.22
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y, W: sideW, H: r.H}, ui.LinearGradient{
		From: ui.RGBA(4, 11, 18, 0.28*fade), To: ui.Transparent, Angle: 90,
	}, 0)
	p.FillGradient(ui.Rect{X: r.X + r.W - sideW, Y: r.Y, W: sideW, H: r.H}, ui.LinearGradient{
		From: ui.Transparent, To: ui.RGBA(4, 11, 18, 0.28*fade), Angle: 90,
	}, 0)

	glow(p, r.X+r.W*.5, r.Y+r.H*.42, short*.72, pal.glow, fade*.10)
	a.drawOpeningBackdrop(p, r, in, fade, short)
}

// drawOpeningBackdrop 画每个角色的背景建筑/光景/氛围。
func (a *App) drawOpeningBackdrop(p *ui.Painter, r ui.Rect, in *visual.Intro, fade, short float32) {
	t := in.T
	switch a.char.Opening {
	case visual.OpeningGanyu:
		drawGanyuAtmosphere(p, r, in, fade, short)
	case visual.OpeningYoimiya:
		drawWarmHorizon(p, r, fade)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	case visual.OpeningYae:
		drawToriiGate(p, r.X+r.W*.10, r.Y+r.H*.18, r.H*.58, r.W*.12, fade*.30)
		drawToriiGate(p, r.X+r.W*.90, r.Y+r.H*.18, r.H*.58, r.W*.12, fade*.30)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	case visual.OpeningNahida:
		drawSanctuary(p, r, t, fade)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	case visual.OpeningFurina:
		drawCurtains(p, r, t, fade)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	case visual.OpeningFaruzan:
		drawArchive(p, r, t, fade)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	case visual.OpeningZhongli:
		drawStoneHall(p, r, t, fade)
		drawOpeningMotes(p, r, a.char.Opening, t, fade, short)
	}
}

// drawOpeningScene 画标识周围的专属演出，在标识之前绘制。
func (a *App) drawOpeningScene(p *ui.Painter, r ui.Rect, pose visual.MarkPose, sceneA float32) {
	if sceneA <= 0.002 {
		return
	}
	t := a.intro.T
	switch a.char.Opening {
	case visual.OpeningGanyu:
		sceneGanyu(p, pose, t, sceneA)
	case visual.OpeningYoimiya:
		sceneYoimiya(p, pose, t, sceneA)
	case visual.OpeningYae:
		sceneYae(p, pose, t, sceneA)
	case visual.OpeningNahida:
		sceneNahida(p, pose, t, sceneA)
	case visual.OpeningFurina:
		sceneFurina(p, pose, t, sceneA)
	case visual.OpeningFaruzan:
		sceneFaruzan(p, pose, t, sceneA)
	case visual.OpeningZhongli:
		sceneZhongli(p, pose, t, sceneA)
	}
}

// ---------- 通用绘制原语 ----------

func openingSmooth(t, a, b float32) float32 { return visual.Smooth(t, a, b) }

func openingWindow(t, a, b, c, d float32) float32 {
	return visual.Smooth(t, a, b) * (1 - visual.Smooth(t, c, d))
}

func openingHash(i int) float32 {
	x := uint32(i)*747796405 + 2891336453
	x = ((x >> ((x >> 28) + 4)) ^ x) * 277803737
	return float32((x>>22^x)&0xffffff) / 16777216
}

func openingStar(p *ui.Painter, cx, cy, r float32, col ui.Color, alpha, rot float32) {
	if alpha <= .003 || r <= .2 {
		return
	}
	path := &ui.Path{}
	for j := 0; j < 8; j++ {
		a := float64(rot) + float64(j)*tau/8
		d := r
		if j%2 == 1 {
			d = r * .17
		}
		x := cx + float32(math.Cos(a))*d
		y := cy + float32(math.Sin(a))*d
		if j == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	p.FillPath(path, col.Alpha(alpha))
}

func openingDiamond(p *ui.Painter, cx, cy, r, rot float32, col ui.Color, alpha float32) {
	if alpha <= .003 || r <= .2 {
		return
	}
	c, s := float32(math.Cos(float64(rot))), float32(math.Sin(float64(rot)))
	q := func(x, y float32) (float32, float32) { return cx + r*(x*c-y*s), cy + r*(x*s+y*c) }
	a, b := q(0, -1)
	cx2, cy2 := q(.62, 0)
	d, e := q(0, 1)
	f, g := q(-.62, 0)
	path := polyPath(a, b, cx2, cy2, d, e, f, g)
	path.Close()
	p.FillPath(path, col.Alpha(alpha*.30))
	p.StrokePath(path, .8, col.Alpha(alpha))
}

func openingRibbon(p *ui.Painter, a, b, c, d [2]float32, width float32, col ui.Color, alpha float32) {
	if alpha <= .002 {
		return
	}
	path := &ui.Path{}
	path.MoveTo(a[0], a[1])
	path.CubeTo(b[0], b[1], c[0], c[1], d[0], d[1])
	p.StrokePath(path, width*3, col.Alpha(alpha*.22))
	p.StrokePath(path, width, col.Alpha(alpha))
}

func openingEllipse(cx, cy, rx, ry float32) *ui.Path {
	const n = 96
	path := &ui.Path{}
	for i := 0; i <= n; i++ {
		a := float64(i) * tau / n
		x, y := cx+rx*float32(math.Cos(a)), cy+ry*float32(math.Sin(a))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	return path
}

func openingFilledPetal(p *ui.Painter, cx, cy, ang, length, width float32, col ui.Color) {
	tipX := cx + length*float32(math.Cos(float64(ang)))
	tipY := cy + length*float32(math.Sin(float64(ang)))
	nx := -float32(math.Sin(float64(ang))) * width
	ny := float32(math.Cos(float64(ang))) * width
	midX := cx + (tipX-cx)*.55
	midY := cy + (tipY-cy)*.55
	path := &ui.Path{}
	path.MoveTo(cx, cy)
	path.CubeTo(midX+nx, midY+ny, tipX+nx*.35, tipY+ny*.35, tipX, tipY)
	path.CubeTo(tipX-nx*.35, tipY-ny*.35, midX-nx, midY-ny, cx, cy)
	path.Close()
	p.FillPath(path, col)
}

func openingPearl(p *ui.Painter, cx, cy, r, t, alpha float32, ice bool) {
	if alpha <= .003 || r <= .2 {
		return
	}
	blue := ui.Hex("#68BFEA")
	green := ui.Hex("#86D9AD")
	col := green
	if ice {
		col = blue
	}
	glow(p, cx, cy, r*2.9, col, alpha*.25)
	for k := 10; k >= 1; k-- {
		rr := r * float32(k) / 10
		cc := col.Alpha(alpha * (.12 + .025*float32(10-k)))
		circle(p, cx+rr*.18, cy+rr*.22, rr, cc)
	}
	circle(p, cx-r*.28, cy-r*.32, r*.20, ui.RGBA(240, 252, 255, alpha*.65))
	path := openingEllipse(cx, cy, r, r)
	p.StrokePath(path, .8, col.Alpha(alpha*.65))
}

// ---------- 背景：甘雨 / 宵宫 / 八重 / 纳西妲 / 芙宁娜 / 珐露珊 / 钟离 ----------

func drawGanyuAtmosphere(p *ui.Painter, r ui.Rect, in *visual.Intro, fade, short float32) {
	for i := 0; i < visual.LantCount; i++ {
		x, y, al, size := in.LanternAt(i, short)
		if al <= .004 || size <= .3 {
			continue
		}
		drawLantern(p, r.X+r.W*x, r.Y+r.H*y, size, al*fade)
	}
	for i := 0; i < visual.SnowCount; i++ {
		x, y, al, ln := in.SnowAt(i, short)
		if al <= .01 || ln <= 0 {
			continue
		}
		cx, cy := r.X+r.W*x, r.Y+r.H*y
		w := max32(.65, ln*.45)
		seg(p, cx, cy, cx+ln*.45, cy-ln*1.8, w, ui.RGBA(208, 235, 243, al*fade))
	}
	for j := 0; j < 55; j++ {
		u := float32(math.Mod(float64(openingHash(j+7)+in.T*(.04+openingHash(j+13)*.04)), 1))
		qx := r.X + r.W*openingHash(j+21) + sinf(float64(in.T*.4+float32(j)))*short*.012
		qy := r.Y + u*r.H
		k := fade * openingSmooth(in.T, .6, 1.7) * (.12 + openingHash(j+43)*.45)
		if j%7 == 0 {
			openingStar(p, qx, qy, short*.003, ui.Hex("#BFE4F7"), k, math.Pi*.25)
		} else {
			circle(p, qx, qy, short*(.0006+openingHash(j+56)*.0010), ui.Hex("#D0ECFF").Alpha(k))
		}
	}
}

func drawWarmHorizon(p *ui.Painter, r ui.Rect, fade float32) {
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y + r.H*.56, W: r.W, H: r.H * .44}, ui.LinearGradient{
		From: ui.Transparent, To: ui.RGBA(180, 76, 34, .18*fade), Angle: 180,
	}, 0)
	for i := 0; i < 12; i++ {
		y := r.Y + r.H*(.60+openingHash(i+3)*.34)
		w := r.W * (.14 + openingHash(i+9)*.24)
		x := r.X + r.W*openingHash(i+15) - w*.5
		p.FillGradient(ui.Rect{X: x, Y: y, W: w, H: 2}, ui.LinearGradient{
			From: ui.Transparent, To: ui.RGBA(255, 148, 76, .10*fade), Angle: 0,
		}, 0)
	}
}

func drawToriiGate(p *ui.Painter, x, top, h, w, alpha float32) {
	if alpha <= .003 {
		return
	}
	col := ui.RGBA(210, 120, 170, alpha)
	post := w * .09
	left := x - w*.5
	right := x + w*.5
	p.StrokePath(polyPath(left, top+h*.24, left, top+h), post, col)
	p.StrokePath(polyPath(right, top+h*.24, right, top+h), post, col)
	// 双层横梁，顶端微微上翘。
	beam := func(y, dx float32, width float32, a float32) {
		path := &ui.Path{}
		path.MoveTo(left-dx, y)
		path.CubeTo(left+w*.24, y-2, right-w*.24, y-2, right+dx, y)
		p.StrokePath(path, width, ui.RGBA(210, 120, 170, a))
	}
	beam(top+h*.26, w*.11, post*1.3, alpha*.85)
	beam(top+h*.43, w*.03, post*.8, alpha*.55)
}

func drawSanctuary(p *ui.Painter, r ui.Rect, t, fade float32) {
	cx := r.X + r.W*.5
	cy := r.Y + r.H*.72
	open := openingSmooth(t, .62, 2.04)
	for j := 0; j < 5; j++ {
		rx := r.W * (.12 + float32(j)*.075) * (.4 + .6*open)
		ry := r.H * (.36 + float32(j)*.055) * (.4 + .6*open)
		path := &ui.Path{}
		path.MoveTo(cx-rx, cy)
		path.CubeTo(cx-rx, cy-ry*1.3, cx+rx, cy-ry*1.3, cx+rx, cy)
		col := ui.RGBA(150, 198, 128, fade*(.16-float32(j)*.018))
		p.StrokePath(path, 1.0, col)
		// 立柱：把拱门和地面连接起来。
		p.StrokePath(polyPath(cx-rx, cy-ry*.35, cx-rx, cy), .7, col)
		p.StrokePath(polyPath(cx+rx, cy-ry*.35, cx+rx, cy), .7, col)
	}
	for j := 0; j < 8; j++ {
		a := float64(j) * tau / 8
		x := cx + float32(math.Cos(a))*r.W*.34*open
		y := cy - r.H*.10 + float32(math.Sin(a))*r.H*.12
		openingDiamond(p, x, y, r.H*.014, 0, ui.Hex("#D5E9B2"), fade*.35)
	}
}

func drawCurtains(p *ui.Painter, r ui.Rect, t, fade float32) {
	open := openingSmooth(t, .10, 1.86)
	// 开幕时帘子从中间向两侧拉开：inner 从 0.48W 收到 0.09W。
	inner := r.W * (.48 - open*.39)
	if inner <= 1 {
		return
	}
	const bands = 5
	w := inner / bands
	for _, side := range []float32{-1, 1} {
		for band := 0; band < bands; band++ {
			var x0 float32
			if side < 0 {
				x0 = r.X + float32(band)*w
			} else {
				x0 = r.X + r.W - float32(band+1)*w
			}
			shade := ui.RGBA(uint8(10+band*3), uint8(22+band*5), uint8(48+band*8), fade*.86)
			p.Fill(ui.Rect{X: x0, Y: r.Y, W: w + .6, H: r.H}, shade, 0)
			fold := ui.RGBA(120, 155, 210, fade*.13)
			for lane := 0; lane < 2; lane++ {
				fx := x0 + w*(.25+float32(lane)*.45)
				p.StrokePath(polyPath(fx, r.Y, fx+sinf(float64(t*.4+float32(lane)))*6, r.Y+r.H), .8, fold)
			}
		}
		// 内侧边缘：一条冷色高光，像帘布被拉开后的褶边。
		ex := r.X + inner
		if side > 0 {
			ex = r.X + r.W - inner
		}
		p.StrokePath(polyPath(ex, r.Y, ex+sinf(float64(t*.5))*4, r.Y+r.H), 1.0,
			ui.RGBA(181, 200, 227, fade*.32))
	}
}

func drawArchive(p *ui.Painter, r ui.Rect, t, fade float32) {
	for side := -1; side <= 1; side += 2 {
		baseX := r.X + r.W*.5 + float32(side)*r.W*.35
		for j := 0; j < 4; j++ {
			x := baseX + float32(side)*float32(j)*r.W*.045
			y := r.Y + r.H*(.34+float32(j)*.08)
			w := r.W * .12
			h := r.H * .34
			col := ui.RGBA(20+uint8(j)*4, 62+uint8(j)*5, 62+uint8(j)*4, fade*.34)
			p.Fill(ui.Rect{X: x - w*.5, Y: y, W: w, H: h}, col, 2)
			p.StrokePath(polyPath(x-w*.5, y, x+w*.5, y, x+w*.5, y+h, x-w*.5, y+h), .7,
				ui.RGBA(198, 224, 198, fade*.18))
			for k := 0; k < 3; k++ {
				yy := y + h*(.22+float32(k)*.24)
				p.StrokePath(polyPath(x-w*.36, yy, x+w*.30, yy), .6,
					ui.RGBA(226, 211, 163, fade*.14))
			}
		}
	}
	// 斜向的风痕，像书架之间穿过的气流。
	for j := 0; j < 5; j++ {
		x := r.X + r.W*.14 + float32(j)*r.W*.17
		y := r.Y + r.H*.30 + sinf(float64(t*.3+float32(j)))*r.H*.02
		p.StrokePath(polyPath(x, y, x+r.W*.08, y-r.H*.10, x+r.W*.16, y-r.H*.02), .8,
			ui.RGBA(160, 230, 200, fade*.10))
	}
}

func drawStoneHall(p *ui.Painter, r ui.Rect, t, fade float32) {
	for side := -1; side <= 1; side += 2 {
		for j := 0; j < 3; j++ {
			x := r.X + r.W*.5 + float32(side)*r.W*(.34+float32(j)*.085)
			y := r.Y + r.H*(.91+float32(j)*.03)
			h := r.H * (.28 + float32(j)*.10)
			drawPillar(p, x, y, h, r.W*(.026+float32(j)*.003), fade*(.68-float32(j)*.15))
		}
	}
	for j := 0; j < 7; j++ {
		x := r.X + r.W*.5 + (openingHash(j+180)-.5)*r.W*.62
		y := r.Y + r.H*(.30+openingHash(j+190)*.32) + sinf(float64(t*.3+float32(j)))*8
		openingDiamond(p, x, y, r.H*(.010+openingHash(j+200)*.012), t*.08, ui.Hex("#E3B45D"), fade*.28)
	}
}

func drawPillar(p *ui.Painter, x, baseY, h, w, alpha float32) {
	if alpha <= .003 {
		return
	}
	top := baseY - h
	left := polyPath(x-w, baseY, x-w, top+w*.36, x, top+w*.74, x, baseY+w*.35)
	right := polyPath(x, baseY+w*.35, x, top+w*.74, x+w, top+w*.36, x+w, baseY)
	cap := polyPath(x-w, top+w*.36, x, top, x+w, top+w*.36, x, top+w*.74)
	p.FillPath(left, ui.RGBA(75, 53, 34, alpha*.90))
	p.FillPath(right, ui.RGBA(38, 30, 24, alpha*.95))
	p.FillPath(cap, ui.RGBA(146, 96, 55, alpha*.65))
	p.StrokePath(left, .8, ui.RGBA(213, 163, 80, alpha*.38))
	p.StrokePath(right, .7, ui.RGBA(195, 138, 56, alpha*.30))
	for j := 0; j < 4; j++ {
		y := top + h*(.18+float32(j)*.19)
		p.StrokePath(polyPath(x-w*.72, y, x-w*.72, y+h*.055, x-w*.26, y+h*.075, x-w*.26, y+h*.025),
			1, ui.RGBA(244, 195, 109, alpha*.55*(.55+.45*sinf(float64(float32(j)*1.7)))))
	}
}

// ---------- 氛围粒子 ----------

func drawOpeningMotes(p *ui.Painter, r ui.Rect, o visual.Opening, t, fade, short float32) {
	var col ui.Color
	switch o {
	case visual.OpeningYoimiya:
		col = ui.Hex("#FFB66D")
	case visual.OpeningYae:
		col = ui.Hex("#D6A2DF")
	case visual.OpeningNahida:
		col = ui.Hex("#BCF4A4")
	case visual.OpeningFurina:
		col = ui.Hex("#9BE3FF")
	case visual.OpeningFaruzan:
		col = ui.Hex("#90E4BD")
	case visual.OpeningZhongli:
		col = ui.Hex("#F0AB34")
	}
	for j := 0; j < 46; j++ {
		u := float32(math.Mod(float64(openingHash(j+700)+t*(.05+openingHash(j+710)*.05)), 1))
		x := r.X + r.W*openingHash(j+720)
		y := r.Y + r.H*(1-u)
		a := fade * openingSmooth(t, .3, 1.4) * float32(math.Sin(float64(u*math.Pi))) *
			(.12 + openingHash(j+730)*.35)
		if a <= .004 {
			continue
		}
		switch o {
		case visual.OpeningYae:
			sakura(p, x, y, short*(.012+openingHash(j+740)*.012), col.Alpha(a), t*.2+float32(j))
		case visual.OpeningNahida:
			openingFilledPetal(p, x, y, -math.Pi/2+float32(openingHash(j+750)-.5), short*.018, short*.006, col.Alpha(a*.7))
		case visual.OpeningFaruzan:
			p.StrokePath(polyPath(x-short*.008, y+short*.010, x, y-short*.006, x+short*.008, y+short*.010),
				.8, col.Alpha(a))
		case visual.OpeningZhongli:
			openingDiamond(p, x, y, short*.006, t*.2, col, a)
		case visual.OpeningFurina:
			circle(p, x, y, max32(.7, short*.0015), col.Alpha(a))
			circle(p, x, y, max32(1.2, short*.004), col.Alpha(a*.18))
		default:
			seg(p, x, y, x, y-short*.008, max32(.7, short*.0014), col.Alpha(a))
		}
	}
}

// ---------- 各角色前景场景 ----------

func sceneGanyu(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	lift := openingSmooth(t, .85, 3.25)
	orb := [2]float32{x, y + r*.28*(1-lift)}
	pearlA := openingWindow(t, .08, .48, 1.35, 2.45) * a
	openingPearl(p, orb[0], orb[1], r*(.12+.08*openingSmooth(t, .1, 1.1)), t, pearlA, true)
	glow(p, x, y+r*(1-lift)*.72, r*1.45, ui.Hex("#2979B8"), openingWindow(t, .12, .6, 2.2, 3.3)*a*.30)

	for lane := 0; lane < 3; lane++ {
		age := t - .8 - float32(lane)*.10
		if age < 0 {
			continue
		}
		k := openingWindow(t, .8, 1.35, 2.7, 3.4) * a * .28
		yy := .65 - age*.32
		openingRibbon(p,
			[2]float32{x - r*1.5, y + r*(yy+.18)},
			[2]float32{x - r*.5, y + r*(yy-.3)},
			[2]float32{x + r*.7, y + r*(yy+.4)},
			[2]float32{x + r*1.5, y + r*(yy-.12)},
			r*.012, ui.Hex("#8ED5FF"), k)
	}

	release := openingWindow(t, 1.95, 2.18, 3.7, 4.5) * a
	for j := 0; j < 70; j++ {
		age := max32(0, t-1.98-openingHash(j+9)*.26)
		angle := openingHash(j+18) * tau
		d := .20 + age*(.6+openingHash(j+31)*.6)
		xx := float32(math.Cos(float64(angle))) * d
		yy := float32(math.Sin(float64(angle)))*d + age*age*.30
		k := release * (1 - openingSmooth(age, .65, 2.2)) * (.35 + .65*openingHash(j+41))
		if j%4 == 0 {
			openingStar(p, x+xx*r, y+yy*r, r*(.012+openingHash(j+52)*.012), ui.Hex("#C7F3FF"), k, math.Pi*.25)
		} else {
			qx, qy := x+xx*r, y+yy*r
			seg(p, qx, qy, qx+r*.006, qy+r*.035, .8, ui.Hex("#B5E5FF").Alpha(k))
		}
	}
	flash := float32(math.Exp(-math.Pow(float64((t-2.42)/.30), 2))) * a
	glow(p, x, y, r*1.9, ui.Hex("#7CCBFF"), flash*.20)
}

func fireworkBurst(p *ui.Painter, cx, cy, r, age, alpha float32, seed, count int) {
	if age < 0 || age > 2.6 || alpha <= .002 {
		return
	}
	onset := openingSmooth(age, 0, .045)
	life := 1 - openingSmooth(age, .70, 2.4)
	glow(p, cx, cy, r*(.25+age*.6), ui.Hex("#FF9C58"), alpha*float32(math.Exp(float64(-age*8)))*.55)
	for j := 0; j < count; j++ {
		angle := float64(j)*tau/float64(count) + float64(openingHash(j+seed))*.035
		speed := .65 + openingHash(j+seed+77)*.5
		point := func(v float32) (float32, float32) {
			d := (1 - float32(math.Exp(float64(-v*1.6)))) * r * speed
			return cx + float32(math.Cos(angle))*d, cy + float32(math.Sin(angle))*d + v*v*r*.15
		}
		for k := 0; k < 12; k++ {
			u := float32(k) / 12
			v := float32(k+1) / 12
			x0, y0 := point(max32(0, age-.30*(1-u)))
			x1, y1 := point(max32(0, age-.30*(1-v)))
			col := ui.Hex("#FFCF86")
			if j%4 == 0 {
				col = ui.Hex("#E96A4C")
			}
			aa := alpha * onset * life * u * (.65 + .35*sinf(float64(age*8+float32(j)*2)))
			seg(p, x0, y0, x1, y1, r*.016, col.Alpha(aa*.12))
			seg(p, x0, y0, x1, y1, max32(.65, r*.0035), ui.Hex("#FFE7A4").Alpha(aa))
		}
		qx, qy := point(age)
		if j%4 == 0 {
			openingStar(p, qx, qy, r*.009, ui.Hex("#FFF4CD"), alpha*onset*life*.7, math.Pi*.25)
		}
	}
}

func sceneYoimiya(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	ascent := openingSmooth(t, .16, 1.40)
	launch := openingWindow(t, .1, .35, 1.38, 1.62) * a
	path := func(u float32) (float32, float32) {
		return x + sinf(float64(u*math.Pi))*.19*r, y + r*2.4*(1-u)
	}
	for j := 0; j < 32; j++ {
		u, v := float32(j)/32, float32(j+1)/32
		x0, y0 := path(max32(0, ascent-.24*(1-u)))
		x1, y1 := path(max32(0, ascent-.24*(1-v)))
		seg(p, x0, y0, x1, y1, r*.035, ui.Hex("#FFB66D").Alpha(launch*u*.13))
		seg(p, x0, y0, x1, y1, 1.1, ui.Hex("#FFF0C1").Alpha(launch*u))
	}
	sx, sy := path(ascent)
	openingStar(p, sx, sy, r*.035, ui.Hex("#FFF3D0"), launch, math.Pi*.25)

	fireworkBurst(p, x, y, r*1.72, t-1.40, a*.83, 73, 64)
	orbit := openingWindow(t, .75, 1.30, 2.72, 3.55) * a
	for _, side := range []float32{-1, 1} {
		angle := t*1.25 + float32(side)*float32(math.Pi)*.5
		d := 1.38 - .25*openingSmooth(t, 1.3, 2.8)
		qx := x + float32(math.Cos(float64(angle)))*d*r
		qy := y + float32(math.Sin(float64(angle)))*d*r*.57
		goldfish(p, qx, qy, r*.34, ui.RGBA(255, 196, 110, orbit*.85), t)
		for k := 0; k < 38; k++ {
			u := float32(k) / 38
			ang := angle - u*1.4
			x0 := x + float32(math.Cos(float64(ang)))*d*r
			y0 := y + float32(math.Sin(float64(ang)))*d*r*.57
			x1 := x + float32(math.Cos(float64(ang-.025)))*d*r
			y1 := y + float32(math.Sin(float64(ang-.025)))*d*r*.57
			seg(p, x0, y0, x1, y1, max32(.7, r*.009*(1-u)), ui.Hex("#FFC26D").Alpha(orbit*(1-u)*.40))
		}
	}
	fireworkBurst(p, x-r*2.1, y-r*.95, r*.85, t-2.52, a*.50, 170, 44)
	fireworkBurst(p, x+r*2.0, y-r*.60, r*.72, t-2.90, a*.48, 299, 38)
	glow(p, x, y, r*1.8, ui.Hex("#F77E3F"), float32(math.Exp(-math.Pow(float64((t-2.15)/.35), 2)))*a*.27)
}

func sakuraTotem(p *ui.Painter, cx, cy, r, t, a float32) {
	if a <= .002 {
		return
	}
	glow(p, cx, cy-r*.13, r*.75, ui.Hex("#B664D3"), a*.22)
	// 神樱的柱体与左右分支。
	stem := polyPath(cx-r*.12, cy+r*.70, cx-r*.09, cy-r*.25, cx, cy-r*.52,
		cx+r*.09, cy-r*.25, cx+r*.12, cy+r*.70)
	stem.Close()
	p.FillPath(stem, ui.Hex("#7E459A").Alpha(a*.38))
	p.StrokePath(stem, 1.0, ui.Hex("#D19AD8").Alpha(a*.78))
	star := &ui.Path{}
	star.MoveTo(cx, cy-r*.52)
	star.LineTo(cx, cy+r*.70)
	star.LineTo(cx+r*.12, cy+r*.70)
	star.Close()
	p.FillPath(star, ui.Hex("#E4ABD8").Alpha(a*.22))
	p.StrokePath(polyPath(cx, cy-r*.18, cx, cy+r*.61), .75, ui.Hex("#F1D8EE").Alpha(a*.58))
	for level := 0; level < 3; level++ {
		yy := cy - r*(.15-float32(level)*.17)
		openingDiamond(p, cx, yy, r*.035, 0, ui.Hex("#EAC9E7"), a*.62)
	}
	for _, side := range []float32{-1, 1} {
		p.StrokePath(polyPath(cx, cy-r*.15, cx+side*r*.25, cy+r*.02, cx+side*r*.36, cy+r*.36, cx+side*r*.25, cy+r*.56),
			r*.014, ui.Hex("#C891DC").Alpha(a*.78))
		p.StrokePath(polyPath(cx+side*r*.48, cy+r*.38, cx+side*r*.25, cy+r*.48, cx, cy+r*.45),
			1.2, ui.Hex("#F1D5ED").Alpha(a*.86))
	}
	sakura(p, cx, cy-r*.88, r*.30, ui.Hex("#E8AAD9").Alpha(a*.92), t*.12)
	openingStar(p, cx, cy-r*.88, r*.07, ui.Hex("#F8E5FA"), a, 0)
}

func lightning(p *ui.Painter, ax, ay, bx, by, age, alpha float32, seed int) {
	grown := openingSmooth(age, 0, .30)
	if alpha <= .002 || grown <= 0 {
		return
	}
	dx, dy := bx-ax, by-ay
	length := max32(.01, float32(math.Hypot(float64(dx), float64(dy))))
	pts := make([]float32, 0, 66)
	for j := 0; j <= 32; j++ {
		u := float32(j) / 32 * grown
		off := (sinf(float64(u*43+float32(seed)))*.043 + sinf(float64(u*91+float32(seed*4)))*.025) *
			sinf(float64(u*math.Pi)) * length
		pts = append(pts, ax+dx*u-dy/length*off, ay+dy*u+dx/length*off)
	}
	p.StrokePath(polyPath(pts...), 3, ui.Hex("#E9B7FF").Alpha(alpha*.18))
	p.StrokePath(polyPath(pts...), 1.0, ui.Hex("#E9B7FF").Alpha(alpha))
}

func sceneYae(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	locs := [][3]float32{{0, -1.12, 0}, {1.42, .72, .22}, {-1.42, .72, .22}}
	for j, l := range locs {
		qx, qy := x+l[0]*r, y+l[1]*r
		birth := .14 + float32(j)*.22
		up := openingSmooth(t, birth, birth+.7)
		k := up * (1 - openingSmooth(t, 2.7+float32(j)*.13, 3.7+float32(j)*.13)) * a
		sakuraTotem(p, qx, qy+r*.25*(1-up), r*.49, t, k)
		discharge := openingWindow(t, 1.42+float32(j)*.15, 1.77+float32(j)*.15,
			2.6+float32(j)*.13, 3.4+float32(j)*.13) * a
		lightning(p, qx, qy-r*.34, x, y, t-1.42-float32(j)*.15, discharge*.82, j+7)
	}
	pulse := float32(math.Exp(-math.Pow(float64((t-2.38)/.32), 2))) * a
	glow(p, x, y, r*2.0, ui.Hex("#AA5BCC"), pulse*.33)
	fade := openingWindow(t, .6, 1.3, 3.75, 4.6) * a
	for j := 0; j < 38; j++ {
		age := max32(0, t-.40-openingHash(j+40)*1.25)
		angle := openingHash(j+48)*tau + age*.26
		d := .6 + openingHash(j+90)*1.6
		qx := x + float32(math.Cos(float64(angle)))*d*r
		qy := y + float32(math.Sin(float64(angle)))*d*r*.64 + age*.065*r
		sakura(p, qx, qy, r*(.035+openingHash(j+65)*.035), ui.Hex("#D6A2DF").Alpha(fade*(.22+openingHash(j+72)*.5)), angle+t*.4)
	}
}

func sceneNahida(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	awake := openingWindow(t, .12, .65, 3.8, 4.65) * a
	baseX, baseY := x, y+r*1.04
	bloom := openingSmooth(t, .18, 1.72)
	glow(p, baseX, baseY, r*.8, ui.Hex("#8CD9A0"), awake*.28)
	for i := 0; i < 8; i++ {
		ang := -math.Pi/2 + float64(i)*tau/8
		openingFilledPetal(p, baseX, baseY, float32(ang),
			r*(.55+.35*bloom), r*.16, ui.Hex("#54B98C").Alpha(awake*.30))
		openingFilledPetal(p, baseX, baseY, float32(ang)+.2,
			r*(.45+.30*bloom), r*.12, ui.Hex("#B9EB9C").Alpha(awake*.16))
	}
	for i := 0; i < 6; i++ {
		ang := -math.Pi/2 + (float64(i)+.5)*tau/6
		openingFilledPetal(p, baseX, baseY, float32(ang),
			r*(.30+.22*bloom), r*.10, ui.Hex("#D5FBC5").Alpha(awake*.30))
	}
	// 梦境宫殿的八块拱面。
	open := openingSmooth(t, .62, 2.04)
	cx, cy := x, y+r*.30
	for j := 0; j < 8; j++ {
		az := float64(j)*tau/8 - math.Pi/2
		px := cx + float32(math.Cos(az))*r*1.35*open
		py := cy - r*.35 + float32(math.Sin(az))*r*.32*open
		path := &ui.Path{}
		path.MoveTo(px-r*.12, py+r*.45)
		path.CubeTo(px-r*.12, py, px+r*.12, py, px+r*.12, py+r*.45)
		p.StrokePath(path, .85, ui.Hex("#D4E5B7").Alpha(awake*.22))
	}
	orbit := openingWindow(t, 1.0, 1.6, 3.45, 4.45) * a
	for j := 0; j < 10; j++ {
		angle := float64(j)*tau/10 + float64(t)*.12
		qx := x + float32(math.Cos(angle))*r*1.46
		qy := y + float32(math.Sin(angle))*r*1.08 - r*.05
		openingFilledPetal(p, qx, qy, float32(angle)-.6, r*.25, r*.08, ui.Hex("#B9E89C").Alpha(orbit*.50))
		if j%2 == 0 {
			openingStar(p, qx, qy, r*.018, ui.Hex("#E5FBC4"), orbit*.7, math.Pi*.25)
		}
	}
	glow(p, x, y, r*1.65, ui.Hex("#81D29C"), float32(math.Exp(-math.Pow(float64((t-2.42)/.4), 2)))*a*.30)
}

func sceneFurina(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	stage := openingWindow(t, .30, 1.10, 4.02, 4.65) * a
	waterY := y + r*1.08
	glow(p, x, waterY, r*2.4, ui.Hex("#2D81B5"), stage*.31)
	for tier := 2; tier >= 0; tier-- {
		assemble := openingSmooth(t, .25, 1.25)
		rx := r * (1.87 + float32(tier)*.13) * assemble
		ry := r * (.31 + float32(tier)*.038) * assemble
		cy := waterY + float32(tier)*r*.075
		path := openingEllipse(x, cy, rx, ry)
		p.FillPath(path, ui.RGBA(10, 30, 52, stage*(.30-float32(tier)*.06)))
		p.StrokePath(path, .9, ui.RGBA(180, 215, 232, stage*(.32-float32(tier)*.06)))
		for lane := 0; lane < 3; lane++ {
			yy := cy - ry*.35 + float32(lane)*ry*.32
			p.StrokePath(polyPath(x-rx*.8, yy, x-rx*.3, yy-ry*.18, x+rx*.3, yy+ry*.12, x+rx*.8, yy),
				.7, ui.RGBA(125, 182, 210, stage*.12))
		}
	}
	for j := 0; j < 3; j++ {
		age := t - .99 - float32(j)*.19
		if age < 0 {
			continue
		}
		fade := stage * openingSmooth(age, 0, .25) * (1 - openingSmooth(age, 1.25, 3.15))
		path := openingEllipse(x, waterY, r*(.16+age*.59), r*(.025+age*.095))
		p.StrokePath(path, .8, ui.Hex("#A1E8FB").Alpha(fade*(.27-float32(j)*.045)))
	}
	drop := openingWindow(t, .15, .38, 1.02, 1.17) * a
	fall := openingSmooth(t, .15, 1.07)
	dy := -1.6 + (1.08+1.6)*fall
	droplet(p, x, y+r*dy, r*.08, ui.Hex("#BFEFFF").Alpha(drop))
	openingRibbon(p, [2]float32{x, y + r*(dy-.55)}, [2]float32{x - r*.03, y + r*(dy-.29)},
		[2]float32{x + r*.02, y + r*(dy-.15)}, [2]float32{x, y + r*dy},
		r*.028, ui.Hex("#BFEFFF"), drop*.3)
	burst := openingWindow(t, 1.05, 1.22, 2.0, 2.85) * a
	for j := 0; j < 40; j++ {
		age := max32(0, t-1.07)
		angle := float64(j) * tau / 40
		d := age * (.65 + openingHash(j+802)*.65)
		h := -float32(math.Sin(angle))*d*.75 + age*age*.40
		qx := x + float32(math.Cos(angle))*d*r
		qy := waterY + h*r
		glow(p, qx, qy, r*.023, ui.Hex("#9BE3FF"), burst*.35)
		circle(p, qx, qy, max32(.6, r*.008), ui.Hex("#C2F1FF").Alpha(burst*.7))
	}
	glow(p, x, y, r*1.8, ui.Hex("#65C0EC"), float32(math.Exp(-math.Pow(float64((t-2.7)/.43), 2)))*a*.21)
}

func sceneFaruzan(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	awake := openingSmooth(t, .06, 1.1)
	open := openingSmooth(t, 1.06, 2.45)
	glow(p, x, y, r*2, ui.Hex("#47D8AC"), openingSmooth(t, .15, .85)*(1-openingSmooth(t, 2.15, 2.85))*a*.25)

	// 六片机关叶：围绕中心旋转、向外展开。
	for j := 0; j < 6; j++ {
		ang := float64(j)*tau/6 + float64(t)*.37 + float64(open)*.22
		inner, outer := r*.26, r*(.86+.30*open)
		x0, y0 := x+float32(math.Cos(ang))*inner, y+float32(math.Sin(ang))*inner
		x1, y1 := x+float32(math.Cos(ang-.42))*outer, y+float32(math.Sin(ang-.42))*outer
		x2, y2 := x+float32(math.Cos(ang+.42))*outer, y+float32(math.Sin(ang+.42))*outer
		col := ui.Hex("#82D4B7").Alpha(awake * (1 - openingSmooth(t, 2.05, 2.90)) * a * .55)
		path := polyPath(x0, y0, x1, y1, x2, y2)
		path.Close()
		p.FillPath(path, col)
		p.StrokePath(path, .8, ui.Hex("#E7D19D").Alpha(awake*(1-openingSmooth(t, 2.05, 2.90))*a*.80))
		// 叶面内环与中心刻线。
		ix1, iy1 := x+(x1-x)*.72, y+(y1-y)*.72
		ix2, iy2 := x+(x2-x)*.72, y+(y2-y)*.72
		p.StrokePath(polyPath(ix1, iy1, ix2, iy2), .7, ui.Hex("#C0F2D8").Alpha(awake*a*.55))
	}
	// 风带：三道螺旋弧。
	for lane := 0; lane < 3; lane++ {
		radius := r * (1.18 + float32(lane)*.17 + .27*openingSmooth(t, 2.6, 3.8))
		head := t*(1.7+float32(lane)*.11) + float32(lane)*tau/3
		active := openingSmooth(t, .55, 1.2) * (1 - openingSmooth(t, 3.55, 4.3)) * a
		arcLine(p, x, y+sinf(float64(head))*.05*r, radius, float64(head)-4.9, float64(head),
			r*(.020-float32(lane)*.003), ui.Hex("#B1FFE1").Alpha(active*(.55-float32(lane)*.09)))
	}
	crest := float32(math.Exp(-math.Pow(float64((t-3.24)/.28), 2))) * a
	glow(p, x, y, r*2.15, ui.Hex("#57D5AB"), crest*.24)
	for j := 0; j < 70; j++ {
		u := float32(math.Mod(float64(openingHash(j+51)+t*(.14+openingHash(j+61)*.10)), 1))
		angle := u*tau + openingHash(j+71)*tau
		distance := (1.62 - u*.95) * (1 + .30*openingSmooth(t, 3, 4))
		qx := x + float32(math.Cos(float64(angle)))*r*distance
		qy := y + float32(math.Sin(float64(angle)))*r*distance*.60
		k := openingSmooth(t, .5, 1.25) * (1 - openingSmooth(t, 4.0, 4.6)) * a *
			float32(math.Sin(float64(u*math.Pi))) * (.2 + openingHash(j+91)*.6)
		openingStar(p, qx, qy, r*(.004+openingHash(j+101)*.006), ui.Hex("#DDFFEC"), k, angle)
	}
}

func sceneZhongli(p *ui.Painter, pose visual.MarkPose, t, a float32) {
	x, y, r := pose.X, pose.Y, pose.Radius
	descend := openingSmooth(t, .06, 1.12)
	open := openingSmooth(t, 1.04, 1.88)
	meteorA := openingSmooth(t, .04, .38) * (1 - openingSmooth(t, 1.27, 2.05)) * a
	cy := y - r*1.65*(1-descend)
	for j := 0; j < 3; j++ {
		tail := .28 + float32(j)*.18
		glow(p, x-r*tail*.12, cy-r*tail, r*(.28-float32(j)*.045), ui.Hex("#D3922B"),
			meteorA*(1-descend)*(.15-float32(j)*.035))
	}
	drawMeteorCubes(p, x, cy, r*(.37+.06*open), t, meteorA)
	crest := float32(math.Exp(-math.Pow(float64((t-3.15)/.25), 2))) * a
	glow(p, x, y, r*2.3, ui.Hex("#D79222"), crest*.27)
	rupture := float32(math.Exp(-math.Pow(float64((t-1.48)/.25), 2))) * a
	glow(p, x, y, r*1.3, ui.Hex("#FFCE73"), rupture*.18)
	release := openingSmooth(t, 2.8, 3.20) * (1 - openingSmooth(t, 4.1, 4.6)) * a
	for layer := 0; layer < 2; layer++ {
		u := openingSmooth(t, 2.82+float32(layer)*.16, 4.25+float32(layer)*.12)
		rr := r * (.82 + u*(1.45+float32(layer)*.28))
		path := polyPath(x, y-rr*.52, x+rr, y, x, y+rr*.52, x-rr, y)
		path.Close()
		p.StrokePath(path, .8, ui.Hex("#E8B45B").Alpha(release*(1-u)*.48))
	}
	motes := openingSmooth(t, .35, 1.2) * (1 - openingSmooth(t, 4.1, 4.65)) * a
	for j := 0; j < 68; j++ {
		u := float32(math.Mod(float64(openingHash(j+10)+t*(.11+openingHash(j+20)*.06)), 1))
		angle := openingHash(j+30) * tau
		d := .35 + u*1.45
		qx := x + float32(math.Cos(float64(angle)))*r*d
		qy := y + float32(math.Sin(float64(angle)))*r*d
		k := motes * float32(math.Sin(float64(u*math.Pi))) * (.15 + openingHash(j+40)*.6)
		openingDiamond(p, qx, qy, r*(.004+openingHash(j+50)*.010), angle+t*.15, ui.Hex("#F0AB34"), k)
	}
}

// drawMeteorCubes 画八枚下坠的岩方块：三个可见面 + 金边。
func drawMeteorCubes(p *ui.Painter, cx, cy, r, t, alpha float32) {
	if alpha <= .003 {
		return
	}
	yaw := .61 + t*.18
	pitch := -.48 + t*.05
	roll := .24
	rotate := func(x, y, z float32) (float32, float32, float32) {
		x1 := x*float32(math.Cos(float64(yaw))) + z*float32(math.Sin(float64(yaw)))
		z1 := -x*float32(math.Sin(float64(yaw))) + z*float32(math.Cos(float64(yaw)))
		y1 := y*float32(math.Cos(float64(pitch))) - z1*float32(math.Sin(float64(pitch)))
		z2 := y*float32(math.Sin(float64(pitch))) + z1*float32(math.Cos(float64(pitch)))
		return x1*float32(math.Cos(float64(roll))) - y1*float32(math.Sin(float64(roll))),
			x1*float32(math.Sin(float64(roll))) + y1*float32(math.Cos(float64(roll))), z2
	}
	type face struct {
		pts [4][2]float32
		z   float32
		col ui.Color
	}
	var faces []face
	colors := []ui.Color{ui.Hex("#6D4B27"), ui.Hex("#B7853F"), ui.Hex("#F0C46E"),
		ui.Hex("#554022"), ui.Hex("#C79B52"), ui.Hex("#8C622F")}
	indices := [6][4]int{{0, 1, 3, 2}, {4, 6, 7, 5}, {0, 4, 5, 1}, {2, 3, 7, 6}, {0, 2, 6, 4}, {1, 5, 7, 3}}
	gap := .052 + .85*openingSmooth(t, 1.04, 1.88)
	for cube := 0; cube < 8; cube++ {
		var verts [8][3]float32
		for k := 0; k < 8; k++ {
			sx, sy, sz := float32(1), float32(1), float32(1)
			if cube&1 != 0 {
				sx = -1
			}
			if cube&2 != 0 {
				sy = -1
			}
			if cube&4 != 0 {
				sz = -1
			}
			bx, by, bz := float32(0), float32(0), float32(0)
			if k&1 != 0 {
				bx = 1
			} else {
				bx = -1
			}
			if k&2 != 0 {
				by = 1
			} else {
				by = -1
			}
			if k&4 != 0 {
				bz = 1
			} else {
				bz = -1
			}
			verts[k] = [3]float32{sx*(.46+gap) + bx*.43, sy*(.46+gap) + by*.43, sz*(.46+gap) + bz*.43}
		}
		for j := 0; j < 6; j++ {
			f := face{col: colors[j]}
			for k := 0; k < 4; k++ {
				v := verts[indices[j][k]]
				rx, ry, rz := rotate(v[0], v[1], v[2])
				persp := 4 / (4 + rz*.28)
				f.pts[k] = [2]float32{cx + rx*r*persp, cy + ry*r*persp}
				f.z += rz * .25
			}
			// 背面剔除（屏幕空间有向面积）。
			ax, ay := f.pts[0][0], f.pts[0][1]
			bx, by := f.pts[1][0], f.pts[1][1]
			cx2, cy2 := f.pts[2][0], f.pts[2][1]
			if (bx-ax)*(cy2-ay)-(by-ay)*(cx2-ax) < 0 {
				continue
			}
			faces = append(faces, f)
		}
	}
	for i := 0; i < len(faces); i++ {
		for j := i + 1; j < len(faces); j++ {
			if faces[j].z > faces[i].z {
				faces[i], faces[j] = faces[j], faces[i]
			}
		}
	}
	for _, f := range faces {
		path := polyPath(f.pts[0][0], f.pts[0][1], f.pts[1][0], f.pts[1][1],
			f.pts[2][0], f.pts[2][1], f.pts[3][0], f.pts[3][1])
		path.Close()
		p.FillPath(path, f.col.Alpha(alpha))
		p.StrokePath(path, .8, ui.Hex("#FFE1A0").Alpha(alpha*.75))
	}
}
