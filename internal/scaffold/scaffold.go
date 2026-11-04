// Package scaffold 以当前仓库为模板生成一个新项目。
//
// 做法是「复制 + 文本改写」，不做 AST 裁剪：
//   - Go 模块路径、产品名/标题/标识、README、mygo.json 全量改写；
//   - 侧栏/页面按 registry.go 里的条目裁剪（页面代码保留，避免删文件导致引用断裂）；
//   - 素材可选：none（纯色降级）/ demo（一个角色）/ full（全部）。
//
// 生成后默认跑一次 go build ./...，确保交付的是「能编译」的工程。
package scaffold

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// templateFS 是仓库快照（scripts/snapshot-template.py 生成）；从安装包/无源码目录
// 生成新项目时用它，仓库内开发时直接读源码树。
//
//go:embed all:testdata/template
var templateFS embed.FS

const templateDir = "testdata/template"

// Options 生成参数。
type Options struct {
	Root       string   // 模板仓库根目录（含 go.mod）
	Dir        string   // 目标目录（必须不存在或为空）
	Module     string   // 新模块路径，如 example.com/acme/app
	Name       string   // 产品名/slug（配置目录、窗口状态键）
	Title      string   // 窗口标题/品牌中文名
	Brand      string   // 侧栏短品牌名（默认取 Name）
	Identifier string   // 反向 DNS 应用标识
	Assets     string   // none / demo / full
	DemoAsset  string   // demo 模式保留的角色目录（默认 ganyu）
	KeepPages  []string // 保留的页面 id；空 = 全部
	SkipBuild  bool     // 跳过编译自检（测试/调试用）
}

// Result 生成结果。
type Result struct {
	Dir         string
	Files       int
	Pages       []string
	Assets      string
	BuildOutput string
	Warnings    []string
}

var pageIDRe = regexp.MustCompile(`//gen:id=([a-z0-9-]+)`)

// Generate 复制并改写模板。
func Generate(o Options) (Result, error) {
	var res Result
	if strings.TrimSpace(o.Root) == "" {
		root, cleanup, err := materializeTemplate()
		if err != nil {
			return res, err
		}
		defer cleanup()
		o.Root = root
	}
	if strings.TrimSpace(o.Dir) == "" {
		return res, errors.New("未指定目标目录")
	}
	if o.Module == "" || !strings.Contains(o.Module, "/") {
		return res, fmt.Errorf("模块路径不合法: %q", o.Module)
	}
	if o.Name == "" || o.Title == "" || o.Identifier == "" {
		return res, errors.New("Name / Title / Identifier 都不能为空")
	}
	if o.Brand == "" {
		o.Brand = o.Name
	}
	if o.Assets == "" {
		o.Assets = "demo"
	}
	switch o.Assets {
	case "none", "demo", "full":
	default:
		return res, fmt.Errorf("未知素材模式: %q", o.Assets)
	}
	if o.DemoAsset == "" {
		o.DemoAsset = "ganyu"
	}

	if err := checkTarget(o.Dir); err != nil {
		return res, err
	}
	oldModule, err := readModule(o.Root)
	if err != nil {
		return res, err
	}
	allPages, err := readPages(filepath.Join(o.Root, "internal", "view", "registry.go"))
	if err != nil {
		return res, err
	}
	keep := o.KeepPages
	if len(keep) == 0 {
		keep = allPages
	}
	keepSet := map[string]bool{}
	for _, p := range keep {
		if !contains(allPages, p) {
			return res, fmt.Errorf("未知页面: %q（可选：%s）", p, strings.Join(allPages, ", "))
		}
		keepSet[p] = true
	}
	if !keepSet["settings"] {
		return res, errors.New("必须保留 settings 页面（否则没有设置入口）")
	}

	if err := copyTree(o.Root, o.Dir, o.Assets, o.DemoAsset); err != nil {
		return res, err
	}
	if err := copyAssets(o.Root, o.Dir, o.Assets, o.DemoAsset); err != nil {
		return res, err
	}
	if err := rewriteTree(o.Dir, oldModule, o); err != nil {
		return res, err
	}
	if err := pruneRegistry(filepath.Join(o.Dir, "internal", "view", "registry.go"), keepSet); err != nil {
		return res, err
	}
	if err := rewriteAppmeta(filepath.Join(o.Dir, "internal", "appmeta", "appmeta.go"), o); err != nil {
		return res, err
	}
	if err := rewriteMygoJSON(filepath.Join(o.Dir, "mygo.json"), o); err != nil {
		return res, err
	}
	if err := rewriteReadme(filepath.Join(o.Dir, "README.md"), o); err != nil {
		return res, err
	}
	if err := writeGeneratedCI(o.Dir); err != nil {
		return res, err
	}

	res.Dir = o.Dir
	res.Pages = keep
	res.Assets = o.Assets
	if n, err := countFiles(o.Dir); err == nil {
		res.Files = n
	}

	if !o.SkipBuild {
		out, err := build(o.Dir)
		res.BuildOutput = out
		if err != nil {
			return res, fmt.Errorf("生成成功但编译失败：%w\n%s", err, out)
		}
	}
	return res, nil
}

