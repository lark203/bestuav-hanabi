// Package assets 读取本地美术资源（不入版本库，见 .gitignore）。
//
// 每帧都会请求同一张图，所以解码结果必须缓存：1200 万像素的 JPG 每帧解码
// 会直接把帧率打到个位数。加载失败返回 nil，界面自动降级为纯色背景。
package assets

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
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
	// thumbs 单独一张缓存：缩略图体积小（每张几百 KB），可以长期留着，
	// 主题库要同时显示七个角色，不能为此解码七张 4K 原图。
	thumbs = map[string]*ui.Bitmap{}
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

// Thumb 返回 assets/<rel> 的缩略图（宽度至多 maxW），失败返回 nil。
//
// 自己解码再降采样，而不是先 Bitmap 再缩放：缩略图只需要小尺寸，
// 没必要为此把 4K 图整张解码进内存。
func Thumb(rel string, maxW int) *ui.Bitmap {
	mu.Lock()
	defer mu.Unlock()

	key := rel + "@" + itoa(maxW)
	if b, ok := thumbs[key]; ok {
		return b
	}

	data, err := os.ReadFile(filepath.Join(assetsDir, rel))
	if err != nil {
		thumbs[key] = nil
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		thumbs[key] = nil
		return nil
	}
	b := ui.NewBitmap(scaleDown(img, maxW))
	thumbs[key] = b
	return b
}

// Keep 只保留仍需要的位图，其余的释放像素并移出缓存。
//
// 一张 4K 壁纸解码后约 33 MB（3840×2160×4）。七个角色的壁纸全留在缓存里
// 就是 200 MB 起——切换主题时必须清掉不用的。保留原始文件数据的话
// （DecodeBitmap 的语义），切回来会重新解码，代价可接受。
func Keep(keys ...string) {
	mu.Lock()
	defer mu.Unlock()

	for rel, bmp := range cache {
		wanted := false
		for _, k := range keys {
			if k == rel {
				wanted = true
				break
			}
		}
		if wanted {
			continue
		}
		if bmp != nil {
			bmp.ReleasePixels()
		}
		delete(cache, rel)
	}
}

// scaleDown 用盒式平均把图缩到宽度 maxW（保持比例）。
//
// 只用在一处：主题库缩略图。降采样的质量在这里够用，也不想为此引入依赖。
// image.RGBA 是预乘 alpha，At() 给出的也是预乘值，逐通道平均后写回即可。
func scaleDown(src image.Image, maxW int) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == 0 || sh == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if maxW > sw {
		maxW = sw
	}
	w := maxW
	h := sh * w / sw
	if h < 1 {
		h = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		y0 := b.Min.Y + y*sh/h
		y1 := b.Min.Y + (y+1)*sh/h
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < w; x++ {
			x0 := b.Min.X + x*sw/w
			x1 := b.Min.X + (x+1)*sw/w
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var sr, sg, sb, sa, n uint64
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					r, g, bl, a := src.At(xx, yy).RGBA()
					sr, sg, sb, sa = sr+uint64(r), sg+uint64(g), sb+uint64(bl), sa+uint64(a)
					n++
				}
			}
			if n == 0 {
				continue
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(sr / n >> 8), G: uint8(sg / n >> 8),
				B: uint8(sb / n >> 8), A: uint8(sa / n >> 8),
			})
		}
	}
	return dst
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
