// Package wallpaper 把静态插画渲染成"人物在呼吸、发丝在飘"的动态壁纸。
//
// 硬约束：建筑、天空完全不动。靠人物轮廓遮罩（Person）当总闸门实现——
// 它以外的地方严格等于原图，不产生任何位移。
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

// Flow 是局部动态的素材。
//
// Person 是人物区域遮罩（灰度，非零处才动）——它是闸门，决定"谁能动"。
// 甘雨实测覆盖 x 34.2%~59.4% / y 22.2%~73.7%，一条竖直的人形带，仅 4.9%。
//
// Veil 是发丝基础层（可选，整片发区），Wave 是细缕层（可选）。
// 两者都会被 Person 裁到人物之内——这正是"建筑天空不动"的保证。
// 未绑定（nil）读为透明，该层自动不参与。
type Flow struct {
	Person *ui.Bitmap // 人物区域（闸门，必需）
	Veil   *ui.Bitmap // 发丝基础层（可选）
	Wave   *ui.Bitmap // 发丝细缕层（可选）
	// Breath 是呼吸幅度（图片 UV 的垂直分量）；0 = 不呼吸。
	Breath float32
	// Sway 是发丝飘动幅度（图片 UV）；0 = 发丝不动。
	Sway float32
}

// Draw 在 r 处绘制壁纸。
//
//   - t 为动效秒数，play 为动效强度 0..1（0 = 静止）；
//   - tint 是主题的元素色，amount 是染色比例 0..1（0 = 原色）；
//   - flow 是可选的局部动态素材，nil 表示该主题完全静态。
//
// 染色在着色器里做（保亮度），所以换主题时壁纸的暗部不会发灰——
// 这一点用半透明色层覆盖做不到。
func Draw(p *ui.Painter, bg *ui.Bitmap, r ui.Rect, t float64, play float32, tint ui.Color, amount float32, flow *Flow) {
	if bg == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	var person, veil, wave *ui.Bitmap
	var breath, sway float32
	if flow != nil && flow.Person != nil {
		person, veil, wave = flow.Person, flow.Veil, flow.Wave
		breath, sway = flow.Breath, flow.Sway
	}
	// image0=背景，image1=人物闸门，image2=发丝基础层，image3=发丝细缕。
	// 空缺槽位读透明（权重 0），对应层自动不参与。
	fx.Draw(p, r, shader.Options{
		Uniforms: map[string]any{
			"Time":           t,
			"Play":           play,
			"Tint":           tint,
			"TintAmount":     amount,
			"BreathStrength": breath,
			"FlowStrength":   sway,
		},
		Images: [4]*ui.Bitmap{bg, person, veil, wave},
	})
}
