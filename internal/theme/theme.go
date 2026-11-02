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
	// MotifLeaf 草：三片尖叶成扇形展开 + 短茎。
	MotifLeaf
	// MotifWave 水：水珠 + 波浪。
	MotifWave
	// MotifWind 风：回旋气流。
	MotifWind
)

// Art 是主题卡片右侧的装饰线稿：每个角色一幅，形状与人物气质呼应。
//
// 卡片不铺壁纸缩略图（缩略图缩小后只是一团色块），改为纯色底 + 线稿：
// 纯色保证一排卡片色彩分明，线稿让每张卡片有各自的图形记忆点。
type Art uint8

const (
	// ArtBlossom 霜花 + 风铃。
	ArtBlossom Art = iota
	// ArtSakura 樱枝 + 御守。
	ArtSakura
	// ArtFirework 烟花 + 灯笼。
	ArtFirework
	// ArtLeaf 叶枝 + 新芽。
	ArtLeaf
	// ArtSail 帆船 + 水波。
	ArtSail
	// ArtCrystal 棱晶 + 书页。
	ArtCrystal
	// ArtRing 岩环 + 六方。
	ArtRing
	// ArtNone 表示这张卡片由 Ornament 的专属纹样接管，不再画通用线稿。
	ArtNone
)

// Chrome 是窗口按钮（最小化 / 最大化 / 关闭）的造型。
//
// 参考项目每个角色都有一套和服饰、元素匹配的窗口按钮；这里按主题声明，
// 绘制层只按枚举分支，不认角色名。
type Chrome uint8

const (
	// ChromePlain 是通用造型：圆角按钮 + SVG 图标。
	ChromePlain Chrome = iota
	// ChromeCrystal 是冰晶切角：八边形切角 + 淡金线 + 冰晶化的三个图标（甘雨）。
	ChromeCrystal
)

// Ornament 是侧栏、卡片与内容区上的角色专属纹样。
type Ornament uint8

const (
	// OrnamentNone 只有统一的元素标识。
	OrnamentNone Ornament = iota
	// OrnamentQingxin 是璃月的清心纹 + 中国结 + 风纹（甘雨）。
	OrnamentQingxin
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
	Danger     ui.Color // 破坏性操作的提示色（关闭按钮等），暖红玫色

	Radius float32 // 全局圆角
}

