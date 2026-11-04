package view

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// TestIntroLayerFollowsViewport 开场播放期间窗口尺寸变化（全屏/最大化）时，
// introLayer 必须把当前绘制区域同步给 Intro；否则 Pose 会用旧尺寸算中心。
func TestIntroLayerFollowsViewport(t *testing.T) {
	a := NewApp()
	a.introDone = false
	a.intro = visual.New(160, 100)
	ui.NewTester(a.View, 480, 320).Image()
	if a.intro.W != 480 || a.intro.H != 320 {
		t.Fatalf("intro 未跟随视口：got %.0fx%.0f, want 480x320", a.intro.W, a.intro.H)
	}
}
