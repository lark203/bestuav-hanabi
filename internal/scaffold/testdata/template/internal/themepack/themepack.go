// Package themepack 导入「外部角色主题包」：一个 zip，里面有 theme.json 与
// 该角色目录下的素材。
//
// 与 assetpack 的区别：assetpack 只给已有角色补图；themepack 会新增一个主题
// （角色 id、名称、令牌、纹样、开场分镜、壁纸/遮罩），运行时注册进 visual。
// 主题包只能复用内建的绘制法（元素/纹样/窗饰/开场/线稿都是枚举），不能带代码。
package themepack

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/theme"
)

// Format 当前主题包格式版本。
const Format = 1

// MaxBytes 单包大小上限。
const MaxBytes = 300 << 20

var (
	idRe  = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	hexRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
)

// Pack 是 theme.json 的结构。
type Pack struct {
	Format        int    `json:"format"`
	ID            string `json:"id"`
	NameCN        string `json:"nameCN"`
	NameEN        string `json:"nameEN"`
	Epithet       string `json:"epithet"`
	Motto         string `json:"motto"`
	Constellation string `json:"constellation"`
	Scene         string `json:"scene"`
	SceneEN       string `json:"sceneEN"`

	Element  string `json:"element"`  // cryo/pyro/electro/dendro/hydro/anemo/geo
	Motif    string `json:"motif"`    // crystal/flame/bolt/geo/leaf/wave/wind
	Opening  string `json:"opening"`  // ganyu/yoimiya/yae/nahida/furina/faruzan/zhongli
	Chrome   string `json:"chrome"`   // plain/crystal/pyro/electro/dendro/hydro/anemo/geo
	Ornament string `json:"ornament"` // none/goldfish/sakura/tophat/polyhedron/qingxin/sprig/stoneSeal
	Art      string `json:"art"`      // blossom/sakura/firework/leaf/sail/crystal/ring/none

	CardBg string     `json:"cardBg"`
	Bg     string     `json:"bg"`
	Icon   string     `json:"icon"`
	Tokens TokensPack `json:"tokens"`
	Motion MotionPack `json:"motion"`
	// 壁纸染色/压暗；省略时用 0.15 / 0.18。
	WallTint float32 `json:"wallTint"`
	Scrim    float32 `json:"scrim"`
}

// TokensPack 颜色都用 #RRGGBB / #RRGGBBAA。
type TokensPack struct {
	WinFill, WinStroke, Sidebar, Card, CardStroke string
	Text, TextMuted, TextFaint                    string
	Accent, AccentSoft, Good, Danger              string
	Radius                                        float32
}

// MotionPack 动态壁纸参数（路径相对资源根）。
type MotionPack struct {
	Person, Hair, Ambient string
	Breath, Sway, Drift   float32
	DriftAngle            float64
	DriftSpeed            float32
}

// Install 校验并安装一个主题包，返回可直接注册的 Character。
//
// themesDir 保存 theme.json，assetsDir 保存图片；两者通常是用户目录下的
// themes/ 与 assets/。known 是内建角色 id（不允许覆盖）。
func Install(zipPath, themesDir, assetsDir string, known map[string]bool) (theme.Character, error) {
	var ch theme.Character
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return ch, fmt.Errorf("打不开主题包: %w", err)
	}
	defer zr.Close()

	var pack Pack
	found := false
	var images []*zip.File
	var total int64
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
			return ch, fmt.Errorf("主题包包含非法路径: %s", f.Name)
		}
		if name == "theme.json" {
			rc, err := f.Open()
			if err != nil {
				return ch, err
			}
			err = json.NewDecoder(io.LimitReader(rc, 1<<20)).Decode(&pack)
			rc.Close()
			if err != nil {
				return ch, fmt.Errorf("theme.json 解析失败: %w", err)
			}
			found = true
			continue
		}
		if !imageExt(name) {
			return ch, fmt.Errorf("主题包里只能放图片: %s", name)
		}
		total += int64(f.UncompressedSize64)
		if total > MaxBytes {
			return ch, errors.New("主题包过大（>300MB）")
		}
		images = append(images, f)
	}
	if !found {
		return ch, errors.New("主题包缺少 theme.json")
	}
	if pack.Format <= 0 || pack.Format > Format {
		return ch, fmt.Errorf("主题包格式版本 %d 不受支持", pack.Format)
	}
	if !idRe.MatchString(pack.ID) {
		return ch, fmt.Errorf("非法角色 id: %q", pack.ID)
	}
	if known[pack.ID] {
		return ch, fmt.Errorf("角色 id %q 与内建主题冲突", pack.ID)
	}
	if strings.TrimSpace(pack.NameCN) == "" {
		return ch, errors.New("nameCN 不能为空")
	}
	if !strings.HasPrefix(pack.Bg, pack.ID+"/") {
		return ch, fmt.Errorf("bg 必须位于 %s/ 下", pack.ID)
	}

	// 先解到临时目录，校验 bg 存在，再一次性复制。
	tmp, err := os.MkdirTemp("", "hanabi-themepack-*")
	if err != nil {
		return ch, err
	}
	defer os.RemoveAll(tmp)

	seen := map[string]bool{}
	for _, f := range images {
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		rel := strings.TrimPrefix(name, pack.ID+"/")
		if rel == "" || rel == name || strings.Contains(rel, "..") {
			return ch, fmt.Errorf("图片应放在 %s/ 下: %s", pack.ID, name)
		}
		if err := extract(f, filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			return ch, err
		}
		seen[name] = true
	}
	if !seen[path.Clean(pack.Bg)] {
		return ch, fmt.Errorf("theme.json 里的 bg 不在包内: %s", pack.Bg)
	}

	ch, err = pack.character()
	if err != nil {
		return ch, err
	}
	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		return ch, err
	}
	data, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return ch, err
	}
	if err := os.WriteFile(filepath.Join(themesDir, pack.ID+".json"), append(data, '\n'), 0o644); err != nil {
		return ch, err
	}
	if err := copyTree(tmp, filepath.Join(assetsDir, pack.ID)); err != nil {
		return ch, err
	}
	return ch, nil
}

