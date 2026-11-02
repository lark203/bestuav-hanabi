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
		// 呼吸与发丝幅度都要极小：超过 0.02 就不像呼吸，像整幅画面在晃。
		if c.PersonBreath < 0 || c.PersonBreath > 0.01 {
			t.Errorf("%s 的 PersonBreath 越界：%.4f（超过 0.01 已不是呼吸的量级）", c.ID, c.PersonBreath)
		}
		if c.HairSway < 0 || c.HairSway > 0.02 {
			t.Errorf("%s 的 HairSway 越界：%.4f（超过 0.02 会被看成整幅画面在晃）", c.ID, c.HairSway)
		}
		// 没有人物闸门却配了动态参数，等于配了套永远不会生效的东西——
		// 静默失效比报错更难查，而且作者会误以为"已经动起来了"。
		if c.PersonMask == "" && (c.PersonBreath != 0 || c.HairSway != 0 || c.HairVeil != "" || c.HairWave != "") {
			t.Errorf("%s 没有 PersonMask 却配了呼吸/发丝参数，不会生效", c.ID)
		}
		// 有人物闸门却没配任何动态，属于漏填。
		if c.PersonMask != "" && c.PersonBreath == 0 && c.HairSway == 0 {
			t.Errorf("%s 有 PersonMask 但呼吸与发丝幅度都是 0，壁纸实际是静态的", c.ID)
		}
		// 发丝幅度必须在有发丝层时才有意义。
		if c.HairSway != 0 && c.HairVeil == "" && c.HairWave == "" {
			t.Errorf("%s 设了 HairSway=%.4f 但没有任何发丝遮罩", c.ID, c.HairSway)
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

// TestPresentationData 钉住顶栏与角色专属造型需要的数据，漏填时界面会静默变空。
func TestPresentationData(t *testing.T) {
	for _, c := range All() {
		if c.Constellation == "" {
			t.Errorf("%s 没有命之座，顶栏中央会空着", c.ID)
		}
		if c.Motto == "" {
			t.Errorf("%s 没有侧栏副标题", c.ID)
		}
	}
	g, ok := ByID("ganyu")
	if !ok {
		t.Fatal("找不到甘雨")
	}
	if g.Chrome != ChromeCrystal {
		t.Errorf("甘雨的窗口按钮应走冰晶切角，实际 %d", g.Chrome)
	}
	if g.Ornament != OrnamentQingxin {
		t.Errorf("甘雨的纹样应为清心纹，实际 %d", g.Ornament)
	}
	if g.Motto != "循霜而行" {
		t.Errorf("甘雨的副标题应为「循霜而行」，实际 %q", g.Motto)
	}
	if g.Constellation != "S I N U S   U N I C O R N I S" {
		t.Errorf("甘雨的命之座不对：%q", g.Constellation)
	}
}

// TestThemeEffects 钉住每个角色的窗口按钮与纹样，避免加角色时忘配。
func TestThemeEffects(t *testing.T) {
	want := map[string]struct {
		chrome   Chrome
		ornament Ornament
		opening  Opening
	}{
		"ganyu":   {ChromeCrystal, OrnamentQingxin, OpeningGanyu},
		"yoimiya": {ChromePyro, OrnamentGoldfish, OpeningYoimiya},
		"yae":     {ChromeElectro, OrnamentSakura, OpeningYae},
		"nahida":  {ChromeDendro, OrnamentSprig, OpeningNahida},
		"furina":  {ChromeHydro, OrnamentTopHat, OpeningFurina},
		"faruzan": {ChromeAnemo, OrnamentPolyhedron, OpeningFaruzan},
		"zhongli": {ChromeGeo, OrnamentStoneSeal, OpeningZhongli},
	}
	for id, w := range want {
		c, ok := ByID(id)
		if !ok {
			t.Errorf("找不到角色 %s", id)
			continue
		}
		if c.Chrome != w.chrome {
			t.Errorf("%s 的 Chrome 应为 %d，实际 %d", id, w.chrome, c.Chrome)
		}
		if c.Ornament != w.ornament {
			t.Errorf("%s 的 Ornament 应为 %d，实际 %d", id, w.ornament, c.Ornament)
		}
		if c.Opening != w.opening {
			t.Errorf("%s 的 Opening 应为 %d，实际 %d", id, w.opening, c.Opening)
		}
	}
}
