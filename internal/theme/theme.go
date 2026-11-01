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

// Motif 是元素标识的画法。绘制层按它分支，而不是按角色名。
type Motif uint8

const (
	// MotifCrystal 六角冰晶：主干 + V 形分叉 + 菱形尖端。
	MotifCrystal Motif = iota
	// MotifFlame 火焰/烟花：放射火舌 + 末端火星，成形时向外炸开。
	MotifFlame
	// MotifBolt 雷：折线闪电 + 外侧小分叉。
	MotifBolt
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
	ID      string
	NameCN  string
	NameEN  string
	Element Element
	Motif   Motif
	Tok     Tokens

	// WallTint 是壁纸染色的强度 0..1：0 = 保持原图颜色。
	// 元素色离壁纸原色越远，这个值要越大才看得出氛围变化。
	WallTint float32
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
		TextFaint: accent.Mix(ui.Hex("#FFFFFF"), 0.22),

		Accent:     accent,
		AccentSoft: accent.Alpha(0.14),
		Good:       ui.Hex("#7BE0B0"),
		Radius:     12,
	}
}

// all 是所有可用角色。第一条是默认主题。
var all = []Character{
	{
		ID: "ganyu", NameCN: "甘雨", NameEN: "Ganyu",
		Element: Cryo, Motif: MotifCrystal,
		WallTint: 0.28, Tok: ganyuTokens(),
	},
	{
		ID: "yoimiya", NameCN: "宵宫", NameEN: "Yoimiya",
		Element: Pyro, Motif: MotifFlame,
		WallTint: 0.62, Tok: yoimiyaTokens(),
	},
	{
		ID: "raiden", NameCN: "雷电将军", NameEN: "Raiden",
		Element: Electro, Motif: MotifBolt,
		WallTint: 0.55, Tok: raidenTokens(),
	},
	// 以下角色先复用冰晶画法，等有对应素材再补各自 Motif。
	{ID: "nahida", NameCN: "纳西妲", NameEN: "Nahida", Element: Dendro, Motif: MotifCrystal, WallTint: 0.55, Tok: derive(ui.Hex("#A8E6A0"))},
	{ID: "furina", NameCN: "芙宁娜", NameEN: "Furina", Element: Hydro, Motif: MotifCrystal, WallTint: 0.32, Tok: derive(ui.Hex("#8FD0FF"))},
	{ID: "faruzan", NameCN: "珐露珊", NameEN: "Faruzan", Element: Anemo, Motif: MotifCrystal, WallTint: 0.52, Tok: derive(ui.Hex("#9EE8D6"))},
	{ID: "zhongli", NameCN: "钟离", NameEN: "Zhongli", Element: Geo, Motif: MotifCrystal, WallTint: 0.60, Tok: derive(ui.Hex("#E8C87A"))},
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
	t.TextFaint = ui.Hex("#6C819B")
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
	t.TextFaint = ui.Hex("#9A7A63")
	return t
}

// raidenTokens 雷电将军（雷）：紫，文字偏冷白。
func raidenTokens() Tokens {
	t := derive(ui.Hex("#C9A6FF"))
	t.WinFill = ui.RGBA(14, 11, 24, 0.56)
	t.Sidebar = ui.RGBA(17, 13, 28, 0.44)
	t.Card = ui.RGBA(180, 160, 230, 0.11)
	t.CardStroke = ui.RGBA(210, 190, 255, 0.22)
	t.Text = ui.Hex("#F0E9FF")
	t.TextMuted = ui.Hex("#B6A8D4")
	t.TextFaint = ui.Hex("#7E719A")
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
