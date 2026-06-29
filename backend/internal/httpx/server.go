package httpx

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"biu-panel/backend/internal/config"
	"biu-panel/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	cfg   config.Config
	store *store.Store
}

const (
	maxNavGroupNameLength = 10
	maxNavItemTitleLength = 15
	maxNavURLLength       = 2048
)

func New(cfg config.Config, st *store.Store) *Server { return &Server{cfg: cfg, store: st} }

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/setup/status", s.setupStatus)
	mux.HandleFunc("POST /api/setup", s.setup)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.HandleFunc("GET /api/auth/me", s.me)
	mux.HandleFunc("GET /api/navigation", s.navigation)
	mux.HandleFunc("POST /api/navigation/groups", s.createNavGroup)
	mux.HandleFunc("PUT /api/navigation/groups", s.updateNavGroup)
	mux.HandleFunc("DELETE /api/navigation/groups", s.deleteNavGroup)
	mux.HandleFunc("POST /api/navigation/items", s.createNavItem)
	mux.HandleFunc("PUT /api/navigation/items", s.updateNavItem)
	mux.HandleFunc("DELETE /api/navigation/items", s.deleteNavItem)
	mux.HandleFunc("GET /api/bookmark/folders", s.bookmarkFolders)
	mux.HandleFunc("POST /api/bookmark/folders", s.createBookmarkFolder)
	mux.HandleFunc("PUT /api/bookmark/folders", s.updateBookmarkFolder)
	mux.HandleFunc("DELETE /api/bookmark/folders", s.deleteBookmarkFolder)
	mux.HandleFunc("GET /api/bookmarks", s.bookmarks)
	mux.HandleFunc("POST /api/bookmarks", s.createBookmark)
	mux.HandleFunc("PUT /api/bookmarks", s.updateBookmark)
	mux.HandleFunc("DELETE /api/bookmarks", s.deleteBookmark)
	mux.HandleFunc("GET /api/bookmark/search", s.bookmarkSearch)
	mux.HandleFunc("GET /api/metadata", s.metadata)
	mux.HandleFunc("GET /api/backup/download", s.downloadBackup)
	mux.HandleFunc("POST /api/backup/restore", s.restoreBackup)
	mux.HandleFunc("GET /api/navigation/backup", s.downloadNavigationBackup)
	mux.HandleFunc("POST /api/navigation/restore", s.restoreNavigationBackup)
	mux.HandleFunc("POST /api/s3/test", s.testS3)
	mux.HandleFunc("POST /api/assets/upload", s.uploadAsset)
	mux.HandleFunc("GET /api/bookmark/export", s.exportBookmarks)
	mux.HandleFunc("POST /api/bookmark/import", s.importBookmarks)
	mux.HandleFunc("GET /api/settings", s.getSettings)
	mux.HandleFunc("PUT /api/settings", s.saveSettings)
	return withCORS(s.withStatic(mux))
}

func (s *Server) withStatic(api http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/uploads/") {
			http.ServeFile(w, r, filepath.Join(s.cfg.DataDir, filepath.Clean(r.URL.Path)))
			return
		}
		path := filepath.Join(s.cfg.StaticDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			http.ServeFile(w, r, path)
			return
		}
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		http.ServeFile(w, r, filepath.Join(s.cfg.StaticDir, "index.html"))
	})
}

func CreateInitialUser(st *store.Store, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return st.CreateUser(username, string(hash))
}

func StartCleanupTask(st *store.Store, retentionDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		if err := st.CleanupExpiredData(retentionDays); err != nil {
			log.Printf("[CleanupTask] Failed to cleanup data: %v\n", err)
		} else {
			log.Println("[CleanupTask] Successfully cleaned up expired sessions and old login logs.")
		}
	}
}

