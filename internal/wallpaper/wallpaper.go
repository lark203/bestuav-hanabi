// Package wallpaper 把静态插画渲染成"局部在流动"的动态壁纸，并按主题染色。
//
// 模型与参考项目（HanaBi-ui-Genshin 的 LayeredWallpaper）一致：一张图 + 遮罩
// 限制的局部形变，形变后仍采样同一张图。人物主体不动，只有遮罩内（发丝等）
// 在流动。不用"叠第二张图"的做法——那会产生双影。
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

// Flow 是驱动局部流动的素材：一张区域遮罩 + 一个主方向。
//
// Mask 是区域权重（只有非零处才形变），来自与背景同源、构图对齐的预处理图；
// 甘雨实测只覆盖头部发区，所以"只有发型在飘"。
//
// 方向是常量而非逐像素：原始 flow-rg 实测 95% 像素为中性色、仅约 5% 有方向，
// 逐像素驱动只会零星抖动；参考项目也把方向放在每层的配置里。
type Flow struct {
	Mask     *ui.Bitmap // flow-mask：区域权重
	Dir      ui.Point   // 主方向（单位向量，y 向下）；零值按 (1,0)
	Strength float32    // 最大位移，图片 UV（0.01 = 图宽的 1%）
}

// Draw 在 r 处绘制壁纸。
//
//   - t 为动效秒数，play 为流动强度 0..1（0 = 静止）；
//   - tint 是主题的元素色，amount 是染色比例 0..1（0 = 原色）；
//   - flow 是可选的局部流动素材，nil 表示该主题不流动（其余主题走不到形变分支）。
//
// 染色在着色器里做（保亮度），所以换主题时壁纸的暗部不会发灰——
// 这一点用半透明色层覆盖做不到。
func Draw(p *ui.Painter, bg *ui.Bitmap, r ui.Rect, t float64, play float32, tint ui.Color, amount float32, flow *Flow) {
	if bg == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	strength := float32(0)
	var mask *ui.Bitmap
	dir := ui.Point{X: 1, Y: 0}
	if flow != nil && flow.Mask != nil {
		strength = flow.Strength
		mask = flow.Mask
		if flow.Dir.X != 0 || flow.Dir.Y != 0 {
			dir = flow.Dir
		}
	}
	// image0=背景，image1=区域遮罩；其余槽位留空（读透明）。
	// FlowDir 要传 [2]float32：shader 的 vec 型 uniform 只收数字数组或 ui.Color，
	// 传 ui.Point 会在设值时 panic（"not a number, numbers or a ui.Color"）。
	fx.Draw(p, r, shader.Options{
		Uniforms: map[string]any{
			"Time":         t,
			"Play":         play,
			"Tint":         tint,
			"TintAmount":   amount,
			"FlowStrength": strength,
			"FlowDir":      [2]float32{dir.X, dir.Y},
		},
		Images: [4]*ui.Bitmap{bg, mask},
	})
}