// Character 是一个角色主题包。
type Character struct {
	ID     string
	NameCN string
	NameEN string
	// Epithet 是卡片上的属性词，如「甘雨·循霜」里的「循霜」。
	Epithet string
	// Motto 是侧栏品牌区副标题（随主题切换，参考图里是「循霜而行」）。
	Motto string
	// Constellation 是顶栏中央的命之座拉丁名，如甘雨的
	// 「S I N U S   U N I C O R N I S」。参考项目把它放在窗口正中，
	// HanaBi 字样只留在状态栏；空表示不画。
	Constellation string
	// Scene / SceneEN 是侧栏底部的场景名（参考图里的「璃月 · 海灯夜 / A QUIET MOMENT」）。
	Scene   string
	SceneEN string
	Element Element
	Motif   Motif
	// Chrome 决定三个窗口按钮的造型；Ornament 决定侧栏、卡片上的角色纹样。
	Chrome   Chrome
	Ornament Ornament
	Tok      Tokens

	// CardBg 是主题卡片的纯色底。每张卡片一块纯色，选中时加元素色描边。
	CardBg ui.Color
	// Art 是卡片右侧的装饰线稿（与左上的元素徽记是两个不同的图标）。
	Art Art

	// Bg 是壁纸在 assets/ 下的相对路径（每个角色一张，均从原项目导入）。
	Bg string
	// Icon 是头像的相对路径；空表示没有头像素材，界面用元素色圆点占位。
	Icon string

	// PersonMask 是**人物区域**遮罩（灰度，非零处才算人物）——它是所有壁纸动态的
	// 总闸门：它以外的地方严格等于原图，建筑与天空纹丝不动。
	//
	// 必须与 Bg 同源、构图对齐。甘雨实测覆盖 x 34.2%~59.4% / y 22.2%~73.7%，
	// 是一条竖直的人形带，仅占画面 4.9%。
	//
	// 空表示该角色壁纸完全静态：不做呼吸、不做流动。
	PersonMask string
	// PersonBreath 是呼吸起伏的幅度，单位是图片 UV 的垂直分量。
	// 0.0022 约等于 4K 图上 4.8px——只是几乎察觉不到的微动，符合静息呼吸。
	//
	// 只走垂直方向且全人物同步：呼吸时胸腔整体上下起伏，不是各部位乱飘。
	// 周期约 4.5 秒（着色器里 sin(Time*1.4)）。
	PersonBreath float32
	// HairVeil 是发丝**基础层**遮罩（可选，整片发区）。会被 PersonMask 裁到人物之内。
	//
	// 空表示没有发丝流动。甘雨的 flow-mask 实测覆盖 x 31.5%~81.6% / y 0%~42%，
	// 这个范围横铺到了天空和建筑上——所以它只能当"发丝层"用，绝不能当总闸门，
	// 否则就会看到建筑和天空在动。
	HairVeil string
	// HairWave 是发丝**细缕层**遮罩（可选），覆盖基础层内的一小片。
	//
	// 分层次靠这个：两层遮罩范围不同、频率相位方向各偏开一点，累加后细缕处比
	// 周围多一份位移，看起来才是一缕一缕分开飘。甘雨的 wave-a 实测覆盖
	// x 46.6%~59.2% / y 26.5%~42.7%（1.08%），落在人物内。
	HairWave string
	// HairSway 是发丝飘动的幅度，单位是图片 UV。细缕层在着色器里乘 1.3。
	HairSway float32

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
		// 警告红玫色：在冷（甘雨）暖（宵宫）各主题上都能跳出来；
		// 参考项目给关闭按钮单独染玫红 (223,140,159)，这里取稍饱和一点的暖红。
		Danger: ui.Hex("#FF6B7A"),
		Radius: 12,
	}
}

