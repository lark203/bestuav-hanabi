package intro

import (
	"math"
	"testing"
)

// TestResizeCentersPose 全屏/最大化切换后，Pose 必须按当前画面尺寸重新居中。
func TestResizeCentersPose(t *testing.T) {
	in := New(1280, 800)
	in.T = 1.0 // 上升阶段，避免 dock 影响
	small := in.Pose(40, 84)
	if math.Abs(float64(small.X-640)) > 0.01 {
		t.Fatalf("1280 宽时中心应在 640，实际 %.2f", small.X)
	}

	in.Resize(1920, 1080)
	if in.T != 1.0 {
		t.Fatalf("Resize 不应重置时间轴：T=%.2f", in.T)
	}
	large := in.Pose(40, 84)
	if math.Abs(float64(large.X-960)) > 0.01 {
		t.Errorf("1920 宽时中心应在 960，实际 %.2f", large.X)
	}
	if !(large.Y > small.Y) {
		t.Errorf("1080 高时标识应比 800 高时更靠下：%.1f vs %.1f", large.Y, small.Y)
	}
}

// TestResizeAfterDock 停靠后尺寸变化仍应落在品牌位，而不是旧画面中心。
func TestResizeAfterDock(t *testing.T) {
	in := New(1280, 800)
	in.T = Duration
	in.Resize(1920, 1080)
	p := in.Pose(40, 84)
	if math.Abs(float64(p.X-40)) > 0.01 || math.Abs(float64(p.Y-84)) > 0.01 {
		t.Errorf("停靠点应保持 (40,84)，实际 (%.2f,%.2f)", p.X, p.Y)
	}
}