func (s *Server) exportBookmarks(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	folders, err := s.store.ListAllFolders()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	bookmarks, err := s.store.ListAllBookmarks()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	children := map[int64][]store.Folder{}
	roots := []store.Folder{}
	for _, f := range folders {
		if f.ParentID == nil {
			roots = append(roots, f)
		} else {
			children[*f.ParentID] = append(children[*f.ParentID], f)
		}
	}
	byFolder := map[int64][]store.Bookmark{}
	for _, b := range bookmarks {
		byFolder[b.FolderID] = append(byFolder[b.FolderID], b)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=bookmarks.html")
	_, _ = fmt.Fprintln(w, `<!DOCTYPE NETSCAPE-Bookmark-file-1>`)
	_, _ = fmt.Fprintln(w, `<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">`)
	_, _ = fmt.Fprintln(w, `<TITLE>Bookmarks</TITLE><H1>Bookmarks</H1><DL><p>`)
	var writeFolder func(store.Folder, int)
	writeFolder = func(f store.Folder, depth int) {
		indent := strings.Repeat("    ", depth)
		_, _ = fmt.Fprintf(w, "%s<DT><H3>%s</H3>\n%s<DL><p>\n", indent, html.EscapeString(f.Name), indent)
		for _, child := range children[f.ID] {
			writeFolder(child, depth+1)
		}
		for _, b := range byFolder[f.ID] {
			_, _ = fmt.Fprintf(w, "%s    <DT><A HREF=\"%s\" ICON=\"%s\">%s</A>\n", indent, html.EscapeString(b.URL), html.EscapeString(b.Favicon), html.EscapeString(b.Title))
			if b.Note != "" {
				_, _ = fmt.Fprintf(w, "%s    <DD>%s\n", indent, html.EscapeString(b.Note))
			}
		}
		_, _ = fmt.Fprintf(w, "%s</DL><p>\n", indent)
	}
	for _, root := range roots {
		writeFolder(root, 1)
	}
	_, _ = fmt.Fprintln(w, `</DL><p>`)
}

func (s *Server) importBookmarks(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	if err := r.ParseMultipartForm(20 * 1024 * 1024); err != nil {
		writeError(w, 400, "导入文件最大支持 20MB")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "请选择书签 HTML 文件")
		return
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, 20*1024*1024+1))
	if err != nil || len(body) > 20*1024*1024 {
		writeError(w, 400, "读取导入文件失败或文件过大")
		return
	}
	result, err := s.parseBookmarkHTML(string(body))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, result)
}

func stripTags(v string) string {
	return regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(v, "")
}

func (s *Server) writeDataTar(tw *tar.Writer) error {
	return filepath.WalkDir(s.cfg.DataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(s.cfg.DataDir, path)
		if err != nil {
			return nil
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}
		header.Name = filepath.ToSlash(filepath.Join("data", rel))
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
}

func (s *Server) restoreBackup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	if err := r.ParseMultipartForm(64 * 1024 * 1024); err != nil {
		writeError(w, 400, "备份文件最大支持 64MB")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "请选择备份文件")
		return
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		writeError(w, 400, "备份文件格式错误")
		return
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	restored := 0
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeError(w, 400, "读取备份失败")
			return
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Clean(header.Name)
		if !strings.HasPrefix(name, "data"+string(filepath.Separator)) {
			continue
		}
		rel, err := filepath.Rel("data", name)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		target := filepath.Join(s.cfg.DataDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		out, err := os.Create(target)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		_, copyErr := io.Copy(out, io.LimitReader(tr, 64*1024*1024+1))
		closeErr := out.Close()
		if copyErr != nil || closeErr != nil {
			writeError(w, 500, "恢复文件失败")
			return
		}
		restored++
	}
	writeJSON(w, 200, map[string]int{"files": restored})
}

func (s *Server) downloadBackup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	fileName := "biu-panel-backup-" + time.Now().Format("20060102-150405") + ".tar.gz"
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	gz := gzip.NewWriter(w)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	_ = s.writeDataTar(tw)
}

type navigationBackupFile struct {
	Version   int              `json:"version"`
	CreatedAt string           `json:"createdAt"`
	Groups    []store.NavGroup `json:"groups"`
	Items     []store.NavItem  `json:"items"`
}

