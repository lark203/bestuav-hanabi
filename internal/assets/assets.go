// Package assets 读取本地美术资源（不入版本库，见 .gitignore）。
//
// 每帧都会请求同一张图，所以解码结果必须缓存：1200 万像素的 JPG 每帧解码
// 会直接把帧率打到个位数。加载失败返回 nil，界面自动降级为纯色背景。
//
// 内存：原始素材是 4K（3840×2160，解码成 RGBA 就是 33MB 一张），但窗口通常
// 1280~1920 宽，4K 纹理只是白占内存——切换主题时新旧两张同时驻留，峰值会很高。
// 这里统一**解码时就降到 maxWidth 宽**：壁纸 1920（8.3MB），遮罩本来就只有
// 960 宽（生成脚本的 MASK_W，2MB 一张），一个主题的常驻约 14MB。
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

// maxWidth 是解码后的最大宽度。
//
// 1920 是权衡：1280 窗口下是 1.5 倍超采样，1920 窗口正好 1:1；再大的窗口会略软，
// 换来的是每张壁纸从 33MB 降到 8.3MB。遮罩只有 960 宽，不会触发缩放。
const maxWidth = 1920

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
	// userDir 是用户导入的素材目录（优先于安装目录里的 assets/）。
	userDir string
)

// Dir 返回只读资源根目录（安装包/源码树里的 assets/）。
func Dir() string { return assetsDir }

// SetUserDir 设置可写的用户素材目录；为空时只用 Dir()。
func SetUserDir(dir string) { userDir = dir }

// UserDir 返回导入素材应写入的目录（用户目录优先，否则资源目录）。
func UserDir() string {
	if userDir != "" {
		return userDir
	}
	return assetsDir
}

// Clear 清空解码缓存；导入/替换素材后调用，下次读取会重新解码。
func Clear() {
	mu.Lock()
	cache = map[string]*ui.Bitmap{}
	mu.Unlock()
}

// Bitmap 返回 assets/<rel> 解码后的位图，失败返回 nil（并记住失败，不重试）。
func Bitmap(rel string) *ui.Bitmap {
	mu.Lock()
	defer mu.Unlock()

	if bmp, ok := cache[rel]; ok {
		return bmp
	}
	data, err := os.ReadFile(filepath.Join(userDirOr(assetsDir), rel))
	if err != nil && userDir != "" {
		data, err = os.ReadFile(filepath.Join(assetsDir, rel))
	}
	if err != nil {
		cache[rel] = nil
		return nil
	}
	img, err := decodeScaled(data, maxWidth)
	if err != nil {
		cache[rel] = nil
		return nil
	}
	// 用 NewBitmap 而不是 DecodeBitmap：后者会把压缩数据和"重新解码"的闭包
	// 一起留着，切换主题时既多占内存，又会在 ReleasePixels 之后反复解码。
	bmp := ui.NewBitmap(img)
	cache[rel] = bmp
	return bmp
}

// userDirOr 返回用户素材目录（未设置时回退到资源目录）。
func userDirOr(fallback string) string {
	if userDir != "" {
		return userDir
	}
	return fallback
}

// Keep 只保留仍需要的位图，其余移出缓存。
//
// 一个主题的壁纸 + 三层遮罩约 14MB；七个角色全留着就是 100MB 起，切换主题时
// 必须清掉不用的。位图用 NewBitmap 创建（像素在本进程里），移出缓存后由 GC
// 回收；切回来会重新解码，代价可以接受。
func Keep(keys ...string) {
	mu.Lock()
	defer mu.Unlock()

	for rel := range cache {
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
		delete(cache, rel)
	}
}

// decodeScaled 解码图片并把宽度压到 maxW 以内（等比）。宽度已经够小就原样返回。
//
// 为什么不用 x/image/draw 的 CatmullRom：它对非 RGBA 源（JPEG 解出来是 YCbCr）
// 走通用路径，会额外分配两张 float32 的中间图——一张 4K 图缩到 1920 峰值要
// 139MB 垃圾，切几次主题进程就涨到几百 MB。这里自己按面积平均，只分配目标图。
func decodeScaled(data []byte, maxW int) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= maxW || maxW <= 0 {
		return img, nil
	}
	w := maxW
	h := sh * maxW / sw
	if h < 1 {
		h = 1
	}
	return downscale(img, w, h), nil
}

// downscale 把 src 按面积平均缩到 w×h，只分配目标图。
//
// 绝大多数素材是 4:2:0 的 JPEG 且恰好缩一半，走 fast2x420（4 个 Y + 1 个 Cb/Cr，
// 不经过接口调用）；其它比例（比如钟离的 2559 宽）走通用面积平均。
func downscale(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if y, ok := src.(*image.YCbCr); ok && sw == 2*w && sh == 2*h &&
		y.SubsampleRatio == image.YCbCrSubsampleRatio420 {
		fast2x420(y, dst)
		return dst
	}
	for dy := 0; dy < h; dy++ {
		y0 := b.Min.Y + dy*sh/h
		y1 := b.Min.Y + (dy+1)*sh/h
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for dx := 0; dx < w; dx++ {
			x0 := b.Min.X + dx*sw/w
			x1 := b.Min.X + (dx+1)*sw/w
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, bl, a, n uint32
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					cr, cg, cb, ca := src.At(xx, yy).RGBA()
					r, g, bl, a = r+cr, g+cg, bl+cb, a+ca
					n++
				}
			}
			dst.SetRGBA(dx, dy, color.RGBA{
				R: uint8(r / n >> 8), G: uint8(g / n >> 8),
				B: uint8(bl / n >> 8), A: uint8(a / n >> 8),
			})
		}
	}
	return dst
}

// fast2x420 是 4:2:0 YCbCr 恰好缩小一半的专用路径：
// 每个目标像素取 2×2 的 Y 求平均、共享同一个 Cb/Cr。
func fast2x420(src *image.YCbCr, dst *image.RGBA) {
	b := src.Bounds()
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	for dy := 0; dy < h; dy++ {
		y0 := b.Min.Y + 2*dy
		row := dst.Pix[dy*dst.Stride:]
		for dx := 0; dx < w; dx++ {
			x := b.Min.X + 2*dx
			yy := (uint32(src.Y[src.YOffset(x, y0)]) +
				uint32(src.Y[src.YOffset(x+1, y0)]) +
				uint32(src.Y[src.YOffset(x, y0+1)]) +
				uint32(src.Y[src.YOffset(x+1, y0+1)]) + 2) / 4
			ci := src.COffset(x, y0)
			r, g, bl := color.YCbCrToRGB(uint8(yy), src.Cb[ci], src.Cr[ci])
			row[4*dx], row[4*dx+1], row[4*dx+2], row[4*dx+3] = r, g, bl, 255
		}
	}
}
