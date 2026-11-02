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

// Opening 是开场动画的专属分镜。
//
// 参考项目每个角色一个 Opening：甘雨霜莲、宵宫金鱼烟火、八重神樱、
// 纳西妲梦境圣所、芙宁娜水上舞台、珐露珊机关、钟离岩柱。
type Opening uint8

const (
	OpeningGanyu Opening = iota
	OpeningYoimiya
	OpeningYae
	OpeningNahida
	OpeningFurina
	OpeningFaruzan
	OpeningZhongli
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
	// ChromePyro 暖金圆角 + 顶部金线 + 圆弧图标（宵宫）。
	ChromePyro
	// ChromeElectro 紫调圆角 + 金弧 + 樱瓣/菱形图标（八重神子）。
	ChromeElectro
	// ChromeDendro 深绿大圆角 + 叶弧 + 方/点图标（纳西妲）。
	ChromeDendro
	// ChromeHydro 蓝调圆角 + 银弧 + 水滴/方图标（芙宁娜）。
	ChromeHydro
	// ChromeAnemo 切角八边形 + 金线 + 顶部三角（珐露珊）。
	ChromeAnemo
	// ChromeGeo 切角八边形 + 金线 + 底部横线（钟离）。
	ChromeGeo
)

// Ornament 是侧栏、卡片与内容区上的角色专属纹样。
type Ornament uint8

