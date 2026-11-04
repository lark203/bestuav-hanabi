// Package settings 保存跨启动的界面设置：启动主题、字号、动效开关与壁纸显现。
//
// 文件放在 mygo 的 PathUserData（%AppData%/<应用名>/settings.json）。读写失败
// 一律退回默认值，界面宁可恢复默认，也不能因为一个坏 JSON 打不开。
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
)

// Startup 是“下次启动用哪个主题”。
type Startup int

const (
	// StartupFollow 跟随上次使用的角色主题。
	StartupFollow Startup = iota
	// StartupFixed 固定使用 Settings.StartupTheme。
	StartupFixed
	// StartupRandom 每次从角色主题里随机一个（黑白基础主题不在列表里）。
	StartupRandom
)

// 设置范围。字号与参考项目一致：100%–125%，默认 112%。
const (
	FontMin     = 100
	FontMax     = 125
	FontDefault = 112

	// WallpaperMin/Max 是“壁纸显现”的范围，参考项目 0.2–1.0。
	WallpaperMin = 0.2
	WallpaperMax = 1.0
)

// Settings 是一份可持久化的界面设置。
type Settings struct {
	// Theme 是最近一次使用的角色主题（ StartupFollow 用它）。
	Theme string `json:"theme"`
	// FontPercent 是整体字号百分比，100–125。
	FontPercent float64 `json:"fontPercent"`
	// StartupMode 决定启动时选哪个主题。
	StartupMode Startup `json:"startupMode"`
	// StartupTheme 是 StartupFixed 时的主题 id。
	StartupTheme string `json:"startupTheme"`

	// WallpaperPlay 控制壁纸的呼吸/发丝动效；false 时完全静止。
	WallpaperPlay bool `json:"wallpaperPlay"`
	// WallpaperVis 是壁纸显现程度 0.2–1：越小，压暗越重。
	WallpaperVis float64 `json:"wallpaperVis"`
	// IntroPlay 控制启动时是否播放 6.2 秒开场。
	IntroPlay bool `json:"introPlay"`
	// ReplayIntro 控制切换角色主题时是否重播开场。
	ReplayIntro bool `json:"replayIntro"`
	// Animations 控制页面/主题切换与浮层动效；false 时直接到位。
	Animations bool `json:"animations"`
}

// Default 返回默认设置。
func Default() Settings {
	return Settings{
		Theme:         "ganyu",
		FontPercent:   FontDefault,
		StartupMode:   StartupFollow,
		StartupTheme:  "ganyu",
		WallpaperPlay: true,
		WallpaperVis:  1,
		IntroPlay:     true,
		Animations:    true,
	}
}

// Normalize 修掉越界或缺失的值；Load 与 UI 修改后都走这里。
func (s Settings) Normalize() Settings {
	if s.FontPercent == 0 {
		s.FontPercent = FontDefault
	}
	s.FontPercent = clamp(s.FontPercent, FontMin, FontMax)
	if s.WallpaperVis == 0 {
		s.WallpaperVis = 1
	}
	s.WallpaperVis = clamp(s.WallpaperVis, WallpaperMin, WallpaperMax)
	if s.StartupMode < StartupFollow || s.StartupMode > StartupRandom {
		s.StartupMode = StartupFollow
	}
	if s.Theme == "" {
		s.Theme = "ganyu"
	}
	if s.StartupTheme == "" {
		s.StartupTheme = s.Theme
	}
	return s
}

// Scale 把 100% 基准的字号换算成当前字号乘数。
func (s Settings) Scale() float32 {
	p := s.FontPercent
	if p == 0 {
		p = FontDefault
	}
	p = clamp(p, FontMin, FontMax)
	return float32(p) / 100
}

// EffectiveScrim 把“壁纸显现”叠加到角色自带的压暗值上。
//
// 每个角色的 Scrim 是按壁纸实测亮度定的；把显现拉到 0.2 时再额外压暗
// 0.45，避免甘雨这种本来就亮的壁纸在“低显现”下仍然抢字。
func (s Settings) EffectiveScrim(base float32) float32 {
	vis := s.WallpaperVis
	if vis == 0 {
		vis = 1
	}
	vis = clamp(vis, WallpaperMin, WallpaperMax)
	extra := float32(1-vis) * 0.45
	return clampF(base+extra, 0, 0.95)
}

// Load 读取默认位置；文件不存在或损坏时返回默认设置。
func Load() Settings {
	path, err := DefaultPath()
	if err != nil {
		return Default()
	}
	return LoadFrom(path)
}

// LoadFrom 从指定文件读取；测试用它避开真实用户目录。
func LoadFrom(path string) Settings {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Default()
	}
	return s.Normalize()
}

// Save 保存到默认位置。
func Save(s Settings) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return SaveTo(path, s)
}

// SaveTo 保存到指定文件（先写临时文件再改名，避免半个 JSON）。
func SaveTo(path string, s Settings) error {
	s = s.Normalize()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		// 某些 Windows 情况下改名会被占用；退化为直接覆盖，至少不丢设置。
		return os.WriteFile(path, data, 0o644)
	}
	return nil
}

// Reset 把偏好恢复默认，旧文件改名成 settings.json.bak-<时间>（能找回）。
// 返回备份路径；文件本来不存在时返回空。
func Reset() (string, error) {
	p, err := DefaultPath()
	if err != nil {
		return "", err
	}
	backup := ""
	if _, statErr := os.Stat(p); statErr == nil {
		backup = p + ".bak-" + time.Now().Format("20060102-150405")
		if err := os.Rename(p, backup); err != nil {
			return "", err
		}
	}
	if err := Save(Default()); err != nil {
		return backup, err
	}
	return backup, nil
}

// DefaultPath 返回 settings.json 的默认路径。
func DefaultPath() (string, error) {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", errors.New("settings: 用户目录为空")
	}
	return filepath.Join(dir, "settings.json"), nil
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
