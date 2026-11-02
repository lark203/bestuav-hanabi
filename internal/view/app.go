// Package view 是界面根：App 持有状态，View 每帧重建元素树（mygo 是即时模式）。
package view

import (
	"math/rand/v2"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/intro"
	"bestuav-hanabi/internal/settings"
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
	tok  theme.Tokens    // 当前显示用的令牌（切换时是 from→char 的过渡值）
	page string

	// settings 是跨启动保存的界面设置；persist=false 时不写盘（测试用）。
	settings settings.Settings
	persist  bool
	dirty    bool
	dirtyAt  time.Time
	// clock 让过渡在测试里可复现；默认 time.Now。
	clock func() time.Time

	// 动效：壁纸的流动由 p.Now() 驱动，用 Painter.AnimationFrameRate 请求下一帧——
	// 它只重绘这一块、不重建整棵元素树，元素不可见时也不会请求（比自建 ticker 省得多）。
	// 12fps 足够让低频漂移看起来连续。
	t0       time.Time
	animPlay float32

	// filter 是主题库的元素筛选：-1 表示「全部」，否则是 theme.Element 的数值。
	filter int

	// themeView 是主题页的两个视图：0 主题库 / 1 外观与启动。
	themeView int
	// viewAt 是当前视图出现的时间，切换视图时内容淡入。
	viewAt time.Time

	// pageAt 是当前页第一次出现的时间，用来做 0.22 秒的入场位移与淡入。
	pageAt time.Time

	// 主题切换交叉淡入：旧主题的壁纸/令牌保留 0.42 秒，与新主题叠化。
	fromChar    theme.Character
	fromTok     theme.Tokens
	fromBg      *ui.Bitmap
	fromFlow    *wallpaper.Flow
	switching   bool
	switchAt    time.Time
	pendingKeep []string

	// 开场动画：在主界面之上叠一层，播完自行淡出并让视图重建一次。
	intro     *intro.Intro
	introAt   time.Time
	introDone bool
	notify    func() // 仅用于开场结束时请求一次重建

	// 自动更新的检查/安装状态；网络与文件操作在 update.go 里放到 goroutine。
	upd updateState
}

// animFPS 是壁纸动效的重绘频率：这类缓慢漂移在这个帧率下已经连续。
const animFPS = 12

// 界面动效时长。
const (
	themeFade = 420 * time.Millisecond // 参考 Shell::SetTheme 的 0.42s
	pageFade  = 220 * time.Millisecond // 页面入场：位移 + 淡入
	viewFade  = 180 * time.Millisecond // 主题页两个视图切换：淡入
	saveDelay = 300 * time.Millisecond // 设置落盘防抖
)

// NewApp 用默认设置创建应用，默认停在主题库页（测试与无配置启动都走这里）。
func NewApp() *App { return newApp(settings.Default(), false) }

// NewAppWithSettings 按持久化设置选择启动主题、字号与动效；main 用它。
func NewAppWithSettings(s settings.Settings) *App { return newApp(s.Normalize(), true) }

func newApp(s settings.Settings, persist bool) *App {
	c := theme.Default()
	switch s.StartupMode {
	case settings.StartupFixed:
		if fixed, ok := theme.ByID(s.StartupTheme); ok {
			c = fixed
		}
	case settings.StartupRandom:
		all := theme.All()
		if len(all) > 0 {
			c = all[rand.IntN(len(all))]
		}
	default:
		if last, ok := theme.ByID(s.Theme); ok {
			c = last
		}
	}

	play := float32(0)
	if s.WallpaperPlay {
		play = 1
	}
	a := &App{
		char:      c,
		tok:       c.Tok,
		page:      "library",
		settings:  s,
		persist:   persist,
		t0:        time.Now(),
		clock:     time.Now,
		animPlay:  play,
		filter:    -1,
		viewAt:    time.Now().Add(-time.Hour), // 启动视图不做淡入
		pageAt:    time.Now().Add(-time.Hour), // 启动页不做入场动画
		introDone: !s.IntroPlay,
		upd:       updateState{events: make(chan updateEvent, 4)},
	}
	a.settings.Theme = c.ID
	if _, ok := theme.ByID(a.settings.Theme); !ok {
		a.settings.Theme = c.ID
	}
	if _, ok := theme.ByID(a.settings.StartupTheme); !ok {
		a.settings.StartupTheme = c.ID
	}
	return a
}