func checkTarget(dir string) error {
	st, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("目标已存在且不是目录: %s", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("目标目录非空（为避免覆盖，请换一个目录）: %s", dir)
	}
	return nil
}

func readModule(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("读取 go.mod 失败: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", errors.New("go.mod 里没有 module 行")
}

func readPages(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 registry.go 失败: %w", err)
	}
	var out []string
	for _, m := range pageIDRe.FindAllStringSubmatch(string(data), -1) {
		out = append(out, m[1])
	}
	if len(out) == 0 {
		return nil, errors.New("registry.go 里没有页面条目")
	}
	return out, nil
}

// copyTree 复制模板；assets 单独处理（可裁剪）。
func copyTree(root, dst, assets, demo string) error {
	skipDirs := map[string]bool{
		".git": true, ".idea": true, ".workbuddy": true, "build": true,
		"assets": true, // 单独复制
		// 生成器本身不进入模板（否则内嵌快照会递归）。
		filepath.Join("internal", "scaffold"):  true,
		filepath.Join("scripts", "newproject"): true,
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if skipDirs[rel] {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		// 仓库自己的 CI / 快照脚本依赖模板仓库结构；生成项目换成通用 CI。
		if rel == filepath.Join(".github", "workflows", "ci.yml") ||
			rel == filepath.Join("scripts", "snapshot-template.py") {
			return nil
		}
		if err := copyFile(path, filepath.Join(dst, rel)); err != nil {
			return err
		}
		return nil
	})
}

