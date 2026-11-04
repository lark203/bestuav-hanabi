package view

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// 像素基准：本地可选（字体栅格化因平台而异，CI 不跑）。
//
//	$env:HANABI_GOLDEN="1"           # 启用比较
//	$env:HANABI_GOLDEN_UPDATE="1"    # 生成/更新基准（再跑一次比较）
//	go test ./internal/view -run TestGoldenDefault
//
// 场景刻意去掉壁纸/头像素材，保证没有 assets 的机器也能复现。
func TestGoldenDefault(t *testing.T) {
	if os.Getenv("HANABI_GOLDEN") == "" {
		t.Skip("设置 HANABI_GOLDEN=1 才跑像素基准")
	}
	a := goldenApp()
	tt := ui.NewTester(a.View, 900, 560)
	tt.Frame()
	img := tt.Image()
	if img == nil {
		t.Fatal("渲染失败")
	}

	path := filepath.Join("testdata", "golden", "default.png")
	if os.Getenv("HANABI_GOLDEN_UPDATE") != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		t.Log("已更新基准:", path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		t.Skip("没有基准图，先设置 HANABI_GOLDEN_UPDATE=1 生成")
	}
	defer f.Close()
	// PNG 编码对同一像素是确定性的，但离屏渲染有个别抗锯齿像素会不同；
	// 两边都用同一条 PNG 解码路径，按非预乘 RGBA 比较，允许少量抗锯齿差异。
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	cur, err := png.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	want, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if want.Bounds() != cur.Bounds() {
		t.Fatalf("尺寸不一致: 基准 %v，当前 %v", want.Bounds(), cur.Bounds())
	}
	diff := 0
	const tol = 20
	for y := 0; y < cur.Bounds().Dy(); y++ {
		for x := 0; x < cur.Bounds().Dx(); x++ {
			wr, wg, wb, _ := color.NRGBAModel.Convert(want.At(x, y)).(color.NRGBA).RGBA()
			r, g, b, _ := color.NRGBAModel.Convert(cur.At(x, y)).(color.NRGBA).RGBA()
			if absDiff(wr, r) > tol || absDiff(wg, g) > tol || absDiff(wb, b) > tol {
				diff++
			}
		}
	}
	limit := cur.Bounds().Dx() * cur.Bounds().Dy() / 50 // 2%
	if diff > limit {
		t.Errorf("与像素基准差异过大：%d 像素（上限 %d）；确认是有意改动后用 HANABI_GOLDEN_UPDATE=1 更新", diff, limit)
	} else if diff > 0 {
		t.Logf("像素差异 %d（在容差内）", diff)
	}
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// goldenApp 固定场景：去掉素材与过渡、固定时钟，只保留界面结构。
func goldenApp() *App {
	fixed := time.Date(2026, 11, 4, 12, 0, 0, 0, time.Local)
	a := benchApp()
	a.clock = func() time.Time { return fixed }
	a.t0 = fixed.Add(-time.Hour)
	a.pageAt = fixed.Add(-time.Hour)
	a.viewAt = fixed.Add(-time.Hour)
	a.introAt = fixed.Add(-time.Hour)
	a.SetPage(libraryPageID)
	return a
}
