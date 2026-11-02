package view

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// 角色纹样的几何原语（参考项目 ui/theme_ornaments.cpp）。
//
// 每个函数只画形状、不决定颜色；调用方按角色给色与透明度。坐标用 r 归一化，
// 和参考项目的 cpp 保持一致，方便逐项对照。

// fade 按比例缩放颜色已有的 alpha（Color.A 是 0..255）。
func fade(c ui.Color, k float32) ui.Color {
	return c.Alpha(float32(c.A) / 255 * k)
}

// polyPath 用扁平坐标（x0,y0,x1,y1,…）连一条折线。
func polyPath(pts ...float32) *ui.Path {
	path := &ui.Path{}
	if len(pts) < 2 {
		return path
	}
	path.MoveTo(pts[0], pts[1])
	for i := 2; i+1 < len(pts); i += 2 {
		path.LineTo(pts[i], pts[i+1])
	}
	return path
}

// polyPoints 用 [2]float32 顶点连一条折线。
func polyPoints(pts ...[2]float32) *ui.Path {
	path := &ui.Path{}
	if len(pts) == 0 {
		return path
	}
	path.MoveTo(pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i][0], pts[i][1])
	}
	return path
}

// sakura 画五瓣樱花：每瓣三段三次贝塞尔，中心一个小圆。
func sakura(p *ui.Painter, cx, cy, r float32, col ui.Color, rot float32) {
	if r <= 0.5 {
		return
	}
	for i := 0; i < 5; i++ {
		a := float64(rot) + float64(i)*math.Pi*.4 - math.Pi*.5
		ca, sa := float32(math.Cos(a)), float32(math.Sin(a))
		pt := func(x, y float32) (float32, float32) {
			return cx + (ca*x-sa*y)*r, cy + (sa*x+ca*y)*r
		}
		var path ui.Path
		x0, y0 := pt(.13, 0)
		path.MoveTo(x0, y0)
		c1x, c1y := pt(.65, -.49)
		c2x, c2y := pt(1.06, -.38)
		x1, y1 := pt(.93, -.03)
		path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		c1x, c1y = pt(.78, -.03)
		c2x, c2y = pt(.80, .05)
		x1, y1 = pt(1, .11)
		path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		c1x, c1y = pt(.96, .50)
		c2x, c2y = pt(.40, .28)
		x1, y1 = pt(.13, 0)
		path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		p.StrokePath(&path, .9, col)
	}
	var core ui.Path
	core.Circle(cx, cy, r*.09)
	p.StrokePath(&core, .8, col)
}

// firework 画十二道放射火舌：长短交错，末端一点火星，中心实心圆。
func firework(p *ui.Painter, cx, cy, r float32, col ui.Color, rot float32) {
	if r <= 0.5 {
		return
	}
	for i := 0; i < 12; i++ {
		a := float64(rot) + float64(i)*math.Pi/6
		length := float32(1)
		if i%2 == 1 {
			length = .74
		}
		ca, sa := float32(math.Cos(a)), float32(math.Sin(a))
		pt := func(x, y float32) (float32, float32) {
			return cx + (ca*x-sa*y)*r, cy + (sa*x+ca*y)*r
		}
		var path ui.Path
		x0, y0 := pt(.23, 0)
		path.MoveTo(x0, y0)
		c1x, c1y := pt(.45, -.03)
		c2x, c2y := pt(length*.83, -.08)
		x1, y1 := pt(length, -.15)
		path.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		p.StrokePath(&path, 1, col)
		circle(p, x1, y1, max32(.7, r*.027), col)
	}
	circle(p, cx, cy, max32(.8, r*.07), col)
}

