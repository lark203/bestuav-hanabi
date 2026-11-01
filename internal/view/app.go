// Package view 是界面根：App 持有状态，View 每帧重建元素树（mygo 是即时模式）。
package view

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/intro"
	"bestuav-hanabi/internal/theme"
	"bestuav-hanabi/internal/ui/icons"
	"bestuav-hanabi/internal/wallpaper"
)

// nav 侧栏条目：id、标题、图标。
var nav = []struct {
	id, label string
	icon      *ui.SVG
}{
	{"library", "主题库", icons.Palette},
	{"wallpaper", "壁纸", icons.Picture},
	{"settings", "设置", icons.Sliders},
}

// App 是应用状态与视图根。
type App struct {
	win  *mygo.Window
	char theme.Character // 当前角色主题包
	tok  theme.Tokens    // char.Tok 的副本，绘制时少一层解引用
	page string

	// 动效：壁纸的流动由 p.Now() 驱动，用 Painter.AnimationFrameRate 请求下一帧——
	// 它只重绘这一块、不重建整棵元素树，元素不可见时也不会请求（比自建 ticker 省得多）。
	// 12fps 足够让低频漂移看起来连续。
	t0       time.Time
	animPlay float32

	// filter 是主题库的元素筛选：-1 表示「全部」，否则是 theme.Element 的数值。
	filter int

	// 开场动画：在主界面之上叠一层，播完自行淡出并让视图重建一次。
	intro     *intro.Intro
	introAt   time.Time
	introDone bool
	notify    func() // 仅用于开场结束时请求一次重建
}

// animFPS 是壁纸动效的重绘频率：这类缓慢漂移在这个帧率下已经连续。
const animFPS = 12

// NewApp 创建应用，默认停在主题库页。
func NewApp() *App {
	c := theme.Default()
	return &App{char: c, tok: c.Tok, page: "library", t0: time.Now(), animPlay: 1, filter: -1}
}

// setCharacter 切换角色主题：令牌与标识画法一起换，界面下一帧即生效。
func (a *App) setCharacter(id string) {
	c, ok := theme.ByID(id)
	if !ok || c.ID == a.char.ID {
		return
	}
	a.char = c
	a.tok = c.Tok
	// 释放上一套的壁纸/头像像素：每张 4K 约 33MB，六套全留着会到 200MB。
	assets.Keep(c.Bg, c.Icon)
}

// SetWindow 由 main 注入窗口句柄。
func (a *App) SetWindow(w *mygo.Window) { a.win = w }

// 壁纸与压暗层。壁纸路径来自当前角色（assets 内部缓存），压暗层保证半透明 UI 上文字可读。
var (
	// 顶部压深带：壁纸上部常有建筑、灯火的亮部，大标题压在上面读不清。
	// 从顶部 0.72 的黑蓝渐隐到透明（190 DIP 处收干净），既托住标题又不压掉通透感。
	// 强度给得比较高是因为有的壁纸本身就是浅色立绘（纳西妲、八重神子）。
	topScrim = ui.LinearGradient{
		From:  ui.RGBA(6, 12, 22, 0.72),
		To:    ui.Transparent,
		Angle: 180, // 从上到下
	}
	topScrimH = float32(190)
)

// View 每帧构建根视图：壁纸在最底 → 中间区域（侧栏 + 右侧内容）→ 整宽底部状态栏。
func (a *App) View(c *ui.Context) {
	t := a.tok.UITheme()
	c.SetTheme(t)

	// 壁纸画在根容器的 Draw 回调里：它在自身背景之后、子元素之前执行，
	// 所以 UI 天然压在壁纸上，不需要额外的层叠容器。
	root := ui.Column(c).Fill().Background(a.tok.WinFill)
	if bg := assets.Bitmap(a.char.Bg); bg != nil {
		root = root.Draw(func(p *ui.Painter, r ui.Rect) {
			// 用帧时间而不是 time.Now()：重绘帧不重建 view，时间也来自同一个时钟。
			elapsed := p.Now().Sub(a.t0).Seconds()
			// 着色器驱动：Cover 铺满 + 缓慢流动 + 按角色元素色保亮度染色。
			wallpaper.Draw(p, bg, r, elapsed, a.animPlay, a.tok.Accent, a.char.WallTint)
			p.Fill(r, ui.RGBA(6, 12, 22, a.char.Scrim), 0) // 按壁纸亮度压暗（见 Character.Scrim）

			// 顶部再压一层渐变，托住标题区（在玻璃条之下，不影响玻璃本身的通透）。
			p.FillGradient(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: topScrimH}, topScrim, 0)

			// 下一帧：只重绘这块，不重建元素树。
			p.AnimationFrameRate(animFPS)
		})
	}

	root.Children(func() {
		// 状态栏放在右侧内容列里，不横穿侧栏——侧栏一直通到窗口底部，
		// 「本地工作空间」属于内容区（参考图就是这样）。
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			a.sidebar(c)
			ui.Row(c).Width(1).Background(a.tok.CardStroke)
			ui.Column(c).Grow(1).Children(func() {
				a.titlebar(c)
				a.content(c)
				a.statusbar(c)
			})
		})
	})

	// 开场：主界面照常构建，开场作为绝对定位的一层压在上面，随进度淡出。
	// 这样结束时不切视图，主界面是「透出来」的，没有突兀的跳变。
	if !a.introDone {
		if a.notify == nil {
			a.notify = c.Invalidate
		}
		ov := ui.Box(c).Absolute().Fill()
		ov.Draw(func(p *ui.Painter, r ui.Rect) { a.introLayer(p, r) })
	}
}