const (
	// OrnamentNone 只有统一的元素标识。
	OrnamentNone Ornament = iota
	// OrnamentQingxin 是璃月的清心纹 + 中国结 + 风纹（甘雨）。
	OrnamentQingxin
	// OrnamentGoldfish 金鱼 + 烟花（宵宫）。
	OrnamentGoldfish
	// OrnamentSakura 樱花 + 御守（八重神子）。
	OrnamentSakura
	// OrnamentSprig 叶枝 + 新芽（纳西妲）。
	OrnamentSprig
	// OrnamentTopHat 礼帽 + 水波（芙宁娜）。
	OrnamentTopHat
	// OrnamentPolyhedron 棱晶 + 古书（珐露珊）。
	OrnamentPolyhedron
	// OrnamentStoneSeal 岩印 + 回纹（钟离）。
	OrnamentStoneSeal
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

// Mix 在两组令牌之间插值（主题切换的 0.42 秒交叉淡入用）。
// k=0 返回 t，k=1 返回 other。
func (t Tokens) Mix(other Tokens, k float32) Tokens {
	if k <= 0 {
		return t
	}
	if k >= 1 {
		return other
	}
	mix := func(a, b ui.Color) ui.Color { return a.Mix(b, k) }
	return Tokens{
		WinFill:    mix(t.WinFill, other.WinFill),
		WinStroke:  mix(t.WinStroke, other.WinStroke),
		Sidebar:    mix(t.Sidebar, other.Sidebar),
		Card:       mix(t.Card, other.Card),
		CardStroke: mix(t.CardStroke, other.CardStroke),
		Text:       mix(t.Text, other.Text),
		TextMuted:  mix(t.TextMuted, other.TextMuted),
		TextFaint:  mix(t.TextFaint, other.TextFaint),
		Accent:     mix(t.Accent, other.Accent),
		AccentSoft: mix(t.AccentSoft, other.AccentSoft),
		Good:       mix(t.Good, other.Good),
		Danger:     mix(t.Danger, other.Danger),
		Radius:     t.Radius + (other.Radius-t.Radius)*k,
	}
}

// Motion 描述一张壁纸怎么动：三层遮罩 + 三组幅度。
//
// 这是「人物会呼吸、发丝会飘、环境在流动」的统一模型：每个角色只填参数，
// 着色器一家管完（见 internal/wallpaper）。三层各自独立，缺素材（空路径）的层
// 自动退化成不参与，所以加角色时可以先只填 person。
//
// 位移单位都是**图片 UV**（不是屏幕像素）：这样窗口放大缩小时，动的幅度
// 相对画面恒定，和参考项目按 3840×2160 标尺写死位移的做法观感一致。
type Motion struct {
	// Person 是人物区域遮罩（灰度，非零处才算人物）——呼吸的总闸门：
	// 它以外的地方严格等于原图，建筑与天空纹丝不动。空表示没有呼吸。
	//
	// 甘雨实测覆盖 x 34.2%~59.4% / y 22.2%~73.7%，是一条竖直的人形带（4.9%）。
	Person string
	// Breath 是呼吸幅度：0.0022 约合 4K 图 4.8px、周期约 4.5 秒，是静息呼吸的量级。
	Breath float32
	// Hair 是发丝/衣摆遮罩，R 通道是基础层、G 通道是细缕层（一张图两个通道）。
	//
	// 分两层是为了「一缕一缕分开飘」：两层频率、速度、方向都不同，细缕处比
	// 周围多一份位移。两层都会被各自通道的权重压住，不会跑到人物之外。
	Hair string
	// Sway 是发丝飘动的幅度（图片 UV）。细缕层在着色器里再乘 1.3。
	Sway float32
	// Ambient 是环境漂移遮罩：只有大块柔和的区域（天空、光、雾、水）才动，
	// 硬边建筑基本不动——它和 Person 一样是"闸门"思路，只是闸门换成"柔不柔"。
	Ambient string
	// Drift 是环境漂移的幅度（图片 UV）。0.0015 约合 4K 图 3.2px。
	Drift float32
	// DriftAngle 是漂移的方向（弧度，0 = 向右，π/2 = 向下）。
	DriftAngle float64
	// DriftSpeed 是漂移的相位速度倍率（1 = 标准，0 视为 0.05）。水慢、风快。
	DriftSpeed float32
	// Blink 是眨眼用的闭眼补丁（RGBA，与 Bg 同构图、同归一化坐标；空表示不眨眼）。
	//
	// 补丁只在不透明处生效，运行时按整张壁纸铺上去、用 alpha 淡入淡出，
	// 所以不需要额外的位置参数。素材由 scripts/assetgen 从 `.hbc` 的闭眼帧烘出来。
	Blink string
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
	// Opening 决定开场动画的专属分镜；绘制层按枚举分发，不认角色名。
	Opening Opening
	// Chrome 决定三个窗口按钮的造型；Ornament 决定侧栏、卡片上的角色纹样。
	Chrome   Chrome
	Ornament Ornament
	Tok      Tokens

	// CardBg 是主题卡片的纯色底。每张卡片一块纯色，选中时加元素色描边。
	CardBg ui.Color
	// Art 是卡片右侧的装饰线稿（与左上的元素徽记是两个不同的图标）。
	Art Art

	// Bg 是壁纸在 assets/ 下的相对路径（每个角色一张，均从原项目导入）。
	//
	// 八重神子与宵宫的原图是**空场景**：人物在原项目里是运行时单独一层网格
	// （.hbc）画的，所以这两张 bg.jpg 是脚本把人物合成进去后重新导出的
	// （见 scripts/assetgen）。
	Bg string
	// Icon 是头像的相对路径；空表示没有头像素材，界面用元素色圆点占位。
	Icon string

	// Motion 是壁纸的动态配置（呼吸 / 发丝 / 环境漂移）。零值 = 完全静态。
	Motion Motion

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
		Element: Cryo, Motif: MotifCrystal, Opening: OpeningGanyu, Chrome: ChromeCrystal, Ornament: OrnamentQingxin,
		WallTint: 0.15, Scrim: 0.32,
		CardBg: ui.Hex("#14314A"), Art: ArtNone,
		Bg: "ganyu/bg.jpg", Icon: "ganyu/icon.png", Tok: ganyuTokens(),
		// 甘雨壁纸动态，分两层机制：
		//
		// 1) 呼吸感：Person（人物轮廓）当总闸门，只做垂直微起伏。
		//    0.0022 约合 4K 图 4.8px、周期约 4.5 秒——静息呼吸的量级。
		//    闸门之外（建筑、天空）严格等于原图，一个像素都不动。
		//
		// 2) 发丝分层：Hair 的 R（整片发区）+ G（一缕细层）叠出层次。
		//    两者都会被人物闸门裁到人物之内——这是关键：flow-mask 本身覆盖
		//    y 0%~42% 的横带、含天空与建筑，只能当发丝层，绝不能当总闸门。
		//
		// 关于 wave-b：实测 bbox y 0%~20.3%，落在人物轮廓（y 从 22.2% 起）之上，
		// 被闸门裁掉后几乎不剩，故不启用。
		//
		// 甘雨**不配 Ambient**：璃月的建筑和天空是硬边，整幅环境漂移会露馅。
		Motion: Motion{
			Person: "ganyu/person.png", Breath: 0.0022,
			Hair: "ganyu/hair.png", Sway: 0.0016,
		},
	},
	{
		ID: "yoimiya", NameCN: "宵宫", NameEN: "Yoimiya", Epithet: "烟火",
		Motto: "琉焰华舞", Constellation: "C A R A S S I U S   A U R A T U S",
		Scene: "稻妻 · 夏祭", SceneEN: "FIREWORKS AT DUSK",
		Element: Pyro, Motif: MotifFlame, Opening: OpeningYoimiya, Chrome: ChromePyro, Ornament: OrnamentGoldfish,
		WallTint: 0.18, Scrim: 0.47,
		CardBg: ui.Hex("#3E2418"), Art: ArtNone,
		Bg: "yoimiya/bg.jpg", Tok: yoimiyaTokens(),
		// 宵宫：人物从 .hbc 网格烤进背景图（原图的背景是没有人的空场景）。
		//
		// 她的 .hbc 每层遮罩都是空的（1×1），拿不到发丝权重，所以发丝区按
		// 「头顶一团 + 左侧发梢」两个椭圆取，再乘人物遮罩，保证不会飘出人物。
		// 环境漂移给 0.0014：夏祭的暖光在流动，纸烟花本身是平的，动起来不违和。
		Motion: Motion{
			Person: "yoimiya/person.png", Breath: 0.0018,
			Hair: "yoimiya/hair.png", Sway: 0.0024,
			Ambient: "yoimiya/ambient.png", Drift: 0.0014, DriftAngle: 0.9, DriftSpeed: 0.6,
		},
	},
	{
		ID: "yae", NameCN: "八重神子", NameEN: "Yae Miko", Epithet: "神樱",
		Motto: "浮世笑百姿", Constellation: "D I V I N A   V U L P E S",
		Scene: "鸣神 · 神樱", SceneEN: "SACRED SAKURA",
		Element: Electro, Motif: MotifBolt, Opening: OpeningYae, Chrome: ChromeElectro, Ornament: OrnamentSakura,
		// 压暗值按合成后壁纸的实测亮度给（0.70 → 0.32+0.38×0.55 ≈ 0.53，取 0.56 留一点余量）。
		// 原背景是一整片亮樱色（0.81），烤进人物后暗部变多，0.66 会把人物压得看不清。
		WallTint: 0.15, Scrim: 0.56,
		CardBg: ui.Hex("#37203F"), Art: ArtNone,
		Bg: "yae/bg.jpg", Tok: yaeTokens(),
		// 八重神子：原图的背景只有樱花枝（人物也是 .hbc 网格，已烤进 bg.jpg）。
		//
		// 遮罩全部从 .hbc 的图层推出来：人物 = 所有图层的 alpha 并集；发丝基础层 =
		// 各图层自带的 mask.r（原项目就是拿它做发丝位移的）按 alpha 加权取并集；
		// 细缕层 = 其中权重最大的一层（实测是右侧那条长辫，正好是最该飘的地方）。
		// 所以她的发丝权重是「数据驱动」的，不是手画椭圆。
		//
		// 环境漂移 0.0018：整幅都是虚化的樱色天空，风一吹整片花瓣都在走。
		//
		// 眨眼（Blink）：`.hbc` 里本来就有一段闭眼帧，生成资源时把「眼睛层用闭眼帧、
		// 其余图层保持睁眼帧」烘成一张补丁。没有它，再大的位移也只是一张会晃的图。
		Motion: Motion{
			Person: "yae/person.png", Breath: 0.0022,
			Hair: "yae/hair.png", Sway: 0.0032,
			Ambient: "yae/ambient.png", Drift: 0.0022, DriftAngle: 0.45, DriftSpeed: 0.62,
			Blink: "yae/eyes.png",
		},
	},
	// 以下角色有自己的壁纸与专属 Motif（草叶 / 水滴 / 风纹）。
	{
		ID: "nahida", NameCN: "纳西妲", NameEN: "Nahida", Epithet: "新芽",
		Motto: "白草净华", Constellation: "S A P I E N T I A   O R O M A S D I S",
		Scene: "须弥 · 世界树", SceneEN: "DREAM OF WISDOM",
		Element: Dendro, Motif: MotifLeaf, Opening: OpeningNahida, Chrome: ChromeDendro, Ornament: OrnamentSprig,
		WallTint: 0.15, Scrim: 0.66,
		CardBg: ui.Hex("#1F3A22"), Art: ArtNone,
		Bg: "nahida/bg.jpg", Tok: derive(ui.Hex("#A8E6A0")),
		// 纳西妲：整幅都是她的特写。人物遮罩从「细节强度」推（虚化的背景细节少、
		// 脸和手细节多），发丝直接用原项目做横向波动的 nahida-wave.png。
		// 梦境本来就该慢慢流动，环境漂移给 0.0012、速度 0.35。
		Motion: Motion{
			Person: "nahida/person.png", Breath: 0.0016,
			Hair: "nahida/hair.png", Sway: 0.0020,
			Ambient: "nahida/ambient.png", Drift: 0.0012, DriftAngle: 0.15, DriftSpeed: 0.35,
		},
	},
	{
		ID: "furina", NameCN: "芙宁娜", NameEN: "Furina", Epithet: "水幕",
		Motto: "不休独舞", Constellation: "A N I M U L A   C H O R A G I",
		Scene: "枫丹 · 歌剧院", SceneEN: "THE STAGE AWAITS",
		Element: Hydro, Motif: MotifWave, Opening: OpeningFurina, Chrome: ChromeHydro, Ornament: OrnamentTopHat,
		WallTint: 0.14, Scrim: 0.35,
		CardBg: ui.Hex("#15294A"), Art: ArtNone,
		Bg: "furina/bg.jpg", Tok: derive(ui.Hex("#8FD0FF")),
		// 芙宁娜：发丝遮罩用原项目的 furina-sway.png（就是她发梢的形状），
		// 人物遮罩从细节强度推。水下光线本来就慢，漂移速度压到 0.3。
		Motion: Motion{
			Person: "furina/person.png", Breath: 0.0018,
			Hair: "furina/hair.png", Sway: 0.0024,
			Ambient: "furina/ambient.png", Drift: 0.0014, DriftAngle: 0.1, DriftSpeed: 0.3,
		},
	},
	{
		ID: "faruzan", NameCN: "珐露珊", NameEN: "Faruzan", Epithet: "风卷",
		Motto: "百巧千机", Constellation: "F L O S C U L I   I M P L E X I",
		Scene: "须弥 · 教令院", SceneEN: "MACHINA OF WISDOM",
		Element: Anemo, Motif: MotifWind, Opening: OpeningFaruzan, Chrome: ChromeAnemo, Ornament: OrnamentPolyhedron,
		WallTint: 0.16, Scrim: 0.47,
		CardBg: ui.Hex("#17342F"), Art: ArtNone,
		Bg: "faruzan/bg.jpg", Tok: derive(ui.Hex("#9EE8D6")),
		// 珐露珊：原项目只有一张流向图（没有遮罩），发丝区按「头顶 + 左侧马尾」
		// 两个椭圆取，再乘人物遮罩。风元素，漂移速度给到 0.5、方向偏右上。
		Motion: Motion{
			Person: "faruzan/person.png", Breath: 0.0016,
			Hair: "faruzan/hair.png", Sway: 0.0024,
			Ambient: "faruzan/ambient.png", Drift: 0.0015, DriftAngle: 0.55, DriftSpeed: 0.5,
		},
	},
	{
		// 钟离的壁纸原项目只有 .hbw 打包格式，这里是从中提取图层合成出的静态图
		// （金色渐变 + 岩印 + 立绘 + 沙漠）。色彩本身很亮，压暗给到 0.62。
		ID: "zhongli", NameCN: "钟离", NameEN: "Zhongli", Epithet: "岩契",
		Motto: "尘世闲游", Constellation: "L A P I S   D E I",
		Scene: "璃月 · 孤云阁", SceneEN: "STONE HARBOUR",
		Element: Geo, Motif: MotifGeo, Opening: OpeningZhongli, Chrome: ChromeGeo, Ornament: OrnamentStoneSeal,
		WallTint: 0.18, Scrim: 0.62,
		CardBg: ui.Hex("#392C14"), Art: ArtNone,
		Bg: "zhongli/bg.jpg", Tok: zhongliTokens(),
		// 钟离：壁纸是 .hbw 分层容器，人物是第 7 层（id=85），发丝权重取该层
		// 自带的 mask.r；环境漂移取「柔光区域 + 胸口金色漩涡」——漩涡转起来，
		// 沙漠的硬边和岩印不动。
		Motion: Motion{
			Person: "zhongli/person.png", Breath: 0.0016,
			Hair: "zhongli/hair.png", Sway: 0.0020,
			Ambient: "zhongli/ambient.png", Drift: 0.0016, DriftAngle: 0, DriftSpeed: 0.3,
		},
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
