package view

import (
	"math"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/settings"
	"bestuav-hanabi/internal/theme"
)

// appearanceView 是主题页的第二个视图：外观与启动。
//
// 信息架构按参考项目：主题库负责选角色；外观（字号、壁纸显现、动效）和
// 启动（跟随/固定/随机、开场）与当前角色强相关，留在主题页里，不跳设置。
func (a *App) appearanceView(c *ui.Context) {
	ui.Row(c).Grow(1).Gap(16).AlignItems(ui.Stretch).Children(func() {
		a.appearancePanel(c)
		a.startupPanel(c)
	})
}

// appearancePanel 左侧：场景与文字。
func (a *App) appearancePanel(c *ui.Context) {
	a.panel(c, "场景与文字", "壁纸、动效与整体字号", func() {
		// 壁纸显现：0.2–1.0，越小压暗越重。
		a.sliderRow(c, "壁纸显现", &a.settings.WallpaperVis, settings.WallpaperMin, settings.WallpaperMax,
			func(v float64) string { return percent(v) }, func(v float64) {
				a.settings.WallpaperVis = v
				a.markDirty(c)
			})

		// 背景流动：关掉后壁纸完全静止（wallpaper.Draw 的 play=0）。
		beforePlay := a.settings.WallpaperPlay
		a.switchRow(c, "背景流动", "人物的呼吸与发丝漂动", &a.settings.WallpaperPlay)
		if a.settings.WallpaperPlay != beforePlay {
			if a.settings.WallpaperPlay {
				a.animPlay = 1
			} else {
				a.animPlay = 0
			}
			a.markDirty(c)
		}

		// 开场动画：下次启动生效。
		beforeIntro := a.settings.IntroPlay
		a.switchRow(c, "启动时播放开场", "下次启动生效 · 6.2 秒", &a.settings.IntroPlay)
		if a.settings.IntroPlay != beforeIntro {
			a.markDirty(c)
		}

		// 整体字号：100–125，默认 112；滑杆拖到小数时取整。
		// 默认按钮放在同一行，省下一整行高度——1000×640、125% 时刚好不溢出。
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, "整体字号").FontSize(a.fs(13)).TextColor(a.tok.TextMuted).Grow(1)
			ui.Text(c, fontPercent(a.settings.FontPercent)).
				FontSize(a.fs(11)).TextColor(a.tok.TextFaint)
			if ui.Button(c, "默认").Clicked() {
				a.settings.FontPercent = settings.FontDefault
				a.markDirty(c)
			}
		})
		beforeFont := a.settings.FontPercent
		ui.Row(c).Children(func() {
			ui.Slider(c, &a.settings.FontPercent, settings.FontMin, settings.FontMax).Label("整体字号").Grow(1)
		})
		if a.settings.FontPercent != beforeFont {
			a.settings.FontPercent = math.Round(a.settings.FontPercent)
			a.markDirty(c)
		}

		ui.Spacer(c)

		ui.Text(c, "当前主题："+a.char.NameCN+" · "+a.char.Epithet).
			FontSize(a.fs(11)).TextColor(a.tok.TextMuted)
	})
}

