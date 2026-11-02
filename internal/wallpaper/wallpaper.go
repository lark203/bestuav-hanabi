// Package wallpaper 把静态插画渲染成"局部在流动"的动态壁纸，并按主题染色。
//
// 模型与参考项目（HanaBi-ui-Genshin 的 LayeredWallpaper）一致：一张图 + 多层
// 遮罩限制的局部形变，形变后仍采样同一张图。人物主体不动，只有遮罩内（发丝等）
// 在流动。不用"叠第二张图"的做法——那会产生双影。
//
// 分层次（一缕一缕分开飘）靠的是遮罩本身：整片发区一层、两处局部各一层，
// 三层用不同的空间频率与方向累加，相邻发缕的位移量因此不同。
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

// Flow 是驱动局部流动的素材：一张整片遮罩、两层局部细缕遮罩，加一个主方向。
//
// 三层覆盖范围逐级收窄（实测甘雨：整片 9.51%、下侧一缕 1.08%、顶部一缕 1.57%），
// 三层相乘累加后，细缕处比周围多一份位移，于是看起来是一缕一缕分开飘。
//
// WaveA / WaveB 可为 nil：未绑定读为透明，该层自动不参与，退化成单层。
type Flow struct {
	Mask     *ui.Bitmap // 整片区域（基础层）
	WaveA    *ui.Bitmap // 局部细缕（可选）
	WaveB    *ui.Bitmap // 局部细缕（可选）
	Dir      ui.Point   // 主方向（单位向量，y 向下）；零值按 (1,0)
	Strength float32    // 基础层最大位移，图片 UV（0.01 = 图宽的 1%）
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
	dir := ui.Point{X: 1, Y: 0}
	var mask, waveA, waveB *ui.Bitmap
	if flow != nil && flow.Mask != nil {
		strength = flow.Strength
		mask, waveA, waveB = flow.Mask, flow.WaveA, flow.WaveB
		if flow.Dir.X != 0 || flow.Dir.Y != 0 {
			dir = flow.Dir
		}
	}
	// image0=背景，image1=整片遮罩，image2/image3=两层细缕；空缺槽位读透明。
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
		Images: [4]*ui.Bitmap{bg, mask, waveA, waveB},
	})
}
