package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

// crystalArm 画一片「刃状主臂」：宽基、尖锋，根部带一对后掠倒刺——
// 参考原神冰符号的厚刃六角星。三段线宽递减叠出刃的锥度。
// grow 0..1 控制生长（开场用），静态调用传 1。
func crystalArm(p *ui.Painter, x, y, ang, rad, grow, wBase float32, line ui.Color) {
	ca := float32(math.Cos(float64(ang)))
	sa := float32(math.Sin(float64(ang)))
	l1 := rad * 0.30 * grow
	l2 := rad * 0.62 * grow
	l3 := rad * grow
	seg(p, x, y, x+l1*ca, y+l1*sa, wBase, line)
	if grow > 0.34 {
		seg(p, x+l1*ca, y+l1*sa, x+l2*ca, y+l2*sa, wBase*0.62, line)
	}
	if grow > 0.67 {
		seg(p, x+l2*ca, y+l2*sa, x+l3*ca, y+l3*sa, wBase*0.36, line)
	}
	if grow > 0.34 {
		bx, by := x+l1*ca, y+l1*sa
		bl := rad * 0.26 * grow
		for _, sgn := range []float64{-1, 1} {
			ba := float64(ang) + sgn*0.62
			seg(p, bx, by,
				bx+bl*float32(math.Cos(ba)), by+bl*float32(math.Sin(ba)),
				wBase*0.55, line)
		}
	}
}

// crystalCore 画冰徽记的中心六角核。
func crystalCore(p *ui.Painter, x, y, rot, rad float32, w float32, line ui.Color) {
	hr := rad * 0.18
	for i := 0; i < 6; i++ {
		a0 := float64(rot) - math.Pi/2 + float64(i)*math.Pi/3
		a1 := float64(rot) - math.Pi/2 + float64(i+1)*math.Pi/3
		seg(p,
			x+hr*float32(math.Cos(a0)), y+hr*float32(math.Sin(a0)),
			x+hr*float32(math.Cos(a1)), y+hr*float32(math.Sin(a1)), w, line)
	}
}

// glyph 画元素标识的静态版本：品牌图标、侧栏底部、主题卡片的左上角都用它。
//
// 开场动画里的标识（draw*Mark）带生长与旋转，这里不需要时间参数——
// 同一个 motif 在界面上多处出现，几何必须一致，所以共用一套画法。
// spin 只影响晶体的侧枝朝向（0 与 1 相差 60°，用于让卡片图标不完全一样）。
func glyph(p *ui.Painter, m theme.Motif, x, y, rad float32, accent ui.Color, alpha, spin float32) {
	if alpha <= 0.004 || rad <= 0.5 {
		return
	}
	line := accent.Mix(ui.Hex("#FFFFFF"), 0.62).Alpha(alpha * 0.92)
	w := max32(1.4, rad*0.09)

	switch m {
	case theme.MotifFlame:
		// 八条火舌 + 火星。
		for i := 0; i < 8; i++ {
			ang := float64(i) * tau / 8
			ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
			l := rad * 0.86
			seg(p, x+ca*rad*0.30, y+sa*rad*0.30, x+ca*l, y+sa*l, w, line)
			circle(p, x+ca*l, y+sa*l, w*0.9, line)
		}
		circle(p, x, y, rad*0.26, accent.Alpha(alpha*0.9))

	case theme.MotifBolt:
		// 四道折线闪电。
		for i := 0; i < 4; i++ {
			ang := float64(i)*math.Pi/2 + math.Pi/4 + float64(spin)*math.Pi/8
			ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
			tx, ty := -sa, ca
			px, py := x, y
			const n = 3
			for k := 1; k <= n; k++ {
				f := float32(k) / n
				j := sinf(float64(i)*2.3+float64(k)*1.9) * rad * 0.13 * (1 - f)
				nx := x + ca*rad*f + tx*j
				ny := y + sa*rad*f + ty*j
				seg(p, px, py, nx, ny, w, line)
				px, py = nx, ny
			}
			circle(p, px, py, w*0.9, line)
		}
		circle(p, x, y, rad*0.22, accent.Alpha(alpha*0.9))

	case theme.MotifGeo:
		// 正六边形 + 内嵌菱形。
		hexR := rad * 0.92
		for i := 0; i < 6; i++ {
			a0 := -math.Pi/2 + float64(i)*math.Pi/3
			a1 := -math.Pi/2 + float64(i+1)*math.Pi/3
			seg(p,
				x+hexR*float32(math.Cos(a0)), y+hexR*float32(math.Sin(a0)),
				x+hexR*float32(math.Cos(a1)), y+hexR*float32(math.Sin(a1)), w, line)
		}
		dm := rad * 0.46
		for i := 0; i < 4; i++ {
			a0 := math.Pi/4 + float64(i)*math.Pi/2
			a1 := math.Pi/4 + float64(i+1)*math.Pi/2
			seg(p,
				x+dm*float32(math.Cos(a0)), y+dm*float32(math.Sin(a0)),
				x+dm*float32(math.Cos(a1)), y+dm*float32(math.Sin(a1)), w*0.85, line)
		}
		circle(p, x, y, rad*0.12, line)

	default:
		// 冰元素徽记：六片刃状主臂 + 臂间短刺 + 中心六角核。
		// 参考原神冰符号——是「厚刃的六角星」，不是细线雪花：
		// 三段收窄的线宽叠出锥度，根部后掠倒刺构成箭镞轮廓。
		rot := float64(spin) * math.Pi / 6
		wBase := max32(1.6, rad*0.16)
		for i := 0; i < 6; i++ {
			crystalArm(p, x, y, float32(rot+float64(i)*math.Pi/3), rad, 1, wBase, line)
		}
		// 臂间短刺：补齐六角星的锯齿轮廓
		for i := 0; i < 6; i++ {
			ang := rot + (float64(i)+0.5)*math.Pi/3
			ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))
			seg(p, x+rad*0.24*ca, y+rad*0.24*sa, x+rad*0.50*ca, y+rad*0.50*sa, wBase*0.45, line)
		}
		crystalCore(p, x, y, float32(rot), rad, wBase*0.7, line)
	}
}
