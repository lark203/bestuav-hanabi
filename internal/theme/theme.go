// Package theme 存放角色主题：一套调色板（Tokens）、一种元素、一种标识画法（Motif）。
//
// 骨架阶段只有甘雨，现在是「角色包」结构：新增角色只要往 all 里加一条，
// 配好元素色即可——绘制层（侧栏/卡片/开场标识）全部按 Motif 与 Tokens 走，不写死任何角色。
package theme

import "github.com/egoist/mygo/ui"

// Element 是元素类型，决定标识形状与氛围色。
type Element uint8

const (
	Cryo Element = iota
	Pyro
	Electro
	Dendro
	Hydro
	Anemo
	Geo
)

// Label 返回元素的中文名。
func (e Element) Label() string {
	switch e {
	case Cryo:
		return "冰元素"
	case Pyro:
		return "火元素"
	case Electro:
		return "雷元素"
	case Dendro:
		return "草元素"
	case Hydro:
		return "水元素"
	case Anemo:
		return "风元素"
	case Geo:
		return "岩元素"
	}
	return "未知"
}

// Short 返回元素的单字名（筛选条上用的「冰 / 水 / 雷 …」）。
func (e Element) Short() string {
	switch e {
	case Cryo:
		return "冰"
	case Pyro:
		return "火"
	case Electro:
		return "雷"
	case Dendro:
		return "草"
	case Hydro:
		return "水"
	case Anemo:
		return "风"
	case Geo:
		return "岩"
	}
	return "?"
}

// Motif 是元素标识的画法。绘制层按它分支，而不是按角色名。
type Motif uint8

const (
	// MotifCrystal 六角冰晶：主干 + V 形分叉 + 菱形尖端。
	MotifCrystal Motif = iota
	// MotifFlame 火焰/烟花：放射火舌 + 末端火星，成形时向外炸开。
	MotifFlame
	// MotifBolt 雷：折线闪电 + 外侧小分叉。
	MotifBolt
	// MotifGeo 岩：六边形岩印 + 内嵌菱形。
	MotifGeo
)

// Tokens 是一套角色主题的调色板。
type Tokens struct {
	WinFill    ui.Color // 窗口内容底色
	WinStroke  ui.Color // 窗口描边
	Sidebar    ui.Color // 侧栏底色
	Card       ui.Color // 卡片底色
	CardStroke ui.Color // 卡片描边

	Text      ui.Color // 主文字
	TextMuted ui.Color // 次要文字
	TextFaint ui.Color // 辅助/说明文字

	Accent     ui.Color // 强调色（元素色）
	AccentSoft ui.Color // 强调色的低透明底
	Good       ui.Color // 正向状态

	Radius float32 // 全局圆角
}

// Character 是一个角色主题包。
type Character struct {
	ID     string
	NameCN string
	NameEN string
	// Epithet 是卡片上的属性词，如「甘雨·循霜」里的「循霜」。
	Epithet string
	Element Element
	Motif   Motif
	Tok     Tokens

	// Bg 是壁纸在 assets/ 下的相对路径（每个角色一张，均从原项目导入）。
	Bg string
	// Icon 是头像的相对路径；空表示没有头像素材，界面用元素色圆点占位。
	Icon string

	// WallTint 是壁纸染色的强度 0..1：0 = 保持原图颜色。
	// 各自的壁纸色调已经贴合角色，这里只做轻微统一，通常 0.1~0.2。
	WallTint float32

	// Scrim 是壁纸上的压暗强度 0..1，保证白字压在上面仍可读。
	//
	// 不能所有角色用同一个值：实测平均亮度从甘雨的 0.27 到八重/纳西妲的 0.80，
	// 同一个 0.38 对前者刚好、对后者等于没压。取值按 0.32+max(0,luma-0.32)×0.55 换算。
	Scrim float32
}

