package httpx

import (
	"biu-panel/backend/internal/store"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) bookmarkFolders(w http.ResponseWriter, r *http.Request) {
	parent, ok := optionalInt(r.URL.Query().Get("parentId"))
	if !ok && r.URL.Query().Get("parentId") != "" {
		writeError(w, 400, "parentId 格式错误")
		return
	}
	folders, err := s.store.ListFolders(parent)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"folders": folders})
}
func (s *Server) createBookmarkFolder(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var f store.Folder
	if !decodeJSON(w, r, &f) {
		return
	}
	if f.Name == "" {
		writeError(w, 400, "文件夹名称不能为空")
		return
	}
	id, err := s.store.CreateFolder(f)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	f.ID = id
	writeJSON(w, 201, f)
}
func (s *Server) updateBookmarkFolder(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var f store.Folder
	if !decodeJSON(w, r, &f) {
		return
	}
	if f.ID == 0 || f.Name == "" {
		writeError(w, 400, "文件夹 ID 和名称必填")
		return
	}
	if err := s.store.UpdateFolder(f); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, f)
}
func (s *Server) deleteBookmarkFolder(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	id, err := idFromQuery(r)
	if err != nil {
		writeError(w, 400, "id 必填")
		return
	}
	if err := s.store.DeleteFolder(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) bookmarks(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("folderId"), 10, 64)
	if err != nil {
		writeError(w, 400, "folderId 必填")
		return
	}
	items, err := s.store.ListBookmarks(id)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (s *Server) createBookmark(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var b store.Bookmark
	if !decodeJSON(w, r, &b) {
		return
	}
	if b.FolderID == 0 || b.Title == "" || b.URL == "" {
		writeError(w, 400, "文件夹、标题、URL 必填")
		return
	}
	id, err := s.store.CreateBookmark(b)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	b.ID = id
	writeJSON(w, 201, b)
}
func (s *Server) updateBookmark(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var b store.Bookmark
	if !decodeJSON(w, r, &b) {
		return
	}
	if b.ID == 0 || b.FolderID == 0 || b.Title == "" || b.URL == "" {
		writeError(w, 400, "ID、文件夹、标题、URL 必填")
		return
	}
	if err := s.store.UpdateBookmark(b); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, b)
}
func (s *Server) deleteBookmark(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	id, err := idFromQuery(r)
	if err != nil {
		writeError(w, 400, "id 必填")
		return
	}
	if err := s.store.DeleteBookmark(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) bookmarkSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, 200, map[string]any{"items": []store.Bookmark{}})
		return
	}
	items, err := s.store.SearchBookmarks(q)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
