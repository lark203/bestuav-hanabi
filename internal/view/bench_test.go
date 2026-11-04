package view

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/visual"
)

// 基准用的 App：去掉开场、壁纸与过渡，只测「界面构建」本身。
func benchApp() *App {
	a := stripped(NewApp())
	a.introDone = true
	a.settings.Animations = false
	a.settings.WallpaperPlay = false
	a.animPlay = 0
	a.clock = time.Now
	return a
}

// BenchmarkViewFrame 1280×800 每帧的构建/布局/绘制成本（含离屏栅格化）。
//
//	go test ./internal/view -bench ViewFrame -benchmem
func BenchmarkViewFrame(b *testing.B) {
	a := benchApp()
	tt := ui.NewTester(a.View, 1280, 800)
	tt.Frame() // 预热字体/位图缓存
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.Frame()
	}
}

// BenchmarkViewFrameSettings 设置页（面板与控件更多）的单帧成本。
func BenchmarkViewFrameSettings(b *testing.B) {
	a := benchApp()
	a.SetPage(settingsPageID)
	tt := ui.NewTester(a.View, 1280, 800)
	tt.Frame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.Frame()
	}
}

// BenchmarkThemeSwitch 主题切换的纯状态成本（不含壁纸解码）。
func BenchmarkThemeSwitch(b *testing.B) {
	a := benchApp()
	ids := make([]string, 0, len(visual.All()))
	for _, c := range visual.All() {
		ids = append(ids, c.ID)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.setCharacter(ids[i%len(ids)])
		a.finishSwitch()
	}
}
