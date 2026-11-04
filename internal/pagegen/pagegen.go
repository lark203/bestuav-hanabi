// Package pagegen 新增一个页面：生成页面文件 + 在 registry.go 里插入导航条目。
//
// 约定与 mygo 版一致：页面 id 用小写连字符，函数名用驼峰，注册表保持字面量；
// 只做「文本插入 + gofmt」，不做 AST 手术，生成失败会回滚已写的文件。
package pagegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Options 生成参数。
type Options struct {
	Root      string // 仓库根
	ID        string // 页面 id，如 notes
	Label     string // 侧栏/页头标题
	Icon      string // icons 包里的名字，默认 Square
	SkipBuild bool
}

var (
	idRe     = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	iconRe   = regexp.MustCompile(`(?m)^\s*([A-Z][A-Za-z0-9]*)\s*=\s*ui\.MustParseSVG`)
	settings = "//gen:id=settings"
)

// Add 生成页面并插入注册表，返回页面文件路径。
func Add(o Options) (string, error) {
	if o.Root == "" {
		return "", errors.New("未指定仓库根目录")
	}
	if !idRe.MatchString(o.ID) {
		return "", fmt.Errorf("页面 id 只能是小写字母/数字/连字符，且以字母开头: %q", o.ID)
	}
	if strings.TrimSpace(o.Label) == "" {
		return "", errors.New("标题不能为空")
	}
	if strings.ContainsAny(o.Label, "\"\n\r") {
		return "", errors.New("标题里不能有引号或换行")
	}
	if o.Icon == "" {
		o.Icon = "Square"
	}
	if err := checkIcon(o.Root, o.Icon); err != nil {
		return "", err
	}

	registryPath := filepath.Join(o.Root, "internal", "view", "registry.go")
	reg, err := os.ReadFile(registryPath)
	if err != nil {
		return "", fmt.Errorf("读取 registry.go 失败: %w", err)
	}
	if bytes.Contains(reg, []byte("//gen:id="+o.ID)) {
		return "", fmt.Errorf("页面已存在: %s", o.ID)
	}
	pagePath := filepath.Join(o.Root, "internal", "view", "page_"+strings.ReplaceAll(o.ID, "-", "_")+".go")
	if _, err := os.Stat(pagePath); err == nil {
		return "", fmt.Errorf("页面文件已存在: %s", pagePath)
	}

	camel := camel(o.ID)
	fn := lowerFirst(camel)
	src := fmt.Sprintf(`package view

import "github.com/egoist/mygo/ui"

// %sPageID 页面 id（注册表与页面状态共用）。
const %sPageID = %q

// %sPage %s。
func (a *App) %sPage(c *ui.Context) {
	ui.Column(c).Children(func() {
		ui.Text(c, %q).FontSize(a.fs(22)).Bold().TextColor(a.tok.Text)
		ui.Text(c, "该页面由 scripts/newpage 生成，按需替换内容。").
			FontSize(a.fs(11)).TextColor(a.tok.TextFaint)
	})
}
`, camel, camel, o.ID, camel, o.Label, fn, o.Label)

	pageSrc, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("生成页面代码失败: %w", err)
	}

	entry := fmt.Sprintf("\t\t{id: %sPageID, label: %q, icon: icons.%s, build: (*App).%sPage}, //gen:id=%s\n",
		camel, o.Label, o.Icon, fn, o.ID)
	idx := bytes.Index(reg, []byte(settings))
	if idx < 0 {
		return "", errors.New("registry.go 里找不到 //gen:id=settings 标记")
	}
	lineStart := bytes.LastIndexByte(reg[:idx], '\n') + 1
	updated := make([]byte, 0, len(reg)+len(entry))
	updated = append(updated, reg[:lineStart]...)
	updated = append(updated, entry...)
	updated = append(updated, reg[lineStart:]...)
	regSrc, err := format.Source(updated)
	if err != nil {
		return "", fmt.Errorf("插入注册表后格式失败: %w", err)
	}

	if err := os.WriteFile(pagePath, pageSrc, 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(registryPath, regSrc, 0o644); err != nil {
		_ = os.Remove(pagePath)
		return "", err
	}

	if !o.SkipBuild {
		if out, err := build(o.Root); err != nil {
			// 编译失败：回滚，避免留下半个页面。
			_ = os.Remove(pagePath)
			_ = os.WriteFile(registryPath, reg, 0o644)
			return "", fmt.Errorf("编译失败，已回滚：%w\n%s", err, out)
		}
	}
	return pagePath, nil
}

func checkIcon(root, icon string) error {
	data, err := os.ReadFile(filepath.Join(root, "internal", "ui", "icons", "icons.go"))
	if err != nil {
		return fmt.Errorf("读取 icons.go 失败: %w", err)
	}
	for _, m := range iconRe.FindAllStringSubmatch(string(data), -1) {
		if m[1] == icon {
			return nil
		}
	}
	return fmt.Errorf("icons 包里没有 %q（可选：%s）", icon, strings.Join(iconNames(string(data)), ", "))
}

func iconNames(src string) []string {
	var out []string
	for _, m := range iconRe.FindAllStringSubmatch(src, -1) {
		out = append(out, m[1])
	}
	return out
}

// camel 把 notes-list 变成 NotesList。
func camel(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// lowerFirst 把 NotesList 变成 notesList（页面函数用小驼峰）。
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func build(root string) (string, error) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}