// fs 把 100% 基准的字号换算成当前字号百分比。
//
// 参考项目所有文字都过 Fs()，100% 是设计稿尺寸、默认 112%；这里保持同样的
// 语义：文字尺寸变，布局尺寸不跟着变（卡片高度、内边距仍是设计稿）。
func (a *App) fs(base float32) float32 {
	return base * a.settings.Scale()
}

// markDirty 标记设置已改，并在 300ms 后落盘；滑杆拖动时不会每帧写文件。
func (a *App) markDirty(c *ui.Context) {
	a.dirty = true
	a.dirtyAt = a.clock()
	if c != nil {
		c.After(saveDelay)
		c.Invalidate()
	}
}

// saveTick 在防抖时间到后写入设置；persist=false 时什么也不做。
func (a *App) saveTick(c *ui.Context) {
	if !a.persist || !a.dirty {
		return
	}
	if d := a.clock().Sub(a.dirtyAt); d < saveDelay {
		c.After(saveDelay - d)
		c.Invalidate()
		return
	}
	_ = settings.Save(a.settings)
	a.dirty = false
}

// FlushSettings 立即落盘（退出前调用，避免最后 300ms 的拖动丢失）。
func (a *App) FlushSettings() {
	if !a.persist || !a.dirty {
		return
	}
	_ = settings.Save(a.settings)
	a.dirty = false
}

// wallpaperFlowFor 构造指定主题的壁纸动态素材；没有人形遮罩的角色返回 nil。
//
// 各遮罩走 assets 缓存（缺图返回 nil，对应层自动不参与或整体退化为静态）。
func (a *App) wallpaperFlowFor(c theme.Character) *wallpaper.Flow {
	if c.PersonMask == "" {
		return nil
	}
	person := assets.Bitmap(c.PersonMask)
	if person == nil {
		return nil
	}
	f := &wallpaper.Flow{
		Person: person,
		Breath: c.PersonBreath,
		Sway:   c.HairSway,
	}
	if c.HairVeil != "" {
		f.Veil = assets.Bitmap(c.HairVeil)
	}
	if c.HairWave != "" {
		f.Wave = assets.Bitmap(c.HairWave)
	}
	return f
}

// wallpaperFlow 是当前主题的动态素材（开场/正常绘制用）。
func (a *App) wallpaperFlow() *wallpaper.Flow { return a.wallpaperFlowFor(a.char) }

// wallpaperKeys 是 Keep 要保留的资源：当前壁纸、头像与动态遮罩。
func wallpaperKeys(c theme.Character) []string {
	keep := []string{c.Bg, c.Icon}
	for _, p := range []string{c.PersonMask, c.HairVeil, c.HairWave} {
		if p != "" {
			keep = append(keep, p)
		}
	}
	return keep
}

// setCharacter 切换角色主题：令牌、窗口按钮造型、纹样和壁纸一起换。
//
// 有动效时走 0.42 秒交叉淡入：旧的壁纸/令牌留在 from* 里，新主题以
// 透明度 k 叠在上面，同时 a.tok 在两组令牌之间插值。旧壁纸的像素留到
// 过渡结束再释放（pendingKeep），否则第一帧会重新解码一张 4K 图。
func (a *App) setCharacter(id string) {
	c, ok := theme.ByID(id)
	if !ok || c.ID == a.char.ID {
		return
	}
	// 上一次过渡还没走完就切下一个：先收尾，避免旧壁纸像素一直留在缓存里。
	if a.switching {
		a.finishSwitch()
	}

	a.fromChar = a.char
	a.fromTok = a.tok
	a.fromBg = assets.Bitmap(a.char.Bg)
	a.fromFlow = a.wallpaperFlowFor(a.char)

	a.char = c
	a.tok = c.Tok
	a.settings.Theme = c.ID
	a.pendingKeep = wallpaperKeys(c)

	a.switching = true
	a.switchAt = a.clock()
	if !a.settings.Animations {
		a.finishSwitch()
	}

	// 可选：换主题后再看一遍开场。开场层是绝对定位的覆盖层，重新置空后
	// 下一帧会按新的角色主题重建（标识、粒子都用新主题的色）。
	if a.settings.ReplayIntro && a.introDone {
		a.introDone = false
		a.intro = nil
	}

	// 主题切换是低频操作，直接落盘，不参与滑杆的防抖。
	if a.persist {
		_ = settings.Save(a.settings)
	}
}

