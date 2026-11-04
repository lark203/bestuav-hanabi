// Package store 是本地持久化层（SQLite，纯 Go 驱动 modernc.org/sqlite）。
//
// 只存「页面私有状态」（page_pref）：筛选条件、上次选中的视图等。
// 应用偏好仍在 settings.json —— 单机单用户的键值配置用 JSON 更直观，
// 页面状态天然是多页多键，用一张表省去自己管理命名空间。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // 注册 "sqlite" 驱动
)

// schemaVersion 当前库结构版本。
const schemaVersion = "1"

// SchemaVersion 返回当前结构版本。
func SchemaVersion() string { return schemaVersion }

// Store 数据库句柄；方法并发安全（database/sql 串行化）。
type Store struct {
	db   *sql.DB
	path string
}

// Open 打开（必要时创建）数据库并建表。
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("store: 未指定数据库路径")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("store: 创建数据目录失败: %w", err)
	}
	dsn := "file:" + path +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: 打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Path 数据库文件路径。
func (s *Store) Path() string { return s.path }

// Close 关闭数据库。
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS schema_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS page_pref (
			page TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (page, key)
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("store: 建表失败: %w", err)
		}
	}
	_, err := s.db.Exec(
		`INSERT INTO schema_meta(key, value) VALUES('schema', ?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value
		 WHERE CAST(excluded.value AS INTEGER) > CAST(schema_meta.value AS INTEGER)`,
		schemaVersion)
	if err != nil {
		return fmt.Errorf("store: 写 schema 版本失败: %w", err)
	}
	return nil
}

// PagePref 读页面私有状态；不存在返回 ok=false。
func (s *Store) PagePref(page, key string) (value string, ok bool, err error) {
	if s == nil || s.db == nil {
		return "", false, nil
	}
	err = s.db.QueryRow(`SELECT value FROM page_pref WHERE page=? AND key=?`, page, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// SetPagePref 写页面私有状态。
func (s *Store) SetPagePref(page, key, value string) error {
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.Exec(
		`INSERT INTO page_pref(page, key, value, updated_at) VALUES(?, ?, ?, unixepoch())
		 ON CONFLICT(page, key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		page, key, value)
	return err
}

// DeletePagePref 删除一条页面私有状态。
func (s *Store) DeletePagePref(page, key string) error {
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM page_pref WHERE page=? AND key=?`, page, key)
	return err
}

// PagePrefCount 页面状态条数（设置页展示/清理用）。
func (s *Store) PagePrefCount() (int64, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM page_pref`).Scan(&n)
	return n, err
}

// ClearPagePrefs 清空页面状态（不影响其它数据）。
func (s *Store) ClearPagePrefs() error {
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM page_pref`)
	return err
}

// BackupTo 用 VACUUM INTO 导出一致性快照（诊断包用）；目标必须不存在。
func (s *Store) BackupTo(path string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("store: 数据库未打开")
	}
	_, err := s.db.Exec(`VACUUM INTO ?`, path)
	return err
}

// AllPagePrefs 导出全部页面状态（配置包用）。
func (s *Store) AllPagePrefs() (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	if s == nil || s.db == nil {
		return out, nil
	}
	rows, err := s.db.Query(`SELECT page, key, value FROM page_pref ORDER BY page, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var page, key, value string
		if err := rows.Scan(&page, &key, &value); err != nil {
			return nil, err
		}
		if out[page] == nil {
			out[page] = map[string]string{}
		}
		out[page][key] = value
	}
	return out, rows.Err()
}

// ReplacePagePrefs 用给定内容替换全部页面状态（导入配置包用）。
func (s *Store) ReplacePagePrefs(prefs map[string]map[string]string) error {
	if s == nil || s.db == nil {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM page_pref`); err != nil {
		_ = tx.Rollback()
		return err
	}
	for page, kv := range prefs {
		for key, value := range kv {
			if _, err := tx.Exec(
				`INSERT INTO page_pref(page, key, value, updated_at) VALUES(?, ?, ?, unixepoch())`,
				page, key, value); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}
