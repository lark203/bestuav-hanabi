// Package apppath 是配置 / 数据库 / 日志路径的唯一来源。
//
// 统一走 mygo.App.Path(mygo.PathUserData)：mygo 已经把「用户配置目录/应用名」拼好，
// 这里不再追加应用名，避免出现 <配置目录>/hanabi/hanabi。测试可用 mygo.App.SetPath
// 改到临时目录。
package apppath

import (
	"errors"
	"path/filepath"

	"github.com/egoist/mygo"
)

// Dir 返回本应用的 per-user 配置目录。
func Dir() (string, error) {
	base, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return "", err
	}
	if base == "" {
		return "", errors.New("apppath: 用户目录为空")
	}
	return base, nil
}

// SettingsPath 返回 settings.json 的路径。
func SettingsPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "settings.json"), nil
}

// DBPath 返回 SQLite 数据库路径。
func DBPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "data.db"), nil
}

// LogDir 返回日志目录（尚无日志实现时也可用于诊断包）。
func LogDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "logs"), nil
}

// ThemesDir 返回外部主题包目录（运行时加载）。
func ThemesDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "themes"), nil
}
