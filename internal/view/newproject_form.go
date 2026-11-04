package view

// newProjectForm 是「新建项目」页的表单（等价于 scripts/newproject 的参数）。
//
// 类型放在这里而不是 newproject_page.go：生成项目会裁掉那页（没有 scaffold），
// 但 App 里的字段仍要能编译。
type newProjectForm struct {
	dir, module, name, title, brand, id string
	assets, demo, pages                 string
	noIntro, noUpdate, skipBuild        bool
}

func defaultNewProjectForm() newProjectForm {
	return newProjectForm{
		module: "example.com/acme/app",
		name:   "app",
		title:  "新项目",
		brand:  "App",
		id:     "com.acme.app",
		assets: "demo",
		demo:   "ganyu",
	}
}
