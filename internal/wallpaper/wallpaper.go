// Package wallpaper 把静态插画渲染成有缓慢流动感的动态壁纸，并按主题染色。
//
// 形态与原项目一致（着色器驱动，而不是叠贴图），但位移量用解析式正弦生成，
// 不依赖 flow/phase 三张辅助图——省三张 1920×1080 采样，CPU 渲染器也能实时跑。
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

// Draw 在 r 处绘制壁纸。
//
//   - t 为动效秒数，play 为流动强度 0..1（0 = 静止）；
//   - tint 是主题的元素色，amount 是染色比例 0..1（0 = 原色）。
//
// 染色在着色器里做（保亮度），所以换主题时壁纸的暗部不会发灰——
// 这一点用半透明色层覆盖做不到。
func Draw(p *ui.Painter, bg *ui.Bitmap, r ui.Rect, t float64, play float32, tint ui.Color, amount float32) {
	if bg == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	fx.Draw(p, r, shader.Options{
		Uniforms: map[string]any{
			"Time":       t,
			"Play":       play,
			"Tint":       tint,
			"TintAmount": amount,
		},
		Images: [4]*ui.Bitmap{bg},
	})
}