// Inspect 只读 theme.json（校验格式与 id），用于「重新导入」时确认还是同一个角色。
func Inspect(zipPath string) (Pack, error) {
	var pack Pack
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return pack, fmt.Errorf("打不开主题包: %w", err)
	}
	defer zr.Close()
	found := false
	for _, f := range zr.File {
		if path.Clean(strings.ReplaceAll(f.Name, "\\", "/")) != "theme.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return pack, err
		}
		err = json.NewDecoder(io.LimitReader(rc, 1<<20)).Decode(&pack)
		rc.Close()
		if err != nil {
			return pack, fmt.Errorf("theme.json 解析失败: %w", err)
		}
		found = true
		break
	}
	if !found {
		return pack, errors.New("主题包缺少 theme.json")
	}
	if pack.Format <= 0 || pack.Format > Format {
		return pack, fmt.Errorf("主题包格式版本 %d 不受支持", pack.Format)
	}
	if !idRe.MatchString(pack.ID) {
		return pack, fmt.Errorf("非法角色 id: %q", pack.ID)
	}
	return pack, nil
}

// Remove 删除一个已安装的主题包（theme.json 与 assets/<id>）。
func Remove(themesDir, assetsDir, id string) error {
	if !idRe.MatchString(id) {
		return fmt.Errorf("非法角色 id: %q", id)
	}
	if err := os.Remove(filepath.Join(themesDir, id+".json")); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.RemoveAll(filepath.Join(assetsDir, id)); err != nil {
		return err
	}
	return nil
}

// LoadDir 读取目录里的 *.json 主题包；坏文件跳过并返回错误列表（启动时只记日志）。
func LoadDir(dir string) ([]theme.Character, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []error{err}
	}
	var out []theme.Character
	var errs []error
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var pack Pack
		if err := json.Unmarshal(data, &pack); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		ch, err := pack.character()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		out = append(out, ch)
	}
	return out, errs
}

func (p Pack) character() (theme.Character, error) {
	ch := theme.Character{
		ID: p.ID, NameCN: p.NameCN, NameEN: p.NameEN, Epithet: p.Epithet,
		Motto: p.Motto, Constellation: p.Constellation,
		Scene: p.Scene, SceneEN: p.SceneEN,
		Bg: p.Bg, Icon: p.Icon,
		WallTint: p.WallTint, Scrim: p.Scrim,
		Motion: theme.Motion{
			Person: p.Motion.Person, Breath: p.Motion.Breath,
			Hair: p.Motion.Hair, Sway: p.Motion.Sway,
			Ambient: p.Motion.Ambient, Drift: p.Motion.Drift,
			DriftAngle: p.Motion.DriftAngle, DriftSpeed: p.Motion.DriftSpeed,
		},
	}
	var err error
	if ch.Element, err = element(p.Element); err != nil {
		return ch, err
	}
	if ch.Motif, err = motif(p.Motif); err != nil {
		return ch, err
	}
	if ch.Opening, err = opening(p.Opening); err != nil {
		return ch, err
	}
	if ch.Chrome, err = chrome(p.Chrome); err != nil {
		return ch, err
	}
	if ch.Ornament, err = ornament(p.Ornament); err != nil {
		return ch, err
	}
	if ch.Art, err = art(p.Art); err != nil {
		return ch, err
	}
	ch.Tok, err = tokens(p.Tokens)
	if err != nil {
		return ch, err
	}
	if ch.CardBg, err = color(p.CardBg, "cardBg"); err != nil {
		return ch, err
	}
	if ch.WallTint == 0 {
		ch.WallTint = 0.15
	}
	if ch.Scrim == 0 {
		ch.Scrim = 0.18
	}
	return ch, nil
}

