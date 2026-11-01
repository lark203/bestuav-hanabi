// Package assets 读取本地美术资源（不入版本库，见 .gitignore）。
//
// 每帧都会请求同一张图，所以解码结果必须缓存：1200 万像素的 JPG 每帧解码
// 会直接把帧率打到个位数。加载失败返回 nil，界面自动降级为纯色背景。
package assets

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// root 是资源子目录名（相对项目根 / 应用资源目录）。
const root = "assets"

// assetsDir 定位资源根目录。
//
// 三种运行形态，优先级从「打包后」到「源码树」：
//  1. 打包产物（含 `mygo dev`）：资源被复制到 mygo.PathResources，读那里的 assets/；
//  2. `go test` / `go run`：工作目录不固定，用 go.mod 作项目根锚点；
//  3. 兜底：exe 同级。
//
// 之所以用 go.mod 当锚点而不是「有没有 assets 目录」：本项目有同名包目录
// internal/assets，按目录名找会先命中它，导致所有图片静默加载失败（踩过）。
var assetsDir = findAssetsDir()

func findAssetsDir() string {
	// 1) 应用资源目录（打包后 / mygo dev 下有效；纯 go test 时该目录不存在）。
	if dir, err := mygo.App.Path(mygo.PathResources); err == nil {
		cand := filepath.Join(dir, root)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
	}
	// 2) 源码树：向上找 go.mod。
	if dir, err := os.Getwd(); err == nil {
		for i := 0; i < 6; i++ {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(dir, root)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// 3) 兜底。
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), root)
	}
	return root
}

var (
	mu    sync.Mutex
	cache = map[string]*ui.Bitmap{}
)

// Bitmap 返回 assets/<rel> 解码后的位图，失败返回 nil（并记住失败，不重试）。
func Bitmap(rel string) *ui.Bitmap {
	mu.Lock()
	defer mu.Unlock()

	if bmp, ok := cache[rel]; ok {
		return bmp
	}

	data, err := os.ReadFile(filepath.Join(assetsDir, rel))
	if err != nil {
		cache[rel] = nil
		return nil
	}
	bmp, err := ui.DecodeBitmap(data)
	if err != nil {
		cache[rel] = nil
		return nil
	}
	cache[rel] = bmp
	return bmp
}
