package view

import (
	"fmt"

	"github.com/egoist/mygo/ui"

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
		ui.Row(c).Gap(14).AlignItems(ui.Start).Children(func() {
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
		Background(ui.RGBA(8, 14, 24, 0.42))
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
		Background(ui.RGBA(8, 14, 24, 0.42))
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

// characterCard 单张角色卡片：纯色底 + 右侧专属线稿，左上角元素标识，
// 中间英文名，底部「中文名·属性」；当前主题加元素色描边。
//
// 为什么不用壁纸缩略图：一张 4K 插画缩到 246×142 只剩一团色块，
// 七张并排时彼此难分；纯色 + 线稿既保证一排卡片色彩分明，也让每张有图形记忆点。
func (a *App) characterCard(c *ui.Context, ch theme.Character) {
	active := ch.ID == a.char.ID

	// Grow 占满一行；MaxWidth 兜住「最后一行只有一张卡片被拉通栏」的情况。
	card := ui.Column(c).Grow(1).MaxWidth(330).Height(142).Padding(15, 14).
		Radius(a.tok.Radius).Border(1, ui.RGBA(255, 255, 255, 0.10)).
		Background(ch.CardBg)

	card.Draw(func(p *ui.Painter, r ui.Rect) {
		// 顶部提亮一点点：纯色平涂会显得发闷，一层自上而下的微光就够了。
		p.FillGradient(r, ui.LinearGradient{
			From:  ui.RGBA(255, 255, 255, 0.07),
			To:    ui.Transparent,
			Angle: 180,
		}, a.tok.Radius)
		// 顶部内侧 1px 亮线：参考图里每张卡片上缘都有一条细细的高光，
		// 是「卡片是实体」这个质感的重要来源。
		p.Fill(ui.Rect{X: r.X + 1, Y: r.Y + 1, W: r.W - 2, H: 1},
			ui.RGBA(255, 255, 255, 0.16), 0)
		// 线稿画在右半区，越出卡片的部分由圆角裁掉。
		p.Clip(r, a.tok.Radius, func() {
			cardArt(p, ch.Art, ui.Rect{X: r.X + r.W*0.38, Y: r.Y, W: r.W * 0.66, H: r.H},
				ch.Tok.Accent)
		})
	})

	switch {
	case active:
		card.Border(2, ch.Tok.Accent)
	case card.Hovered():
		card.Border(1.4, ch.Tok.Accent.Alpha(0.75))
	}
	if card.Clicked() {
		a.setCharacter(ch.ID)
	}

	card.Children(func() {
		// 左上：元素标识（小尺寸，不抢英文名）
		ui.Row(c).Width(22).Height(22).Draw(func(p *ui.Painter, r ui.Rect) {
			glyph(p, ch.Motif, r.X+r.W/2, r.Y+r.H/2, r.W*0.40,
				ch.Tok.Accent, 0.95, 0)
		})

		ui.Spacer(c)

		ui.Text(c, ch.NameEN).FontSize(26).
			Font("Georgia, Cambria, 'Times New Roman', serif").
			TextColor(ui.Hex("#F4F7FA"))
		ui.Row(c).AlignItems(ui.Center).Gap(7).Children(func() {
			ui.Text(c, ch.NameCN+" · "+ch.Epithet).FontSize(11).
				TextColor(ui.RGBA(255, 255, 255, 0.60))
			// 选中卡的右下角对勾徽章（参考图里也有）：元素色圆 + 深色勾。
			if active {
				ui.Row(c).Width(15).Height(15).Draw(func(p *ui.Painter, r ui.Rect) {
					cx, cy := r.X+r.W/2, r.Y+r.H/2
					circle(p, cx, cy, r.W/2, ch.Tok.Accent)
					seg(p, cx-3.2, cy+0.2, cx-0.9, cy+2.6, 1.5, ui.Hex("#0A1420"))
					seg(p, cx-0.9, cy+2.6, cx+3.4, cy-2.4, 1.5, ui.Hex("#0A1420"))
				})
			}
		})
	})
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