// derive 从元素色推导一套暗色令牌：底色用元素色混入近黑，保证始终是深色底。
func derive(accent ui.Color) Tokens {
	deep := accent.Mix(ui.Hex("#04070C"), 0.86)
	return Tokens{
		WinFill:    deep.Alpha(0.58),
		WinStroke:  accent.Alpha(0.20),
		Sidebar:    deep.Alpha(0.44),
		Card:       accent.Alpha(0.10),
		CardStroke: accent.Alpha(0.22),

		// 文字都掺一点元素色：整块界面会带角色气质，而不是纯灰白。
		Text:      accent.Mix(ui.Hex("#FFFFFF"), 0.80),
		TextMuted: accent.Mix(ui.Hex("#FFFFFF"), 0.44),
		TextFaint: accent.Mix(ui.Hex("#FFFFFF"), 0.34),

		Accent:     accent,
		AccentSoft: accent.Alpha(0.14),
		Good:       ui.Hex("#7BE0B0"),
		Radius:     12,
	}
}

// all 是所有可用角色。第一条是默认主题。
//
// 壁纸与头像都来自 HanaBi-ui-Genshin 的 assets/tex（本地导入，不入库），
// 目录约定为 assets/<id>/bg.jpg 与可选的 assets/<id>/icon.png。
var all = []Character{
	{
		ID: "ganyu", NameCN: "甘雨", NameEN: "Ganyu", Epithet: "循霜",
		Element: Cryo, Motif: MotifCrystal, WallTint: 0.15, Scrim: 0.32,
		Bg: "ganyu/bg.jpg", Icon: "ganyu/icon.png", Tok: ganyuTokens(),
	},
	{
		ID: "yoimiya", NameCN: "宵宫", NameEN: "Yoimiya", Epithet: "烟火",
		Element: Pyro, Motif: MotifFlame, WallTint: 0.18, Scrim: 0.47,
		Bg: "yoimiya/bg.jpg", Tok: yoimiyaTokens(),
	},
	{
		ID: "yae", NameCN: "八重神子", NameEN: "Yae Miko", Epithet: "神樱",
		Element: Electro, Motif: MotifBolt, WallTint: 0.15, Scrim: 0.66,
		Bg: "yae/bg.jpg", Tok: yaeTokens(),
	},
	// 以下角色有自己的壁纸，但标识先复用冰晶画法，等补了专属 Motif 再换。
	{
		ID: "nahida", NameCN: "纳西妲", NameEN: "Nahida", Epithet: "新芽",
		Element: Dendro, Motif: MotifCrystal, WallTint: 0.15, Scrim: 0.66,
		Bg: "nahida/bg.jpg", Tok: derive(ui.Hex("#A8E6A0")),
	},
	{
		ID: "furina", NameCN: "芙宁娜", NameEN: "Furina", Epithet: "水幕",
		Element: Hydro, Motif: MotifCrystal, WallTint: 0.14, Scrim: 0.35,
		Bg: "furina/bg.jpg", Tok: derive(ui.Hex("#8FD0FF")),
	},
	{
		ID: "faruzan", NameCN: "珐露珊", NameEN: "Faruzan", Epithet: "风卷",
		Element: Anemo, Motif: MotifCrystal, WallTint: 0.16, Scrim: 0.47,
		Bg: "faruzan/bg.jpg", Tok: derive(ui.Hex("#9EE8D6")),
	},
	{
		// 钟离的壁纸原项目只有 .hbw 打包格式，这里是从中提取图层合成出的静态图
		// （金色渐变 + 岩印 + 立绘 + 沙漠）。色彩本身很亮，压暗给到 0.46。
		ID: "zhongli", NameCN: "钟离", NameEN: "Zhongli", Epithet: "岩契",
		Element: Geo, Motif: MotifGeo, WallTint: 0.18, Scrim: 0.62,
		Bg: "zhongli/bg.jpg", Tok: zhongliTokens(),
	},
}

// All 返回全部角色（顺序即界面顺序）。
func All() []Character { return all }

// ByID 按 id 查找角色。
func ByID(id string) (Character, bool) {
	for _, c := range all {
		if c.ID == id {
			return c, true
		}
	}
	return Character{}, false
}

// Default 返回默认角色。
func Default() Character { return all[0] }

