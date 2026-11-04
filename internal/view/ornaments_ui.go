package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 角色纹样在界面上的三处落点：内容区右下角的大淡纹、侧栏底部、导航选中项。
// 卡片在 library.go 的 cardOrnament 里单独排（它要用卡片坐标）。

// ornamentSignature 画一个纹样的主形；颜色由调用方给。
func ornamentSignature(p *ui.Painter, o visual.Ornament, cx, cy, r float32,
	col ui.Color, t float32) {
	switch o {
	case visual.OrnamentQingxin:
		qingxin(p, cx, cy, r, col, .1)
	case visual.OrnamentGoldfish:
		goldfish(p, cx, cy, r, col, t)
		firework(p, cx+r*.72, cy-r*.80, r*.28, fade(col, .55), .2)
	case visual.OrnamentSakura:
		sakura(p, cx, cy, r, col, .15)
	case visual.OrnamentSprig:
		sprig(p, cx, cy, r, col, t)
	case visual.OrnamentTopHat:
		topHat(p, cx, cy, r, col, t)
	case visual.OrnamentPolyhedron:
		archiveBook(p, cx, cy, r*.92, col, t)
	case visual.OrnamentStoneSeal:
		stoneSeal(p, cx, cy, r, col)
	}
}

// drawChromeOrnament 是内容区右下角的大幅淡纹（参考各角色 _ui.cpp 的 Chrome）。
func (a *App) drawChromeOrnament(p *ui.Painter, r ui.Rect) {
	t := float32(p.Now().Sub(a.t0).Seconds())
	accent := a.tok.Accent
	gold := ui.Hex("#E5C385")
	right := r.X + r.W
	bottom := r.Y + r.H
	switch a.char.Ornament {
	case visual.OrnamentQingxin:
		qingxin(p, right-45, bottom-40, 61, ui.RGBA(176, 217, 235, .11), .1)
	case visual.OrnamentGoldfish:
		goldfish(p, right-83, bottom-48, 47, ui.RGBA(255, 204, 140, .14), t)
		firework(p, right-202, bottom-14, 25, ui.RGBA(255, 218, 162, .11), .2)
	case visual.OrnamentSakura:
		sakura(p, right-51, bottom-44, 57, fade(accent, .15), .15)
		for j := 0; j < 3; j++ {
			y := bottom - 4 + float32(j)*14
			waveLine(p, r.X+r.W*.62, y, r.W*.38, 18+float32(j)*4, 0, fade(gold, .08), .7)
		}
	case visual.OrnamentSprig:
		sprig(p, right-65, bottom-41, 47, ui.RGBA(215, 232, 159, .17), t)
		for side := float32(-1); side <= 1; side += 2 {
			curveLine(p,
				right-200, bottom-2,
				right-150, bottom-6,
				right-120, bottom-60,
				right-65+side*30, bottom-64,
				fade(ui.RGBA(197, 222, 164, .10), 1), .7)
		}
	case visual.OrnamentTopHat:
		topHat(p, right-72, bottom-46, 40, ui.RGBA(195, 222, 241, .16), t)
		for j := 0; j < 3; j++ {
			curveLine(p,
				right-270, bottom-1-float32(j)*10,
				right-160, bottom+24-float32(j)*12,
				right-170, bottom-56-float32(j)*12,
				right-25, bottom-22-float32(j)*8,
				ui.RGBA(164, 213, 243, .09), .7)
		}
	case visual.OrnamentPolyhedron:
		archiveBook(p, right-80, bottom-40, 40, ui.RGBA(227, 211, 163, .17), t)
		for j := 0; j < 3; j++ {
			x := right - 240 + float32(j)*21
			y := bottom - 15 - float32(j)*10
			p.StrokePath(polyPath(x, y, x+30, y-16), .7, ui.RGBA(173, 225, 203, .11))
			p.StrokePath(polyPath(x+30, y-16, x+60, y), .7, ui.RGBA(173, 225, 203, .11))
		}
	case visual.OrnamentStoneSeal:
		stoneSeal(p, right-75, bottom-42, 43, ui.RGBA(230, 184, 104, .19))
		for j := 0; j < 3; j++ {
			x := right - 260 + float32(j)*26
			y := bottom + 6
			p.StrokePath(polyPath(x, y, x, y-28, x+28, y-28, x+28, y-10), .8,
				ui.RGBA(222, 178, 98, .10))
		}
	}
}

// navOrnament 是导航选中项右侧的小标识。
func (a *App) navOrnament(p *ui.Painter, cx, cy, r float32) {
	accent := a.tok.Accent
	switch a.char.Ornament {
	case visual.OrnamentQingxin:
		qingxin(p, cx, cy, r, accent.Alpha(.5), 0)
	case visual.OrnamentGoldfish:
		firework(p, cx, cy, r*.9, accent.Alpha(.5), .2)
	case visual.OrnamentSakura:
		sakura(p, cx, cy, r, accent.Alpha(.5), 0)
	case visual.OrnamentSprig:
		sprig(p, cx, cy, r*.8, accent.Alpha(.5), 0)
	case visual.OrnamentTopHat:
		droplet(p, cx, cy, r*.8, accent.Alpha(.55))
	case visual.OrnamentPolyhedron:
		polyhedron(p, cx, cy, r*.95, accent.Alpha(.65), 0)
	case visual.OrnamentStoneSeal:
		qingxin(p, cx, cy, r, accent.Alpha(.5), 0)
	default:
		glyph(p, a.char.Motif, cx, cy, r*.88, accent, .85, 0)
	}
}