// finishSwitch 结束交叉淡入：把 from* 归位到当前主题并释放旧像素。
func (a *App) finishSwitch() {
	if !a.switching {
		return
	}
	a.switching = false
	a.tok = a.char.Tok
	a.fromChar = a.char
	a.fromTok = a.char.Tok
	a.fromBg = assets.Bitmap(a.char.Bg)
	a.fromFlow = a.wallpaperFlowFor(a.char)
	if len(a.pendingKeep) > 0 {
		assets.Keep(a.pendingKeep...)
		a.pendingKeep = nil
	}
}

// themeProgress 返回主题切换进度 0..1；到 1 时顺手收尾。
func (a *App) themeProgress(now time.Time) float32 {
	if !a.switching {
		return 1
	}
	k := float32(now.Sub(a.switchAt).Seconds() * 1000 / float64(themeFade/time.Millisecond))
	if k >= 1 {
		a.finishSwitch()
		return 1
	}
	if k < 0 {
		k = 0
	}
	// smootherstep，和开场/页面动效保持一致。
	return k * k * k * (k*(k*6-15) + 10)
}

// pageProgress 返回当前页入场进度 0..1。
func (a *App) pageProgress(now time.Time) float32 {
	if !a.settings.Animations {
		return 1
	}
	k := float32(now.Sub(a.pageAt).Seconds() * 1000 / float64(pageFade/time.Millisecond))
	if k >= 1 {
		return 1
	}
	if k < 0 {
		k = 0
	}
	return k * k * k * (k*(k*6-15) + 10)
}

// viewProgress 返回主题页视图切换的淡入进度 0..1。
func (a *App) viewProgress(now time.Time) float32 {
	if !a.settings.Animations {
		return 1
	}
	k := float32(now.Sub(a.viewAt).Seconds() * 1000 / float64(viewFade/time.Millisecond))
	if k >= 1 {
		return 1
	}
	if k < 0 {
		k = 0
	}
	return k * k * k * (k*(k*6-15) + 10)
}

