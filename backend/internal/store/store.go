package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct{ DB *sql.DB }

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	st := &Store{DB: db}
	return st, st.Migrate()
}

func (s *Store) Migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, created_at TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, user_id INTEGER NOT NULL, expires_at TEXT, remember INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS login_logs (id INTEGER PRIMARY KEY, username TEXT NOT NULL, success INTEGER NOT NULL, ip TEXT, message TEXT, created_at TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS nav_groups (id INTEGER PRIMARY KEY, name TEXT NOT NULL, sort INTEGER NOT NULL DEFAULT 0);`,
		`CREATE TABLE IF NOT EXISTS nav_items (id INTEGER PRIMARY KEY, group_id INTEGER NOT NULL, name TEXT NOT NULL, icon TEXT, lan_url TEXT, wan_url TEXT, url_mode TEXT NOT NULL DEFAULT 'wan', sort INTEGER NOT NULL DEFAULT 0, FOREIGN KEY(group_id) REFERENCES nav_groups(id) ON DELETE CASCADE);`,
		`CREATE TABLE IF NOT EXISTS bookmark_folders (id INTEGER PRIMARY KEY, parent_id INTEGER, name TEXT NOT NULL, sort INTEGER NOT NULL DEFAULT 0, FOREIGN KEY(parent_id) REFERENCES bookmark_folders(id) ON DELETE CASCADE);`,
		`CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY, folder_id INTEGER NOT NULL, title TEXT NOT NULL, url TEXT NOT NULL, favicon TEXT, note TEXT, sort INTEGER NOT NULL DEFAULT 0, FOREIGN KEY(folder_id) REFERENCES bookmark_folders(id) ON DELETE CASCADE);`,
		`CREATE TABLE IF NOT EXISTS assets (id INTEGER PRIMARY KEY, name TEXT, source TEXT NOT NULL, path TEXT NOT NULL, mime TEXT, size INTEGER, created_at TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS storage_configs (id INTEGER PRIMARY KEY, kind TEXT NOT NULL, name TEXT NOT NULL, config_json TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 0);`,
		`CREATE TABLE IF NOT EXISTS backup_records (id INTEGER PRIMARY KEY, file_name TEXT NOT NULL, target TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_nav_items_group_sort ON nav_items(group_id, sort, id);`,
		`CREATE INDEX IF NOT EXISTS idx_bookmark_folders_parent_sort ON bookmark_folders(parent_id, sort, id);`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_folder_sort ON bookmarks(folder_id, sort, id);`,
	}
	for _, stmt := range stmts {
		if _, err := s.DB.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := s.DB.Exec(`UPDATE nav_items SET url_mode='wan' WHERE url_mode='' OR url_mode='auto'`); err != nil {
		return err
	}
	return nil
}
