package store

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct{ DB *sql.DB }

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}
type NavGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}
type NavItem struct {
	ID      int64  `json:"id"`
	GroupID int64  `json:"groupId"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	LANURL  string `json:"lanUrl"`
	WANURL  string `json:"wanUrl"`
	URLMode string `json:"urlMode"`
	Sort    int    `json:"sort"`
}
type Folder struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parentId"`
	Name        string `json:"name"`
	Sort        int    `json:"sort"`
	HasChildren bool   `json:"hasChildren"`
}
type Bookmark struct {
	ID       int64  `json:"id"`
	FolderID int64  `json:"folderId"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Favicon  string `json:"favicon"`
	Note     string `json:"note"`
	Sort     int    `json:"sort"`
	Path     string `json:"path,omitempty"`
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
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

func (s *Store) HasUser() (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n > 0, err
}
func (s *Store) CreateUser(username, hash string) error {
	_, err := s.DB.Exec(`INSERT INTO users(username,password_hash,created_at) VALUES(?,?,?)`, username, hash, time.Now().Format(time.RFC3339))
	return err
}
func (s *Store) FindUser(username string) (User, error) {
	var u User
	err := s.DB.QueryRow(`SELECT id,username,password_hash FROM users WHERE username=?`, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	return u, err
}
func (s *Store) LogLogin(username string, success bool, ip, msg string) {
	val := 0
	if success {
		val = 1
	}
	_, _ = s.DB.Exec(`INSERT INTO login_logs(username,success,ip,message,created_at) VALUES(?,?,?,?,?)`, username, val, ip, msg, time.Now().Format(time.RFC3339))
}
func (s *Store) FailedLoginsSince(username string, since time.Time) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM login_logs WHERE username=? AND success=0 AND created_at>?`, username, since.Format(time.RFC3339)).Scan(&n)
	return n, err
}
func (s *Store) CleanupExpiredData(logsRetentionDays int) error {
	nowStr := time.Now().Format(time.RFC3339)
	oldLogsStr := time.Now().AddDate(0, 0, -logsRetentionDays).Format(time.RFC3339)

	// Clean expired sessions
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE expires_at != '' AND expires_at < ?`, nowStr)
	if err != nil {
		log.Printf("[Store] Failed to cleanup expired sessions: %v\n", err)
	}

	// Clean old login logs
	_, err = s.DB.Exec(`DELETE FROM login_logs WHERE created_at < ?`, oldLogsStr)
	if err != nil {
		log.Printf("[Store] Failed to cleanup old login_logs: %v\n", err)
	}

	return nil
}

func (s *Store) SaveSession(token string, userID int64, expires *time.Time, remember bool) error {
	rem := 0
	var exp any
	if remember {
		rem = 1
	}
	if expires != nil {
		exp = expires.Format(time.RFC3339)
	}
	_, err := s.DB.Exec(`INSERT INTO sessions(token,user_id,expires_at,remember,created_at) VALUES(?,?,?,?,?)`, token, userID, exp, rem, time.Now().Format(time.RFC3339))
	return err
}
func (s *Store) DeleteSession(token string) {
	_, _ = s.DB.Exec(`DELETE FROM sessions WHERE token=?`, token)
}
func (s *Store) UserBySession(token string) (User, error) {
	var u User
	var exp sql.NullString
	err := s.DB.QueryRow(`SELECT u.id,u.username,u.password_hash,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token=?`, token).Scan(&u.ID, &u.Username, &u.PasswordHash, &exp)
	if err != nil {
		return u, err
	}
	if exp.Valid {
		t, _ := time.Parse(time.RFC3339, exp.String)
		if time.Now().After(t) {
			s.DeleteSession(token)
			return u, errors.New("session expired")
		}
	}
	return u, nil
}

func (s *Store) ListNavigation() ([]NavGroup, map[int64][]NavItem, error) {
	rows, err := s.DB.Query(`SELECT id,name,sort FROM nav_groups ORDER BY sort,id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	groups := []NavGroup{}
	for rows.Next() {
		var g NavGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Sort); err != nil {
			return nil, nil, err
		}
		groups = append(groups, g)
	}
	itemRows, err := s.DB.Query(`SELECT id,group_id,name,COALESCE(icon,''),COALESCE(lan_url,''),COALESCE(wan_url,''),url_mode,sort FROM nav_items ORDER BY sort,id`)
	if err != nil {
		return nil, nil, err
	}
	defer itemRows.Close()
	items := map[int64][]NavItem{}
	for itemRows.Next() {
		var it NavItem
		if err := itemRows.Scan(&it.ID, &it.GroupID, &it.Name, &it.Icon, &it.LANURL, &it.WANURL, &it.URLMode, &it.Sort); err != nil {
			return nil, nil, err
		}
		items[it.GroupID] = append(items[it.GroupID], it)
	}
	return groups, items, rows.Err()
}

