// Package visual 是界面的「视觉层」边界：角色主题、令牌、纹样、壁纸与开场。
//
// view 只依赖这个包，不直接 import theme / wallpaper / intro；将来换实现（例如
// 主题包改为运行时加载）时，view 不需要改。当前是薄封装 + 类型别名，行为不变。
package visual

import (
	"math"
	"sync"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/intro"
	"bestuav-hanabi/internal/theme"
	"bestuav-hanabi/internal/wallpaper"
)

// 类型别名：让 view 无需感知内部包，同时保留原有字段与方法。
type (
	Character = theme.Character
	Tokens    = theme.Tokens
	Element   = theme.Element
	Motif     = theme.Motif
	Ornament  = theme.Ornament
	Chrome    = theme.Chrome
	Opening   = theme.Opening
	Art       = theme.Art
	Motion    = theme.Motion
	Flow      = wallpaper.Flow
	Options   = wallpaper.Options
	Intro     = intro.Intro
	MarkPose  = intro.MarkPose
)

// 元素 / 纹样 / 窗饰 / 开场 / 标识绘制法的常量。
const (
	Anemo   = theme.Anemo
	Cryo    = theme.Cryo
	Dendro  = theme.Dendro
	Electro = theme.Electro
	Geo     = theme.Geo
	Hydro   = theme.Hydro
	Pyro    = theme.Pyro

	MotifBolt  = theme.MotifBolt
	MotifFlame = theme.MotifFlame
	MotifGeo   = theme.MotifGeo
	MotifLeaf  = theme.MotifLeaf
	MotifWave  = theme.MotifWave
	MotifWind  = theme.MotifWind

	OrnamentGoldfish   = theme.OrnamentGoldfish
	OrnamentNone       = theme.OrnamentNone
	OrnamentPolyhedron = theme.OrnamentPolyhedron
	OrnamentQingxin    = theme.OrnamentQingxin
	OrnamentSakura     = theme.OrnamentSakura
	OrnamentSprig      = theme.OrnamentSprig
	OrnamentStoneSeal  = theme.OrnamentStoneSeal
	OrnamentTopHat     = theme.OrnamentTopHat

	ChromeAnemo   = theme.ChromeAnemo
	ChromeCrystal = theme.ChromeCrystal
	ChromeDendro  = theme.ChromeDendro
	ChromeElectro = theme.ChromeElectro
	ChromeGeo     = theme.ChromeGeo
	ChromeHydro   = theme.ChromeHydro
	ChromePlain   = theme.ChromePlain
	ChromePyro    = theme.ChromePyro

	OpeningFaruzan = theme.OpeningFaruzan
	OpeningFurina  = theme.OpeningFurina
	OpeningGanyu   = theme.OpeningGanyu
	OpeningNahida  = theme.OpeningNahida
	OpeningYae     = theme.OpeningYae
	OpeningYoimiya = theme.OpeningYoimiya
	OpeningZhongli = theme.OpeningZhongli

	ArtCrystal  = theme.ArtCrystal
	ArtFirework = theme.ArtFirework
	ArtLeaf     = theme.ArtLeaf
	ArtNone     = theme.ArtNone
	ArtRing     = theme.ArtRing
	ArtSail     = theme.ArtSail
	ArtSakura   = theme.ArtSakura
)

// 开场时间轴的常量与方法转发。
const (
	TurnStart = intro.TurnStart
	TurnEnd   = intro.TurnEnd
	SnowCount = intro.SnowCount
	LantCount = intro.LantCount
)

// All / ByID / Default 主题清单。
func All() []Character {
	base := theme.All()
	extMu.RLock()
	defer extMu.RUnlock()
	if len(external) == 0 {
		return base
	}
	out := make([]Character, 0, len(base)+len(external))
	out = append(out, base...)
	out = append(out, external...)
	return out
}

func ByID(id string) (Character, bool) {
	if c, ok := theme.ByID(id); ok {
		return c, true
	}
	extMu.RLock()
	defer extMu.RUnlock()
	for _, c := range external {
		if c.ID == id {
			return c, true
		}
	}
	return Character{}, false
}

func Default() Character { return theme.Default() }

var (
	extMu    sync.RWMutex
	external []Character
)

