package view

import (
	"testing"
	"time"

	"bestuav-hanabi/internal/settings"
	"bestuav-hanabi/internal/theme"
)

// TestStartupSettings 钉住启动时的设置应用：主题、字号、开场、动效与压暗。
func TestStartupSettings(t *testing.T) {
	s := settings.Default()
	s.StartupMode = settings.StartupFixed
	s.StartupTheme = "zhongli"
	s.FontPercent = 125
	s.IntroPlay = false
	s.WallpaperPlay = false
	s.WallpaperVis = 0.5

	a := NewAppWithSettings(s)
	if a.char.ID != "zhongli" {
		t.Errorf("启动主题应为 zhongli，实际 %s", a.char.ID)
	}
	if got := a.fs(10); got < 12.49 || got > 12.51 {
		t.Errorf("125%% 下 10px 应缩放成 12.5，实际 %.2f", got)
	}
	if !a.introDone {
		t.Error("关闭开场后应直接进入主界面")
	}
	if a.animPlay != 0 {
		t.Errorf("关闭壁纸流动后 animPlay 应为 0，实际 %.1f", a.animPlay)
	}
	if !a.persist {
		t.Error("NewAppWithSettings 应开启持久化")
	}
	if got := a.settings.EffectiveScrim(0.32); got < 0.54 || got > 0.55 {
		t.Errorf("显现 0.5 时压暗应约 0.545，实际 %.3f", got)
	}
}

// TestFollowStartup 跟随当前时用上次保存的主题 id。
func TestFollowStartup(t *testing.T) {
	s := settings.Default()
	s.Theme = "yoimiya"
	s.StartupMode = settings.StartupFollow
	a := NewAppWithSettings(s)
	if a.char.ID != "yoimiya" {
		t.Errorf("跟随当前应启动 yoimiya，实际 %s", a.char.ID)
	}
}

// TestRandomStartup 随机启动永远落在角色列表里。
func TestRandomStartup(t *testing.T) {
	s := settings.Default()
	s.StartupMode = settings.StartupRandom
	for i := 0; i < 8; i++ {
		a := NewAppWithSettings(s)
		if _, ok := theme.ByID(a.char.ID); !ok {
			t.Fatalf("随机到了不存在的主题 %q", a.char.ID)
		}
	}
}

// TestThemeTransition 0.42 秒过渡：中点未完成、超时后收尾。
func TestThemeTransition(t *testing.T) {
	base := time.Now()
	a := NewApp()
	a.introDone = true
	a.clock = func() time.Time { return base }

	a.setCharacter("yoimiya")
	if !a.switching {
		t.Fatal("切换主题后应进入过渡")
	}
	a.switchAt = base.Add(-210 * time.Millisecond)
	if k := a.themeProgress(base); k <= 0.05 || k >= 0.95 {
		t.Errorf("0.21s 时应过渡到一半，实际 k=%.3f", k)
	}
	a.switchAt = base.Add(-500 * time.Millisecond)
	if k := a.themeProgress(base); k != 1 {
		t.Errorf("超过 0.42s 后应到 1，实际 %.3f", k)
	}
	if a.switching {
		t.Error("过渡结束后 switcher 应归位")
	}
	if a.tok != a.char.Tok {
		t.Error("过渡结束后令牌应回到目标主题")
	}
}

// TestPageTransition 页面入场 0.22 秒；关闭动效时直接到位。
func TestPageTransition(t *testing.T) {
	base := time.Now()
	a := NewApp()
	a.clock = func() time.Time { return base }
	a.pageAt = base.Add(-110 * time.Millisecond)
	if k := a.pageProgress(base); k <= 0.05 || k >= 0.95 {
		t.Errorf("0.11s 时页面应入场到一半，实际 k=%.3f", k)
	}
	a.pageAt = base.Add(-time.Second)
	if k := a.pageProgress(base); k != 1 {
		t.Errorf("1s 后应到 1，实际 %.3f", k)
	}
	a.settings.Animations = false
	a.pageAt = base
	if k := a.pageProgress(base); k != 1 {
		t.Errorf("关闭动效后应立即到位，实际 %.3f", k)
	}
}

// TestViewTransition 主题页两个视图切换：0.18 秒淡入，关动效直接到位。
func TestViewTransition(t *testing.T) {
	base := time.Now()
	a := NewApp()
	a.clock = func() time.Time { return base }
	a.viewAt = base.Add(-90 * time.Millisecond)
	if k := a.viewProgress(base); k <= 0.05 || k >= 0.95 {
		t.Errorf("0.09s 时视图应淡入到一半，实际 k=%.3f", k)
	}
	a.viewAt = base.Add(-time.Second)
	if k := a.viewProgress(base); k != 1 {
		t.Errorf("1s 后应到 1，实际 %.3f", k)
	}
	a.settings.Animations = false
	a.viewAt = base
	if k := a.viewProgress(base); k != 1 {
		t.Errorf("关闭动效后应立即到位，实际 %.3f", k)
	}
}
