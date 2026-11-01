// Package intro 复刻 HanaBi 的 6.20 秒开场：时间轴、粒子运动与元素标识的姿态。
//
// 这里只有纯逻辑（不含任何绘制），以便单测：给定时间就能算出每个粒子在哪、
// 标识多大、整体淡到多少。数值全部照搬原项目 src/intro/intro.cpp，
// 连随机数发生器都一致（LCG，种子 0x85A16B3D），所以粒子分布与原版逐帧对应。
package intro

import "math"

// 时间轴（秒），与原项目 Intro 的常量一致。
const (
	Duration  = 6.20 // 总时长
	TurnStart = 1.45 // 标识开始转向
	TurnEnd   = 3.65 // 转向结束（成形）
	DockStart = 4.65 // 开始停靠到侧栏
)

// 粒子数量。
const (
	SnowCount = 32 // 雪（斜向短划）
	LantCount = 20 // 孔明灯
)

const tau = 2 * math.Pi

// Snow 是一片雪，坐标均为归一化（0..1）。
type Snow struct {
	X, Y  float32 // 初始位置
	Size  float32 // 相对短边
	Depth float32 // 景深：影响下落速度与不透明度
	Ph    float32 // 相位
}

// Lantern 是一盏孔明灯，坐标均为归一化（0..1）。
type Lantern struct {
	X, Y0 float32 // 初始位置
	Size  float32 // 相对短边
	Spd   float32 // 上升速度（归一化每秒）
	Ph    float32 // 相位
}

// MarkPose 是元素标识在某一帧的位置、半径与不透明度。
type MarkPose struct {
	X, Y    float32
	Radius  float32
	Opacity float32
}

// Intro 持有开场动画的状态。
type Intro struct {
	T    float32 // 已播放秒数
	W, H float32

	snow [SnowCount]Snow
	lant [LantCount]Lantern

	finished bool
	skipped  bool
	skipAt   float32 // 触发跳过时的 T
	skipDur  float32
}

// New 创建开场并按原项目算法铺好粒子（确定性，同一尺寸每次结果相同）。
func New(w, h float32) *Intro {
	in := &Intro{W: max32(w, 1), H: max32(h, 1)}
	in.reset()
	return in
}

func (in *Intro) reset() {
	in.T = 0
	in.finished, in.skipped = false, false

	// 与原项目同一个线性同余发生器，保证粒子分布一致。
	r := lcg(0x85A16B3D)
	for i := range in.snow {
		in.snow[i] = Snow{
			X:     r.next(),
			Y:     r.next(),
			Size:  0.0005 + r.next()*0.0013,
			Depth: r.next(),
			Ph:    r.next() * tau,
		}
	}
	for i := 0; i < LantCount; i++ {
		depth := float32(i) / float32(LantCount-1)
		side := float32(-1)
		if i%2 == 1 {
			side = 1
		}
		// 注意取值顺序：原项目初始化列表里 Random 的调用顺序是 x → y0 → ph。
		x := 0.5 + side*(0.20+r.next()*0.27)
		y0 := 0.18 + r.next()*0.95
		ph := r.next() * tau
		in.lant[i] = Lantern{
			X:    x,
			Y0:   y0,
			Size: 0.0028 + depth*depth*0.027,
			Spd:  0.007 + depth*0.019,
			Ph:   ph,
		}
	}
	// 原项目把最后两盏固定在构图的两侧（一大一小），作为前景点缀。
	in.lant[LantCount-2] = Lantern{X: 0.075, Y0: 0.76, Size: 0.066, Spd: 0.038, Ph: 0.7}
	in.lant[LantCount-1] = Lantern{X: 0.910, Y0: 0.40, Size: 0.050, Spd: 0.030, Ph: 2.1}
}

// lcg 是原项目用的线性同余发生器（乘数 1664525、增量 1013904223、取高 24 位）。
type lcg uint32

func (l *lcg) next() float32 {
	*l = *l*1664525 + 1013904223
	return float32(*l>>8) / 16777216.0
}

// Update 推进时间；dt 会被夹到 0..0.20，避免暂停后恢复时一帧跳太多。
func (in *Intro) Update(dt float32) {
	if dt < 0 {
		dt = 0
	} else if dt > 0.20 {
		dt = 0.20
	}
	if in.finished {
		return
	}

	if in.skipped {
		// 跳过：用不超过 0.9 秒把剩余进度快速走完，而不是硬切。
		remain := Duration - in.skipAt
		dur := min32(0.9, remain)
		in.skipDur += dt
		in.T = min32(Duration, in.skipAt+in.skipDur+(remain-dur)*Smooth(in.skipDur, 0, dur))
	} else {
		in.T = min32(Duration, in.T+dt)
	}
	in.finished = in.T >= Duration
}

