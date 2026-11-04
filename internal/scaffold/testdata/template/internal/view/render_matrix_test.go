package view

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 离屏渲染矩阵：每个主题 × 每个页面都要能画出来，且页面标志性文字出现。
//
// 比像素基准稳：字体/平台差异不会让 CI 红，但漏画整页、panic、页面分发写错会立刻抓到。
func TestRenderMatrix(t *testing.T) {
	pages := []struct {
		id   string
		text string
	}{
		{libraryPageID, "主题库"},
		{wallpaperPageID, "壁纸"},
		{settingsPageID, "数据与关于"},
	}
	for _, ch := range visual.All() {
		for _, pg := range pages {
			a := NewApp()
			a.introDone = true
			a.settings.Animations = false
			a.settings.WallpaperPlay = false
			a.animPlay = 0
			if ch.ID != visual.Default().ID {
				a.setCharacter(ch.ID)
				a.finishSwitch()
			}
			stripped(a)
			a.SetPage(pg.id)
			tt := ui.NewTester(a.View, 1000, 640)
			tt.Frame()
			img := tt.Image()
			if img == nil || img.Bounds().Dx() != 1000 {
				t.Fatalf("%s / %s 渲染失败", ch.ID, pg.id)
			}
			if !tt.HasText(pg.text) {
				t.Errorf("%s / %s 缺少文字 %q", ch.ID, pg.id, pg.text)
			}
			// 非纯背景像素：防止整页静默空白。
			lit := 0
			for i := 0; i+3 < len(img.Pix); i += 4 {
				if img.Pix[i] > 40 || img.Pix[i+1] > 40 || img.Pix[i+2] > 40 {
					lit++
				}
			}
			if lit < 2000 {
				t.Errorf("%s / %s 渲染过空：亮像素 %d", ch.ID, pg.id, lit)
			}
		}
	}
}

// TestRenderSizes 常见窗口尺寸（含最小尺寸）都不崩。
func TestRenderSizes(t *testing.T) {
	for _, sz := range [][2]int{{1000, 640}, {1280, 800}, {1920, 1080}} {
		a := benchApp()
		a.SetPage(settingsPageID)
		tt := ui.NewTester(a.View, sz[0], sz[1])
		tt.Frame()
		if img := tt.Image(); img == nil || img.Bounds().Dx() != sz[0] {
			t.Fatalf("%dx%d 渲染失败", sz[0], sz[1])
		}
	}
}
