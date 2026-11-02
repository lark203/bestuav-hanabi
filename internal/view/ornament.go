package view

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// 角色专属纹样（参考项目 ui/theme_ornaments.cpp）。
//
// 全部用 ui.Path + Painter.FillPath/StrokePath 画：任意角度的折线、贝塞尔和
// 闭合多边形都能直接描边/填充，不用再像早期那样拿 Transform + Fill 拼胶囊
// （那样每段都有圆头，折点会鼓起来）。纹样只负责几何，颜色由调用方给。

// qingxin 画一朵六瓣清心纹：花瓣是两段三次贝塞尔围成的尖瓣，中心一个小圆。
//
// 参考 ornament::Flower：六瓣按 60° 展开，每瓣从 0.12r 起笔、在 1.2r 处
// 收尖再回来——这是一朵"线描"的花，不是实心花瓣，所以只描边不填充。
func qingxin(p *ui.Painter, cx, cy, r float32, col ui.Color, rot float32) {
	if r <= 0.5 {
		return
	}
	var path ui.Path
	for i := 0; i < 6; i++ {
		a := float64(rot) + float64(i)*math.Pi/3
		ca, sa := float32(math.Cos(a)), float32(math.Sin(a))
		pt := func(x, y float32) (float32, float32) {
			return cx + (ca*x-sa*y)*r, cy + (sa*x+ca*y)*r
		}
		x0, y0 := pt(.12, 0)
		c1x, c1y := pt(.65, -.44)
		c2x, c2y := pt(1.2, -.23)
		x1, y1 := pt(1, 0)
		c3x, c3y := pt(1.2, .23)
		c4x, c4y := pt(.65, .44)
		path.MoveTo(x0, y0)
		path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		path.CubeTo(c3x, c3y, c4x, c4y, x0, y0)
	}
	p.StrokePath(&path, 1, col)

	var core ui.Path
	core.Circle(cx, cy, r*.12)
	p.StrokePath(&core, 1, col)
}

// windLines 画卡片上的三条风纹（参考 ganyu_ui.cpp Card 里的三根贝塞尔）。
//
// 线条从左缘横贯到右缘，越往下越亮一点，像风从卡片下面扫过去。
func windLines(p *ui.Painter, r ui.Rect) {
	for j := 0; j < 3; j++ {
		yy := r.Y + 90 - float32(j)*11
		var path ui.Path
		path.MoveTo(r.X+12, yy)
		path.CubeTo(
			r.X+r.W*.40, yy-24,
			r.X+r.W*.70, yy+15,
			r.X+r.W-12, yy-15,
		)
		col := ui.RGBA(156, 211, 233, float32(30+j*8)/255)
		p.StrokePath(&path, 1, col)
	}
}

