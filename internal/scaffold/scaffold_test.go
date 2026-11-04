package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot 返回 bestuav-hanabi 仓库根（本测试在 internal/scaffold 下）。
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("找不到模板根目录: %v", err)
	}
	return root
}

func TestGenerateRewrites(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "newapp")
	res, err := Generate(Options{
		Root:       repoRoot(t),
		Dir:        dst,
		Module:     "example.com/acme/newapp",
		Name:       "newapp",
		Title:      "新应用",
		Brand:      "NewApp",
		Identifier: "com.acme.newapp",
		Assets:     "none",
		KeepPages:  []string{"library", "settings"},
		SkipBuild:  true,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Files == 0 {
		t.Error("没有复制文件")
	}

	// 模块路径。
	gomod, _ := os.ReadFile(filepath.Join(dst, "go.mod"))
	if !strings.Contains(string(gomod), "module example.com/acme/newapp") {
		t.Errorf("go.mod 未改写:\n%s", gomod)
	}
	am, _ := os.ReadFile(filepath.Join(dst, "internal", "view", "app.go"))
	if !strings.Contains(string(am), `"example.com/acme/newapp/internal/visual"`) {
		t.Error("Go import 未改写")
	}

	// 产品标识。
	meta, _ := os.ReadFile(filepath.Join(dst, "internal", "appmeta", "appmeta.go"))
	for _, want := range []string{
		`const Name = "newapp"`, `const Title = "新应用"`,
		`const Brand = "NewApp"`, `const Identifier = "com.acme.newapp"`,
	} {
		if !strings.Contains(string(meta), want) {
			t.Errorf("appmeta 缺少 %q", want)
		}
	}

	// mygo.json：版本归零、不继承更新源、发布者改写。
	var mj map[string]any
	data, _ := os.ReadFile(filepath.Join(dst, "mygo.json"))
	if err := json.Unmarshal(data, &mj); err != nil {
		t.Fatal(err)
	}
	if mj["version"] != "0.1.0" || mj["slug"] != "newapp" || mj["identifier"] != "com.acme.newapp" {
		t.Errorf("mygo.json 字段不对: %v", mj)
	}
	if _, ok := mj["updates"]; ok {
		t.Error("不应继承更新源")
	}
	if mj["publisher"] != "新应用" {
		t.Errorf("publisher = %v", mj["publisher"])
	}

	// 页面裁剪：只留 library/settings。
	reg, _ := os.ReadFile(filepath.Join(dst, "internal", "view", "registry.go"))
	if strings.Contains(string(reg), "gen:id=wallpaper") {
		t.Error("wallpaper 页面应被裁掉")
	}
	if !strings.Contains(string(reg), "gen:id=library") || !strings.Contains(string(reg), "gen:id=settings") {
		t.Error("保留页面缺失")
	}

	// 素材 none：assets 目录不存在。
	if _, err := os.Stat(filepath.Join(dst, "assets")); !os.IsNotExist(err) {
		t.Errorf("assets 应被清空: %v", err)
	}

	// README 标题；生成项目里不应再出现「如何生成新项目」一节。
	readme, _ := os.ReadFile(filepath.Join(dst, "README.md"))
	if !strings.HasPrefix(string(readme), "# 新应用（newapp）") {
		t.Errorf("README 标题未改写:\n%s", firstLine(string(readme)))
	}
	if strings.Contains(string(readme), "作为新项目基座") {
		t.Error("README 里不应保留模板使用说明")
	}
}

// TestGenerateFromEmbeddedSnapshot 不依赖源码树（安装包场景）也能生成。
func TestGenerateFromEmbeddedSnapshot(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "embedded")
	res, err := Generate(Options{
		Root: "", Dir: dst, Module: "example.com/acme/emb",
		Name: "emb", Title: "内置模板", Identifier: "com.acme.emb",
		Assets: "none", SkipBuild: true,
	})
	if err != nil {
		t.Fatalf("Generate(embedded): %v", err)
	}
	if res.Files == 0 {
		t.Error("没有复制文件")
	}
	if _, err := os.Stat(filepath.Join(dst, "go.mod")); err != nil {
		t.Errorf("go.mod 未还原: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "internal", "scaffold")); !os.IsNotExist(err) {
		t.Error("模板不应包含生成器自身")
	}
}

func TestGenerateRejectsNonEmptyTarget(t *testing.T) {
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(dst, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Generate(Options{
		Root: repoRoot(t), Dir: dst, Module: "example.com/a/b",
		Name: "b", Title: "B", Identifier: "com.a.b", Assets: "none", SkipBuild: true,
	})
	if err == nil || !strings.Contains(err.Error(), "非空") {
		t.Errorf("非空目标应拒绝: %v", err)
	}
}

func TestGenerateRequiresSettings(t *testing.T) {
	_, err := Generate(Options{
		Root: repoRoot(t), Dir: filepath.Join(t.TempDir(), "x"),
		Module: "example.com/a/b", Name: "b", Title: "B", Identifier: "com.a.b",
		Assets: "none", KeepPages: []string{"library"}, SkipBuild: true,
	})
	if err == nil || !strings.Contains(err.Error(), "settings") {
		t.Errorf("缺 settings 应拒绝: %v", err)
	}
}

// TestGeneratedProjectBuilds 端到端：生成后真的 go build。
// 默认跳过（耗时），CI 或本地用 `go test -run TestGeneratedProjectBuilds` 跑。
func TestGeneratedProjectBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("short 模式跳过真编译")
	}
	dst := filepath.Join(t.TempDir(), "built")
	_, err := Generate(Options{
		Root:       repoRoot(t),
		Dir:        dst,
		Module:     "example.com/acme/built",
		Name:       "built",
		Title:      "编译测试",
		Identifier: "com.acme.built",
		Assets:     "none",
	})
	if err != nil {
		t.Fatalf("生成或编译失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "go.mod")); err != nil {
		t.Fatal(err)
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