// Skip 请求跳过；重复调用无效果。
func (in *Intro) Skip() {
	if in.finished || in.skipped {
		return
	}
	in.skipped, in.skipAt, in.skipDur = true, in.T, 0
}

// Finished 报告开场是否已结束。
func (in *Intro) Finished() bool { return in.finished }

// Opacity 是整体淡出（4.65s 起，5.55s 完全透明）。
func (in *Intro) Opacity() float32 { return 1 - Smooth(in.T, DockStart, 5.55) }

// HandoffK 是「交给主界面」的进度：标识从中央飞到侧栏并缩到图标大小。
func (in *Intro) HandoffK() float32 { return Smooth(in.T, DockStart, Duration) }

// Progress 是整体进度 0..1。
func (in *Intro) Progress() float32 { return clamp01(in.T / Duration) }

// Pose 给出元素标识此刻的位置与半径。
//
// dockX/dockY 是停靠终点（侧栏品牌头像的位置）：标识收缩成图标大小停在那里，
// 主界面随后出现，视觉上就是「开场标识变成应用图标」。
func (in *Intro) Pose(dockX, dockY float32) MarkPose {
	t := in.T
	rise := Smooth(t, 0.20, TurnStart)
	dock := in.HandoffK()

	// 起始点在画面中央偏下，随 rise 从 66% 高度升到 45%。
	sx := in.W * 0.5
	sy := in.H * lerp(0.66, 0.45, rise)

	// 成形瞬间的缩放脉冲：先微收、再弹大、回落，对应 1.45→4.18s 那段。
	impact := 1 - 0.025*Smooth(t, 3.33, 3.58) + 0.065*Smooth(t, 3.58, 3.78) - 0.040*Smooth(t, 3.78, 4.18)

	// 直线飞向品牌位：标识一路缩小，落进品牌位即交接。
	// 之前这里有一条向上拱的抛物线（避免直线平移显得机械），但实测它会
	// 先掠过品牌位上方再折回来——王工要求直接落到品牌处，去掉。
	x := lerp(sx, dockX, dock)
	y := lerp(sy, dockY, dock)

	return MarkPose{
		X:       x,
		Y:       y,
		Radius:  lerp(min32(in.W, in.H)*0.155*impact, 16, dock),
		Opacity: Smooth(t, 0.18, 0.67),
	}
}

// SnowAt 返回第 i 片雪的归一化位置、不透明度（0..1）与长度（DIP）。
func (in *Intro) SnowAt(i int, short float32) (x, y, alpha, length float32) {
	p := in.snow[i]
	t := in.T
	x = p.X + 0.006*sin32(t*0.40+p.Ph)
	// fmod 保证循环：+2 避免负数取模。
	y = fmod32(p.Y-t*(0.009+0.012*p.Depth)+2, 1)
	a := Smooth(t, 0.80, 2.2) * (0.10 + 0.25*p.Depth) * Smooth(y, 0, 0.08) * (1 - Smooth(y, 0.92, 1))
	return x, y, a, short * p.Size
}

// LanternAt 返回第 i 盏灯的归一化位置、不透明度与尺寸（DIP）。
func (in *Intro) LanternAt(i int, short float32) (x, y, alpha, size float32) {
	p := in.lant[i]
	t := in.T
	x = p.X + 0.007*sin32(t*0.40+p.Ph)
	y = p.Y0 - t*p.Spd
	// 上下边缘淡入淡出：灯从画面底部升起、从顶部飘走。
	edge := Smooth(y, -0.1, 0.05) * (1 - Smooth(y, 1.05, 1.2))
	a := edge * Smooth(t, 0, 0.70) * 0.85
	return x, y, a, short * p.Size
}

// Smooth 是 smootherstep（6u⁵-15u⁴+10u³），与原项目的 S() 一致。
func Smooth(t, a, b float32) float32 {
	if b == a {
		return 0
	}
	u := clamp01((t - a) / (b - a))
	return u * u * u * (u*(u*6-15) + 10)
}

func sin32(v float32) float32 { return float32(math.Sin(float64(v))) }

func fmod32(a, b float32) float32 {
	return a - b*float32(math.Floor(float64(a/b)))
}

func clamp01(v float32) float32    { return min32(1, max32(0, v)) }
func lerp(a, b, k float32) float32 { return a + (b-a)*k }
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
