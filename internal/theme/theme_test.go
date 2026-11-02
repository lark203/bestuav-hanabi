package theme

import "testing"

// TestCharacters 钉住角色包的基本约束：id 唯一、可见信息齐全、强调色互不相同。
//
// 这些约束看着琐碎，但加角色时最容易犯的就是复制粘贴改了名字忘了改色——
// 那样主题库里会出现两张一模一样、点哪个都没反应的卡片。
func TestCharacters(t *testing.T) {
	seen := map[string]bool{}
	accents := map[uint32]string{}
	motifs := map[Motif]bool{}

	for _, c := range All() {
		if c.ID == "" || c.NameCN == "" || c.NameEN == "" {
			t.Errorf("角色信息不完整：%+v", c)
		}
		if seen[c.ID] {
			t.Errorf("角色 id 重复：%s", c.ID)
		}
		seen[c.ID] = true

		if c.Tok.Accent == (Tokens{}).Accent {
			t.Errorf("%s 的强调色为空", c.ID)
		}
		key := uint32(c.Tok.Accent.R)<<16 | uint32(c.Tok.Accent.G)<<8 | uint32(c.Tok.Accent.B)
		if other, dup := accents[key]; dup {
			t.Errorf("%s 与 %s 的强调色完全相同（#%06X）", c.ID, other, key)
		}
		accents[key] = c.ID

		if c.WallTint < 0 || c.WallTint > 1 {
			t.Errorf("%s 的 WallTint 越界：%.2f", c.ID, c.WallTint)
		}
		if c.OverlayStrength < 0 || c.OverlayStrength > 1 {
			t.Errorf("%s 的 OverlayStrength 越界：%.2f", c.ID, c.OverlayStrength)
		}
		// Overlay 与遮罩必须成对：有颜色层就该有遮罩，避免叠层静默不显。
		if c.Overlay != "" && c.OverlayMask == "" {
			t.Errorf("%s 有 Overlay 但缺 OverlayMask", c.ID)
		}
		motifs[c.Motif] = true
	}

	// 至少要有两种画法，否则「按 Motif 分发」这件事没有被验证过。
	if len(motifs) < 2 {
		t.Errorf("只有 %d 种标识画法，至少要有 2 种", len(motifs))
	}
}

// TestByID 检查查找与默认值。
func TestByID(t *testing.T) {
	d := Default()
	if _, ok := ByID(d.ID); !ok {
		t.Errorf("默认角色 %s 查不到", d.ID)
	}
	if _, ok := ByID("no-such-id"); ok {
		t.Error("不存在的 id 不应查得到")
	}
}
