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
	if !ok || c.NameCN != "演示" {
		t.Errorf("ByID(demo) = %+v, %v", c, ok)
	}
	if _, ok := ByID("ganyu"); !ok {
		t.Error("内建主题应仍可查")
	}
	if !BuiltinIDs()["ganyu"] || BuiltinIDs()["demo"] {
		t.Error("BuiltinIDs 不应包含外部主题")
	}
}
