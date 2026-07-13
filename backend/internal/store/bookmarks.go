package store

import "database/sql"

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

func (s *Store) GetBookmark(id int64) (Bookmark, error) {
	var b Bookmark
	err := s.DB.QueryRow(`SELECT id,folder_id,title,url,COALESCE(favicon,''),COALESCE(note,''),sort FROM bookmarks WHERE id=?`, id).Scan(&b.ID, &b.FolderID, &b.Title, &b.URL, &b.Favicon, &b.Note, &b.Sort)
	return b, err
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

func (s *Store) UpdateBookmarkFavicon(id int64, favicon string) error {
	_, err := s.DB.Exec(`UPDATE bookmarks SET favicon=? WHERE id=? AND (favicon IS NULL OR TRIM(favicon)='' OR (favicon NOT LIKE '/uploads/%' AND favicon NOT LIKE 'http://%' AND favicon NOT LIKE 'https://%' AND favicon NOT LIKE 'data:image/%'))`, favicon, id)
	return err
}

func (s *Store) DeleteBookmark(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM bookmarks WHERE id=?`, id)
	return err
}
