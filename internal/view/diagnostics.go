package view

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/egoist/mygo"

	"bestuav-hanabi/internal/appmeta"
	"bestuav-hanabi/internal/apppath"
	"bestuav-hanabi/internal/settings"
)

// 诊断包：设置、数据库快照与环境信息，便于排查「别人的机器上复现不了」的问题。
// 不含角色素材（体积大且与问题无关）。
func (a *App) exportDiagnostics() {
	st := a.st
	go func() {
		dir, _ := settings.DefaultPath()
		out, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Title:       "导出诊断包",
			DefaultPath: filepath.Join(filepath.Dir(dir), "hanabi-diagnostics.zip"),
			ButtonLabel: "导出",
			Filters:     []mygo.FileFilter{{Name: "zip", Extensions: []string{"zip"}}},
		})
		if err != nil || out == "" {
			return
		}
		if filepath.Ext(out) == "" {
			out += ".zip"
		}
		err = writeDiagnostics(out, st)
		mygo.RunOnMain(func() {
			if err != nil {
				a.notice = "诊断包导出失败：" + err.Error()
			} else {
				a.notice = "诊断包已导出：" + filepath.Base(out)
			}
			a.updateNotice()
		})
	}()
}

func writeDiagnostics(out string, st interface {
	BackupTo(string) error
}) error {
	tmp, err := os.MkdirTemp("", "hanabi-diag-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// 设置。
	if p, err := settings.DefaultPath(); err == nil {
		if data, rerr := os.ReadFile(p); rerr == nil {
			_ = os.WriteFile(filepath.Join(tmp, "settings.json"), data, 0o644)
		}
	}
	// 数据库快照。
	if st != nil {
		dbTmp := filepath.Join(tmp, "data.db")
		if err := st.BackupTo(dbTmp); err != nil {
			_ = os.WriteFile(filepath.Join(tmp, "db-error.txt"), []byte(err.Error()), 0o644)
		}
	}
	// 环境。
	env := fmt.Sprintf("app: %s\nversion: %s\nplatform: %s/%s\ntime: %s\n",
		appmeta.Identifier, appmeta.BuildVersion(), runtime.GOOS, runtime.GOARCH,
		time.Now().Format(time.RFC3339))
	if d, err := apppath.Dir(); err == nil {
		env += "config: " + d + "\n"
	}
	_ = os.WriteFile(filepath.Join(tmp, "env.txt"), []byte(env), 0o644)

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	entries, _ := os.ReadDir(tmp)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(tmp, e.Name())
		w, werr := zw.Create(e.Name())
		if werr != nil {
			zw.Close()
			f.Close()
			return werr
		}
		data, rerr := os.ReadFile(src)
		if rerr != nil {
			continue
		}
		if _, werr := w.Write(data); werr != nil {
			zw.Close()
			f.Close()
			return werr
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
