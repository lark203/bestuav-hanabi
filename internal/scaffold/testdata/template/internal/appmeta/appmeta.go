// Package appmeta 集中存放产品标识：名称、标题、应用标识与构建版本。
//
// 改名时改这里，并在 mygo.json 里同步 name / slug / identifier。
package appmeta

import "runtime/debug"

// Name 产品名与配置目录名（mygo.PathUserData 下的子目录），也用作窗口状态键。
const Name = "hanabi"

// Title 窗口标题与界面品牌名。
const Title = "花火 HanaBi"

// Brand 侧栏/状态栏里的短品牌名（英文或简称）。
const Brand = "HanaBi"

// Identifier 应用唯一标识（打包标识、单实例锁）。
const Identifier = "com.bestuav.hanabi"

// BuildVersion 返回构建版本：模块版本 + 短提交号；本地 go run/build 通常是 dev。
func BuildVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	v := bi.Main.Version
	if v == "" || v == "(devel)" {
		v = "dev"
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 7 {
			return v + " · " + s.Value[:7]
		}
	}
	return v
}