func (s *Server) downloadNavigationBackup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	groups, itemsByGroup, err := s.store.ListNavigation()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	items := []store.NavItem{}
	for _, group := range groups {
		items = append(items, itemsByGroup[group.ID]...)
	}
	fileName := "biu-panel-navigation-" + time.Now().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	_ = json.NewEncoder(w).Encode(navigationBackupFile{Version: 1, CreatedAt: time.Now().Format(time.RFC3339), Groups: groups, Items: items})
}

func (s *Server) restoreNavigationBackup(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	if err := r.ParseMultipartForm(8 * 1024 * 1024); err != nil {
		writeError(w, 400, "导航备份文件最大支持 8MB")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "请选择导航备份文件")
		return
	}
	defer file.Close()
	var backup navigationBackupFile
	decoder := json.NewDecoder(io.LimitReader(file, 8*1024*1024+1))
	if err := decoder.Decode(&backup); err != nil {
		writeError(w, 400, "导航备份文件格式错误")
		return
	}
	if backup.Version != 1 {
		writeError(w, 400, "不支持的导航备份版本")
		return
	}
	if len(backup.Groups) > 100 || len(backup.Items) > 5000 {
		writeError(w, 400, "导航备份数据过大")
		return
	}
	groupNames := map[string]bool{}
	groupItems := map[int64][]store.NavItem{}
	for index := range backup.Groups {
		group := backup.Groups[index]
		group.Name = strings.TrimSpace(group.Name)
		if group.ID == 0 {
			writeError(w, 400, "导航分组 ID 必填")
			return
		}
		if group.Name == "" {
			writeError(w, 400, "导航分组名称不能为空")
			return
		}
		if len([]rune(group.Name)) > maxNavGroupNameLength {
			writeError(w, 400, "导航分组名称不能超过 10 个字")
			return
		}
		if group.Sort < 0 {
			group.Sort = 0
		}
		if groupNames[group.Name] {
			writeError(w, 400, "导航分组名称重复")
			return
		}
		backup.Groups[index] = group
		groupNames[group.Name] = true
		groupItems[group.ID] = []store.NavItem{}
	}
	for _, item := range backup.Items {
		item.Name = strings.TrimSpace(item.Name)
		item.LANURL = strings.TrimSpace(item.LANURL)
		item.WANURL = strings.TrimSpace(item.WANURL)
		item.URLMode = strings.TrimSpace(item.URLMode)
		if item.Name == "" || item.GroupID == 0 {
			writeError(w, 400, "导航卡片标题和分组必填")
			return
		}
		if len([]rune(item.Name)) > maxNavItemTitleLength {
			writeError(w, 400, "导航卡片标题不能超过 15 个字")
			return
		}
		if item.WANURL == "" {
			writeError(w, 400, "导航卡片公网地址不能为空")
			return
		}
		if len(item.LANURL) > maxNavURLLength || len(item.WANURL) > maxNavURLLength {
			writeError(w, 400, "导航卡片地址不能超过 2048 个字符")
			return
		}
		if _, ok := groupItems[item.GroupID]; !ok {
			writeError(w, 400, "导航卡片引用了不存在的分组")
			return
		}
		if item.URLMode == "" || item.URLMode == "auto" {
			item.URLMode = "wan"
		}
		if item.URLMode != "lan" && item.URLMode != "wan" {
			writeError(w, 400, "导航卡片打开方式不支持")
			return
		}
		if item.Sort < 0 {
			item.Sort = 0
		}
		groupItems[item.GroupID] = append(groupItems[item.GroupID], item)
	}
	if err := s.store.ReplaceNavigation(backup.Groups, groupItems); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]int{"groups": len(backup.Groups), "items": len(backup.Items)})
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return strings.TrimSpace(strings.Split(v, ",")[0])
	}
	return r.RemoteAddr
}

func idFromQuery(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
}

func optionalInt(v string) (*int64, bool) {
	if v == "" {
		return nil, true
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, false
	}
	return &n, true
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, 400, "请求格式错误")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": status < 400, "data": data})
}
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": msg})
}
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