// RegisterExternal 注册运行时导入的主题（重复 id/与内建冲突的跳过）。
//
// 在创建 App 之前调用（main 启动时），保证主题库与启动主题都能看到它们。
func RegisterExternal(chs []Character) {
	extMu.Lock()
	defer extMu.Unlock()
	known := map[string]bool{}
	for _, c := range theme.All() {
		known[c.ID] = true
	}
	for _, c := range chs {
		if c.ID == "" || known[c.ID] {
			continue
		}
		replaced := false
		for i := range external {
			if external[i].ID == c.ID {
				external[i] = c // 重新导入同 id：覆盖
				replaced = true
				break
			}
		}
		if !replaced {
			external = append(external, c)
		}
	}
}

// ExternalThemes 返回已注册的外部主题（导入/删除管理用）。
func ExternalThemes() []Character {
	extMu.RLock()
	defer extMu.RUnlock()
	out := make([]Character, len(external))
	copy(out, external)
	return out
}

// RemoveExternal 从运行时注册表移除外部主题。
func RemoveExternal(id string) {
	extMu.Lock()
	defer extMu.Unlock()
	for i, c := range external {
		if c.ID == id {
			external = append(external[:i], external[i+1:]...)
			return
		}
	}
}

// BuiltinIDs 返回内建主题 id（导入主题包时用于防冲突）。
func BuiltinIDs() map[string]bool {
	out := map[string]bool{}
	for _, c := range theme.All() {
		out[c.ID] = true
	}
	return out
}

// Background / Icon 角色素材（缺素材时返回 nil，界面退化为纯色）。
func Background(c Character) *ui.Bitmap { return assets.Bitmap(c.Bg) }
func Icon(c Character) *ui.Bitmap       { return assets.Bitmap(c.Icon) }

// Keep 保留素材缓存（切换主题的过渡期用）。
func Keep(paths ...string) { assets.Keep(paths...) }

// WallpaperFlowFor 组装一个角色的动态壁纸参数；没有人物遮罩时返回 nil。
func WallpaperFlowFor(c Character) *Flow {
	m := c.Motion
	if m.Person == "" {
		return nil
	}
	person := assets.Bitmap(m.Person)
	if person == nil {
		return nil
	}
	f := &Flow{
		Person:     person,
		Breath:     m.Breath,
		Sway:       m.Sway,
		Drift:      m.Drift,
		DriftSpeed: m.DriftSpeed,
		DriftDir:   [2]float32{float32(math.Cos(m.DriftAngle)), float32(math.Sin(m.DriftAngle))},
	}
	if m.Hair != "" {
		f.Hair = assets.Bitmap(m.Hair)
	}
	if m.Ambient != "" {
		f.Ambient = assets.Bitmap(m.Ambient)
	}
	return f
}

// WallpaperKeys 过渡期要保留的素材：壁纸、头像与三层动态遮罩。
func WallpaperKeys(c Character) []string {
	keep := []string{c.Bg, c.Icon}
	for _, p := range []string{c.Motion.Person, c.Motion.Hair, c.Motion.Ambient} {
		if p != "" {
			keep = append(keep, p)
		}
	}
	return keep
}

// Draw 绘制动态壁纸（转发 wallpaper.Draw）。
func Draw(p *ui.Painter, bg *ui.Bitmap, r ui.Rect, o Options) { wallpaper.Draw(p, bg, r, o) }

// New 创建开场动画。
func New(w, h float32) *Intro { return intro.New(w, h) }

// Smooth 开场用的平滑插值。
func Smooth(t, a, b float32) float32 { return intro.Smooth(t, a, b) }

// Engine 视觉层接口：view 只需要这些能力；当前实现是上面的包函数。
type Engine interface {
	All() []Character
	ByID(string) (Character, bool)
	Default() Character
	Background(Character) *ui.Bitmap
	Icon(Character) *ui.Bitmap
	Wallpaper(Character) *Flow
	WallpaperKeys(Character) []string
	NewIntro(w, h float32) *Intro
}

// DefaultEngine 默认实现。
var DefaultEngine Engine = engine{}

type engine struct{}

func (engine) All() []Character                   { return All() }
func (engine) ByID(id string) (Character, bool)   { return ByID(id) }
func (engine) Default() Character                 { return Default() }
func (engine) Background(c Character) *ui.Bitmap  { return Background(c) }
func (engine) Icon(c Character) *ui.Bitmap        { return Icon(c) }
func (engine) Wallpaper(c Character) *Flow        { return WallpaperFlowFor(c) }
func (engine) WallpaperKeys(c Character) []string { return WallpaperKeys(c) }
func (engine) NewIntro(w, h float32) *Intro       { return New(w, h) }
