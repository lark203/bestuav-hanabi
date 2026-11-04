// Package wallpaper 把静态插画渲染成"人物在呼吸、发丝在飘、环境在流动"的动态壁纸。
//
// 三层结构，每层一张遮罩、一个幅度：
//
//	person   人物区域 —— 呼吸的总闸门；它以外严格等于原图
//	hair     发丝/衣摆 —— R 基础层（慢、小）、G 细缕层（快、略大）
//	ambient  环境 —— 只有大块柔和的区域（天空、光、雾、水）在漂移
//
// 模型取自参考项目（HanaBi-ui-Genshin）background.cpp 的 shake 分支：
// 先按遮罩混合"原始采样"与"位移后采样"的颜色，而不是位移 UV。
// 这样遮罩外一个量化误差都不会有。
package wallpaper

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
)

//go:embed wallpaper.kage
var source string

// fx 在包初始化时编译：着色器写错会直接 panic（与 blackhole 演示同样的策略），
// 好处是错误在启动时就暴露，而不是等某次绘制静默不出图。
var fx = shader.MustCompile(source)

// Flow 是一张壁纸的动态素材与参数。
//
// 位移幅度的单位是**图片 UV**（不是屏幕像素），所以窗口缩放时观感一致。
// 任何一个遮罩为 nil 时，对应层读作透明（权重 0），自动不参与——缺素材的
// 主题会安静地退化成静态，而不是报错。
type Flow struct {
	// Person 是人物区域遮罩（灰度，非零处才算人物）——呼吸的总闸门。
	Person *ui.Bitmap
	// Hair 是发丝/衣摆遮罩：R = 基础层，G = 细缕层。
	Hair *ui.Bitmap
	// Ambient 是环境漂移遮罩。
	Ambient *ui.Bitmap

	// Breath 是呼吸幅度（图片 UV 的垂直分量）；0 = 不呼吸。
	//
	// 只走垂直方向且全人物同步：呼吸时胸腔整体上下起伏，不是各部位乱飘。
	// 周期由着色器的 sin(Time*1.4) 决定，约 4.5 秒一次，对应静息呼吸频率。
	Breath float32
	// Sway 是发丝飘动的幅度（图片 UV）；0 = 发丝不动。
	Sway float32
	// Drift 是环境漂移的幅度（图片 UV）；0 = 环境静止。
	Drift float32
	// DriftDir 是漂移方向（单位向量）。零值等于向右。
	DriftDir [2]float32
	// DriftSpeed 是漂移的相位速度倍率；0 视为 0.05（放着不动）。
	DriftSpeed float32
}

// Options 是一次壁纸绘制的全部参数。
type Options struct {
	// Time 是动效秒数，Play 是动效强度 0..1（0 = 静止）。
	Time float64
	Play float32
	// Tint 是主题元素色，TintAmount 是染色比例 0..1（0 = 原色）。
	Tint       ui.Color
	TintAmount float32
	// Flow 是可选的动态素材，nil 表示该主题完全静态。
	Flow *Flow
}

// Draw 在 r 处按 o 绘制壁纸。
//
// 染色在着色器里做（保亮度），所以换主题时壁纸的暗部不会发灰——
// 这一点用半透明色层覆盖做不到。
func Draw(p *ui.Painter, bg *ui.Bitmap, r ui.Rect, o Options) {
	if bg == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	var person, hair, ambient *ui.Bitmap
	var breath, sway, drift, speed float32
	dir := [2]float32{1, 0}
	if o.Flow != nil {
		person, hair, ambient = o.Flow.Person, o.Flow.Hair, o.Flow.Ambient
		breath, sway, drift, speed = o.Flow.Breath, o.Flow.Sway, o.Flow.Drift, o.Flow.DriftSpeed
		if o.Flow.DriftDir != [2]float32{} {
			dir = o.Flow.DriftDir
		}
	}
	// image0=背景，image1=人物闸门，image2=发丝（R/G 两层），image3=环境。
	// 空缺槽位读透明（权重 0），对应层自动不参与。
	fx.Draw(p, r, shader.Options{
		Uniforms: map[string]any{
			"Time":           o.Time,
			"Play":           o.Play,
			"Tint":           o.Tint,
			"TintAmount":     o.TintAmount,
			"BreathStrength": breath,
			"FlowStrength":   sway,
			"DriftStrength":  drift,
			"DriftDir":       dir[:],
			"DriftSpeed":     speed,
		},
		Images: [4]*ui.Bitmap{bg, person, hair, ambient},
	})
}