// chineseKnot 画一枚淡金中国结（参考 ornament::ChineseKnot）。
//
// topX/topY 是结顶（挂绳的起点），scale 缩放，t 让结与流苏轻微摇摆。
// 参考项目在卡片右上角用它，作为甘雨的中国结。
func chineseKnot(p *ui.Painter, topX, topY, scale, t float32) {
	if scale <= 0 {
		return
	}
	shade := ui.RGBA(65, 27, 49, .96)
	silk := ui.RGBA(197, 67, 90, 1)
	shine := ui.RGBA(250, 149, 144, .90)
	gold := ui.RGBA(229, 195, 133, .94)

	sway := sinf(float64(t)*0.75) * 0.018
	cs, sn := float32(math.Cos(float64(sway))), float32(math.Sin(float64(sway)))
	// P 以结顶为原点，绕 (0,13) 轻微旋转——和参考的 p() 一致。
	P := func(x, y float32) (float32, float32) {
		return topX + (x*cs-(y-13)*sn)*scale, topY + (13+x*sn+(y-13)*cs)*scale
	}
	// rope 先画一层更宽的深色底，再画丝线本体与一道偏左上的高光：
	// 细线在深色卡片上才有立体感（参考 theme_ornaments.cpp 的 rope）。
	rope := func(path *ui.Path, w float32, col ui.Color) {
		p.StrokePath(path, (w+1.25)*scale, shade)
		p.StrokePath(path, w*scale, col)
		p.Transform(ui.Translation(-.35*scale, -.32*scale), 0, 0, func() {
			p.StrokePath(path, .55*scale, shine)
		})
	}

	// 挂绳：两根金色细绳从顶部汇到珠子。
	var cord ui.Path
	cx0, cy0 := P(-1, 0)
	c1x, c1y := P(-2, 6)
	c2x, c2y := P(-2, 10)
	cx1, cy1 := P(0, 14)
	cord.MoveTo(cx0, cy0)
	cord.CubeTo(c1x, c1y, c2x, c2y, cx1, cy1)
	p.StrokePath(&cord, 1.15*scale, gold)

	var cord2 ui.Path
	cx0, cy0 = P(1, 0)
	c1x, c1y = P(2, 6)
	c2x, c2y = P(2, 10)
	cord2.MoveTo(cx0, cy0)
	cord2.CubeTo(c1x, c1y, c2x, c2y, cx1, cy1)
	p.StrokePath(&cord2, .7*scale, gold)

	bx, by := P(0, 15)
	circle(p, bx, by, 2.2*scale, gold)

	// 结体：两组圆角方环在 45° 上交错，是盘长结的基本骨架。
	q := func(x, y float32) (float32, float32) {
		return P((x-y)*.70710678, 31+(x+y)*.70710678)
	}
	for axis := 0; axis < 2; axis++ {
		v := func(x, y float32) (float32, float32) {
			if axis == 1 {
				x, y = y, x
			}
			return q(x, y)
		}
		var path ui.Path
		ax, ay := v(-7, -6)
		path.MoveTo(ax, ay)
		ax, ay = v(7, -6)
		path.LineTo(ax, ay)
		c1x, c1y = v(13, -6)
		c2x, c2y = v(13, 0)
		ax, ay = v(7, 0)
		path.CubeTo(c1x, c1y, c2x, c2y, ax, ay)
		ax, ay = v(-7, 0)
		path.LineTo(ax, ay)
		c1x, c1y = v(-13, 0)
		c2x, c2y = v(-13, 6)
		ax, ay = v(-7, 6)
		path.CubeTo(c1x, c1y, c2x, c2y, ax, ay)
		ax, ay = v(7, 6)
		path.LineTo(ax, ay)
		rope(&path, 2.3, silk)
	}

	// 3×3 的斜向编绳：交错方向，读起来才是"编"出来的。
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			x := -6 + float32(col)*6
			y := -6 + float32(row)*6
			var ax, ay, bx2, by2 float32
			if (row+col)%2 == 1 {
				ax, ay = q(x-2, y)
				bx2, by2 = q(x+2, y)
			} else {
				ax, ay = q(x, y-2)
				bx2, by2 = q(x, y+2)
			}
			var path ui.Path
			path.MoveTo(ax, ay)
			path.LineTo(bx2, by2)
			rope(&path, 2.2, silk)
		}
	}

	// 底珠、金箍与九根流苏。
	bx, by = P(0, 47)
	circle(p, bx, by, 2.4*scale, silk)
	rx0, ry0 := P(-2.5, 49)
	rx1, ry1 := P(2.5, 53)
	if rx1 < rx0 {
		rx0, rx1 = rx1, rx0
	}
	if ry1 < ry0 {
		ry0, ry1 = ry1, ry0
	}
	p.Fill(ui.Rect{X: rx0, Y: ry0, W: rx1 - rx0, H: ry1 - ry0}, gold, scale)

	for i := -4; i <= 4; i++ {
		x := float32(i) * .85
		fi := float32(i)
		if fi < 0 {
			fi = -fi
		}
		tip := 72 - fi*.65
		drift := sinf(float64(t)*0.9+float64(i)*0.12) * .5
		x0, y0 := P(x*.5, 53)
		x1, y1 := P(x, 59)
		x2, y2 := P(x+drift, 66)
		x3, y3 := P(x+drift, tip)
		var path ui.Path
		path.MoveTo(x0, y0)
		path.CubeTo(x1, y1, x2, y2, x3, y3)
		col := silk
		if (i+4)%3 == 0 {
			col = shine
		}
		p.StrokePath(&path, .85*scale, col)
	}

	gx0, gy0 := P(-3.1, 55)
	gx1, gy1 := P(3.1, 55)
	var band ui.Path
	band.MoveTo(gx0, gy0)
	band.LineTo(gx1, gy1)
	p.StrokePath(&band, .75*scale, gold)
}

// octagonPoints 返回切角八边形的 8 个顶点（顺时针，从左上角开始）。
//
// cut 是每个角切掉的长度；half 是中心到边的距离。
func octagonPoints(cx, cy, half, cut float32) [8][2]float32 {
	return [8][2]float32{
		{cx - half + cut, cy - half},
		{cx + half - cut, cy - half},
		{cx + half, cy - half + cut},
		{cx + half, cy + half - cut},
		{cx + half - cut, cy + half},
		{cx - half + cut, cy + half},
		{cx - half, cy + half - cut},
		{cx - half, cy - half + cut},
	}
}

// octagonFrom 把 8 个顶点连成闭合路径。
func octagonFrom(pts [8][2]float32) *ui.Path {
	path := &ui.Path{}
	path.MoveTo(pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i][0], pts[i][1])
	}
	path.Close()
	return path
}
