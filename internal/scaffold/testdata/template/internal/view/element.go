package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

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

	case theme.MotifLeaf:
		// 草：三片尖叶自中心呈扇形展开 + 向下的短茎。
		// 叶形就是「两端尖的梭形」——petal 的两条对称二次曲线。
		for _, ang := range []float64{-math.Pi / 2, -math.Pi/2 - 0.85, -math.Pi/2 + 0.85} {
			petal(p, x, y, float32(ang), rad*0.95, rad*0.26, w, line)
		}
		// 茎与基点：让叶丛看起来是「长」出来的，而不是浮在原地。
		seg(p, x, y, x, y+rad*0.42, w*0.8, line)
		circle(p, x, y+rad*0.50, w*0.9, line)

	case theme.MotifWave:
		// 水：一滴水珠（上尖下圆）+ 底下一道波浪。
		topY := y - rad*0.55
		cy := y - rad*0.10
		rr := rad * 0.46
		// 轮廓：左右两条对称曲线，落到圆弧两端。
		arcL := math.Asin(0.55) // 弧端点的仰角
		ex := rr * float32(math.Cos(arcL))
		ey := cy + rr*float32(math.Sin(arcL))
		quadLine(p, x, topY, x-rr*0.95, cy-rad*0.28, x-ex, ey, w, line)
		quadLine(p, x, topY, x+rr*0.95, cy-rad*0.28, x+ex, ey, w, line)
		arcLine(p, x, cy, rr, arcL, math.Pi-arcL, w, line)
		// 内部一点 + 底下的波浪。
		circle(p, x, y-rad*0.06, rad*0.09, line)
		wy := y + rad*0.62
		quadLine(p, x-rad*0.62, wy, x-rad*0.31, wy-rad*0.14, x, wy, w*0.8, line)
		quadLine(p, x, wy, x+rad*0.31, wy+rad*0.14, x+rad*0.62, wy, w*0.8, line)

	case theme.MotifWind:
		// 风：两道回旋气流盘向中心，带一条甩出去的尾迹。
		arcLine(p, x, y, rad*0.88, math.Pi*0.62, math.Pi*1.62, w, line)
		arcLine(p, x-rad*0.06, y+rad*0.04, rad*0.52, math.Pi*1.72, math.Pi*2.86, w*0.9, line)
		// 外弧起点的切向尾巴，给出「流动」的方向感。
		tx := x + rad*0.88*float32(math.Cos(math.Pi*0.62))
		ty := y + rad*0.88*float32(math.Sin(math.Pi*0.62))
		seg(p, tx, ty, tx-rad*0.28, ty+rad*0.10, w, line)
		circle(p, x, y, rad*0.10, line)

	default:
		// 冰晶：六条主臂 + 三对分叉 + 中心六边形 + 臂尖菱形。
		//
		// 原版的冰元素符号是个"厚"的六角雪花：靠粗线、分叉密度与中心镂空立住；
		// 只有主干加两点分叉会读成"米字"，小尺寸下尤其明显。
		rot := float64(spin) * math.Pi / 6
		for i := 0; i < 6; i++ {
			ang := rot + float64(i)*math.Pi/3
			ca, sa := float32(math.Cos(ang)), float32(math.Sin(ang))

			seg(p, x, y, x+rad*ca, y+rad*sa, w, line)

			for j := 1; j <= 3; j++ {
				f := 0.26 + 0.22*float64(j)
				bx := x + rad*float32(f)*ca
				by := y + rad*float32(f)*sa
				bl := rad * float32(0.30-0.07*float64(j))
				for _, sgn := range []float64{-1, 1} {
					ba := ang + sgn*(math.Pi/3)
					seg(p, bx, by,
						bx+bl*float32(math.Cos(ba)), by+bl*float32(math.Sin(ba)), w*0.62, line)
				}
			}

			// 臂尖小菱形
			tx, ty := x+rad*ca, y+rad*sa
			tr := rad * 0.12
			ax, ay := tx+ca*tr*0.8, ty+sa*tr*0.8
			bx2, by2 := tx-sa*tr*0.5, ty+ca*tr*0.5
			cx2, cy2 := tx-ca*tr*0.4, ty-sa*tr*0.4
			dx2, dy2 := tx+sa*tr*0.5, ty-ca*tr*0.5
			seg(p, ax, ay, bx2, by2, w*0.55, line)
			seg(p, bx2, by2, cx2, cy2, w*0.55, line)
			seg(p, cx2, cy2, dx2, dy2, w*0.55, line)
			seg(p, dx2, dy2, ax, ay, w*0.55, line)
		}

		// 中心六边形，留出镂空感
		hr := rad * 0.26
		for i := 0; i < 6; i++ {
			a0 := rot - math.Pi/2 + float64(i)*math.Pi/3
			a1 := rot - math.Pi/2 + float64(i+1)*math.Pi/3
			seg(p,
				x+hr*float32(math.Cos(a0)), y+hr*float32(math.Sin(a0)),
				x+hr*float32(math.Cos(a1)), y+hr*float32(math.Sin(a1)),
				w*0.8, line)
		}
	}
}
