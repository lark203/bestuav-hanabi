package view

import (
	"testing"

	"bestuav-hanabi/internal/wallpaper"
)

// TestSpringProgress 钉住主题切换的弹簧曲线：
// 起点 0、单调、比匀速更早到达（0.1s 已经过半），到 0.42s 基本到 1。
func TestSpringProgress(t *testing.T) {
	if v := springProgress(0); v != 0 {
		t.Errorf("t=0 应为 0，实际 %v", v)
	}
	if v := springProgress(0.42); v < 0.999 {
		t.Errorf("0.42s 应基本到位，实际 %v", v)
	}
	prev := float32(0)
	for i := 1; i <= 42; i++ {
		v := springProgress(float64(i) / 100)
		if v < prev-1e-6 {
			t.Fatalf("t=%.2fs 出现回退：%v < %v", float64(i)/100, v, prev)
		}
		prev = v
	}
	// 弹簧起步比 slowerstep 快：0.1s 时已经过半（匀速是 0.24）。
	if v := springProgress(0.1); v < 0.6 {
		t.Errorf("0.1s 应已过半（弹簧起步快），实际 %v", v)
	}
}

// TestFramingMix 钉住取景插值：端点原样，中点居中。
func TestFramingMix(t *testing.T) {
	from := wallpaper.Framing{Zoom: 1.22, FocusX: 0.395, FocusY: 0.485}
	to := wallpaper.Framing{Zoom: 1.02, FocusX: 0.605, FocusY: 0.485}

	if got := from.Mix(to, 0); got != from {
		t.Errorf("k=0 应等于起点，实际 %+v", got)
	}
	if got := from.Mix(to, 1); got != to {
		t.Errorf("k=1 应等于终点，实际 %+v", got)
	}
	mid := from.Mix(to, 0.5)
	if d := mid.Zoom - 1.12; d > 1e-5 || d < -1e-5 {
		t.Errorf("中点 zoom 应为 1.12，实际 %v", mid.Zoom)
	}
	if d := mid.FocusX - 0.5; d > 1e-5 || d < -1e-5 {
		t.Errorf("中点 focusX 应为 0.5，实际 %v", mid.FocusX)
	}
	// 零值要补成默认取景（zoom 1、居中），否则着色器会拿到 0 缩放。
	def := wallpaper.Framing{}.Norm()
	if def.Zoom != 1 || def.FocusX != 0.5 || def.FocusY != 0.5 {
		t.Errorf("零值应归一成 zoom=1/居中，实际 %+v", def)
	}
}
