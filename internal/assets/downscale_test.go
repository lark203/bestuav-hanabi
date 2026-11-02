package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// TestDownscale 钉住面积平均的尺寸与颜色：
// 2×2 纯色块缩一半后颜色不变，尺寸按比例取整。
func TestDownscale(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			// 左半边红、右半边蓝
			c := color.RGBA{R: 255, A: 255}
			if x >= 4 {
				c = color.RGBA{B: 255, A: 255}
			}
			src.SetRGBA(x, y, c)
		}
	}
	dst := downscale(src, 4, 2)
	if dst.Bounds().Dx() != 4 || dst.Bounds().Dy() != 2 {
		t.Fatalf("尺寸应为 4x2，实际 %v", dst.Bounds().Size())
	}
	if got := dst.RGBAAt(0, 0); got.R < 250 || got.B > 5 {
		t.Errorf("左上应为红色，实际 %v", got)
	}
	if got := dst.RGBAAt(3, 1); got.B < 250 || got.R > 5 {
		t.Errorf("右下应为蓝色，实际 %v", got)
	}
}

// TestDecodeScaledKeepsSmall 检查宽度已达标时不做缩放（避免无谓的内存与失真）。
func TestDecodeScaledKeepsSmall(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	img, err := decodeScaled(buf.Bytes(), 1920)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 2 {
		t.Fatalf("小图不应缩放，实际 %v", img.Bounds().Size())
	}
}

// TestBitmapMissing 检查缺图时返回 nil 且不 panic（界面要能降级）。
func TestBitmapMissing(t *testing.T) {
	if b := Bitmap("no-such/theme.png"); b != nil {
		t.Error("缺图应返回 nil")
	}
}