// cardOrnament 画主题卡片上的角色纹样（参考各角色 _ui.cpp 的 Card）。
func cardOrnament(p *ui.Painter, r ui.Rect, ch visual.Character, t float32) {
	right, bottom := r.X+r.W, r.Y+r.H
	switch ch.Ornament {
	case visual.OrnamentQingxin:
		windLines(p, r)
		qingxin(p, right-81, r.Y+57, 36, ui.RGBA(176, 223, 236, .35), -.3+.025*sinf(float64(t)*.25))
		chineseKnot(p, right-29, r.Y+12, 1, t)

	case visual.OrnamentGoldfish:
		firework(p, right-53, r.Y+43, 27, ui.RGBA(255, 214, 144, .35), .15+.02*sinf(float64(t)*.25))
		goldfish(p, right-61, r.Y+84, 29, ui.RGBA(255, 189, 118, .37), t)
		for j := 0; j < 2; j++ {
			y := r.Y + 88 + float32(j)*6
			curveLine(p, r.X+14, y, r.X+r.W*.35, y-12, r.X+r.W*.6, y+20, right-14, y+2,
				ui.RGBA(255, 198, 130, .18), .8)
		}

	case visual.OrnamentSakura:
		sakura(p, right-83, r.Y+56, 35, ui.RGBA(239, 181, 214, .41), .1+.025*sinf(float64(t)*.25))
		sakura(p, right-125, r.Y+27, 10, ui.RGBA(227, 193, 149, .47), -.2)
		for j := 0; j < 3; j++ {
			y := r.Y + 90 - float32(j)*8
			curveLine(p, r.X+12, y, r.X+r.W*.4, y-39, r.X+r.W*.7, y+8, right-12, y-21,
				ui.RGBA(229, 168, 207, float32(30+j*10)/255), .8)
		}
		p.StrokePath(polyPath(right-25, r.Y+17, right-25, r.Y+66), .8, ui.RGBA(224, 188, 142, .72))
		charm := &ui.Path{}
		charm.MoveTo(right-25, r.Y+40)
		charm.LineTo(right-19, r.Y+49)
		charm.LineTo(right-25, r.Y+58)
		charm.LineTo(right-31, r.Y+49)
		charm.Close()
		p.StrokePath(charm, .8, ui.RGBA(224, 188, 142, .72))
		p.StrokePath(polyPath(right-28, r.Y+66, right-22, r.Y+72, right-28, r.Y+78), 1,
			ui.RGBA(224, 188, 142, .72))

	case visual.OrnamentSprig:
		sprig(p, right-35, r.Y+62, 31, ui.RGBA(211, 231, 161, .43), t)
		curveLine(p, r.X+14, r.Y+89, r.X+r.W*.4, r.Y+58, r.X+r.W*.7, r.Y+103, right-14, r.Y+82,
			ui.RGBA(219, 230, 163, .20), .8)

	case visual.OrnamentTopHat:
		topHat(p, right-33, r.Y+65, 27, ui.RGBA(192, 221, 243, .45), t)
		for j := 0; j < 3; j++ {
			y := r.Y + 86 + float32(j)*6
			curveLine(p, r.X+12, y, r.X+r.W*.4, y-20, r.X+r.W*.6, y+22, right-12, y-4,
				ui.RGBA(162, 217, 248, float32(32+j*9)/255), .8)
		}

	case visual.OrnamentPolyhedron:
		polyhedron(p, right-29, r.Y+42, 23, ui.RGBA(233, 215, 169, .39), t)
		archiveBook(p, right-33, r.Y+84, 21, ui.RGBA(179, 222, 201, .25), t)

	case visual.OrnamentStoneSeal:
		stoneSeal(p, right-35, r.Y+61, 36, ui.RGBA(249, 205, 121, .41))
		for j := 0; j < 3; j++ {
			x := r.X + 20 + float32(j)*28
			y := r.Y + 91
			p.StrokePath(polyPath(x, y, x, y-11, x+16, y-11, x+16, y-3, x+25, y-3), .8,
				ui.RGBA(235, 193, 113, .22))
		}
	}
	_ = bottom
}

// curveLine 用三次贝塞尔画一条装饰线。
func curveLine(p *ui.Painter, x0, y0, c1x, c1y, c2x, c2y, x1, y1 float32, col ui.Color, w float32) {
	var path ui.Path
	path.MoveTo(x0, y0)
	path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&path, w, col)
}

// waveLine 画一条轻微的波浪线（用于 Chrome 底部的长装饰线）。
func waveLine(p *ui.Painter, x, y, w, amp, phase float32, col ui.Color, width float32) {
	const steps = 18
	var path ui.Path
	for i := 0; i <= steps; i++ {
		t := float32(i) / steps
		px := x + t*w
		py := y + amp*float32(math.Sin(float64(t)*math.Pi*2+float64(phase)))
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	p.StrokePath(&path, width, col)
}
