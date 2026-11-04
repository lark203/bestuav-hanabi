package assets

import "testing"

func TestUserDir(t *testing.T) {
	base := UserDir()
	dir := t.TempDir()
	SetUserDir(dir)
	defer SetUserDir("")
	if got := UserDir(); got != dir {
		t.Errorf("UserDir = %q，期望 %q", got, dir)
	}
	SetUserDir("")
	if got := UserDir(); got != Dir() {
		t.Errorf("清空后应回退到 Dir(): %q vs %q", got, Dir())
	}
	_ = base
}
