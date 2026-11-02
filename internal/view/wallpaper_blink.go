package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/theme"
)

// 眨眼：把 `.hbc` 里的闭眼帧做成一张补丁，按时间叠在壁纸上。
//
// 参考项目的角色是网格动画，本来就会眨眼（八重神子的 .hbc 在 12 秒循环里
// 有一段闭眼帧）；Go 版是静态烤图，眨眼信息会丢。这里在生成资源时把"眼睛层
// 用闭眼帧、其余图层保持睁眼帧"单独烘一张 `eyes.png`（见 scripts/assetgen），
// 运行时按 0→1→0 的曲线把它贴上去，就得到一次眨眼。
//
// 补丁只在不透明处生效（眼睛与周围一点皮肤、刘海），所以直接铺满整个壁纸矩形，
// 不需要额外做定位；只需要跟人物的呼吸位移一起上下挪一下，避免眼皮"浮"在脸上。
const (
	// blinkPeriod 是两次眨眼的基准间隔（秒）；奇数轮会提前一点，避免机械感。
	blinkPeriod = 4.6
	// blinkClose / blinkHold / blinkOpen 是闭合、保持、睁开三段的时长。
	blinkClose = 0.07
	blinkHold  = 0.05
	blinkOpen  = 0.10
	// blinkJitter 是奇数轮提前的时间（秒），让间隔在 3.7~5.5 秒之间交替。
	blinkJitter = 0.9
)

// blinkAmount 返回 t 时刻的眨眼强度 0..1（0 = 睁眼，1 = 完全闭合）。
//
// 曲线：快速闭合 → 短暂保持 → 稍慢睁开。用 smoothstep 收尾，避免一帧跳变。
func blinkAmount(t float64) float32 {
	if t <= 0 {
		return 0
	}
	cycle := math.Floor(t / blinkPeriod)
	offset := 0.0
	if int64(cycle)%2 == 1 {
		offset = blinkJitter
	}
	p := t - (cycle*blinkPeriod + offset)
	switch {
	case p < 0 || p >= blinkClose+blinkHold+blinkOpen:
		return 0
	case p < blinkClose:
		return smooth01(float32(p / blinkClose))
	case p < blinkClose+blinkHold:
		return 1
	default:
		return 1 - smooth01(float32((p-blinkClose-blinkHold)/blinkOpen))
	}
}

// drawBlink 在壁纸 r 上叠一次眨眼补丁；k 是眨眼强度（0 = 不画）。
//
// bg 用于和着色器保持同一套 Cover 映射：补丁与壁纸同归一化坐标，铺满同一个
// 矩形即可；再按呼吸的垂直位移挪一下，和人物一起起伏。
func (a *App) drawBlink(p *ui.Painter, r ui.Rect, bg *ui.Bitmap, ch theme.Character, elapsed float64, k float32) {
	if k <= 0.004 || ch.Motion.Blink == "" {
		return
	}
	patch := assets.Bitmap(ch.Motion.Blink)
	if patch == nil || bg == nil {
		return
	}
	bw, bh := bg.Size()
	if bw <= 0 || bh <= 0 {
		return
	}
	// 与着色器同一套 Cover 映射：等比放大到铺满矩形、居中裁切。
	// 不能用 r 直接当目标矩形——窗口长宽比和 16:9 不同，直接拉伸会把眼睛拉歪。
	s := max32(r.W/float32(bw), r.H/float32(bh))
	dw, dh := float32(bw)*s, float32(bh)*s
	dx := r.X + (r.W-dw)*0.5
	// 呼吸只走垂直方向（见 wallpaper.kage），这里按人物闸门内 ≈1 处理。
	dy := r.Y + (r.H-dh)*0.5 + float32(math.Sin(elapsed*1.4))*ch.Motion.Breath*s*float32(bh)

	pw, ph := patch.Size()
	// 预乘 alpha 的补丁要连 RGB 一起淡，否则闭眼的皮肤会发亮。
	v := uint8(255*k + 0.5)
	tint := ui.Color{R: v, G: v, B: v, A: v}
	p.ImageRegion(patch,
		ui.Rect{W: float32(pw), H: float32(ph)},
		ui.Rect{X: dx, Y: dy, W: dw, H: dh},
		ui.ImageOptions{Tint: tint},
	)
}
