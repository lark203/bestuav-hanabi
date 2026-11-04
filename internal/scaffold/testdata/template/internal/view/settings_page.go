package view

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/appmeta"
	"bestuav-hanabi/internal/settings"
)

// settingsPage 是应用级设置：版本与更新、数据与存储、关于。
//
// 外观与启动不在这里——它们跟当前角色强相关，放在主题页的第二个视图
// （appearance.go）。这样侧栏「设置」不会再和主题页的分段控件抢同一个
// 页面，设置也能继续长更新、存储、快捷键这些内容。
func (a *App) settingsPage(c *ui.Context) {
	ui.Row(c).AlignItems(ui.End).Children(func() {
		ui.Column(c).Gap(3).Grow(1).Children(func() {
			ui.Row(c).AlignItems(ui.End).Gap(9).Children(func() {
				ui.Text(c, "设置").FontSize(a.fs(27)).Bold().TextColor(a.tok.Text)
				ui.Text(c, "PREFERENCES").FontSize(a.fs(9)).TextColor(a.tok.TextFaint).
					Padding(0, 0, 4, 0)
			})
			ui.Text(c, "版本、更新、数据与关于。").FontSize(a.fs(12)).TextColor(a.tok.TextMuted)
		})
	})

	ui.Row(c).Grow(1).Gap(16).AlignItems(ui.Stretch).Children(func() {
		a.versionPanel(c)
		a.dataPanel(c)
	})
}

// versionPanel 左侧：版本与更新。
func (a *App) versionPanel(c *ui.Context) {
	a.panel(c, "版本与更新", "当前构建与自动更新状态", func() {
		ui.Text(c, appmeta.Title).FontSize(a.fs(19)).Bold().TextColor(a.tok.Text)
		ui.Text(c, "版本 "+appVersion()).FontSize(a.fs(11)).TextColor(a.tok.TextMuted)
		ui.Text(c, "以角色主题为核心的桌面 UI。").
			FontSize(a.fs(11)).TextColor(a.tok.TextFaint)

		ui.Spacer(c)

		ui.Text(c, "自动更新").FontSize(a.fs(13)).TextColor(a.tok.Text)
		ui.Text(c, a.updateStatusLine()).FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		a.updateControls(c)
		if !a.updateEnabled() {
			ui.Text(c, "开发和测试构建不联网检查；发布构建会带着更新公钥与更新源。").
				FontSize(a.fs(10)).TextColor(a.tok.TextFaint.Alpha(.8))
		}
	})
}

// dataPanel 右侧：设置文件、恢复默认与关于。
func (a *App) dataPanel(c *ui.Context) {
	a.panel(c, "数据与关于", "设置保存在用户目录，关闭应用后仍然有效", func() {
		path, err := settings.DefaultPath()
		if err != nil || path == "" {
			path = "（无法定位用户目录）"
		}

		ui.Text(c, "设置文件").FontSize(a.fs(13)).TextColor(a.tok.Text)
		ui.Text(c, path).FontSize(a.fs(10)).TextColor(a.tok.TextFaint).
			SingleLine().Ellipsis("…").Tooltip(path)

		ui.Row(c).Gap(8).Wrap().Children(func() {
			if ui.Button(c, "复制路径").Clicked() {
				c.WriteClipboard(path)
			}
			if ui.Button(c, "恢复默认设置").Clicked() {
				a.resetSettings(c)
			}
			if ui.Button(c, "导出配置…").Clicked() {
				a.exportProfile()
			}
			if ui.Button(c, "导入配置…").Clicked() {
				a.importProfile()
			}
			if ui.Button(c, "导出诊断包").Clicked() {
				a.exportDiagnostics()
			}
			if ui.Button(c, "导入素材包…").Clicked() {
				a.importAssetPack()
			}
		})
		if a.notice != "" {
			ui.Text(c, a.notice).FontSize(a.fs(10)).TextColor(a.tok.Accent)
		}
		ui.Text(c, "恢复默认会重置字号、壁纸显现、动效与启动方式，保留当前主题。").
			FontSize(a.fs(10)).TextColor(a.tok.TextFaint.Alpha(.8))

		ui.Spacer(c)

		ui.Text(c, "关于").FontSize(a.fs(13)).TextColor(a.tok.Text)
		ui.Text(c, "角色名称与相关视觉元素归原版权方所有，本项目仅作技术学习与本地演示。").
			FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		ui.Row(c).Children(func() {
			ui.Link(c, "参考 HanaBi-ui-Genshin",
				"https://github.com/xiaoyeuzixi/HanaBi-ui-Genshin").FontSize(a.fs(11))
		})
	})
}

// panel 是设置类页面的一块面板：标题 + 副标题 + 内容。
func (a *App) panel(c *ui.Context, title, sub string, body func()) {
	ui.Column(c).Grow(1).MaxHeight(450).Padding(18, 16).Gap(12).Radius(a.tok.Radius).
		Border(1, a.tok.CardStroke).Background(a.tok.WinFill.Mix(ui.Hex("#05080E"), .30)).
		Children(func() {
			ui.Text(c, title).FontSize(a.fs(15)).Bold().TextColor(a.tok.Text)
			if sub != "" {
				ui.Text(c, sub).FontSize(a.fs(11)).TextColor(a.tok.TextFaint)
			}
			body()
		})
}

// resetSettings 恢复默认，但保留当前正在看的角色主题，避免界面突然换人。
func (a *App) resetSettings(c *ui.Context) {
	id := a.char.ID
	a.settings = settings.Default()
	a.settings.Theme = id
	a.settings.StartupTheme = id
	a.animPlay = 1
	a.markDirty(c)
}

// appVersion 读 mygo.json 里的版本；测试/未打包时给一个可读的兜底。
func appVersion() string {
	if v := mygo.App.Version(); v != "" {
		return v
	}
	// go test 下没有 mygo.json；mygo dev/build 会读到真实版本。
	return "0.1.0"
}
