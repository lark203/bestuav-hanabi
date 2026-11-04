// Package cli 解析启动参数。
//
// 桌面程序也需要一点命令行能力：--version 给打包/排障用，--page 给单实例转发用，
// --reset-settings 在配置写坏时救命，--safe-mode 跳过数据库与自动更新。
package cli

import (
	"flag"
	"fmt"
	"io"
)

// Options 解析后的启动选项。
type Options struct {
	Version       bool   // 打印版本后退出
	Page          string // 启动后打开哪个页面（页面 id；空 = 默认首页）
	ResetSettings bool   // 启动前重置偏好
	SafeMode      bool   // 安全模式：不打开数据库、不检查更新
	Args          []string
}

// Parse 解析参数；args 不含程序名。
func Parse(args []string, out io.Writer) (Options, error) {
	fs := flag.NewFlagSet("hanabi", flag.ContinueOnError)
	fs.SetOutput(out)
	var (
		version = fs.Bool("version", false, "打印版本并退出")
		page    = fs.String("page", "", "启动后打开的页面 id（library / wallpaper / settings）")
		reset   = fs.Bool("reset-settings", false, "启动前把用户偏好恢复默认")
		safe    = fs.Bool("safe-mode", false, "安全模式：不打开数据库、不检查更新")
	)
	fs.Usage = func() {
		fmt.Fprintln(out, "用法: hanabi [选项]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	return Options{
		Version:       *version,
		Page:          *page,
		ResetSettings: *reset,
		SafeMode:      *safe,
		Args:          args,
	}, nil
}

// PageFromArgs 从任意参数里取 --page（单实例转发用），解析失败返回空。
func PageFromArgs(args []string) string {
	for i, a := range args {
		if v, ok := cutPrefix(a, "--page="); ok {
			return v
		}
		if v, ok := cutPrefix(a, "-page="); ok {
			return v
		}
		if (a == "--page" || a == "-page") && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// AttachConsole 把 Windows GUI 子系统的标准输出接回启动它的终端；其他平台空操作。
func AttachConsole() { attachConsole() }

func cutPrefix(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):], true
	}
	return "", false
}
