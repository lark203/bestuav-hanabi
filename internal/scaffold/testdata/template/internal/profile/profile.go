// Package profile 导出/导入「设置 + 页面状态」配置包（JSON）。
//
// 不含素材与角色版权资源：换机时把配置带过去即可，壁纸仍由本机 assets 提供。
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bestuav-hanabi/internal/settings"
	"bestuav-hanabi/internal/store"
)

// Format 当前包格式版本。
const Format = 1

// Bundle 一个配置包。
type Bundle struct {
	Format    int                          `json:"format"`
	App       string                       `json:"app"`
	Version   string                       `json:"version"`
	Exported  time.Time                    `json:"exportedAt"`
	Settings  settings.Settings            `json:"settings"`
	PagePrefs map[string]map[string]string `json:"pagePrefs,omitempty"`
}

// Export 写配置文件；prefs 为空时也可以（无数据库）。
func Export(path, app, version string, s settings.Settings, prefs map[string]map[string]string) error {
	b := Bundle{
		Format:    Format,
		App:       app,
		Version:   version,
		Exported:  time.Now(),
		Settings:  s.Normalize(),
		PagePrefs: prefs,
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Import 读取并校验配置文件。
func Import(path string) (Bundle, error) {
	var b Bundle
	data, err := os.ReadFile(path)
	if err != nil {
		return b, err
	}
	if err := json.Unmarshal(data, &b); err != nil {
		return b, fmt.Errorf("不是合法的配置文件: %w", err)
	}
	if b.Format <= 0 || b.Format > Format {
		return b, fmt.Errorf("配置包格式版本 %d 不受支持", b.Format)
	}
	b.Settings = b.Settings.Normalize()
	return b, nil
}

// Prefs 从数据库收集页面状态；st 为 nil 时返回空。
func Prefs(st *store.Store) map[string]map[string]string {
	if st == nil {
		return nil
	}
	prefs, err := st.AllPagePrefs()
	if err != nil {
		return nil
	}
	return prefs
}
