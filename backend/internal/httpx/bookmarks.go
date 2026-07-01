package httpx

import (
	"biu-panel/backend/internal/store"
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (s *Server) bookmarkFolders(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
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
	if !s.requireAuth(w, r) {
		return
	}
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
func (s *Server) refreshBookmarkFavicon(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	id, err := bookmarkIDFromRefreshRequest(w, r)
	if err != nil {
		return
	}
	if id == 0 {
		writeError(w, 400, "id 必填")
		return
	}
	bookmark, err := s.store.GetBookmark(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, 404, "书签不存在")
			return
		}
		writeError(w, 500, err.Error())
		return
	}
	if isBookmarkFaviconImage(bookmark.Favicon) {
		writeJSON(w, 200, map[string]any{"ok": true, "favicon": bookmark.Favicon})
		return
	}
	favicon := bookmarkDefaultFaviconURLFromBookmark(bookmark.URL)
	if favicon == "" {
		writeJSON(w, 200, map[string]bool{"ok": true})
		return
	}
	if err := s.store.UpdateBookmarkFavicon(bookmark.ID, favicon); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "favicon": favicon})
}

func bookmarkIDFromRefreshRequest(w http.ResponseWriter, r *http.Request) (int64, error) {
	if rawID := strings.TrimSpace(r.URL.Query().Get("id")); rawID != "" {
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			writeError(w, 400, "id 格式错误")
			return 0, err
		}
		return id, nil
	}
	var payload struct {
		ID int64 `json:"id"`
	}
	if !decodeJSON(w, r, &payload) {
		return 0, http.ErrBodyReadAfterClose
	}
	return payload.ID, nil
}

func bookmarkMetadataURLCandidates(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return []string{value}
	}
	return []string{"https://" + value, "http://" + value}
}

func bookmarkDefaultFaviconURLFromBookmark(raw string) string {
	for _, candidate := range bookmarkMetadataURLCandidates(raw) {
		if favicon := bookmarkDefaultFaviconURL(candidate); favicon != "" {
			return favicon
		}
	}
	return ""
}

func bookmarkDefaultFaviconURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
}

func isBookmarkFaviconImage(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.HasPrefix(trimmed, "/uploads/") ||
		strings.HasPrefix(trimmed, "http://") ||
		strings.HasPrefix(trimmed, "https://") ||
		strings.HasPrefix(trimmed, "data:image/")
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
	if !s.requireAuth(w, r) {
		return
	}
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
