package view

import (
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
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

// characterCard 单张角色卡片：用该角色的壁纸做缩略图，压暗后叠文字。
//
// 主题库的卡片不追求"玻璃感"，而是要让人一眼看出这套主题长什么样——
// 所以缩略图铺底，比半透明玻璃更能表达差异。
func (a *App) characterCard(c *ui.Context, ch theme.Character) {
	active := ch.ID == a.char.ID

	card := ui.Column(c).Width(246).Height(136).Padding(16).Gap(8).
		Radius(a.tok.Radius).Background(a.tok.Card).Border(1, a.tok.CardStroke)

	// 缩略图在自身背景之后、子元素之前绘制：先 Cover 裁切铺满，再压暗 + 元素色薄染。
	card.Draw(func(p *ui.Painter, r ui.Rect) {
		b := assets.Thumb(ch.Bg, 360)
		if b == nil {
			return
		}
		p.Clip(r, a.tok.Radius, func() {
			p.Image(b, r, ui.Cover)
		})
		// 压暗：卡片上的文字必须能读（有的壁纸是浅色立绘）。
		p.Fill(r, ui.RGBA(8, 12, 20, 0.72), a.tok.Radius)
		// 元素色薄一层，六张卡片各有各的调子。
		p.Fill(r, ch.Tok.Accent.Alpha(0.16), a.tok.Radius)
	})

	if card.Hovered() && !active {
		card.Border(1, ch.Tok.Accent.Alpha(0.75))
	}
	if active {
		card.Border(1.5, ch.Tok.Accent)
	}
	if card.Clicked() {
		a.setCharacter(ch.ID)
	}

	card.Children(func() {
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Width(10).Height(10).Radius(5).Background(ch.Tok.Accent)
			ui.Text(c, ch.NameCN).FontSize(17).Bold().TextColor(ui.Hex("#FFFFFF"))
			if active {
				ui.Text(c, "当前").FontSize(9).TextColor(ch.Tok.Accent)
			}
		})
		ui.Text(c, ch.Element.Label()).FontSize(11).TextColor(ch.Tok.Accent)
		ui.Spacer(c)
		ui.Text(c, "开场 6.2s · "+motifName(ch.Motif)).FontSize(10).
			TextColor(ui.RGBA(255, 255, 255, 0.62))
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
