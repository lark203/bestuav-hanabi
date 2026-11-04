// Package assetpack 导入「角色素材包」：一个 zip，里面有 manifest.json 与
// 该角色目录下的图片（bg.jpg / icon.png / 遮罩等）。
//
// 只补素材，不新增角色：角色 id 必须是已编译进 theme 的 id；新增角色仍需改
// theme.go 数据（后续 P2 的外部主题包再做）。
package assetpack

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Format 当前素材包格式版本。
const Format = 1

// MaxBytes 单包大小上限（壁纸 4K 原图也不该超过这个量级）。
const MaxBytes = 300 << 20

// Manifest 素材包清单。
type Manifest struct {
	Format int    `json:"format"`
	ID     string `json:"id"`   // 角色 id，如 ganyu
	Name   string `json:"name"` // 展示名（仅提示用）
}

// Result 导入结果。
type Result struct {
	ID    string
	Files int
	Bytes int64
}

// Import 把 zip 里的角色素材复制到 assetsDir/<id>/。
//
// known 是允许的角色 id 集合（来自主题表）；任何路径穿越、未知 id、
// 非图片文件或超限包都会被拒绝，且不会修改现有素材。
func Import(zipPath, assetsDir string, known map[string]bool) (Result, error) {
	var res Result
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return res, fmt.Errorf("打不开素材包: %w", err)
	}
	defer zr.Close()

	var manifest Manifest
	found := false
	var total int64
	var images []*zip.File
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		if name == "manifest.json" {
			rc, err := f.Open()
			if err != nil {
				return res, err
			}
			err = json.NewDecoder(io.LimitReader(rc, 1<<20)).Decode(&manifest)
			rc.Close()
			if err != nil {
				return res, fmt.Errorf("manifest.json 解析失败: %w", err)
			}
			found = true
			continue
		}
		if strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") || strings.Contains(name, "/../") {
			return res, fmt.Errorf("素材包包含非法路径: %s", f.Name)
		}
		if !imageExt(name) {
			return res, fmt.Errorf("素材包里只能放图片: %s", name)
		}
		total += int64(f.UncompressedSize64)
		if total > MaxBytes {
			return res, errors.New("素材包过大（>300MB）")
		}
		images = append(images, f)
	}
	if !found {
		return res, errors.New("素材包缺少 manifest.json")
	}
	if manifest.Format <= 0 || manifest.Format > Format {
		return res, fmt.Errorf("素材包格式版本 %d 不受支持", manifest.Format)
	}
	if !known[manifest.ID] {
		return res, fmt.Errorf("未知角色 id: %q（只有主题表里已有的角色能补素材）", manifest.ID)
	}
	if len(images) == 0 {
		return res, errors.New("素材包里没有图片")
	}

	// 先全部解到临时目录并校验角色前缀，再一次性复制过去。
	tmp, err := os.MkdirTemp("", "hanabi-assetpack-*")
	if err != nil {
		return res, err
	}
	defer os.RemoveAll(tmp)

	for _, f := range images {
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		prefix := manifest.ID + "/"
		if !strings.HasPrefix(name, prefix) {
			return res, fmt.Errorf("图片应放在 %s 下: %s", prefix, name)
		}
		rel := strings.TrimPrefix(name, prefix)
		if rel == "" || strings.Contains(rel, "..") {
			return res, fmt.Errorf("非法文件名: %s", name)
		}
		dest := filepath.Join(tmp, filepath.FromSlash(rel))
		if err := extract(f, dest); err != nil {
			return res, err
		}
		res.Files++
	}

	dstDir := filepath.Join(assetsDir, manifest.ID)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return res, err
	}
	if err := copyTree(tmp, dstDir); err != nil {
		return res, err
	}
	res.ID = manifest.ID
	res.Bytes = total
	return res, nil
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
