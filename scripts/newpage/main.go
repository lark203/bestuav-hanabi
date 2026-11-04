// Command newpage 新增一个页面：生成页面文件 + 注册表条目。
//
// 用法（仓库根目录）：
//
//	go run ./scripts/newpage -id notes -label "笔记" -icon Square
//
// -icon 取 internal/ui/icons 里的名字（默认 Square）；-skip-build 只改文件不编译。
package main

import (
	"flag"
	"fmt"
	"os"

	"bestuav-hanabi/internal/pagegen"
)

func main() {
	var (
		root      = flag.String("root", ".", "仓库根目录")
		id        = flag.String("id", "", "页面 id（小写字母/数字/连字符）")
		label     = flag.String("label", "", "侧栏/页头标题")
		icon      = flag.String("icon", "Square", "icons 包里的图标名")
		skipBuild = flag.Bool("skip-build", false, "跳过生成后的 go build 自检")
	)
	flag.Parse()
	if *id == "" || *label == "" {
		flag.Usage()
		os.Exit(2)
	}
	file, err := pagegen.Add(pagegen.Options{
		Root:      *root,
		ID:        *id,
		Label:     *label,
		Icon:      *icon,
		SkipBuild: *skipBuild,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "生成失败:", err)
		os.Exit(1)
	}
	fmt.Println("已生成页面:", file)
}