func tokens(t TokensPack) (theme.Tokens, error) {
	var tok theme.Tokens
	fields := []struct {
		name string
		hex  string
		dst  *ui.Color
	}{
		{"winFill", t.WinFill, &tok.WinFill}, {"winStroke", t.WinStroke, &tok.WinStroke},
		{"sidebar", t.Sidebar, &tok.Sidebar}, {"card", t.Card, &tok.Card},
		{"cardStroke", t.CardStroke, &tok.CardStroke}, {"text", t.Text, &tok.Text},
		{"textMuted", t.TextMuted, &tok.TextMuted}, {"textFaint", t.TextFaint, &tok.TextFaint},
		{"accent", t.Accent, &tok.Accent}, {"accentSoft", t.AccentSoft, &tok.AccentSoft},
		{"good", t.Good, &tok.Good}, {"danger", t.Danger, &tok.Danger},
	}
	for _, f := range fields {
		c, err := color(f.hex, f.name)
		if err != nil {
			return tok, err
		}
		*f.dst = c
	}
	tok.Radius = t.Radius
	if tok.Radius <= 0 {
		tok.Radius = 10
	}
	return tok, nil
}

func color(hex, name string) (ui.Color, error) {
	if !hexRe.MatchString(strings.TrimSpace(hex)) {
		return ui.Color{}, fmt.Errorf("%s 必须是 #RRGGBB 或 #RRGGBBAA（当前 %q）", name, hex)
	}
	return ui.Hex(strings.TrimSpace(hex)), nil
}

func element(s string) (theme.Element, error) {
	switch strings.ToLower(s) {
	case "cryo":
		return theme.Cryo, nil
	case "pyro":
		return theme.Pyro, nil
	case "electro":
		return theme.Electro, nil
	case "dendro":
		return theme.Dendro, nil
	case "hydro":
		return theme.Hydro, nil
	case "anemo":
		return theme.Anemo, nil
	case "geo":
		return theme.Geo, nil
	}
	return 0, fmt.Errorf("未知 element: %q", s)
}

func motif(s string) (theme.Motif, error) {
	switch strings.ToLower(s) {
	case "crystal":
		return theme.MotifCrystal, nil
	case "flame":
		return theme.MotifFlame, nil
	case "bolt":
		return theme.MotifBolt, nil
	case "geo":
		return theme.MotifGeo, nil
	case "leaf":
		return theme.MotifLeaf, nil
	case "wave":
		return theme.MotifWave, nil
	case "wind":
		return theme.MotifWind, nil
	}
	return 0, fmt.Errorf("未知 motif: %q", s)
}

func opening(s string) (theme.Opening, error) {
	switch strings.ToLower(s) {
	case "ganyu":
		return theme.OpeningGanyu, nil
	case "yoimiya":
		return theme.OpeningYoimiya, nil
	case "yae":
		return theme.OpeningYae, nil
	case "nahida":
		return theme.OpeningNahida, nil
	case "furina":
		return theme.OpeningFurina, nil
	case "faruzan":
		return theme.OpeningFaruzan, nil
	case "zhongli":
		return theme.OpeningZhongli, nil
	}
	return 0, fmt.Errorf("未知 opening: %q", s)
}

func chrome(s string) (theme.Chrome, error) {
	switch strings.ToLower(s) {
	case "plain":
		return theme.ChromePlain, nil
	case "crystal":
		return theme.ChromeCrystal, nil
	case "pyro":
		return theme.ChromePyro, nil
	case "electro":
		return theme.ChromeElectro, nil
	case "dendro":
		return theme.ChromeDendro, nil
	case "hydro":
		return theme.ChromeHydro, nil
	case "anemo":
		return theme.ChromeAnemo, nil
	case "geo":
		return theme.ChromeGeo, nil
	}
	return 0, fmt.Errorf("未知 chrome: %q", s)
}

func ornament(s string) (theme.Ornament, error) {
	switch strings.ToLower(s) {
	case "none":
		return theme.OrnamentNone, nil
	case "goldfish":
		return theme.OrnamentGoldfish, nil
	case "sakura":
		return theme.OrnamentSakura, nil
	case "tophat":
		return theme.OrnamentTopHat, nil
	case "polyhedron":
		return theme.OrnamentPolyhedron, nil
	case "qingxin":
		return theme.OrnamentQingxin, nil
	case "sprig":
		return theme.OrnamentSprig, nil
	case "stoneseal":
		return theme.OrnamentStoneSeal, nil
	}
	return 0, fmt.Errorf("未知 ornament: %q", s)
}

func art(s string) (theme.Art, error) {
	switch strings.ToLower(s) {
	case "blossom":
		return theme.ArtBlossom, nil
	case "sakura":
		return theme.ArtSakura, nil
	case "firework":
		return theme.ArtFirework, nil
	case "leaf":
		return theme.ArtLeaf, nil
	case "sail":
		return theme.ArtSail, nil
	case "crystal":
		return theme.ArtCrystal, nil
	case "ring":
		return theme.ArtRing, nil
	case "none":
		return theme.ArtNone, nil
	}
	return 0, fmt.Errorf("未知 art: %q", s)
}

func imageExt(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp":
		return true
	}
	return false
}

func extract(f *zip.File, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, io.LimitReader(rc, MaxBytes)); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}
