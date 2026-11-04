package visual

import "testing"

func TestRegisterExternal(t *testing.T) {
	extMu.Lock()
	old := external
	external = nil
	extMu.Unlock()
	defer func() {
		extMu.Lock()
		external = old
		extMu.Unlock()
	}()

	base := len(All())
	RegisterExternal([]Character{
		{ID: "demo", NameCN: "演示"},
		{ID: "demo", NameCN: "重复"},
		{ID: "ganyu", NameCN: "冲突"},
	})
	if got := len(All()); got != base+1 {
		t.Fatalf("All = %d，期望 %d", got, base+1)
	}
	c, ok := ByID("demo")
	if !ok || c.NameCN != "重复" { // 同批里的后一个覆盖前一个
		t.Errorf("ByID(demo) = %+v, %v", c, ok)
	}
	// 重新导入同 id：覆盖而不是重复。
	RegisterExternal([]Character{{ID: "demo", NameCN: "演示2"}})
	if got := len(All()); got != base+1 {
		t.Fatalf("重复 id 不应新增: %d", got)
	}
	if c, _ := ByID("demo"); c.NameCN != "演示2" {
		t.Errorf("同 id 应覆盖: %+v", c)
	}
	RemoveExternal("demo")
	if _, ok := ByID("demo"); ok {
		t.Error("RemoveExternal 后不应再查到")
	}
	if _, ok := ByID("ganyu"); !ok {
		t.Error("内建主题应仍可查")
	}
	if !BuiltinIDs()["ganyu"] || BuiltinIDs()["demo"] {
		t.Error("BuiltinIDs 不应包含外部主题")
	}
}
