package store

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