func copyAssets(root, dst, mode, demo string) error {
	src := filepath.Join(root, "assets")
	if _, err := os.Stat(src); err != nil {
		return nil // 模板本身没有素材也能生成（纯色降级）
	}
	dstDir := filepath.Join(dst, "assets")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	switch mode {
	case "none":
		return os.RemoveAll(dstDir)
	case "full":
		return copyDir(src, dstDir)
	default: // demo
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				if e.Name() != demo {
					continue
				}
				if err := copyDir(filepath.Join(src, e.Name()), filepath.Join(dstDir, e.Name())); err != nil {
					return err
				}
				continue
			}
			if err := copyFile(filepath.Join(src, e.Name()), filepath.Join(dstDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// rewriteTree 把所有文本文件里的旧模块路径换成新模块路径。
func rewriteTree(dir, oldModule string, o Options) error {
	exts := map[string]bool{
		".go": true, ".md": true, ".json": true, ".yml": true, ".yaml": true,
		".ps1": true, ".py": true, ".txt": true, ".mod": true, ".sum": true,
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !exts[strings.ToLower(filepath.Ext(path))] && filepath.Base(path) != ".gitignore" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		changed := strings.ReplaceAll(string(data), oldModule, o.Module)
		if changed != string(data) {
			return os.WriteFile(path, []byte(changed), 0o644)
		}
		return nil
	})
}

var (
	nameRe  = regexp.MustCompile(`(?m)^const Name = ".*"$`)
	titleRe = regexp.MustCompile(`(?m)^const Title = ".*"$`)
	brandRe = regexp.MustCompile(`(?m)^const Brand = ".*"$`)
	idRe    = regexp.MustCompile(`(?m)^const Identifier = ".*"$`)
)

func rewriteAppmeta(path string, o Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	t := string(data)
	t = nameRe.ReplaceAllString(t, fmt.Sprintf("const Name = %q", o.Name))
	t = titleRe.ReplaceAllString(t, fmt.Sprintf("const Title = %q", o.Title))
	t = brandRe.ReplaceAllString(t, fmt.Sprintf("const Brand = %q", o.Brand))
	t = idRe.ReplaceAllString(t, fmt.Sprintf("const Identifier = %q", o.Identifier))
	return os.WriteFile(path, []byte(t), 0o644)
}

func rewriteMygoJSON(path string, o Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	m["name"] = o.Title
	m["slug"] = o.Name
	m["identifier"] = o.Identifier
	m["version"] = "0.1.0"
	m["publisher"] = o.Title
	m["copyright"] = "Copyright © 2026 " + o.Title
	// 新项目不继承 hanabi 的更新源/公钥。
	delete(m, "updates")
	if o.Assets == "none" {
		delete(m, "icon")
		delete(m, "resources")
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func pruneRegistry(path string, keep map[string]bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		id := ""
		if m := pageIDRe.FindStringSubmatch(line); m != nil {
			id = m[1]
		}
		if id != "" && !keep[id] {
			continue // 删掉这一条导航
		}
		out = append(out, line)
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644)
}

func rewriteReadme(path string, o Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
		lines[0] = fmt.Sprintf("# %s（%s）", o.Title, o.Name)
	}
	body := strings.Join(lines, "\n")
	body = strings.ReplaceAll(body, "花火 HanaBi", o.Title)
	// 生成项目不需要「如何生成新项目」这一节。
	if i := strings.Index(body, "## 作为新项目基座"); i >= 0 {
		rest := body[i:]
		if j := strings.Index(rest[1:], "\n## "); j >= 0 {
			body = body[:i] + rest[j+2:]
		} else {
			body = body[:i]
		}
	}
	// 标题后插入一行来源说明，其余内容原样保留。
	if i := strings.IndexByte(body, '\n'); i >= 0 {
		note := fmt.Sprintf("\n> 由 bestuav-hanabi 生成的项目骨架：`%s`。\n", o.Module)
		body = body[:i+1] + note + body[i+1:]
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

// generatedCI 是生成项目用的通用 CI（不含模板快照检查）。
const generatedCI = `name: ci

on:
  push:
    branches: ["**"]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - run: go vet ./...
      - run: go test ./...

  build-windows:
    runs-on: windows-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - run: go build ./...
`

func writeGeneratedCI(dir string) error {
	path := filepath.Join(dir, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(generatedCI), 0o644)
}

func build(dir string) (string, error) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func countFiles(dir string) (int, error) {
	n := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	return n, err
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// SortedPages 返回 registry 里的页面 id（供 CLI 提示）。
func SortedPages(root string) ([]string, error) {
	pages, err := readPages(filepath.Join(root, "internal", "view", "registry.go"))
	if err != nil {
		return nil, err
	}
	sort.Strings(pages)
	return pages, nil
}

// materializeTemplate 把内嵌快照解到临时目录，返回根目录与清理函数。
func materializeTemplate() (string, func(), error) {
	dst, err := os.MkdirTemp("", "hanabi-template-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dst) }
	err = fs.WalkDir(templateFS, templateDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, templateDir)
		rel = strings.TrimPrefix(rel, "/")
		// 快照里的 go.mod / go.sum 带 .snapshot 后缀（go:embed 不允许目录内含 go.mod）。
		if strings.HasSuffix(rel, ".snapshot") {
			rel = strings.TrimSuffix(rel, ".snapshot")
		}
		target := filepath.Join(dst, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := templateFS.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
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
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("解压模板快照失败: %w", err)
	}
	return dst, cleanup, nil
}