// goldfish 画一条金鱼：身体、尾巴、上鳍与眼睛，整体随时间轻微摆动。
func goldfish(p *ui.Painter, cx, cy, r float32, col ui.Color, t float32) {
	if r <= 0.5 {
		return
	}
	sway := sinf(float64(t)*.7) * .05
	cs, sn := float32(math.Cos(float64(sway))), float32(math.Sin(float64(sway)))
	q := func(x, y float32) (float32, float32) {
		return cx + (x*cs-y*sn)*r, cy + (x*sn+y*cs)*r
	}
	var body ui.Path
	x0, y0 := q(-.65, 0)
	body.MoveTo(x0, y0)
	c1x, c1y := q(-.60, -.58)
	c2x, c2y := q(.15, -.62)
	x1, y1 := q(.44, -.12)
	body.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(.10, .50)
	c2x, c2y = q(-.50, .55)
	x1, y1 = q(-.65, 0)
	body.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&body, 1.1, col)

	var tail ui.Path
	x0, y0 = q(.44, -.12)
	tail.MoveTo(x0, y0)
	c1x, c1y = q(.60, -.50)
	c2x, c2y = q(1.15, -.70)
	x1, y1 = q(.96, -.18)
	tail.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(1.26, .38)
	c2x, c2y = q(.65, .51)
	x1, y1 = q(.44, -.12)
	tail.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&tail, 1, col)

	var fin ui.Path
	x0, y0 = q(-.20, -.40)
	fin.MoveTo(x0, y0)
	c1x, c1y = q(.02, -.92)
	c2x, c2y = q(.38, -.68)
	x1, y1 = q(.18, -.35)
	fin.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&fin, .8, col)

	ex, ey := q(-.42, -.10)
	circle(p, ex, ey, r*.044, col)

	var mouth ui.Path
	x0, y0 = q(-.07, 0)
	mouth.MoveTo(x0, y0)
	c1x, c1y = q(.18, 0)
	c2x, c2y = q(.15, .40)
	x1, y1 = q(-.07, .26)
	mouth.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&mouth, .8, col)
}

// sprig 画一根带叶的枝条：主干 + 两对叶子 + 顶端芽，随时间轻摆。
func sprig(p *ui.Painter, cx, cy, r float32, col ui.Color, t float32) {
	if r <= 0.5 {
		return
	}
	angle := sinf(float64(t)*.5) * .035
	cs, sn := float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))
	q := func(x, y float32) (float32, float32) {
		return cx + (x*cs-y*sn)*r, cy + (x*sn+y*cs)*r
	}
	var stem ui.Path
	x0, y0 := q(0, .95)
	stem.MoveTo(x0, y0)
	c1x, c1y := q(.12, .25)
	c2x, c2y := q(-.10, -.4)
	x1, y1 := q(0, -1.05)
	stem.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&stem, 1, col)

	for _, side := range []float32{-1, 1} {
		for i := 0; i < 2; i++ {
			yy := .38 - float32(i)*.62
			var leaf ui.Path
			x0, y0 = q(0, yy)
			leaf.MoveTo(x0, y0)
			c1x, c1y = q(side*.15, yy-.52)
			c2x, c2y = q(side*.83, yy-.75)
			x1, y1 = q(side*.77, yy-.56)
			leaf.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
			c1x, c1y = q(side*.90, yy-.06)
			c2x, c2y = q(side*.18, yy+.06)
			x1, y1 = q(0, yy)
			leaf.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
			p.StrokePath(&leaf, .85, col)

			var vein ui.Path
			x0, y0 = q(0, yy)
			vein.MoveTo(x0, y0)
			c1x, c1y = q(side*.22, yy-.1)
			c2x, c2y = q(side*.52, yy-.25)
			x1, y1 = q(side*.77, yy-.56)
			vein.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
			p.StrokePath(&vein, .65, fade(col, .6))
		}
	}

	var bud ui.Path
	x0, y0 = q(0, -.62)
	bud.MoveTo(x0, y0)
	c1x, c1y = q(-.39, -.92)
	c2x, c2y = q(-.15, -1.25)
	x1, y1 = q(0, -1.36)
	bud.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(.15, -1.25)
	c2x, c2y = q(.39, -.92)
	x1, y1 = q(0, -.62)
	bud.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&bud, .85, col)
}

