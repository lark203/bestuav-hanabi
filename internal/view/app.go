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
	return &App{char: c, tok: c.Tok, page: "library", t0: time.Now(), animPlay: 1}
}

// setCharacter 切换角色主题：令牌与标识画法一起换，界面下一帧即生效。
func (a *App) setCharacter(id string) {
	c, ok := theme.ByID(id)
	if !ok || c.ID == a.char.ID {
		return
	}
	a.char = c
	a.tok = c.Tok
}

// SetWindow 由 main 注入窗口句柄。
func (a *App) SetWindow(w *mygo.Window) { a.win = w }

// 壁纸与压暗层。壁纸只解一次（assets 内部缓存），压暗层保证半透明 UI 上文字可读。
var (
	wallpaperPath  = "ganyu/ganyu-bg-4k.jpg"
	wallpaperShade = ui.RGBA(6, 12, 22, 0.38)

	// 顶部压深带：壁纸上部常有建筑、灯火的亮部，大标题压在上面读不清。
	// 从顶部 0.55 的黑蓝渐隐到透明（170 DIP 处收干净），既托住标题又不压掉通透感。
	topScrim = ui.LinearGradient{
		From:  ui.RGBA(6, 12, 22, 0.55),
		To:    ui.Transparent,
		Angle: 180, // 从上到下
	}
	topScrimH = float32(170)
)

// View 每帧构建根视图：壁纸在最底 → 中间区域（侧栏 + 右侧内容）→ 整宽底部状态栏。
func (a *App) View(c *ui.Context) {
	t := a.tok.UITheme()
	c.SetTheme(t)

	// 壁纸画在根容器的 Draw 回调里：它在自身背景之后、子元素之前执行，
	// 所以 UI 天然压在壁纸上，不需要额外的层叠容器。
	root := ui.Column(c).Fill().Background(a.tok.WinFill)
	if bg := assets.Bitmap(wallpaperPath); bg != nil {
		root = root.Draw(func(p *ui.Painter, r ui.Rect) {
			// 用帧时间而不是 time.Now()：重绘帧不重建 view，时间也来自同一个时钟。
			elapsed := p.Now().Sub(a.t0).Seconds()
			// 着色器驱动：Cover 铺满 + 缓慢流动 + 按角色元素色保亮度染色。
			wallpaper.Draw(p, bg, r, elapsed, a.animPlay, a.tok.Accent, a.char.WallTint)
			p.Fill(r, wallpaperShade, 0) // 整体压暗

			// 顶部再压一层渐变，托住标题区（在玻璃条之下，不影响玻璃本身的通透）。
			p.FillGradient(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: topScrimH}, topScrim, 0)

			// 下一帧：只重绘这块，不重建元素树。
			p.AnimationFrameRate(animFPS)
		})
	}

	root.Children(func() {
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			a.sidebar(c)
			ui.Row(c).Width(1).Background(a.tok.CardStroke)
			ui.Column(c).Grow(1).Children(func() {
				a.titlebar(c)
				a.content(c)
			})
		})
		a.statusbar(c)
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

// sidebar 左侧栏：品牌（头像 + 双行标题）+ 菜单。
//
// 侧栏用 glass.Clear：大面积面板用 Regular 会把壁纸糊得太狠，
// Clear 只轻微磨砂，壁纸的色块和笔触还能辨认（参考图就是这层质感）。
func (a *App) sidebar(c *ui.Context) {
	ui.Column(c).Width(212).
		Material(glass.Glass{Style: glass.Clear, Clarity: 0.30}).
		Children(func() {
			ui.Row(c).Padding(18, 16, 16, 18).Gap(11).AlignItems(ui.Center).Children(func() {
				if bmp := assets.Bitmap("ganyu/ganyu-icon.png"); bmp != nil {
					ui.Image(c, bmp).Width(38).Height(38).Radius(19).Clip()
				}
				ui.Column(c).Gap(1).Children(func() {
					ui.Text(c, "花火").FontSize(17).Bold().TextColor(a.tok.Text)
					ui.Text(c, "HANABI UI").FontSize(9).TextColor(a.tok.TextFaint)
				})
			})

			ui.Column(c).PaddingX(10).Gap(2).Children(func() {
				for _, it := range nav {
					a.navRow(c, it.id, it.label, it.icon)
				}
			})
		})
}

// navRow 菜单项：左侧竖条 + 图标 + 文字；选中时整体用元素色。
func (a *App) navRow(c *ui.Context, id, label string, icon *ui.SVG) {
	active := a.page == id

	row := ui.Row(c).Height(38).Padding(0, 12, 0, 0).Gap(10).Radius(9).AlignItems(ui.Center)
	bar := ui.Transparent
	color := a.tok.TextMuted
	if active {
		row.Background(a.tok.AccentSoft)
		bar = a.tok.Accent
		color = a.tok.Accent
	}
	if row.Clicked() {
		a.page = id
	}
	row.Children(func() {
		ui.Row(c).Width(3).Height(16).Radius(2).Background(bar)
		ui.Icon(c, icon).FontSize(17).TextColor(color)
		ui.Text(c, label).FontSize(14).TextColor(color)
	})
}

// titlebar 自绘标题栏：参考图里只保留右上角窗口按钮（骨架阶段为视觉占位）。
// 同样铺玻璃，让顶部与左侧栏的材质连续。
func (a *App) titlebar(c *ui.Context) {
	ui.Row(c).Height(44).Padding(16, 12).AlignItems(ui.Center).Gap(6).
		Material(glass.Glass{Style: glass.Clear, Clarity: 0.15}).
		Children(func() {
			ui.Spacer(c)
			for _, s := range []string{"—", "□", "✕"} {
				ui.Row(c).Width(28).Height(28).Radius(7).AlignItems(ui.Center).Justify(ui.Center).
					Children(func() {
						ui.Text(c, s).FontSize(12).TextColor(a.tok.TextMuted)
					})
			}
		})
}

// content 内容区：按当前页分发。
func (a *App) content(c *ui.Context) {
	ui.Column(c).Grow(1).Clip().Children(func() {
		ui.Column(c).Padding(28, 22).Gap(18).Children(func() {
			switch a.page {
			case "library":
				a.libraryPage(c)
			default:
				a.placeholderPage(c)
			}
		})
	})
}

// statusbar 整宽底部：左角色标记、中品牌小字、右工作空间状态。
//
// 状态栏也用玻璃：它是整宽的一条，若只是裸壁纸，会和上方的玻璃侧栏之间
// 出现一条质地不同的接缝。玻璃条把上下两块的材质统一起来。
func (a *App) statusbar(c *ui.Context) {
	ui.Row(c).Height(54).Padding(18, 16).AlignItems(ui.Center).Gap(9).
		Material(glass.Glass{Style: glass.Clear, Clarity: 0.15}).
		Children(func() {
			if bmp := assets.Bitmap("ganyu/ganyu-icon.png"); bmp != nil {
				ui.Image(c, bmp).Width(22).Height(22).Radius(11).Clip()
			}
			ui.Column(c).Gap(1).Children(func() {
				ui.Text(c, a.char.NameCN+" · "+a.char.Element.Label()).FontSize(11).TextColor(a.tok.TextMuted)
				ui.Text(c, "A GENTLE FIGURE").FontSize(8).TextColor(a.tok.TextFaint)
			})
			ui.Spacer(c)
			ui.Text(c, "HANABI / PERSONAL WORKSPACE").FontSize(9).TextColor(a.tok.TextFaint)
			ui.Spacer(c)
			ui.Text(c, "本地工作空间").FontSize(10).TextColor(a.tok.TextFaint)
		})
}
