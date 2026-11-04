package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
)

func TestResetBacksUp(t *testing.T) {
	dir := t.TempDir()
	mygo.App.SetPath(mygo.PathUserData, dir)
	t.Cleanup(func() { mygo.App.SetPath(mygo.PathUserData, "") })

	s := Default()
	s.Theme = "nahida"
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	backup, err := Reset()
	if err != nil {
		t.Fatal(err)
	}
	if backup == "" {
		t.Fatal("应返回备份路径")
	}
	if _, err := os.Stat(backup); err != nil {
		t.Errorf("备份不存在: %v", err)
	}
	if got := Load(); got.Theme != Default().Theme {
		t.Errorf("重置后主题 = %q", got.Theme)
	}
	if filepath.Dir(backup) != dir {
		t.Errorf("备份位置不对: %s", backup)
	}
}
