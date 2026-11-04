package visual

import "testing"

func TestCharacters(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("没有角色主题")
	}
	if Default().ID != all[0].ID {
		t.Error("Default 应为第一个主题")
	}
	if _, ok := ByID(all[0].ID); !ok {
		t.Error("ByID 找不到第一个主题")
	}
	if _, ok := ByID("nope"); ok {
		t.Error("未知 id 不应命中")
	}
}

func TestEngine(t *testing.T) {
	if DefaultEngine == nil {
		t.Fatal("DefaultEngine 为空")
	}
	c := DefaultEngine.Default()
	if got := DefaultEngine.All(); len(got) == 0 {
		t.Error("Engine.All 为空")
	}
	if _, ok := DefaultEngine.ByID(c.ID); !ok {
		t.Error("Engine.ByID 失败")
	}
	// 素材缺失时应返回 nil，而不是 panic（生成项目 assets none 场景）。
	_ = DefaultEngine.Background(c)
	_ = DefaultEngine.Icon(c)
	_ = DefaultEngine.Wallpaper(c)
	if keys := DefaultEngine.WallpaperKeys(c); len(keys) == 0 {
		t.Error("WallpaperKeys 至少应包含壁纸/头像路径")
	}
	in := DefaultEngine.NewIntro(100, 80)
	if in == nil || in.W != 100 || in.H != 80 {
		t.Errorf("NewIntro 尺寸不对: %+v", in)
	}
}

func TestIntroTimeline(t *testing.T) {
	if TurnEnd <= TurnStart {
		t.Errorf("时间轴顺序不对: %v..%v", TurnStart, TurnEnd)
	}
	if SnowCount <= 0 || LantCount <= 0 {
		t.Error("粒子数量应为正")
	}
	if got := Smooth(0, 0, 1); got != 0 {
		t.Errorf("Smooth(0) = %v", got)
	}
	if got := Smooth(1, 0, 1); got != 1 {
		t.Errorf("Smooth(1) = %v", got)
	}
}
