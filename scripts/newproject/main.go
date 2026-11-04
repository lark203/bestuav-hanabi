// Command newproject 以当前 bestuav-hanabi 仓库为模板生成一个新项目。
//
// 用法（在仓库根目录）：
//
//	go run ./scripts/newproject -dir ../my-app -module example.com/acme/my-app \
//	  -name my-app -title "我的应用" -id com.acme.myapp
//
// 默认带一个演示角色的素材（-assets demo），也可 -assets none（纯色降级）或 full。
// 生成后可 -pages library,settings 之类裁剪侧栏（settings 必须保留）。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bestuav-hanabi/internal/scaffold"
)

func main() {
	var (
		root       = flag.String("root", ".", "模板仓库根目录")
		dir        = flag.String("dir", "", "目标目录（必须不存在或为空）")
		module     = flag.String("module", "", "新模块路径，如 example.com/acme/app")
		name       = flag.String("name", "", "产品名/slug（配置目录、窗口状态键）")
		title      = flag.String("title", "", "窗口标题/中文产品名")
		brand      = flag.String("brand", "", "侧栏短品牌名（默认同 -name）")
		identifier = flag.String("id", "", "应用标识，如 com.acme.app")
		assets     = flag.String("assets", "demo", "素材：none / demo / full")
		assetsRoot = flag.String("assets-root", "", "素材源目录（默认 <root>/assets）")
		demo       = flag.String("demo", "ganyu", "demo 模式保留的角色目录")
		pages      = flag.String("pages", "", "保留的页面 id，逗号分隔（默认全部；settings 必须保留）")
		skipBuild  = flag.Bool("skip-build", false, "跳过生成后的 go build 自检")
		noIntro    = flag.Bool("no-intro", false, "去掉开场动画（删 view 的 intro/opening 与相关测试）")
		noUpdate   = flag.Bool("no-update", false, "去掉自动更新（删 update.go，注入空实现）")
	)
	flag.Parse()
	if *dir == "" || *module == "" || *name == "" || *title == "" || *identifier == "" {
		flag.Usage()
		os.Exit(2)
	}
	var err error
	absRoot := ""
	if strings.TrimSpace(*root) != "" {
		absRoot, err = filepath.Abs(*root)
		if err != nil {
			fail(err)
		}
	}
	absAssets := ""
	if strings.TrimSpace(*assetsRoot) != "" {
		absAssets, err = filepath.Abs(*assetsRoot)
		if err != nil {
			fail(err)
		}
	}
	var keep []string
	if strings.TrimSpace(*pages) != "" {
		for _, p := range strings.Split(*pages, ",") {
			if p = strings.TrimSpace(p); p != "" {
				keep = append(keep, p)
			}
		}
	}
	res, err := scaffold.Generate(scaffold.Options{
		Root:       absRoot,
		AssetRoot:  absAssets,
		Dir:        *dir,
		Module:     *module,
		Name:       *name,
		Title:      *title,
		Brand:      *brand,
		Identifier: *identifier,
		Assets:     *assets,
		DemoAsset:  *demo,
		KeepPages:  keep,
		SkipBuild:  *skipBuild,
		NoIntro:    *noIntro,
		NoUpdate:   *noUpdate,
	})
	if err != nil {
		fail(err)
	}
	fmt.Printf("已生成 %s\n", res.Dir)
	fmt.Printf("文件 %d 个；页面 %s；素材 %s\n", res.Files, strings.Join(res.Pages, ", "), res.Assets)
	if res.BuildOutput != "" {
		fmt.Println("编译自检通过")
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "生成失败:", err)
	os.Exit(1)
}
