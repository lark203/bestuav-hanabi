package view

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/egoist/mygo/ui"
)

// 基础绘制辅助：多处共用（开场、窗饰、图标、粒子），所以不放在 intro.go 里——
// 生成器 -no-intro 删除开场文件时，这些仍要保留。

func sinf(v float64) float32 { return float32(math.Sin(v)) }

// tau 是整圆（多处几何计算共用）。
const tau = 2 * math.Pi

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