// all 是所有可用角色。第一条是默认主题。
//
// 壁纸与头像都来自 HanaBi-ui-Genshin 的 assets/tex（本地导入，不入库），
// 目录约定为 assets/<id>/bg.jpg 与可选的 assets/<id>/icon.png。
var all = []Character{
	{
		ID: "ganyu", NameCN: "甘雨", NameEN: "Ganyu", Epithet: "循霜",
		Motto: "循霜而行", Constellation: "S I N U S   U N I C O R N I S",
		Scene: "璃月 · 海灯夜", SceneEN: "A QUIET MOMENT",
		Element: Cryo, Motif: MotifCrystal, Chrome: ChromeCrystal, Ornament: OrnamentQingxin,
		WallTint: 0.15, Scrim: 0.32,
		CardBg: ui.Hex("#14314A"), Art: ArtNone,
		Bg: "ganyu/bg.jpg", Icon: "ganyu/icon.png", Tok: ganyuTokens(),
		// 甘雨壁纸动态，分两层机制：
		//
		// 1) 呼吸感：PersonMask（人物轮廓）当总闸门，只做垂直微起伏。
		//    0.0022 约合 4K 图 4.8px、周期约 4.5 秒——静息呼吸的量级。
		//    闸门之外（建筑、天空）严格等于原图，一个像素都不动。
		//
		// 2) 发丝分层：HairVeil（整片发区）+ HairWave（一缕细层）叠出层次。
		//    两者都会被人物闸门裁到人物之内——这是关键：flow-mask 本身覆盖
		//    y 0%~42% 的横带、含天空与建筑，只能当发丝层，绝不能当总闸门。
		//
		// 关于 wave-b：实测 bbox y 0%~20.3%，落在人物轮廓（y 从 22.2% 起）之上，
		// 被闸门裁掉后几乎不剩，故不启用。
		PersonMask:   "ganyu/ganyu-shake-opacity.png",
		PersonBreath: 0.0022,
		HairVeil:     "ganyu/ganyu-flow-mask.png",
		HairWave:     "ganyu/ganyu-wave-a.png",
		HairSway:     0.0016,
	},
	{
		ID: "yoimiya", NameCN: "宵宫", NameEN: "Yoimiya", Epithet: "烟火",
		Motto: "琉焰华舞", Constellation: "C A R A S S I U S   A U R A T U S",
		Scene: "稻妻 · 夏祭", SceneEN: "FIREWORKS AT DUSK",
		Element: Pyro, Motif: MotifFlame, WallTint: 0.18, Scrim: 0.47,
		CardBg: ui.Hex("#3E2418"), Art: ArtFirework,
		Bg: "yoimiya/bg.jpg", Tok: yoimiyaTokens(),
	},
	{
		ID: "yae", NameCN: "八重神子", NameEN: "Yae Miko", Epithet: "神樱",
		Motto: "浮世笑百姿", Constellation: "D I V I N A   V U L P E S",
		Scene: "鸣神 · 神樱", SceneEN: "SACRED SAKURA",
		Element: Electro, Motif: MotifBolt, WallTint: 0.15, Scrim: 0.66,
		CardBg: ui.Hex("#37203F"), Art: ArtSakura,
		Bg: "yae/bg.jpg", Tok: yaeTokens(),
	},
	// 以下角色有自己的壁纸与专属 Motif（草叶 / 水滴 / 风纹）。
	{
		ID: "nahida", NameCN: "纳西妲", NameEN: "Nahida", Epithet: "新芽",
		Motto: "白草净华", Constellation: "S A P I E N T I A   O R O M A S D I S",
		Scene: "须弥 · 世界树", SceneEN: "DREAM OF WISDOM",
		Element: Dendro, Motif: MotifLeaf, WallTint: 0.15, Scrim: 0.66,
		CardBg: ui.Hex("#1F3A22"), Art: ArtLeaf,
		Bg: "nahida/bg.jpg", Tok: derive(ui.Hex("#A8E6A0")),
	},
	{
		ID: "furina", NameCN: "芙宁娜", NameEN: "Furina", Epithet: "水幕",
		Motto: "不休独舞", Constellation: "A N I M U L A   C H O R A G I",
		Scene: "枫丹 · 歌剧院", SceneEN: "THE STAGE AWAITS",
		Element: Hydro, Motif: MotifWave, WallTint: 0.14, Scrim: 0.35,
		CardBg: ui.Hex("#15294A"), Art: ArtSail,
		Bg: "furina/bg.jpg", Tok: derive(ui.Hex("#8FD0FF")),
	},
	{
		ID: "faruzan", NameCN: "珐露珊", NameEN: "Faruzan", Epithet: "风卷",
		Motto: "百巧千机", Constellation: "F L O S C U L I   I M P L E X I",
		Scene: "须弥 · 教令院", SceneEN: "MACHINA OF WISDOM",
		Element: Anemo, Motif: MotifWind, WallTint: 0.16, Scrim: 0.47,
		CardBg: ui.Hex("#17342F"), Art: ArtCrystal,
		Bg: "faruzan/bg.jpg", Tok: derive(ui.Hex("#9EE8D6")),
	},
	{
		// 钟离的壁纸原项目只有 .hbw 打包格式，这里是从中提取图层合成出的静态图
		// （金色渐变 + 岩印 + 立绘 + 沙漠）。色彩本身很亮，压暗给到 0.62。
		ID: "zhongli", NameCN: "钟离", NameEN: "Zhongli", Epithet: "岩契",
		Motto: "尘世闲游", Constellation: "L A P I S   D E I",
		Scene: "璃月 · 孤云阁", SceneEN: "STONE HARBOUR",
		Element: Geo, Motif: MotifGeo, WallTint: 0.18, Scrim: 0.62,
		CardBg: ui.Hex("#392C14"), Art: ArtRing,
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