// ganyuTokens 甘雨（冰）：手调的一套，比 derive 更冷、更透。
func ganyuTokens() Tokens {
	t := derive(ui.Hex("#7FD7FF"))
	// 带壁纸后，浮层与卡片都用半透明：壁纸从下面透出来，同时保证文字可读。
	// 卡片用「浅色叠加」而不是深色——在暗色壁纸上这是提亮，才接近参考图的通透感。
	t.WinFill = ui.RGBA(8, 14, 24, 0.55)
	t.Sidebar = ui.RGBA(8, 16, 28, 0.42)
	t.Card = ui.RGBA(150, 190, 230, 0.10)
	t.CardStroke = ui.RGBA(180, 215, 255, 0.20)
	t.Text = ui.Hex("#E6F2FF")
	t.TextMuted = ui.Hex("#9FB6D0")
	t.TextFaint = ui.Hex("#8FA6C4")
	return t
}

// yoimiyaTokens 宵宫（火）：暖橙，底色压得更沉，避免整屏发红。
func yoimiyaTokens() Tokens {
	t := derive(ui.Hex("#FFB27A"))
	t.WinFill = ui.RGBA(22, 12, 9, 0.56)
	t.Sidebar = ui.RGBA(26, 14, 10, 0.44)
	t.Card = ui.RGBA(230, 170, 120, 0.11)
	t.CardStroke = ui.RGBA(255, 200, 160, 0.22)
	t.Text = ui.Hex("#FFF0E2")
	t.TextMuted = ui.Hex("#D6B49B")
	t.TextFaint = ui.Hex("#BFA184")
	return t
}

// yaeTokens 八重神子（雷）：紫，文字偏冷白。
func yaeTokens() Tokens {
	t := derive(ui.Hex("#C9A6FF"))
	t.WinFill = ui.RGBA(14, 11, 24, 0.56)
	t.Sidebar = ui.RGBA(17, 13, 28, 0.44)
	t.Card = ui.RGBA(180, 160, 230, 0.11)
	t.CardStroke = ui.RGBA(210, 190, 255, 0.22)
	t.Text = ui.Hex("#F0E9FF")
	t.TextMuted = ui.Hex("#B6A8D4")
	t.TextFaint = ui.Hex("#A99BC4")
	return t
}

// zhongliTokens 钟离（岩）：金褐，文字偏暖白。
func zhongliTokens() Tokens {
	t := derive(ui.Hex("#E8C87A"))
	t.WinFill = ui.RGBA(24, 18, 10, 0.56)
	t.Sidebar = ui.RGBA(28, 21, 12, 0.44)
	t.Card = ui.RGBA(232, 200, 122, 0.11)
	t.CardStroke = ui.RGBA(255, 226, 150, 0.22)
	t.Text = ui.Hex("#FFF6E2")
	t.TextMuted = ui.Hex("#E4D2AE")
	t.TextFaint = ui.Hex("#BFA982")
	return t
}

// UITheme 把令牌转成 mygo 的主题结构：内置控件（按钮、输入框、滚动条）
// 也会跟随这套配色，而不是系统默认的蓝。
func (t Tokens) UITheme() *ui.Theme {
	th := ui.DarkTheme()
	th.Background = t.WinFill
	th.Surface = t.Card
	th.SurfaceHover = t.Card.Mix(t.Accent, 0.12)
	th.SurfacePressed = t.Card.Mix(t.Accent, 0.20)
	th.Border = t.CardStroke
	th.Text = t.Text
	th.TextMuted = t.TextMuted
	th.Accent = t.Accent
	th.AccentHover = t.Accent.Mix(ui.Hex("#FFFFFF"), 0.18)
	th.AccentPressed = t.Accent.Mix(ui.Hex("#000000"), 0.18)
	th.AccentText = ui.Hex("#08131F")
	th.Danger = ui.Hex("#FF7A8A")
	th.Warning = ui.Hex("#F5C56B")
	th.Success = t.Good
	th.Selection = t.Accent.Alpha(0.35)
	th.Focus = t.Accent.Alpha(0.60)
	th.Radius = t.Radius
	th.Spacing = 4
	th.FontSize = 15
	return th
}
