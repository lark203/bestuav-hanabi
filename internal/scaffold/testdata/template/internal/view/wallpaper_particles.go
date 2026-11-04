package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 壁纸粒子：每个元素一套环境粒子，让"动态壁纸"一眼看得出来。
//
// 只靠着色器里的位移，动效是"看久了才发现"的量级（4~6px、周期十几秒）；
// 参考项目每个角色都另有一层环境粒子（樱花 / 火星 / 光尘 / 气泡 / 金屑），
// background.cpp 的 DrawParticles 里那几十行就是干这个的。这里用同一套思路
// 重写：位置由「序号 + 时间」直接算出来（无状态、可复现），不存粒子数组，
// 所以不会因为掉帧而漂移。
//
// 所有粒子都画在壁纸之上、UI 之下（drawWallpaper 的最后一步），强度跟随
// 「背景流动」开关；关掉动效时完全不画。

// wpShape 是粒子的形态。
type wpShape uint8

const (
	wpMote   wpShape = iota // 光尘：一团柔光 + 亮核
	wpPetal                 // 花瓣：会翻面的胶囊
	wpEmber                 // 火星：亮核 + 向上的短尾
	wpBubble                // 气泡：细圆环 + 高光
)

// wpParticles 是一套环境粒子的参数。
//
// 尺寸按窗口高度换算（size×h），这样窗口缩放时粒子不会忽大忽小。
type wpParticles struct {
	count int
	shape wpShape
	glow  ui.Color // 柔光色
	core  ui.Color // 核心色
	size  float32  // 基础半径（相对高度）
	life  [2]float64
	// speed 是每秒的纵向归一化位移（相对高度）；负值表示上升。
	speed [2]float64
	sway  float32 // 横向摆幅（相对宽度）
	alpha float32
}

// 每套粒子的参数：数量、形态、颜色、寿命与漂移。
//
// 取色都从各角色立绘上直接吸（八重的樱粉、宵宫的暖橙、珐露珊的青灰…），
// 透明度压在 0.1~0.35：它们是氛围，不该抢文字和卡片。
var (
	sakuraParticles = wpParticles{
		count: 24, shape: wpPetal,
		glow: ui.Hex("#E9A8C6"), core: ui.Hex("#F7C6DC"),
		size: 0.0060, life: [2]float64{16, 30}, speed: [2]float64{0.030, 0.060}, sway: 0.028,
		alpha: 0.34,
	}
	emberParticles = wpParticles{
		count: 30, shape: wpEmber,
		glow: ui.Hex("#FF9A4D"), core: ui.Hex("#FFD79A"),
		size: 0.0048, life: [2]float64{6, 13}, speed: [2]float64{-0.075, -0.040}, sway: 0.010,
		alpha: 0.36,
	}
	dreamParticles = wpParticles{
		count: 44, shape: wpMote,
		glow: ui.Hex("#D8F5B0"), core: ui.Hex("#F2FFD8"),
		size: 0.0050, life: [2]float64{10, 24}, speed: [2]float64{-0.012, -0.004}, sway: 0.014,
		alpha: 0.26,
	}
	bubbleParticles = wpParticles{
		count: 20, shape: wpBubble,
		glow: ui.Hex("#8ADCF5"), core: ui.Hex("#E8FBFF"),
		size: 0.0060, life: [2]float64{12, 26}, speed: [2]float64{-0.030, -0.014}, sway: 0.008,
		alpha: 0.30,
	}
	dustParticles = wpParticles{
		count: 46, shape: wpMote,
		glow: ui.Hex("#CFE7DF"), core: ui.Hex("#F0FFFA"),
		size: 0.0040, life: [2]float64{4, 10}, speed: [2]float64{-0.020, 0.010}, sway: 0.020,
		alpha: 0.28,
	}
	goldParticles = wpParticles{
		count: 36, shape: wpMote,
		glow: ui.Hex("#E8C87A"), core: ui.Hex("#FFEFC2"),
		size: 0.0044, life: [2]float64{8, 18}, speed: [2]float64{-0.055, -0.025}, sway: 0.012,
		alpha: 0.34,
	}
)

// particlesFor 返回元素对应的粒子；nil 表示该主题不加粒子（甘雨保持原样）。
func particlesFor(e visual.Element) *wpParticles {
	switch e {
	case visual.Electro:
		return &sakuraParticles // 八重神子：神樱落樱
	case visual.Pyro:
		return &emberParticles // 宵宫：烟花火星
	case visual.Dendro:
		return &dreamParticles // 纳西妲：梦境光尘
	case visual.Hydro:
		return &bubbleParticles // 芙宁娜：水汽气泡
	case visual.Anemo:
		return &dustParticles // 珐露珊：机关浮尘
	case visual.Geo:
		return &goldParticles // 钟离：岩金碎屑
	}
	return nil
}

