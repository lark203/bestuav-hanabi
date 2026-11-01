package view

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/theme"
)

// libraryPage 主题库：标题区 + 元素筛选 + 三列角色卡片。
//
// 页面里没有任何角色名或颜色常量——全部来自 theme.All()。
func (a *App) libraryPage(c *ui.Context) {
	a.pageHeader(c)
	a.elementFilter(c)

	all := theme.All()
	shown := make([]theme.Character, 0, len(all))
	for _, ch := range all {
		if a.filter < 0 || int(ch.Element) == a.filter {
			shown = append(shown, ch)
		}
	}

	const cols = 3
	for i := 0; i < len(shown); i += cols {
		ui.Row(c).Gap(14).Children(func() {
			for j := i; j < i+cols && j < len(shown); j++ {
				a.characterCard(c, shown[j])
			}
		})
	}
}

// pageHeader 标题行：大标题 + 英文小标 + 副标题，右上角是分段控件。
func (a *App) pageHeader(c *ui.Context) {
	ui.Row(c).AlignItems(ui.End).Children(func() {
		ui.Column(c).Gap(3).Grow(1).Children(func() {
			ui.Row(c).AlignItems(ui.End).Gap(9).Children(func() {
				ui.Text(c, "主题").FontSize(27).Bold().TextColor(a.tok.Text)
				ui.Text(c, "APPEARANCE").FontSize(9).TextColor(a.tok.TextFaint).
					Padding(0, 0, 4, 0)
			})
			ui.Text(c, "给工作空间，一点自己的气质。").FontSize(12).TextColor(a.tok.TextMuted)
		})
		a.segmented(c, []string{"主题库", "外观与启动"}, 0)
	})
}

// segmented 画一个分段控件（当前只有第一段是可用页，第二段是占位）。
func (a *App) segmented(c *ui.Context, items []string, active int) {
	row := ui.Row(c).Padding(3).Radius(10).Gap(2).
		Background(ui.RGBA(255, 255, 255, 0.07))
	row.Children(func() {
		for i, it := range items {
			on := i == active
			seg := ui.Row(c).Height(26).Padding(0, 15).Radius(7).
				AlignItems(ui.Center).Justify(ui.Center)
			color := a.tok.TextMuted
			if on {
				seg.Background(a.tok.AccentSoft)
				color = a.tok.Text
			}
			seg.Children(func() {
				ui.Text(c, it).FontSize(11).TextColor(color)
			})
		}
	})
}

// elementFilter 元素筛选条：全部 / 冰 / 水 / 雷 / 火 / 草 / 风 / 岩。
func (a *App) elementFilter(c *ui.Context) {
	order := []theme.Element{
		theme.Cryo, theme.Hydro, theme.Electro, theme.Pyro,
		theme.Dendro, theme.Anemo, theme.Geo,
	}
	count := 0
	for _, ch := range theme.All() {
		if a.filter < 0 || int(ch.Element) == a.filter {
			count++
		}
	}

	row := ui.Row(c).Height(40).Padding(4).Radius(10).AlignItems(ui.Center).Gap(2).
		Background(ui.RGBA(255, 255, 255, 0.05))
	row.Children(func() {
		ui.Text(c, "元素").FontSize(10).TextColor(a.tok.TextFaint).Padding(0, 10)

		a.filterChip(c, "全部", a.filter < 0, func() { a.filter = -1 })
		for _, el := range order {
			el := el
			a.filterChip(c, el.Short(), a.filter == int(el), func() { a.filter = int(el) })
		}

		ui.Spacer(c)
		ui.Text(c, fmt.Sprintf("%d 个角色主题", count)).
			FontSize(10).TextColor(a.tok.TextFaint).Padding(0, 12, 0, 0)
	})
}

// filterChip 是筛选条上的一个胶囊。
func (a *App) filterChip(c *ui.Context, label string, on bool, onClick func()) {
	chip := ui.Row(c).Height(26).Padding(0, 14).Radius(8).
		AlignItems(ui.Center).Justify(ui.Center)
	color := a.tok.Text
	if on {
		chip.Background(a.tok.AccentSoft)
		color = a.tok.Accent
	}
	if chip.Clicked() {
		onClick()
	}
	chip.Children(func() {
		ui.Text(c, label).FontSize(11).TextColor(color)
	})
}

