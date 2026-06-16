package httpx

import (
	"net/http"
	"strings"

	"biu-panel/backend/internal/store"
)

func (s *Server) navigation(w http.ResponseWriter, _ *http.Request) {
	groups, items, err := s.store.ListNavigation()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	out := []map[string]any{}
	for _, g := range groups {
		groupItems := items[g.ID]
		if groupItems == nil {
			groupItems = []store.NavItem{}
		}
		out = append(out, map[string]any{"id": g.ID, "name": g.Name, "sort": g.Sort, "collapsed": g.Collapsed, "items": groupItems})
	}
	writeJSON(w, 200, map[string]any{"groups": out})
}
func (s *Server) createNavGroup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var g store.NavGroup
	if !decodeJSON(w, r, &g) {
		return
	}
	g.Name = strings.TrimSpace(g.Name)
	if g.Name == "" {
		writeError(w, 400, "分组名称不能为空")
		return
	}
	if len([]rune(g.Name)) > maxNavGroupNameLength {
		writeError(w, 400, "分组名称不能超过 10 个字")
		return
	}
	if g.Sort < 0 {
		g.Sort = 0
	}
	if exists, err := s.store.NavGroupNameExists(g.Name, 0); err != nil {
		writeError(w, 500, err.Error())
		return
	} else if exists {
		writeError(w, 409, "分组名称已存在")
		return
	}
	id, err := s.store.CreateNavGroup(g)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	g.ID = id
	writeJSON(w, 201, g)
}
func (s *Server) updateNavGroup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var g store.NavGroup
	if !decodeJSON(w, r, &g) {
		return
	}
	g.Name = strings.TrimSpace(g.Name)
	if g.ID == 0 {
		writeError(w, 400, "分组 ID 必填")
		return
	}
	if g.Name == "" {
		writeError(w, 400, "分组名称不能为空")
		return
	}
	if len([]rune(g.Name)) > maxNavGroupNameLength {
		writeError(w, 400, "分组名称不能超过 10 个字")
		return
	}
	if exists, err := s.store.NavGroupExists(g.ID); err != nil {
		writeError(w, 500, err.Error())
		return
	} else if !exists {
		writeError(w, 404, "分组不存在")
		return
	}
	if g.Sort < 0 {
		g.Sort = 0
	}
	if exists, err := s.store.NavGroupNameExists(g.Name, g.ID); err != nil {
		writeError(w, 500, err.Error())
		return
	} else if exists {
		writeError(w, 409, "分组名称已存在")
		return
	}
	if err := s.store.UpdateNavGroup(g); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, g)
}
func (s *Server) deleteNavGroup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	id, err := idFromQuery(r)
	if err != nil {
		writeError(w, 400, "id 必填")
		return
	}
	if exists, err := s.store.NavGroupExists(id); err != nil {
		writeError(w, 500, err.Error())
		return
	} else if !exists {
		writeError(w, 404, "分组不存在")
		return
	}
	items, err := s.store.ListNavItemsByGroup(id)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if len(items) > 0 {
		writeError(w, 400, "分组内存在卡片，无法删除")
		return
	}
	if err := s.store.DeleteNavGroup(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) createNavItem(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var it store.NavItem
	if !decodeJSON(w, r, &it) {
		return
	}
	if !s.validateNavItem(w, &it, false) {
		return
	}
	id, err := s.store.CreateNavItem(it)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	it.ID = id
	writeJSON(w, 201, it)
}
func (s *Server) updateNavItem(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var it store.NavItem
	if !decodeJSON(w, r, &it) {
		return
	}
	if !s.validateNavItem(w, &it, true) {
		return
	}
	if err := s.store.UpdateNavItem(it); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, it)
}
func (s *Server) deleteNavItem(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	id, err := idFromQuery(r)
	if err != nil {
		writeError(w, 400, "id 必填")
		return
	}
	if exists, err := s.store.NavItemExists(id); err != nil {
		writeError(w, 500, err.Error())
		return
	} else if !exists {
		writeError(w, 404, "卡片不存在")
		return
	}
	if err := s.store.DeleteNavItem(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) validateNavItem(w http.ResponseWriter, it *store.NavItem, requireID bool) bool {
	it.Name = strings.TrimSpace(it.Name)
	it.LANURL = strings.TrimSpace(it.LANURL)
	it.WANURL = strings.TrimSpace(it.WANURL)
	it.URLMode = strings.TrimSpace(it.URLMode)
	if requireID && it.ID == 0 {
		writeError(w, 400, "卡片 ID 必填")
		return false
	}
	if it.Name == "" {
		writeError(w, 400, "卡片标题不能为空")
		return false
	}
	if len([]rune(it.Name)) > maxNavItemTitleLength {
		writeError(w, 400, "卡片标题不能超过 15 个字")
		return false
	}
	if it.GroupID == 0 {
		writeError(w, 400, "卡片分组必填")
		return false
	}
	if it.WANURL == "" {
		writeError(w, 400, "公网地址不能为空")
		return false
	}
	if len(it.LANURL) > maxNavURLLength || len(it.WANURL) > maxNavURLLength {
		writeError(w, 400, "卡片地址不能超过 2048 个字符")
		return false
	}
	if it.URLMode == "" {
		it.URLMode = "wan"
	}
	if it.URLMode != "lan" && it.URLMode != "wan" {
		writeError(w, 400, "打开方式不支持")
		return false
	}
	if it.Sort < 0 {
		it.Sort = 0
	}
	if exists, err := s.store.NavGroupExists(it.GroupID); err != nil {
		writeError(w, 500, err.Error())
		return false
	} else if !exists {
		writeError(w, 400, "卡片分组不存在")
		return false
	}
	if requireID {
		if exists, err := s.store.NavItemExists(it.ID); err != nil {
			writeError(w, 500, err.Error())
			return false
		} else if !exists {
			writeError(w, 404, "卡片不存在")
			return false
		}
	}
	return true
}
