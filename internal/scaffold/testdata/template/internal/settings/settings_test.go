package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaults 钉住参考项目的默认值：112%、跟随当前、动效全开。
func TestDefaults(t *testing.T) {
	s := Default()
	if s.FontPercent != 112 {
		t.Errorf("默认字号应为 112%%，实际 %.0f", s.FontPercent)
	}
	if s.StartupMode != StartupFollow {
		t.Errorf("默认启动模式应为跟随当前，实际 %d", s.StartupMode)
	}
	if !s.WallpaperPlay || !s.IntroPlay || !s.Animations {
		t.Error("默认应开启动效、开场与壁纸流动")
	}
	if s.WallpaperVis != 1 {
		t.Errorf("默认壁纸显现应为 1，实际 %.2f", s.WallpaperVis)
	}
}

// TestNormalize 越界与缺省值必须被夹回合法范围，否则会把 UI 画崩。
func TestNormalize(t *testing.T) {
	cases := []struct {
		in                Settings
		wantFont, wantVis float64
		wantStartup       Startup
	}{
		{Settings{FontPercent: 0, WallpaperVis: 0}, 112, 1, StartupFollow},
		{Settings{FontPercent: 10}, 100, 1, StartupFollow},
		{Settings{FontPercent: 300}, 125, 1, StartupFollow},
		{Settings{WallpaperVis: 0.01}, 112, 0.2, StartupFollow},
		{Settings{WallpaperVis: 2}, 112, 1, StartupFollow},
		{Settings{StartupMode: 9}, 112, 1, StartupFollow},
	}
	for _, c := range cases {
		c.in.Theme = "ganyu"
		c.in.StartupTheme = "ganyu"
		got := c.in.Normalize()
		if got.FontPercent != c.wantFont {
			t.Errorf("FontPercent %+v → %.0f，期望 %.0f", c.in, got.FontPercent, c.wantFont)
		}
		if got.WallpaperVis != c.wantVis {
			t.Errorf("WallpaperVis %+v → %.2f，期望 %.2f", c.in, got.WallpaperVis, c.wantVis)
		}
		if got.StartupMode != c.wantStartup {
			t.Errorf("StartupMode %+v → %d，期望 %d", c.in, got.StartupMode, c.wantStartup)
		}
	}
}

// TestScale 100% 基准的 26px 字在默认 112% 下应是 26×1.12。
func TestScale(t *testing.T) {
	s := Default()
	if got := s.Scale(); got < 1.119 || got > 1.121 {
		t.Errorf("默认字号乘数应为 1.12，实际 %.3f", got)
	}
	if got := s.EffectiveScrim(0.32); got < 0.319 || got > 0.321 {
		t.Errorf("显现为 1 时压暗应保持 0.32，实际 %.3f", got)
	}
	if got := s.EffectiveScrim(0.32); got != 0.32 {
		t.Errorf("显现为 1 时压暗必须原样返回，实际 %.3f", got)
	}
	s.WallpaperVis = 0.2
	if got := s.EffectiveScrim(0.32); got < 0.67 || got > 0.69 {
		t.Errorf("显现 0.2 时压暗应约 0.68，实际 %.3f", got)
	}
}

// TestRoundTrip 写出去再读回来必须一致；坏 JSON 退回默认。
func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s := Default()
	s.Theme = "yoimiya"
	s.FontPercent = 118
	s.StartupMode = StartupFixed
	s.StartupTheme = "zhongli"
	s.WallpaperPlay = false
	s.IntroPlay = false
	s.Animations = false
	s.WallpaperVis = 0.6

	if err := SaveTo(path, s); err != nil {
		t.Fatalf("保存失败：%v", err)
	}
	got := LoadFrom(path)
	if got != s {
		t.Errorf("读回不一致：\n got=%+v\nwant=%+v", got, s)
	}

	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := LoadFrom(path); got != Default() {
		t.Errorf("坏 JSON 应退回默认值，实际 %+v", got)
	}

	if got := LoadFrom(filepath.Join(dir, "missing.json")); got != Default() {
		t.Errorf("缺文件应退回默认值，实际 %+v", got)
	}
}