// drawWallpaperParticles 在壁纸区域 r 上画当前主题的粒子。
//
// strength 是 0..1 的整体强度（跟随「背景流动」），t 是动效秒数。
func (a *App) drawWallpaperParticles(p *ui.Painter, r ui.Rect, ch visual.Character, t float64, strength float32) {
	if strength <= 0.01 || r.W <= 0 || r.H <= 0 {
		return
	}
	spec := particlesFor(ch.Element)
	if spec == nil {
		return
	}
	h := r.H

	for i := 0; i < spec.count; i++ {
		seed := uint32(i)*0x9E3779B1 + 0x7F4A7C15
		hx := hash01(seed, 1)
		hy := hash01(seed, 2)
		hl := hash01(seed, 3)
		hs := hash01(seed, 4)
		hp := hash01(seed, 5)

		life := spec.life[0] + float64(hl)*(spec.life[1]-spec.life[0])
		age := math.Mod(t+float64(hy)*life, life)
		u := float32(age / life)

		// 出现与消失：两端各留一段淡入淡出，粒子不会凭空冒出或消失。
		fade := smooth01(u / 0.12)
		fade *= 1 - smooth01((u-0.75)/0.25)
		alpha := fade * spec.alpha * strength
		if alpha <= 0.004 {
			continue
		}

		speed := spec.speed[0] + float64(hs)*(spec.speed[1]-spec.speed[0])
		// 上升的粒子从画面下方进入，下落的从上方进入。
		y0 := 0.5
		if speed < 0 {
			y0 = 1.05
		} else {
			y0 = -0.05
		}
		x := r.X + r.W*(hx+spec.sway*float32(math.Sin(float64(i)*1.7+t*0.35)))
		y := r.Y + h*(float32(y0)+float32(speed)*float32(age))

		rad := spec.size * h * (0.65 + 0.7*hp)

		switch spec.shape {
		case wpPetal:
			// 花瓣：绕自身长轴"翻面"（横向压扁），再按飘落角度旋转。
			flip := 0.30 + 0.70*float32(math.Abs(math.Sin(float64(i)*0.9+float64(u)*6.2)))
			deg := float32(28*math.Sin(float64(i)*1.3)) + 40*float32(math.Sin(float64(i)+float64(u)*2.0))
			w, hh := rad*2.6, rad
			xf := ui.Scaling(flip, 1).Then(ui.Rotation(deg))
			p.Transform(xf, x, y, func() {
				p.Fill(ui.Rect{X: x - w/2, Y: y - hh/2, W: w, H: hh}, spec.core.Alpha(alpha), hh/2)
			})
			glow(p, x, y, rad*2.2, spec.glow, alpha*0.18)
		case wpEmber:
			// 火星：向上拖一条短尾，头上有亮核。
			tail := rad * (2.0 + 3.0*hp)
			seg(p, x, y, x+rad*0.35, y-tail, rad*0.9, spec.core.Alpha(alpha*0.45))
			glow(p, x, y, rad*2.6, spec.glow, alpha*0.30)
			circle(p, x, y, rad*0.55, spec.core.Alpha(alpha))
		case wpBubble:
			// 气泡：细圆环 + 左上高光。
			p.Stroke(ui.Rect{X: x - rad, Y: y - rad, W: rad * 2, H: rad * 2}, spec.glow.Alpha(alpha*0.75), rad, 1.0)
			circle(p, x-rad*0.32, y-rad*0.34, rad*0.30, spec.core.Alpha(alpha*0.9))
		default:
			// 光尘：柔光 + 亮核，亮度跟着序号轻微呼吸。
			tw := 0.75 + 0.25*float32(math.Sin(float64(i)*2.3+t*1.7))
			glow(p, x, y, rad*3.4, spec.glow, alpha*0.32*tw)
			circle(p, x, y, rad*0.7, spec.core.Alpha(alpha*0.85*tw))
		}
	}
}

// hash01 是由（序号, 通道）决定的确定性 0..1 随机数：同一颗粒子每一帧的
// 起点、寿命、尺寸都一样，画面才不会"抖"。
func hash01(seed, lane uint32) float32 {
	v := seed ^ (lane * 0x85EBCA6B)
	v ^= v >> 16
	v *= 0x7FEB352D
	v ^= v >> 15
	v *= 0x846CA68B
	v ^= v >> 16
	return float32(v&0xFFFFFF) / float32(0x1000000)
}

// smooth01 是 0..1 的平滑过渡（smoothstep 0→1）。
func smooth01(t float32) float32 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}