// elementMark 画当前角色的元素标识：品牌图标与侧栏底部共用。
//
// 开场的标识正是停靠到这个位置（introDockX/Y），两边用的是同一套 glyph，
// 所以开场结束时接手不会跳变——这就是"启动页的图标动画移过来"。
func (a *App) elementMark(c *ui.Context, size float32) {
	ui.Row(c).Width(size).Height(size).Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		glow(p, cx, cy, size*0.58, a.tok.Accent, 0.55)
		glyph(p, a.char.Motif, cx, cy, size*0.36, a.tok.Accent, 1, 0)
	})
}

// sidebar 左侧栏：品牌（元素标识 + 名称）→ 分组标签 → 菜单 → 底部当前角色。
func (a *App) sidebar(c *ui.Context) {
	col := ui.Column(c).Width(212)
	// Regular 在暗模式下会压暗底下的壁纸；Clear 太透，亮壁纸（钟离、纳西妲）
	// 上菜单文字会糊。Clarity 0.15 保留一点通透，但足够托住文字。
	col.Material(glass.Glass{Style: glass.Regular, Clarity: 0.15})
	col.Children(func() {
		ui.Row(c).Padding(16, 12, 10, 20).Gap(12).AlignItems(ui.Center).Children(func() {
			a.elementMark(c, 40)
			ui.Column(c).Gap(1).Children(func() {
				ui.Text(c, "HanaBi").FontSize(19).Bold().TextColor(a.tok.Text)
				ui.Text(c, a.char.Motto).FontSize(9).TextColor(a.tok.TextFaint)
			})
		})

		// 分组标签：原版把导航分成「工作空间」等几组。
		ui.Text(c, "工作空间").FontSize(9).TextColor(a.tok.TextFaint).
			Padding(13, 10, 6, 20)

		ui.Column(c).PaddingX(12).Gap(2).Children(func() {
			for _, it := range nav {
				a.navRow(c, it.id, it.label, it.icon)
			}
		})

		ui.Spacer(c)

		// 底部：元素标识 + 身份信息，整体水平居中（参考图里就是居中两行）。
		ui.Column(c).AlignItems(ui.Center).Gap(7).Padding(0, 16, 0, 26).Children(func() {
			a.elementMark(c, 34)
			ui.Text(c, a.char.Scene).FontSize(11).TextColor(a.tok.TextMuted)
			ui.Text(c, a.char.SceneEN).FontSize(8.5).TextColor(a.tok.TextFaint)
		})
	})
}

// navRow 菜单项：图标 + 文字，选中时底色胶囊 + 右侧标出当前主题的元素符号。
func (a *App) navRow(c *ui.Context, id, label string, icon *ui.SVG) {
	active := a.page == id

	row := ui.Row(c).Height(40).Padding(9, 12, 9, 12).Gap(11).
		Radius(9).AlignItems(ui.Center)
	color := a.tok.TextMuted
	if active {
		row.Background(a.tok.AccentSoft)
		color = a.tok.Text
	}
	if row.Clicked() {
		a.page = id
	}
	row.Children(func() {
		// 选中项左侧的竖杠：参考图里导航选中态的标志（圆角短棒，元素色）。
		ui.Row(c).Width(3).Height(18).Radius(1.5).
			Background(a.tok.Accent.Alpha(map[bool]float32{true: 0.95, false: 0}[active]))
		ui.Icon(c, icon).FontSize(17).TextColor(color)
		ui.Text(c, label).FontSize(13).TextColor(color).Grow(1)
		if active {
			ui.Row(c).Width(15).Height(15).Draw(func(p *ui.Painter, r ui.Rect) {
				glyph(p, a.char.Motif, r.X+r.W/2, r.Y+r.H/2, r.W*0.44, a.tok.Accent, 0.85, 0)
			})
		}
	})
}

