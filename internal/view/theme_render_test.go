package view

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

// TestCharacterThemesRender 每个角色的纹样/窗口按钮都真的能画出来。
//
// 这些绘制分支（chrome.go / ornament_primitives.go）没有纯逻辑单测可钉，
// 用离屏渲染做冒烟：哪怕某个 Path 坐标写错，也会在这里 panic 或整屏空白。
func TestCharacterThemesRender(t *testing.T) {
	for _, ch := range theme.All() {
		a := NewApp()
		a.introDone = true
		if ch.ID != theme.Default().ID {
			a.setCharacter(ch.ID)
			a.finishSwitch()
		}
		img := ui.NewTester(a.View, 320, 200).Image()
		if img == nil || img.Bounds().Dx() != 320 {
			t.Fatalf("%s 渲染失败", ch.ID)
		}
		// 统计非纯背景像素，避免所有绘制都静默返回。
		lit := 0
		for i := 0; i+3 < len(img.Pix); i += 4 {
			if img.Pix[i] > 40 || img.Pix[i+1] > 40 || img.Pix[i+2] > 40 {
				lit++
			}
		}
		if lit < 1000 {
			t.Errorf("%s 渲染过于空：亮像素只有 %d", ch.ID, lit)
		}
	}
}