// characterCard 单张角色卡片：壁纸缩略图铺底，左上角元素标识，中间英文名，
// 底部「中文名·属性」；当前主题用元素色描边并在右下角打勾。
func (a *App) characterCard(c *ui.Context, ch theme.Character) {
	active := ch.ID == a.char.ID

	card := ui.Column(c).Width(246).Height(150).Padding(13).
		Radius(a.tok.Radius).Border(1, a.tok.CardStroke)
	if active {
		card.Border(1.6, ch.Tok.Accent)
		card.Draw(func(p *ui.Painter, r ui.Rect) {
			a.cardBacking(p, r, ch)
			// 选中态：元素色薄罩一层
			p.Fill(r, ch.Tok.Accent.Alpha(0.10), a.tok.Radius)
		})
	} else {
		card.Draw(func(p *ui.Painter, r ui.Rect) { a.cardBacking(p, r, ch) })
		if card.Hovered() {
			card.Border(1.4, ch.Tok.Accent.Alpha(0.8))
		}
	}
	if card.Clicked() {
		a.setCharacter(ch.ID)
	}

	card.Children(func() {
		ui.Row(c).AlignItems(ui.Start).Children(func() {
			// 左上：元素标识
			ui.Row(c).Width(36).Height(36).Draw(func(p *ui.Painter, r ui.Rect) {
				cx, cy := r.X+r.W/2, r.Y+r.H/2
				glow(p, cx, cy, r.W*0.62, ch.Tok.Accent, 0.55)
				glyph(p, ch.Motif, cx, cy, r.W*0.36, ch.Tok.Accent, 1, 0)
			})
			ui.Spacer(c)
			// 右上：同一个标识放大、极淡，当装饰纹样用
			ui.Row(c).Width(58).Height(58).Draw(func(p *ui.Painter, r ui.Rect) {
				glyph(p, ch.Motif, r.X+r.W/2, r.Y+r.H/2, r.W*0.40,
					ui.Hex("#FFFFFF"), 0.24, 1)
			})
		})

		ui.Spacer(c)

		ui.Text(c, ch.NameEN).FontSize(22).TextColor(ui.Hex("#FFFFFF"))
		ui.Text(c, ch.NameCN+" · "+ch.Epithet).FontSize(11).
			TextColor(ui.RGBA(255, 255, 255, 0.72))

		if active {
			// 右下角的勾：一圈元素色 + 两条线
			ui.Row(c).Height(0).DrawOver(func(p *ui.Painter, r ui.Rect) {
				cx, cy := r.X+r.W-13, r.Y+r.H-13
				circle(p, cx, cy, 8, ch.Tok.Accent.Alpha(0.95))
				seg(p, cx-3.4, cy+0.2, cx-0.9, cy+2.8, 1.7, ui.Hex("#0A1420"))
				seg(p, cx-0.9, cy+2.8, cx+3.6, cy-2.6, 1.7, ui.Hex("#0A1420"))
			})
		}
	})
}

// cardBacking 画卡片底：壁纸缩略图 + 压暗 + 底部渐变（保证英文名可读）。
func (a *App) cardBacking(p *ui.Painter, r ui.Rect, ch theme.Character) {
	b := assets.Thumb(ch.Bg, 360)
	if b == nil {
		p.Fill(r, a.tok.Card, a.tok.Radius)
		return
	}
	p.Clip(r, a.tok.Radius, func() {
		p.Image(b, r, ui.Cover)
	})
	p.Fill(r, ui.RGBA(6, 10, 16, 0.55), a.tok.Radius)
	// 下半部再压一层，让大号英文名压得住亮壁纸
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y + r.H*0.42, W: r.W, H: r.H * 0.58},
		ui.LinearGradient{
			From:  ui.RGBA(6, 10, 16, 0.14),
			To:    ui.RGBA(6, 10, 16, 0.74),
			Angle: 180,
		}, 0)
}

// motifName 是标识画法的中文说明。
func motifName(m theme.Motif) string {
	switch m {
	case theme.MotifFlame:
		return "焰火标识"
	case theme.MotifBolt:
		return "雷光标识"
	case theme.MotifGeo:
		return "岩印标识"
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