// droplet 画一滴水：上尖下圆，左下一条高光。
func droplet(p *ui.Painter, cx, cy, r float32, col ui.Color) {
	if r <= 0.5 {
		return
	}
	q := func(x, y float32) (float32, float32) { return cx + x*r, cy + y*r }
	var outline ui.Path
	x0, y0 := q(0, -1)
	outline.MoveTo(x0, y0)
	c1x, c1y := q(-.15, -.43)
	c2x, c2y := q(-.85, -.08)
	x1, y1 := q(-.62, .48)
	outline.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(-.33, 1.08)
	c2x, c2y = q(.33, 1.08)
	x1, y1 = q(.62, .48)
	outline.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(.85, -.08)
	c2x, c2y = q(.15, -.43)
	x1, y1 = q(0, -1)
	outline.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&outline, .9, col)

	var shine ui.Path
	x0, y0 = q(-.34, .20)
	shine.MoveTo(x0, y0)
	c1x, c1y = q(-.42, .47)
	c2x, c2y = q(-.20, .66)
	x1, y1 = q(-.06, .61)
	shine.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&shine, .8, fade(col, .7))
}

// topHat 画一顶礼帽：帽檐、帽身、帽带与右侧羽毛，稍微随时间摆动。
func topHat(p *ui.Painter, cx, cy, r float32, col ui.Color, t float32) {
	if r <= 0.5 {
		return
	}
	a := sinf(float64(t)*.4) * .018
	cs, sn := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	q := func(x, y float32) (float32, float32) {
		return cx + (x*cs-y*sn)*r, cy + (x*sn+y*cs)*r
	}
	q2 := func(x, y float32) [2]float32 {
		px, py := q(x, y)
		return [2]float32{px, py}
	}
	var crown ui.Path
	x0, y0 := q(-.60, -.68)
	crown.MoveTo(x0, y0)
	c1x, c1y := q(-.31, -.82)
	c2x, c2y := q(.30, -.82)
	x1, y1 := q(.60, -.68)
	crown.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&crown, 1, col)
	p.StrokePath(polyPoints(q2(-.60, -.68), q2(-.46, .27)), 1, col)
	p.StrokePath(polyPoints(q2(.60, -.68), q2(.46, .27)), 1, col)

	var brim ui.Path
	x0, y0 = q(-.46, .27)
	brim.MoveTo(x0, y0)
	c1x, c1y = q(-1.55, -.06)
	c2x, c2y = q(-1.22, .65)
	x1, y1 = q(0, .60)
	brim.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(1.22, .65)
	c2x, c2y = q(1.55, -.06)
	x1, y1 = q(.46, .27)
	brim.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&brim, 1, col)

	var band ui.Path
	x0, y0 = q(-.50, -.02)
	band.MoveTo(x0, y0)
	c1x, c1y = q(-.18, .12)
	c2x, c2y = q(.18, .12)
	x1, y1 = q(.50, -.02)
	band.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&band, 1.3, col)

	var feather ui.Path
	x0, y0 = q(.38, -.26)
	feather.MoveTo(x0, y0)
	c1x, c1y = q(.34, -1.05)
	c2x, c2y = q(.61, -1.45)
	x1, y1 = q(.99, -1.55)
	feather.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	c1x, c1y = q(1.21, -1.04)
	c2x, c2y = q(.85, -.62)
	x1, y1 = q(.38, -.26)
	feather.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
	p.StrokePath(&feather, .8, col)
	p.StrokePath(polyPoints(q2(.38, -.26), q2(.65, -.62), q2(.74, -1.12), q2(.99, -1.55)), .65, fade(col, .55))

	droplet(p, cx, cy+.11*r, r*.13, col)
}

// polyhedron 画一个轻微自转的棱晶：上下顶点 + 三个腰点。
func polyhedron(p *ui.Painter, cx, cy, r float32, col ui.Color, t float32) {
	if r <= 0.5 {
		return
	}
	top := [2]float32{cx, cy - r}
	bottom := [2]float32{cx, cy + r}
	var ring [3][2]float32
	for j := 0; j < 3; j++ {
		a := float64(t)*.23 + float64(j)*math.Pi*2/3
		ring[j] = [2]float32{cx + float32(math.Cos(a))*r*.84, cy + float32(math.Sin(a))*r*.22}
	}
	for j := 0; j < 3; j++ {
		p.StrokePath(polyPoints(top, ring[j]), .9, col)
		p.StrokePath(polyPoints(bottom, ring[j]), .9, col)
		p.StrokePath(polyPoints(ring[j], ring[(j+1)%3]), .75, fade(col, .55))
	}
	var facet ui.Path
	facet.MoveTo(top[0], top[1])
	facet.LineTo(ring[0][0], ring[0][1])
	facet.LineTo(ring[1][0], ring[1][1])
	facet.Close()
	p.FillPath(&facet, fade(col, .07))
}

