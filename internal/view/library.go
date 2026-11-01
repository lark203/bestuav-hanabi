package view

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

// libraryPage 主题库：三列卡片网格，点一张就切换整套主题。
//
// 卡片全部来自 theme.All()，这里不含任何角色名或颜色常量——
// 加角色只改主题包，页面不用动（这就是这一层抽象要保证的事）。
func (a *App) libraryPage(c *ui.Context) {
	ui.Text(c, "主题库").FontSize(26).Bold().TextColor(a.tok.Text)
	ui.Text(c, "换一套主题，配色、氛围与开场标识一起变").FontSize(12).TextColor(a.tok.TextMuted)

	const cols = 3
	all := theme.All()
	for i := 0; i < len(all); i += cols {
		ui.Row(c).Gap(14).Children(func() {
			for j := i; j < i+cols && j < len(all); j++ {
				a.characterCard(c, all[j])
			}
		})
	}
}

// characterCard 单张角色卡片：毛玻璃底 + 元素色点，当前主题用元素色描边并加 tint。
func (a *App) characterCard(c *ui.Context, ch theme.Character) {
	active := ch.ID == a.char.ID

	// 毛玻璃材质：玻璃的底是壁纸，所以"选中"只能靠 Tint 与描边表达，不能换底色。
	m := glass.Glass{Style: glass.Regular, Clarity: 0.45}
	if active {
		m.Tint = ch.Tok.Accent.Alpha(0.22)
	}
	card := ui.Column(c).Width(246).Height(136).Padding(16).Gap(8).
		Radius(a.tok.Radius).Material(m).Border(1, a.tok.CardStroke)
	if card.Hovered() && !active {
		m.Tint = ch.Tok.Accent.Alpha(0.15)
		card.Material(m)
	}
	if active {
		card.Border(1.5, ch.Tok.Accent)
	}
	if card.Clicked() {
		a.setCharacter(ch.ID)
	}

	card.Children(func() {
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			// 元素色点：一眼看出这套主题的强调色。
			ui.Row(c).Width(10).Height(10).Radius(5).Background(ch.Tok.Accent)
			ui.Text(c, ch.NameCN).FontSize(17).Bold().TextColor(a.tok.Text)
			if active {
				ui.Text(c, "当前").FontSize(9).TextColor(ch.Tok.Accent)
			}
		})
		ui.Text(c, ch.Element.Label()).FontSize(11).TextColor(ch.Tok.Accent)
		ui.Spacer(c)
		ui.Text(c, "开场 6.2s · "+motifName(ch.Motif)).FontSize(10).TextColor(a.tok.TextFaint)
	})
}

// motifName 是标识画法的中文说明，卡片上标出来便于看出主题差异。
func motifName(m theme.Motif) string {
	switch m {
	case theme.MotifFlame:
		return "焰火标识"
	case theme.MotifBolt:
		return "雷光标识"
	default:
		return "冰晶标识"
	}
}

// placeholderPage 尚未实现的页面。
func (a *App) placeholderPage(c *ui.Context) {
	title := "未命名"
	for _, it := range nav {
		if it.id == a.page {
			title = it.label
		}
	}
	ui.Text(c, title).FontSize(22).Bold().TextColor(a.tok.Text)
	ui.Text(c, "该页面尚未实现（骨架阶段）。").FontSize(12).TextColor(a.tok.TextFaint)
}