// titlebar 自绘标题栏：整条可拖动窗口，右侧是最小化 / 最大化 / 关闭。
//
// 系统标题栏与窗口按钮已由 main.go 关掉（TitleBarHidden + HideWindowButtons），
// 否则会和我们这一条叠成两层。双击标题栏最大化由 DragWindow 自带。
func (a *App) titlebar(c *ui.Context) {
	bar := ui.Row(c).Height(40).Padding(0, 10).AlignItems(ui.Center).Gap(4)
	bar.Children(func() {
		ui.Spacer(c)
		a.winButton(c, icons.Minimize, "最小化", func() {
			if a.win != nil {
				a.win.Minimize()
			}
		})
		a.winButton(c, icons.Maximize, "最大化 / 还原", func() {
			if a.win == nil {
				return
			}
			if a.win.IsMaximized() {
				a.win.Unmaximize()
			} else {
				a.win.Maximize()
			}
		})
		a.winButton(c, icons.Close, "关闭", func() {
			if a.win != nil {
				a.win.Close()
			}
		})
	})
	bar.DragWindow()

	// 标题栏中间的工作室小标：用 DrawOver 精确居中，
	// 不影响左右两端的布局（右侧是窗口按钮）。
	bar.DrawOver(func(p *ui.Painter, r ui.Rect) {
		const label = "HANABI / STUDIO"
		w, h := p.MeasureText(0, ui.Span{Text: label, Size: 9})
		p.Text(r.X+(r.W-w)/2, r.Y+(r.H-h)/2, label, 9, a.tok.TextFaint)
	})
}

// winButton 是一个窗口控件按钮：悬停变亮，点击执行 action。
//
// 图标用 SVG 而不是 "—" "□" "✕" 这些字符：字体里不一定有这些字形，
// 缺字形时按钮会变成一个空方块（试过）。
func (a *App) winButton(c *ui.Context, icon *ui.SVG, tip string, action func()) {
	btn := ui.Row(c).Width(32).Height(26).Radius(6).
		AlignItems(ui.Center).Justify(ui.Center)
	if btn.Hovered() {
		btn.Background(ui.RGBA(255, 255, 255, 0.16))
	}
	btn.Tooltip(tip)
	if btn.Clicked() {
		action()
	}
	btn.Children(func() {
		ui.Icon(c, icon).FontSize(16).TextColor(a.tok.Text)
	})
}

// content 内容区：按当前页分发。
//
// 整块内容区压一层很淡的暗色：壁纸在这里是最花的一段，标题、筛选条、
// 卡片说明都直接压在上面，不压一层的话浅色壁纸会把文字吃掉。
// 内容套在 Scroll 里：主题超出可视高度时纵向滚动（状态栏才不会被顶出窗口）。
func (a *App) content(c *ui.Context) {
	ui.Column(c).Grow(1).Clip().
		Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(r, ui.RGBA(6, 11, 18, 0.30), 0)
		}).
		Children(func() {
			ui.Scroll(c).Grow(1).Children(func() {
				ui.Column(c).Padding(28, 22).Gap(18).Children(func() {
					switch a.page {
					case "library":
						a.libraryPage(c)
					default:
						a.placeholderPage(c)
					}
				})
			})
		})
}

// statusbar 内容区底部：中间一行品牌小字，右侧是工作空间状态。
//
// 它只属于右侧内容列（不通到侧栏），高度也收窄到 34——参考图里它是一条
// 很轻的信息条，不是一条横贯全屏的横幅。
func (a *App) statusbar(c *ui.Context) {
	// 不铺任何材质、不加分隔线：参考图里它只是壁纸上的两行淡字，
	// 压一层玻璃反而在亮壁纸上亮出一条横带。
	ui.Row(c).Height(34).Padding(0, 16).AlignItems(ui.Center).
		Children(func() {
			ui.Spacer(c)
			ui.Text(c, "HANABI / PERSONAL WORKSPACE").FontSize(8.5).
				TextColor(a.tok.TextFaint.Alpha(0.72))
			ui.Spacer(c)
			ui.Text(c, "本地工作空间").FontSize(9.5).
				TextColor(a.tok.TextFaint.Alpha(0.80))
		})
}