// archiveBook 画一本摊开的古书：左右页、书脊、三条文字线，上方一个小棱晶。
func archiveBook(p *ui.Painter, cx, cy, r float32, col ui.Color, t float32) {
	if r <= 0.5 {
		return
	}
	wave := sinf(float64(t)*.55) * .035
	q := func(x, y float32) (float32, float32) {
		return cx + x*r, cy + (y+wave*x*x)*r
	}
	q2 := func(x, y float32) [2]float32 {
		px, py := q(x, y)
		return [2]float32{px, py}
	}
	for _, side := range []float32{-1, 1} {
		var page ui.Path
		x0, y0 := q(0, -.35)
		page.MoveTo(x0, y0)
		c1x, c1y := q(side*.3, -.65)
		c2x, c2y := q(side*.72, -.63)
		x1, y1 := q(side*1.04, -.45)
		page.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		page.LineTo(q(side*1.04, .60))
		c1x, c1y = q(side*.65, .40)
		c2x, c2y = q(side*.35, .43)
		x1, y1 = q(0, .70)
		page.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
		p.StrokePath(&page, 1, col)

		for j := 0; j < 3; j++ {
			var line ui.Path
			x0, y0 = q(side*.17, -.10+float32(j)*.20)
			line.MoveTo(x0, y0)
			c1x, c1y = q(side*.4, -.25+float32(j)*.20)
			c2x, c2y = q(side*.6, -.23+float32(j)*.20)
			x1, y1 = q(side*.86, -.10+float32(j)*.20)
			line.CubeTo(c1x, c1y, c2x, c2y, x1, y1)
			p.StrokePath(&line, .65, fade(col, .4))
		}
	}
	p.StrokePath(polyPoints(q2(0, -.35), q2(0, .70)), 1, col)
	_, budY := q(0, -1.25)
	polyhedron(p, cx, budY, r*.35, fade(col, .85), t)
}

// stoneSeal 画一枚岩印：六边形外框、内部回纹与底部山形。
func stoneSeal(p *ui.Painter, cx, cy, r float32, col ui.Color) {
	if r <= 0.5 {
		return
	}
	q := func(x, y float32) (float32, float32) { return cx + r*x, cy + r*y }
	q2 := func(x, y float32) [2]float32 {
		px, py := q(x, y)
		return [2]float32{px, py}
	}
	outer := [6][2]float32{
		q2(0, -1), q2(.72, -.28), q2(.72, .28), q2(0, 1), q2(-.72, .28), q2(-.72, -.28),
	}
	path := &ui.Path{}
	path.MoveTo(outer[0][0], outer[0][1])
	for i := 1; i < len(outer); i++ {
		path.LineTo(outer[i][0], outer[i][1])
	}
	path.Close()
	p.FillPath(path, fade(col, .06))
	p.StrokePath(path, .8, fade(col, .65))

	for _, side := range []float32{-1, 1} {
		pts := [][2]float32{
			q2(side*.42, -.30), q2(0, -.72), q2(-side*.42, -.30),
			q2(side*.17, .28), q2(side*.39, .06), q2(0, -.33),
		}
		path := &ui.Path{}
		path.MoveTo(pts[0][0], pts[0][1])
		for i := 1; i < len(pts); i++ {
			path.LineTo(pts[i][0], pts[i][1])
		}
		p.StrokePath(path, 1, col)
	}

	end := [3][2]float32{q2(-.28, .48), q2(0, .76), q2(.28, .48)}
	p.StrokePath(polyPoints(end[0], end[1], end[2]), .8, fade(col, .65))
}