// goPage 切换页面，并重新开始入场动画。
func (a *App) goPage(c *ui.Context, page string) {
	// 从「外观与启动」回到主题库：只切视图，不换页；点侧栏「主题库」时
	// 用户预期看到的是卡片，而不是停在设置面板上。
	if page == "library" && a.themeView != 0 {
		a.themeView = 0
		if a.page == "library" {
			a.viewAt = a.clock()
			c.Invalidate()
			return
		}
	}
	if a.page == page {
		return
	}
	a.page = page
	a.pageAt = a.clock()
	c.Invalidate()
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
	now := a.clock()
	a.saveTick(c)
	a.pollUpdate(c)

	// 主题交叉淡入：a.tok 在旧、新两组令牌之间插值；壁纸用两层叠加，
	// 新壁纸那一层的 Element.Opacity 会连着色器一起淡入（mygo 的 shader
	// Draw 明确写着“clipped, transformed and faded with the painter”）。
	k := a.themeProgress(now)
	if a.switching {
		a.tok = a.fromTok.Mix(a.char.Tok, k)
		c.Invalidate() // 过渡期间需要重建视图，才能推进 k
	}

	// 页面入场：位移 + 淡入，详见 content。
	if a.pageProgress(now) < 1 {
		c.Invalidate()
	}
	// 主题页视图切换的淡入。
	if a.viewProgress(now) < 1 {
		c.Invalidate()
	}

	t := a.tok.UITheme()
	t.FontSize = a.fs(13) // 内置控件（滑杆/开关/分段/下拉）也跟随整体字号
	c.SetTheme(t)
	if !a.settings.Animations {
		c.DisableAnimations() // 连内置控件的补间也关掉
	}

	// 底层的壁纸：过渡时是旧主题，平时就是当前主题。
	fromChar := a.fromChar
	fromTok := a.fromTok
	fromBg := a.fromBg
	fromFlow := a.fromFlow
	if !a.switching {
		fromChar = a.char
		fromTok = a.tok
		fromBg = assets.Bitmap(a.char.Bg)
		fromFlow = a.wallpaperFlow()
	}

	// 壁纸画在根容器的 Draw 回调里：它在自身背景之后、子元素之前执行，
	// 所以 UI 天然压在壁纸上，不需要额外的层叠容器。
	root := ui.Column(c).Fill().Background(a.tok.WinFill)
	if fromBg != nil {
		root = root.Draw(func(p *ui.Painter, r ui.Rect) {
			a.drawWallpaper(p, r, fromChar, fromBg, fromFlow, fromTok, p.Now())
		})
	}

	root.Children(func() {
		// 过渡中的新壁纸：普通子元素先画，整层按 k 淡入。
		//
		// 不能把它做成 Absolute：mygo 先画普通子元素、最后画 Absolute，
		// 那样壁纸会盖到侧栏和卡片上面（试过，整屏发白）。
		if a.switching {
			if bg := assets.Bitmap(a.char.Bg); bg != nil {
				flow := a.wallpaperFlow()
				ui.Box(c).Fill().Opacity(k).
					Draw(func(p *ui.Painter, r ui.Rect) {
						a.drawWallpaper(p, r, a.char, bg, flow, a.char.Tok, p.Now())
					})
			}
		}

		// UI 整层是 Absolute：它排在普通子元素（新壁纸）之后绘制，
		// 因此壁纸永远在 UI 下面。状态栏在右侧内容列里，不横穿侧栏。
		ui.Row(c).Absolute().Fill().AlignItems(ui.Stretch).Children(func() {
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
		// 跳过：点画面任意处或按 Esc。交给 Intro.Skip 用 0.9 秒快进走完，
		// 不硬切——硬切会让标识瞬间消失、品牌位再突然亮起。
		//
		// 覆盖层本身已经接住了指针（全屏 Box），此前的问题是没人读这次点击，
		// 所以 Skip() 永远调不到，表现就是「点了没反应」。键盘不受 Box 影响，
		// Esc 单独接。
		if a.intro != nil && (ov.Clicked() || c.Shortcut(0, ui.KeyEscape)) {
			a.intro.Skip()
		}
	}
}

// drawWallpaper 画一层壁纸：着色器 + 按壁纸亮度压暗 + 顶部渐变。
//
// 过渡时旧、新两层各画一次，新层靠 Element.Opacity 淡入。壁纸显现
// （settings.WallpaperVis）叠加到角色自带的 Scrim 上，两层各自算。
func (a *App) drawWallpaper(p *ui.Painter, r ui.Rect, ch theme.Character,
	bg *ui.Bitmap, flow *wallpaper.Flow, tok theme.Tokens, now time.Time) {
	elapsed := now.Sub(a.t0).Seconds()
	// 着色器驱动：Cover 铺满 + 按角色元素色保亮度染色；
	// 甘雨这类带流动图层的角色，只有遮罩内的细节（发丝）在摆，主体不动。
	wallpaper.Draw(p, bg, r, elapsed, a.animPlay, tok.Accent, ch.WallTint, flow)
	scrim := a.settings.EffectiveScrim(ch.Scrim)
	p.Fill(r, ui.RGBA(6, 12, 22, scrim), 0) // 按壁纸亮度压暗（见 Character.Scrim）

	// 顶部再压一层渐变，托住标题区（在玻璃条之下，不影响玻璃本身的通透）。
	p.FillGradient(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: topScrimH}, topScrim, 0)

	// 下一帧：只重绘这块，不重建元素树。动效关掉时不必再请求重绘。
	if a.animPlay > 0 {
		p.AnimationFrameRate(animFPS)
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

// elementMarkSoft 底部的大标识：参考图里是一枚很淡的花形线稿，
// 不抢戏——只有淡淡的辉光和半透明线条。带纹样的主题（甘雨）换成清心纹，
// 参考项目侧栏底部画的正是同一朵花。
func (a *App) elementMarkSoft(c *ui.Context, size float32) {
	ui.Row(c).Width(size).Height(size).Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		if a.char.Ornament != theme.OrnamentNone {
			t := float32(p.Now().Sub(a.t0).Seconds())
			if a.char.Ornament == theme.OrnamentQingxin {
				// 甘雨侧栏的清心纹沿用参考的 -90° 朝向和更淡的冰蓝。
				qingxin(p, cx, cy, size*.45, ui.RGBA(188, 218, 232, .35), -1.5707964)
				return
			}
			ornamentSignature(p, a.char.Ornament, cx, cy, size*.46, a.tok.Accent.Alpha(.55), t)
			return
		}
		glow(p, cx, cy, size*0.52, a.tok.Accent, 0.22)
		glyph(p, a.char.Motif, cx, cy, size*0.38, a.tok.Text, 0.55, 0)
	})
}

// sidebar 左侧栏：品牌（元素标识 + 名称）→ 分组标签 → 菜单 → 底部当前角色。
func (a *App) sidebar(c *ui.Context) {
	// 不用 glass 材质：它会画边缘高光，侧栏上下看起来像有白色边框。
	// 参考图就是一块纯净的半透明深色，直接用令牌色铺底。
	col := ui.Column(c).Width(212).Background(a.tok.Sidebar)
	col.Children(func() {
		// 品牌区不贴顶：参考图里图标顶部约占窗口高 8.5%，按 800 高折算约 64px。
		// 品牌图标：启动动画期间留空——开场标识正飞向这个位置，
		// 落定（introDone）后由 elementMark 接手显示，完成「飞入品牌位」的交接。
		// 不播开场（introDone 或尚未创建 intro）时直接显示。
		ui.Row(c).Padding(64, 12, 10, 20).Gap(12).AlignItems(ui.Center).Children(func() {
			if a.intro == nil || a.introDone {
				a.elementMark(c, 40)
			} else {
				ui.Row(c).Width(40).Height(40)
			}
			ui.Column(c).Gap(1).Children(func() {
				ui.Text(c, "HanaBi").FontSize(a.fs(19)).Bold().TextColor(a.tok.Text)
				ui.Text(c, a.char.Motto).FontSize(a.fs(9)).TextColor(a.tok.TextFaint)
			})
		})

		// 分组标签：原版把导航分成「工作空间」等几组。
		ui.Text(c, "工作空间").FontSize(a.fs(9)).TextColor(a.tok.TextFaint).
			Padding(13, 10, 6, 20)

		ui.Column(c).PaddingX(12).Gap(2).Children(func() {
			for _, it := range nav {
				a.navRow(c, it.id, it.label, it.icon)
			}
		})

		ui.Spacer(c)

		// 底部：一枚较大的淡色元素标识 + 场景名，整体居中；下缘距窗口底约 6%
		// （参考图量出来的比例），所以底部留白比菜单侧的 padding 大。
		ui.Column(c).AlignItems(ui.Center).Gap(8).Padding(0, 16, 46, 16).Children(func() {
			a.elementMarkSoft(c, 64)
			ui.Text(c, a.char.Scene).FontSize(a.fs(11)).TextColor(a.tok.TextMuted)
			ui.Text(c, a.char.SceneEN).FontSize(a.fs(8.5)).TextColor(a.tok.TextFaint)
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
		a.goPage(c, id)
	}
	row.Children(func() {
		// 选中项左侧的竖杠：参考图里导航选中态的标志（圆角短棒，元素色）。
		barAlpha := float32(0)
		if active {
			barAlpha = 0.95
		}
		ui.Row(c).Width(3).Height(18).Radius(1.5).
			Background(a.tok.Accent.Alpha(barAlpha))
		ui.Icon(c, icon).FontSize(17).TextColor(color)
		ui.Text(c, label).FontSize(a.fs(13)).TextColor(color).Grow(1)
		if active {
			ui.Row(c).Width(18).Height(18).Draw(func(p *ui.Painter, r ui.Rect) {
				cx, cy := r.X+r.W/2, r.Y+r.H/2
				if a.char.Ornament != theme.OrnamentNone {
					a.navOrnament(p, cx, cy, r.W*0.50)
					return
				}
				glyph(p, a.char.Motif, cx, cy, r.W*0.44, a.tok.Accent, 0.85, 0)
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
		a.winButton(c, 0, icons.Minimize, "最小化", func() {
			if a.win != nil {
				a.win.Minimize()
			}
		})
		a.winButton(c, 1, icons.Maximize, "最大化 / 还原", func() {
			if a.win == nil {
				return
			}
			if a.win.IsMaximized() {
				a.win.Unmaximize()
			} else {
				a.win.Maximize()
			}
		})
		a.winButton(c, 2, icons.Close, "关闭", func() {
			if a.win != nil {
				a.win.Close()
			}
		})
	})
	bar.DragWindow()

	// 顶栏正中是当前角色的命之座拉丁名（参考项目 DrawChrome 的 text.constellation）。
	// 注意要按「整个窗口」居中，而不是标题栏（它从侧栏右边开始）居中——
	// HanaBi 字样只留在状态栏，不再和命之座抢中间。
	bar.DrawOver(func(p *ui.Painter, r ui.Rect) {
		label := a.char.Constellation
		if label == "" {
			return
		}
		w, _ := c.Size()
		size := a.fs(9)
		tw, th := p.MeasureText(0, ui.Span{Text: label, Size: size})
		p.Text(w/2-tw/2, r.Y+(r.H-th)/2, label, size, a.tok.TextFaint.Alpha(.85))
	})
}

// winButton 是一个窗口控件按钮：悬停变亮，点击执行 action。
//
// idx=2 是关闭按钮，走「警告色」分支：悬停染成主题警告红玫色，图标也转成
// 同色——与最小化/最大化区分开。参考项目正是给关闭按钮单独染玫红。
//
// 造型按主题的 Chrome 分发：
//   - ChromePlain：圆角按钮 + SVG 图标。图标不用 "—" "□" "✕" 这些字符，
//     字体里不一定有这些字形，缺字形时按钮会变成一个空方块（试过）。
//   - ChromeCrystal：冰晶切角八边形 + 淡金线 + 冰晶化的三个图标（甘雨）。
//
// 悬停用 140ms 的补间而不是瞬变：参考项目这里是弹簧，颜色要跟着走。
func (a *App) winButton(c *ui.Context, idx int, icon *ui.SVG, tip string, action func()) {
	warn := idx == 2
	btn := ui.Row(c).Width(36).Height(28).Radius(6).
		AlignItems(ui.Center).Justify(ui.Center)
	hovered := btn.Hovered()
	target := float32(0)
	if hovered {
		target = 1
	}
	hv := target
	if a.settings.Animations {
		hv = btn.AnimateTo("hover", target,
			ui.Animation{Duration: 140 * time.Millisecond, Ease: ui.EaseOut})
	}
	btn.Tooltip(tip)
	if btn.Clicked() {
		action()
	}

	if a.char.Chrome != theme.ChromePlain {
		btn.Draw(func(p *ui.Painter, r ui.Rect) { a.chromeButton(p, r, idx, hv) })
		return
	}

	if hv > 0.002 {
		if warn {
			// 警告色提醒：破坏性操作，悬停时背景与图标都染成主题警告红。
			btn.Background(a.tok.Danger.Alpha(0.22 * hv))
		} else {
			btn.Background(ui.RGBA(255, 255, 255, 0.16*hv))
		}
	}
	btn.Children(func() {
		col := a.tok.Text
		if warn {
			col = a.tok.Text.Mix(a.tok.Danger, hv)
		}
		ui.Icon(c, icon).FontSize(16).TextColor(col)
	})
}

// crystalButton 画甘雨的窗口按钮：八边形冰晶切角 + 顶部淡金线 + 冰晶化图标。
//
// 几何照参考 ganyu_ui.cpp::WindowControl：中心、半宽 13、切角 5
// （按钮 36×28，八边形 26×26）。三个图标的画法：
// 最小化是一横 + 菱形箭头，最大化是菱形轮廓 + 中心点，关闭是叉 + 小圆环。
func (a *App) crystalButton(p *ui.Painter, r ui.Rect, idx int, hv float32) {
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	half := min32(r.W, r.H)*0.5 - 1
	cut := half * 0.38

	pts := octagonPoints(cx, cy, half, cut)
	// 冰蓝底：悬停时更实一点，切角才有"水晶块"的体积感。
	p.FillPath(octagonFrom(pts), ui.RGBA(12, 37, 57, .29+.22*hv))

	base := ui.RGBA(151, 219, 240, 1)
	hot := ui.RGBA(220, 238, 249, 1)
	if idx == 2 {
		// 关闭是破坏性操作：悬停从冰蓝过渡到玫红。
		hot = ui.RGBA(223, 140, 159, 1)
	}
	tint := base.Mix(hot, hv)

	// 顶部切面：一条极淡的高光三角，让切角面读起来有折射。
	var facet ui.Path
	facet.MoveTo(pts[0][0], pts[0][1])
	facet.LineTo(pts[1][0], pts[1][1])
	facet.LineTo(cx, cy-2)
	facet.Close()
	p.FillPath(&facet, tint.Alpha(.07))

	// 轮廓与顶部金线。
	p.StrokePath(octagonFrom(pts), .8, tint.Alpha(.30+.50*hv))
	var edge ui.Path
	edge.MoveTo(pts[0][0], pts[0][1])
	edge.LineTo(pts[1][0], pts[1][1])
	p.StrokePath(&edge, .8, ui.RGBA(218, 191, 137, .31+.35*hv))

	col := tint
	z := float32(4.5)
	switch idx {
	case 0: // 最小化：一横 + 小菱形箭头
		var line ui.Path
		line.MoveTo(cx-z, cy)
		line.LineTo(cx+z, cy)
		p.StrokePath(&line, 1.1, col)
		var arrow ui.Path
		arrow.MoveTo(cx+z+1, cy-2)
		arrow.LineTo(cx+z+3, cy)
		arrow.LineTo(cx+z+1, cy+2)
		arrow.LineTo(cx+z-1, cy)
		arrow.Close()
		p.FillPath(&arrow, col)
	case 1: // 最大化：菱形轮廓 + 中心点
		var d ui.Path
		d.MoveTo(cx, cy-z-1)
		d.LineTo(cx+z+1, cy)
		d.LineTo(cx, cy+z+1)
		d.LineTo(cx-z-1, cy)
		d.Close()
		p.StrokePath(&d, 1.1, col)
		circle(p, cx, cy, 1, col)
	case 2: // 关闭：叉 + 小圆环
		var x ui.Path
		x.MoveTo(cx-z, cy-z)
		x.LineTo(cx+z, cy+z)
		x.MoveTo(cx-z, cy+z)
		x.LineTo(cx+z, cy-z)
		p.StrokePath(&x, 1.2, col)
		var ring ui.Path
		ring.Circle(cx, cy, 1.7)
		p.StrokePath(&ring, .7, col)
	}
}

// content 内容区：按当前页分发。
//
// 只压一层很淡的暗色托文字，不再包整块面板——参考图里内容是直接浮在
// 壁纸上的，只有卡片本身是一块块玻璃。
func (a *App) content(c *ui.Context) {
	ui.Column(c).Grow(1).Clip().
		Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(r, ui.RGBA(6, 11, 18, 0.28), 0)
			// 角色主题：内容区右下角一枚很大的淡纹（参考各角色 _ui.cpp 的 Chrome）。
			// 画在暗色底之后、页面之前，所以它垫在卡片与文字下面。
			if a.char.Ornament != theme.OrnamentNone {
				a.drawChromeOrnament(p, r)
			}
		}).
		Children(func() {
			// Grow(1) 必须加：Scroll 只在「父容器高度有界」时才滚动——
			// 这层不带 Grow 的话它按内容自然长高，Scroll 拿到无限高度，
			// 最后一行卡片直接被外层 Clip 裁掉、滚轮无反应（踩过）。
			page := ui.Column(c).Grow(1).Padding(24, 22).Gap(16)
			// 页面入场：从下方 10 DIP 滑入并淡入（参考 pageWeight + VtxApply）。
			if k := a.pageProgress(a.clock()); k < 1 {
				page = page.Opacity(k).Transform(ui.Translation(0, (1-k)*10))
			}
			page.Children(func() {
				switch a.page {
				case "library":
					a.libraryPage(c)
				case "settings":
					a.settingsPage(c)
				default:
					a.placeholderPage(c)
				}
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
			ui.Text(c, "HANABI / PERSONAL WORKSPACE").FontSize(a.fs(8.5)).
				TextColor(a.tok.TextFaint.Alpha(0.72))
			ui.Spacer(c)
			ui.Text(c, "本地工作空间").FontSize(a.fs(9.5)).
				TextColor(a.tok.TextFaint.Alpha(0.80))
		})
}
