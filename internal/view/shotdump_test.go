package view

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/theme"
	"bestuav-hanabi/internal/wallpaper"
)

// 这套测试默认跳过：只有在 HANABI_SHOTS 指向一个目录时才出图。
//
//	go test ./internal/view -run TestWriteShots -timeout 1200s -v
//	$env:HANABI_SHOTS="D:\workbuddy\shots"
//
// 用途：
//   - 生成 README 里的主题截图；
//   - 改壁纸着色器/遮罩后，肉眼比对动效在不同时刻的差别（shader 本身只有
//     冒烟测试，观感必须看图）。
//
// 之所以留在测试里：它需要 App 的非导出状态（当前主题、动效素材），
// 独立命令行程序拿不到。
func TestWriteShots(t *testing.T) {
	dir := os.Getenv("HANABI_SHOTS")
	if dir == "" {
		t.Skip("设置 HANABI_SHOTS=<目录> 才出图")
	}
	const w, h = 1280, 800
	only := os.Getenv("HANABI_SHOTS_ONLY") // 只出这些主题，逗号分隔
	for _, ch := range theme.All() {
		if only != "" && !containsID(only, ch.ID) {
			continue
		}
		a := NewApp()
		a.introDone = true
		a.settings.Animations = false
		if ch.ID != theme.Default().ID {
			a.setCharacter(ch.ID)
			a.finishSwitch()
		}
		// 固定时钟：离屏渲染一帧要几秒，真时钟会把动效推进一大截。
		base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		a.clock = func() time.Time { return base }
		a.t0 = base

		img := ui.NewTester(a.View, w, h).Image()
		savePNG(t, filepath.Join(dir, "theme-"+ch.ID+".png"), img)

		// 壁纸单独出两张：t=0 与 t=6，用来比对动效（去掉 UI 的遮挡）。
		bg := assets.Bitmap(ch.Bg)
		if bg == nil {
			t.Logf("%s 没有壁纸素材，跳过壁纸截图", ch.ID)
			continue
		}
		flow := a.wallpaperFlowFor(ch)
		cam := framingOf(ch)
		for _, tm := range []float64{0, 6} {
			fx := tm
			shot := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
					wallpaper.Draw(p, bg, r, wallpaper.Options{
						Time: fx, Play: 1, Tint: ch.Tok.Accent, TintAmount: ch.WallTint,
						Flow: flow, Framing: cam,
					})
				})
			}, w, h).Image()
			savePNG(t, filepath.Join(dir, fmt.Sprintf("wallpaper-%s-t%.2f.png", ch.ID, tm)), shot)

			// 带压暗层与粒子的完整壁纸（等于主界面上看到的那一层，但没有 UI）。
			full := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
					a.drawWallpaperAt(p, r, ch, bg, flow, ch.Tok, fx, cam)
				})
			}, w, h).Image()
			savePNG(t, filepath.Join(dir, fmt.Sprintf("wallpaperfull-%s-t%.2f.png", ch.ID, tm)), full)
		}
		t.Logf("%s 出图完成（flow=%v）", ch.ID, flow != nil)
	}
}

func containsID(list, id string) bool {
	for _, part := range splitComma(list) {
		if part == id {
			return true
		}
	}
	return false
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' || r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func savePNG(t *testing.T, path string, img *image.RGBA) {
	t.Helper()
	if img == nil {
		t.Fatalf("%s：渲染结果为空", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