// startupPanel 右侧：交互动效与启动主题。
func (a *App) startupPanel(c *ui.Context) {
	a.panel(c, "交互与启动", "下次启动时用哪个角色主题", func() {
		beforeAnimations := a.settings.Animations
		a.switchRow(c, "流畅动效", "页面与主题切换的过渡", &a.settings.Animations)
		if a.settings.Animations != beforeAnimations {
			if !a.settings.Animations {
				a.finishSwitch() // 正在过渡时直接到位
			}
			a.markDirty(c)
		}

		beforeReplay := a.settings.ReplayIntro
		a.switchRow(c, "换主题时重播开场", "每次切换都再看一遍 6.2 秒", &a.settings.ReplayIntro)
		if a.settings.ReplayIntro != beforeReplay {
			a.markDirty(c)
		}

		ui.Text(c, "下次启动").FontSize(a.fs(13)).TextColor(a.tok.TextMuted)
		mode := int(a.settings.StartupMode)
		ui.Row(c).Children(func() {
			ui.Segmented(c, &mode, "跟随当前", "指定主题", "随机角色").Grow(1)
		})
		if mode != int(a.settings.StartupMode) {
			a.settings.StartupMode = settings.Startup(mode)
			a.markDirty(c)
		}

		if a.settings.StartupMode == settings.StartupFixed {
			ui.Row(c).Wrap().Gap(5).Children(func() {
				for _, ch := range theme.All() {
					selected := ch.ID == a.settings.StartupTheme
					chip := ui.Row(c).Height(24).Padding(0, 9).Radius(7).
						AlignItems(ui.Center).Justify(ui.Center)
					color := a.tok.TextMuted
					if selected {
						chip.Background(a.tok.Accent.Alpha(.20))
						chip.Border(1, a.tok.Accent.Alpha(.55))
						color = a.tok.Accent
					} else {
						chip.Background(ui.RGBA(255, 255, 255, 0.06))
						chip.Border(1, a.tok.CardStroke)
					}
					if chip.Clicked() && !selected {
						a.settings.StartupTheme = ch.ID
						a.markDirty(c)
					}
					chip.Children(func() {
						ui.Text(c, ch.NameCN).FontSize(a.fs(11)).TextColor(color)
					})
				}
			})
			ui.Text(c, "启动时使用所选主题及其开场动画。").
				FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		} else if a.settings.StartupMode == settings.StartupRandom {
			ui.Text(c, "每次都从角色主题里随机一个（不含黑白基础主题）。").
				FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		} else {
			ui.Text(c, "记住上次使用的角色主题，下次从它开始。").
				FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		}

		ui.Text(c, "下次启动："+startupSummary(a.settings)).
			FontSize(a.fs(11)).TextColor(a.tok.Accent)

		ui.Spacer(c)

		ui.Text(c, "设置保存在用户目录，关闭应用后仍然有效。").
			FontSize(a.fs(10)).TextColor(a.tok.TextFaint.Alpha(.8))
	})
}

// switchRow 是一行「标签 + 说明 + Switch」；Switch 用 mygo 自带的。
func (a *App) switchRow(c *ui.Context, label, detail string, on *bool) {
	ui.Row(c).AlignItems(ui.Center).Gap(10).Children(func() {
		ui.Column(c).Gap(1).Grow(1).Children(func() {
			ui.Text(c, label).FontSize(a.fs(13)).TextColor(a.tok.Text)
			if detail != "" {
				ui.Text(c, detail).FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
			}
		})
		ui.Switch(c, on).Label(label)
	})
}

// sliderRow 是「标签 + 当前值」一行，下面跟一条 mygo Slider。
// value 直接指向设置字段，拖动时不会丢；onChange 只在值真的变了时调用。
func (a *App) sliderRow(c *ui.Context, label string, value *float64, lo, hi float64,
	format func(float64) string, onChange func(float64)) {
	ui.Row(c).AlignItems(ui.End).Children(func() {
		ui.Text(c, label).FontSize(a.fs(13)).TextColor(a.tok.TextMuted).Grow(1)
		ui.Text(c, format(*value)).FontSize(a.fs(11)).TextColor(a.tok.TextFaint)
	})
	before := *value
	ui.Row(c).Children(func() {
		ui.Slider(c, value, lo, hi).Label(label).Grow(1)
	})
	if *value != before {
		onChange(*value)
	}
}

// startupSummary 用一句人话描述下次启动会选谁。
func startupSummary(s settings.Settings) string {
	switch s.StartupMode {
	case settings.StartupFixed:
		if c, ok := theme.ByID(s.StartupTheme); ok {
			return c.NameCN + " · " + c.Epithet
		}
		return "未选择"
	case settings.StartupRandom:
		return "随机角色"
	default:
		if c, ok := theme.ByID(s.Theme); ok {
			return c.NameCN + " · " + c.Epithet
		}
		return "默认角色"
	}
}

// percent 把 0.82 显示成 82%（壁纸显现的 0..1 值）。
func percent(v float64) string {
	return itoa(int(math.Round(v*100))) + "%"
}

// fontPercent 把 112 显示成 112%（字号本身就是百分数）。
func fontPercent(v float64) string {
	return itoa(int(math.Round(v))) + "%"
}

// itoa 是 strconv.Itoa 的小替身，避免为了一个数字引入 import。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [8]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
