package pagegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	view := filepath.Join(root, "internal", "view")
	icons := filepath.Join(root, "internal", "ui", "icons")
	if err := os.MkdirAll(view, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(icons, 0o755); err != nil {
		t.Fatal(err)
	}
	reg := `package view

import "bestuav-hanabi/internal/ui/icons"

func navRegistry() []navItem {
	return []navItem{
		{id: settingsPageID, label: "设置", icon: icons.Sliders, build: (*App).settingsPage}, //gen:id=settings
	}
}
`
	if err := os.WriteFile(filepath.Join(view, "registry.go"), []byte(reg), 0o644); err != nil {
		t.Fatal(err)
	}
	ic := `package icons

import "github.com/egoist/mygo/ui"

var (
	Sliders = ui.MustParseSVG([]byte("<svg/>"))
	Square  = ui.MustParseSVG([]byte("<svg/>"))
)
`
	if err := os.WriteFile(filepath.Join(icons, "icons.go"), []byte(ic), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestAddPage(t *testing.T) {
	root := fixture(t)
	file, err := Add(Options{Root: root, ID: "notes-list", Label: "笔记", SkipBuild: true})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(file) != "page_notes_list.go" {
		t.Errorf("文件名 = %s", file)
	}
	page, _ := os.ReadFile(file)
	if !strings.Contains(string(page), "const NotesListPageID = \"notes-list\"") ||
		!strings.Contains(string(page), "func (a *App) notesListPage(") {
		t.Errorf("页面内容不对:\n%s", page)
	}
	reg, _ := os.ReadFile(filepath.Join(root, "internal", "view", "registry.go"))
	if !strings.Contains(string(reg), "//gen:id=notes-list") ||
		!strings.Contains(string(reg), "icon: icons.Square") {
		t.Errorf("注册表未插入:\n%s", reg)
	}
	// settings 仍应在最后。
	if strings.Index(string(reg), "gen:id=notes-list") > strings.Index(string(reg), "gen:id=settings") {
		t.Error("新页面应插在 settings 之前")
	}
}

func TestAddPageRejects(t *testing.T) {
	root := fixture(t)
	if _, err := Add(Options{Root: root, ID: "Bad_ID", Label: "x", SkipBuild: true}); err == nil {
		t.Error("非法 id 应报错")
	}
	if _, err := Add(Options{Root: root, ID: "x", Label: "", SkipBuild: true}); err == nil {
		t.Error("空标题应报错")
	}
	if _, err := Add(Options{Root: root, ID: "x", Label: "X", Icon: "Nope", SkipBuild: true}); err == nil {
		t.Error("未知图标应报错")
	}
	if _, err := Add(Options{Root: root, ID: "notes", Label: "笔记", SkipBuild: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(Options{Root: root, ID: "notes", Label: "笔记", SkipBuild: true}); err == nil {
		t.Error("重复页面应报错")
	}
}

func TestCamel(t *testing.T) {
	for in, want := range map[string]string{
		"notes": "Notes", "notes-list": "NotesList", "a-b-c": "ABC",
	} {
		if got := camel(in); got != want {
			t.Errorf("camel(%q) = %q", in, got)
		}
	}
}
