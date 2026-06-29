package httpx

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
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
