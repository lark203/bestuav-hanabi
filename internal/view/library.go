package view

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

// 卡片网格的布局常量：sidebarW/pagePadX 用于把窗口宽度换算成卡片区可用宽度。
const (
	sidebarW = 212
	pagePadX = 22
	cardMinW = 280
	cardGap  = 14
	// scrollReserve 是给纵向滚动条留的右侧空档（条宽 6 + 透气 12）。
	// 滚动条是覆盖式绘制，不给内容让位会压住最右卡片的线稿；
	// 10px 视觉上还是太挤，王工要求再拉开一档。
	scrollReserve = 18
)

// libraryPage 主题库：标题区 + 元素筛选 + 卡片网格（列数随窗口宽度自适应）。
//
// 页面里没有任何角色名或颜色常量——全部来自 theme.All()。
func (a *App) libraryPage(c *ui.Context) {
	// 标题与筛选条固定不动，只有下面的卡片滚动（参考图行为）。
	a.pageHeader(c)
	a.elementFilter(c)

	all := theme.All()
	shown := make([]theme.Character, 0, len(all))
	for _, ch := range all {
		if a.filter < 0 || int(ch.Element) == a.filter {
			shown = append(shown, ch)
		}
	}

	// 列数随窗口宽度自适应（flex-wrap 的效果）：卡片在行内撑满剩余宽度，
	// 窗口变宽列数变多、变窄列数变少，始终没有右侧的空档。
	// 2~4 列：以卡片最小舒适宽度约 280 DIP 划档。
	w, _ := c.Size()
	avail := w - sidebarW - 1 - pagePadX*2 - scrollReserve
	cols := int((avail + cardGap) / (cardMinW + cardGap))
	if cols < 2 {
		cols = 2
	}
	if cols > 4 {
		cols = 4
	}
	// 卡片用显式宽度而不是 Grow(1)：整行正好填满，最后一行不满时
	// 卡片保持同宽（Grow 会把末行那几张拉通栏）。
	cardW := (avail - cardGap*float32(cols-1)) / float32(cols)

	ui.Scroll(c).Grow(1).Children(func() {
		// 右侧让出滚动条的宽度：条覆盖绘制，卡片不能伸到它底下。
		ui.Column(c).Padding(4, scrollReserve, 10, 0).Gap(cardGap).Children(func() {
			for i := 0; i < len(shown); i += cols {
				ui.Row(c).Gap(cardGap).AlignItems(ui.Start).Children(func() {
					for j := i; j < i+cols && j < len(shown); j++ {
						a.characterCard(c, shown[j], cardW)
					}
				})
			}
		})
	})
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
func (a *App) characterCard(c *ui.Context, ch theme.Character, cardW float32) {
	active := ch.ID == a.char.ID

	card := ui.Column(c).Width(cardW).Height(142).Padding(15, 14).
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
		// 对勾徽章贴卡片最右边（参考图位置）：文字 Grow(1) 顶开。
		ui.Row(c).AlignItems(ui.Center).Gap(7).Children(func() {
			ui.Text(c, ch.NameCN+" · "+ch.Epithet).FontSize(11).
				TextColor(ui.RGBA(255, 255, 255, 0.60)).Grow(1)
			if active {
				ui.Row(c).Width(16).Height(16).Draw(func(p *ui.Painter, r ui.Rect) {
					cx, cy := r.X+r.W/2, r.Y+r.H/2
					circle(p, cx, cy, r.W/2, ch.Tok.Accent)
					seg(p, cx-3.4, cy+0.2, cx-1.0, cy+2.8, 1.6, ui.Hex("#0A1420"))
					seg(p, cx-1.0, cy+2.8, cx+3.6, cy-2.6, 1.6, ui.Hex("#0A1420"))
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
