package view

import (
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 各角色的窗口按钮（参考 characters/*/\*_ui.cpp 的 WindowControl）。
//
// 共同结构：一个内缩的按钮面 + 一条装饰线 + 最小化/最大化/关闭三个图标；
// 形状、圆角/切角、冷暖金线和图标细节按角色不同。悬停统一用 hv 0..1 插值。

type chromeLook struct {
	octagon bool
	radius  float32
	cut     float32
	fill    ui.Color
	stroke  ui.Color
	line    ui.Color
	accent  ui.Color
	hot     ui.Color
	// 顶部装饰：0 无 / 1 横线 / 2 弧线 / 3 三角 / 4 底横线
	top   int
	min   int // 0 横线 / 1 弧线 / 2 弧线+点 / 3 横线+点
	max   int // 0 方框 / 1 方框+线 / 2 菱形 / 3 方框+点 / 4 方框+斜线
	close int // 0 叉 / 1 叉+点 / 2 曲线叉
}

func chromeLookFor(style visual.Chrome, hv float32) chromeLook {
	switch style {
	case visual.ChromePyro:
		return chromeLook{
			radius: 6,
			fill:   ui.RGBA(76, 35, 23, .39+.25*hv),
			stroke: ui.RGBA(255, 202, 132, .25+.47*hv),
			line:   ui.RGBA(255, 220, 164, .63+.31*hv),
			accent: ui.RGBA(255, 209, 155, 1),
			hot:    ui.RGBA(255, 154, 117, 1),
			top:    1, min: 1, max: 1, close: 1,
		}
	case visual.ChromeElectro:
		return chromeLook{
			radius: 7,
			fill:   ui.RGBA(69, 33, 66, .31+.23*hv),
			stroke: ui.RGBA(236, 177, 215, .20+.43*hv),
			line:   ui.RGBA(227, 193, 146, .49+.35*hv),
			accent: ui.RGBA(242, 184, 223, 1),
			hot:    ui.RGBA(255, 160, 193, 1),
			top:    2, min: 3, max: 2, close: 2,
		}
	case visual.ChromeDendro:
		return chromeLook{
			radius: 10,
			fill:   ui.RGBA(17, 52, 35, .61+.25*hv),
			stroke: ui.RGBA(210, 229, 165, .37+.39*hv),
			line:   ui.RGBA(225, 221, 158, .49+.35*hv),
			accent: ui.RGBA(195, 237, 166, 1),
			hot:    ui.RGBA(255, 203, 180, 1),
			top:    2, min: 1, max: 3, close: 0,
		}
	case visual.ChromeHydro:
		return chromeLook{
			radius: 7,
			fill:   ui.RGBA(18, 42, 78, .49+.27*hv),
			stroke: ui.RGBA(191, 221, 244, .33+.43*hv),
			line:   ui.RGBA(224, 233, 228, .53+.35*hv),
			accent: ui.RGBA(166, 226, 250, 1),
			hot:    ui.RGBA(251, 186, 207, 1),
			top:    2, min: 2, max: 1, close: 0,
		}
	case visual.ChromeAnemo:
		return chromeLook{
			octagon: true, cut: 4,
			fill:   ui.RGBA(16, 58, 62, .73+.20*hv),
			stroke: ui.RGBA(232, 210, 156, .49+.41*hv),
			line:   ui.RGBA(232, 210, 156, .49+.41*hv),
			accent: ui.RGBA(177, 240, 214, 1),
			hot:    ui.RGBA(255, 197, 163, 1),
			top:    3, min: 0, max: 4, close: 0,
		}
	case visual.ChromeGeo:
		return chromeLook{
			octagon: true, cut: 4,
			fill:   ui.RGBA(42, 32, 20, .69+.25*hv),
			stroke: ui.RGBA(241, 196, 117, .39+.51*hv),
			line:   ui.RGBA(241, 196, 117, .39+.51*hv),
			accent: ui.RGBA(245, 201, 105, 1),
			hot:    ui.RGBA(255, 185, 131, 1),
			top:    4, min: 0, max: 2, close: 0,
		}
	}
	return chromeLook{}
}

// chromeButton 按主题的 Chrome 分派；Plain 走调用方原来的 SVG 图标。
func (a *App) chromeButton(p *ui.Painter, r ui.Rect, idx int, hv float32) {
	if a.char.Chrome == visual.ChromeCrystal {
		a.crystalButton(p, r, idx, hv)
		return
	}
	look := chromeLookFor(a.char.Chrome, hv)
	if look.accent.A == 0 {
		a.crystalButton(p, r, idx, hv)
		return
	}
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	box := inset(r, 6, 4)

	if look.octagon {
		path := cutRect(box, look.cut)
		p.FillPath(path, look.fill)
		p.StrokePath(path, .85, look.stroke)
	} else {
		p.Fill(box, look.fill, look.radius)
		p.Stroke(box, look.stroke, look.radius, .8)
	}

	// 顶部（或底部）装饰。
	switch look.top {
	case 1:
		x0, y0 := box.X+5, box.Y+2
		x1, y1 := box.X+box.W-5, box.Y+2
		p.StrokePath(polyPath(x0, y0, x1, y1), .7, look.line)
	case 2:
		arc := &ui.Path{}
		arc.MoveTo(box.X+3, box.Y+box.H*.20)
		arc.CubeTo(box.X+box.W*.30, box.Y+box.H*.03,
			box.X+box.W*.70, box.Y+box.H*.03,
			box.X+box.W-3, box.Y+box.H*.20)
		p.StrokePath(arc, .8, look.line)
	case 3:
		tri := &ui.Path{}
		tri.MoveTo(cx-2, box.Y+1)
		tri.LineTo(cx+2, box.Y+1)
		tri.LineTo(cx, box.Y+4)
		tri.Close()
		p.FillPath(tri, look.line)
	case 4:
		x0, y0 := box.X+5, box.Y+box.H-2
		x1, y1 := box.X+box.W-5, box.Y+box.H-2
		p.StrokePath(polyPath(x0, y0, x1, y1), .7, look.line)
	}

	col := look.accent.Mix(look.hot, hv)
	z := float32(4.5)

	// 最小化。
	if idx == 0 {
		switch look.min {
		case 1, 2:
			curve := &ui.Path{}
			curve.MoveTo(cx-z, cy)
			curve.CubeTo(cx-z*.3, cy+2, cx+z*.3, cy+2, cx+z, cy)
			p.StrokePath(curve, 1.2, col)
			if look.min == 2 {
				circle(p, cx, cy+4.2, .7, look.line)
			}
		case 3:
			p.StrokePath(polyPath(cx-z, cy, cx+z, cy), 1.2, col)
			circle(p, cx, cy+4.2, .7, look.line)
		default:
			p.StrokePath(polyPath(cx-z, cy, cx+z, cy), 1.2, col)
		}
	}

	// 最大化。
	if idx == 1 {
		switch look.max {
		case 1:
			p.Stroke(ui.Rect{X: cx - z, Y: cy - z, W: z * 2, H: z * 2}, col, 1, 1.1)
			p.StrokePath(polyPath(cx-z, cy-z-2, cx+z, cy-z-2), .8, look.line)
		case 2:
			d := &ui.Path{}
			d.MoveTo(cx, cy-z-1)
			d.LineTo(cx+z+1, cy)
			d.LineTo(cx, cy+z+1)
			d.LineTo(cx-z-1, cy)
			d.Close()
			p.StrokePath(d, 1.1, col)
		case 3:
			p.Stroke(ui.Rect{X: cx - z, Y: cy - z, W: z * 2, H: z * 2}, col, 2, 1.1)
			circle(p, cx, cy, .8, col)
		case 4:
			p.Stroke(ui.Rect{X: cx - z, Y: cy - z, W: z * 2, H: z * 2}, col, .4, 1.1)
			p.StrokePath(polyPath(cx-z, cy-z, cx+z, cy+z), .65, fade(look.line, .5))
		default:
			p.Stroke(ui.Rect{X: cx - z, Y: cy - z, W: z * 2, H: z * 2}, col, 1, 1.1)
		}
	}

	// 关闭。
	if idx == 2 {
		switch look.close {
		case 1:
			p.StrokePath(polyPath(cx-z, cy-z, cx+z, cy+z), 1.2, col)
			p.StrokePath(polyPath(cx-z, cy+z, cx+z, cy-z), 1.2, col)
			circle(p, cx, cy, 1, look.line)
		case 2:
			curve := &ui.Path{}
			curve.MoveTo(cx-z, cy-z)
			curve.CubeTo(cx-z*.2, cy-z*.2, cx+z*.2, cy+z*.2, cx+z, cy+z)
			p.StrokePath(curve, 1.2, col)
			p.StrokePath(polyPath(cx-z, cy+z, cx+z, cy-z), 1.2, col)
		default:
			p.StrokePath(polyPath(cx-z, cy-z, cx+z, cy+z), 1.2, col)
			p.StrokePath(polyPath(cx-z, cy+z, cx+z, cy-z), 1.2, col)
		}
	}
}

func inset(r ui.Rect, dx, dy float32) ui.Rect {
	return ui.Rect{X: r.X + dx, Y: r.Y + dy, W: r.W - dx*2, H: r.H - dy*2}
}

func cutRect(r ui.Rect, cut float32) *ui.Path {
	pts := [8][2]float32{
		{r.X + cut, r.Y},
		{r.X + r.W - cut, r.Y},
		{r.X + r.W, r.Y + cut},
		{r.X + r.W, r.Y + r.H - cut},
		{r.X + r.W - cut, r.Y + r.H},
		{r.X + cut, r.Y + r.H},
		{r.X, r.Y + r.H - cut},
		{r.X, r.Y + cut},
	}
	path := &ui.Path{}
	path.MoveTo(pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i][0], pts[i][1])
	}
	path.Close()
	return path
}
