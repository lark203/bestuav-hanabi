package view

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// 「新建项目」页只在模板仓库里有（生成项目会裁掉），测试也只在这里跑。
func TestNewProjectPageRenders(t *testing.T) {
	a := benchApp()
	a.SetPage(newprojectPageID)
	tt := ui.NewTester(a.View, 1000, 640)
	tt.Frame()
	if !tt.HasText("新建项目") {
		t.Error("页面缺少标题")
	}
	if !tt.HasText("生成项目") {
		t.Error("页面缺少生成按钮")
	}
}