func (s *Store) NavGroupNameExists(name string, excludeID int64) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM nav_groups WHERE name=? AND id<>?`, name, excludeID).Scan(&n)
	return n > 0, err
}

func (s *Store) NavGroupExists(id int64) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM nav_groups WHERE id=?`, id).Scan(&n)
	return n > 0, err
}

func (s *Store) NavItemExists(id int64) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM nav_items WHERE id=?`, id).Scan(&n)
	return n > 0, err
}

func (s *Store) CreateNavGroup(g NavGroup) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO nav_groups(name,sort) VALUES(?,?)`, g.Name, g.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
func (s *Store) CreateNavItem(it NavItem) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO nav_items(group_id,name,icon,lan_url,wan_url,url_mode,sort) VALUES(?,?,?,?,?,?,?)`, it.GroupID, it.Name, it.Icon, it.LANURL, it.WANURL, it.URLMode, it.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListNavItemsByGroup(groupID int64) ([]NavItem, error) {
	rows, err := s.DB.Query(`SELECT id,group_id,name,icon,lan_url,wan_url,url_mode,sort FROM nav_items WHERE group_id=? ORDER BY sort ASC, id ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NavItem
	for rows.Next() {
		var it NavItem
		if err := rows.Scan(&it.ID, &it.GroupID, &it.Name, &it.Icon, &it.LANURL, &it.WANURL, &it.URLMode, &it.Sort); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func (s *Store) ReplaceNavigation(groups []NavGroup, items map[int64][]NavItem) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM nav_items`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM nav_groups`); err != nil {
		return err
	}
	for index, group := range groups {
		sort := group.Sort
		if sort == 0 {
			sort = index + 1
		}
		res, err := tx.Exec(`INSERT INTO nav_groups(name,sort) VALUES(?,?)`, group.Name, sort)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for itemIndex, item := range items[group.ID] {
			itemSort := item.Sort
			if itemSort == 0 {
				itemSort = itemIndex + 1
			}
			urlMode := item.URLMode
			if urlMode == "" || urlMode == "auto" {
				urlMode = "wan"
			}
			if _, err := tx.Exec(`INSERT INTO nav_items(group_id,name,icon,lan_url,wan_url,url_mode,sort) VALUES(?,?,?,?,?,?,?)`, id, item.Name, item.Icon, item.LANURL, item.WANURL, urlMode, itemSort); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *Store) ListFolders(parent *int64) ([]Folder, error) {
	var rows *sql.Rows
	var err error
	if parent == nil {
		rows, err = s.DB.Query(`SELECT f.id,f.parent_id,f.name,f.sort,EXISTS(SELECT 1 FROM bookmark_folders c WHERE c.parent_id=f.id) FROM bookmark_folders f WHERE f.parent_id IS NULL ORDER BY f.sort,f.id`)
	} else {
		rows, err = s.DB.Query(`SELECT f.id,f.parent_id,f.name,f.sort,EXISTS(SELECT 1 FROM bookmark_folders c WHERE c.parent_id=f.id) FROM bookmark_folders f WHERE f.parent_id=? ORDER BY f.sort,f.id`, *parent)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Folder{}
	for rows.Next() {
		var f Folder
		var pid sql.NullInt64
		var has int
		if err := rows.Scan(&f.ID, &pid, &f.Name, &f.Sort, &has); err != nil {
			return nil, err
		}
		if pid.Valid {
			v := pid.Int64
			f.ParentID = &v
		}
		f.HasChildren = has == 1
		out = append(out, f)
	}
	return out, rows.Err()
}
func (s *Store) CreateFolder(f Folder) (int64, error) {
	var p any
	if f.ParentID != nil {
		p = *f.ParentID
	}
	res, err := s.DB.Exec(`INSERT INTO bookmark_folders(parent_id,name,sort) VALUES(?,?,?)`, p, f.Name, f.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
func (s *Store) ListBookmarks(folderID int64) ([]Bookmark, error) {
	rows, err := s.DB.Query(`SELECT id,folder_id,title,url,COALESCE(favicon,''),COALESCE(note,''),sort FROM bookmarks WHERE folder_id=? ORDER BY sort,id`, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Bookmark{}
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.FolderID, &b.Title, &b.URL, &b.Favicon, &b.Note, &b.Sort); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (s *Store) CreateBookmark(b Bookmark) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO bookmarks(folder_id,title,url,favicon,note,sort) VALUES(?,?,?,?,?,?)`, b.FolderID, b.Title, b.URL, b.Favicon, b.Note, b.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListAllFolders() ([]Folder, error) {
	rows, err := s.DB.Query(`SELECT f.id,f.parent_id,f.name,f.sort,EXISTS(SELECT 1 FROM bookmark_folders c WHERE c.parent_id=f.id) FROM bookmark_folders f ORDER BY COALESCE(f.parent_id,0),f.sort,f.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Folder{}
	for rows.Next() {
		var f Folder
		var pid sql.NullInt64
		var has int
		if err := rows.Scan(&f.ID, &pid, &f.Name, &f.Sort, &has); err != nil {
			return nil, err
		}
		if pid.Valid {
			v := pid.Int64
			f.ParentID = &v
		}
		f.HasChildren = has == 1
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) ListAllBookmarks() ([]Bookmark, error) {
	rows, err := s.DB.Query(`SELECT id,folder_id,title,url,COALESCE(favicon,''),COALESCE(note,''),sort FROM bookmarks ORDER BY folder_id,sort,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Bookmark{}
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.FolderID, &b.Title, &b.URL, &b.Favicon, &b.Note, &b.Sort); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) SearchBookmarks(q string) ([]Bookmark, error) {
	like := "%" + q + "%"
	rows, err := s.DB.Query(`WITH RECURSIVE folder_paths(id, path) AS (
		SELECT id, name FROM bookmark_folders WHERE parent_id IS NULL
		UNION ALL
		SELECT f.id, folder_paths.path || ' / ' || f.name FROM bookmark_folders f JOIN folder_paths ON f.parent_id = folder_paths.id
	)
	SELECT b.id,b.folder_id,b.title,b.url,COALESCE(b.favicon,''),COALESCE(b.note,''),b.sort,COALESCE(folder_paths.path, f.name)
	FROM bookmarks b
	JOIN bookmark_folders f ON f.id=b.folder_id
	LEFT JOIN folder_paths ON folder_paths.id=f.id
	WHERE b.title LIKE ? OR b.url LIKE ? OR b.note LIKE ?
	ORDER BY b.title LIMIT 100`, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Bookmark{}
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.FolderID, &b.Title, &b.URL, &b.Favicon, &b.Note, &b.Sort, &b.Path); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) UpdateNavGroup(g NavGroup) error {
	_, err := s.DB.Exec(`UPDATE nav_groups SET name=?, sort=? WHERE id=?`, g.Name, g.Sort, g.ID)
	return err
}
func (s *Store) DeleteNavGroup(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM nav_groups WHERE id=?`, id)
	return err
}
func (s *Store) UpdateNavItem(it NavItem) error {
	_, err := s.DB.Exec(`UPDATE nav_items SET group_id=?, name=?, icon=?, lan_url=?, wan_url=?, url_mode=?, sort=? WHERE id=?`, it.GroupID, it.Name, it.Icon, it.LANURL, it.WANURL, it.URLMode, it.Sort, it.ID)
	return err
}
func (s *Store) DeleteNavItem(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM nav_items WHERE id=?`, id)
	return err
}
func (s *Store) UpdateFolder(f Folder) error {
	var p any
	if f.ParentID != nil {
		p = *f.ParentID
	}
	_, err := s.DB.Exec(`UPDATE bookmark_folders SET parent_id=?, name=?, sort=? WHERE id=?`, p, f.Name, f.Sort, f.ID)
	return err
}
func (s *Store) DeleteFolder(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM bookmark_folders WHERE id=?`, id)
	return err
}
func (s *Store) UpdateBookmark(b Bookmark) error {
	_, err := s.DB.Exec(`UPDATE bookmarks SET folder_id=?, title=?, url=?, favicon=?, note=?, sort=? WHERE id=?`, b.FolderID, b.Title, b.URL, b.Favicon, b.Note, b.Sort, b.ID)
	return err
}
func (s *Store) DeleteBookmark(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM bookmarks WHERE id=?`, id)
	return err
}

func (s *Store) ListSettings() (map[string]string, error) {
	rows, err := s.DB.Query(`SELECT key,value FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (s *Store) SaveSettings(values map[string]string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for k, v := range values {
		if _, err := tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) BookmarkExistsInFolder(folderID int64, url string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`SELECT count(*) FROM bookmarks WHERE folder_id = ? AND url = ?`, folderID, url).Scan(&count)
	return count > 0, err
}
