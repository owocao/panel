package httpx

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
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

func (s *Server) uploadAsset(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	if err := r.ParseMultipartForm(8 * 1024 * 1024); err != nil {
		writeError(w, 400, "文件最大支持 8MB")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "请选择上传文件")
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}
	if !strings.HasPrefix(contentType, "image/") {
		writeError(w, 400, "仅支持图片文件")
		return
	}
	if ext == "" {
		ext = ".bin"
	}
	name := time.Now().Format("20060102-150405") + "-" + randomShort() + ext
	dir := filepath.Join(s.cfg.DataDir, "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	path := filepath.Join(dir, name)
	out, err := os.Create(path)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer out.Close()
	written, err := io.Copy(out, io.LimitReader(file, 8*1024*1024+1))
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if written > 8*1024*1024 {
		_ = os.Remove(path)
		writeError(w, 400, "文件最大支持 8MB")
		return
	}
	publicPath := "/uploads/" + name
	source := "local"
	if settings, err := s.store.ListSettings(); err == nil && settings["s3Enabled"] == "true" {
		prefix := strings.Trim(settings["s3Prefix"], "/")
		key := "uploads/" + name
		if prefix != "" {
			key = prefix + "/" + key
		}
		if err := s.s3PutObject(settings, key, contentType, mustReadFile(path)); err == nil {
			publicPath = s.s3PublicURL(settings, key)
			source = "s3"
		}
	}
	_, _ = s.store.DB.Exec(`INSERT INTO assets(name,source,path,mime,size,created_at) VALUES(?,?,?,?,?,?)`, header.Filename, source, publicPath, contentType, written, time.Now().Format(time.RFC3339))
	writeJSON(w, 201, map[string]any{"url": publicPath, "name": header.Filename, "size": written, "mime": contentType, "source": source})
}

func randomShort() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func mustReadFile(path string) []byte {
	data, _ := os.ReadFile(path)
	return data
}

func (s *Server) s3PublicURL(settings map[string]string, key string) string {
	if base := strings.TrimRight(settings["s3PublicBase"], "/"); base != "" {
		return base + "/" + key
	}
	endpoint := strings.TrimRight(settings["s3Endpoint"], "/")
	bucket := settings["s3Bucket"]
	if settings["s3PathStyle"] == "false" {
		if u, err := url.Parse(endpoint); err == nil {
			u.Host = bucket + "." + u.Host
			u.Path = "/" + key
			return u.String()
		}
	}
	return endpoint + "/" + bucket + "/" + key
}

func (s *Server) testS3(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	settings, err := s.store.ListSettings()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	key := strings.Trim(settings["s3Prefix"], "/")
	if key != "" {
		key += "/"
	}
	key += "test/connection-check.txt"
	payload := []byte("biu-panel s3 connectivity check\n")
	if err := s.s3PutObject(settings, key, "text/plain; charset=utf-8", payload); err != nil {
		writeError(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"key": key, "url": s.s3PublicURL(settings, key), "size": len(payload)})
}

func (s *Server) s3PutObject(settings map[string]string, key, contentType string, payload []byte) error {
	endpoint := strings.TrimRight(settings["s3Endpoint"], "/")
	bucket := settings["s3Bucket"]
	accessKey := settings["s3AccessKey"]
	secretKey := settings["s3SecretKey"]
	region := settings["s3Region"]
	if region == "" {
		region = "auto"
	}
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return errors.New("S3 配置不完整")
	}
	base, err := url.Parse(endpoint)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return errors.New("S3 Endpoint 格式错误")
	}
	pathStyle := settings["s3PathStyle"] != "false"
	objectPath := "/" + bucket + "/" + key
	if !pathStyle {
		base.Host = bucket + "." + base.Host
		objectPath = "/" + key
	}
	base.Path = objectPath
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := sha256Hex(payload)
	headers := map[string]string{
		"host":                 base.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
		"content-type":         contentType,
	}
	canonicalHeaders := "content-type:" + headers["content-type"] + "\n" + "host:" + headers["host"] + "\n" + "x-amz-content-sha256:" + payloadHash + "\n" + "x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{"PUT", uriEncodePath(objectPath), "", canonicalHeaders, signedHeaders, payloadHash}, "\n")
	credentialScope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))
	signature := hex.EncodeToString(hmacSHA256(signingKey(secretKey, dateStamp, region), []byte(stringToSign)))
	authorization := "AWS4-HMAC-SHA256 Credential=" + accessKey + "/" + credentialScope + ", SignedHeaders=" + signedHeaders + ", Signature=" + signature
	req, err := http.NewRequest(http.MethodPut, base.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("Authorization", authorization)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("S3 上传失败：%s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func signingKey(secret, date, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func uriEncodePath(v string) string {
	parts := strings.Split(v, "/")
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(url.QueryEscape(part), "+", "%20")
	}
	return strings.Join(parts, "/")
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

func (s *Server) metadata(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" {
		writeError(w, 400, "url 必填")
		return
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, 400, "仅支持 http/https 地址")
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		writeError(w, 400, "网址格式错误")
		return
	}
	req.Header.Set("User-Agent", "biu-panel/0.1 metadata fetcher")
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, 502, "抓取网页失败")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		writeError(w, 502, "网页返回错误状态")
		return
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		writeError(w, 502, "读取网页失败")
		return
	}
	html := string(body)
	title := extractFirst(html, `(?is)<title[^>]*>(.*?)</title>`)
	favicon := extractFirst(html, `(?is)<link[^>]+rel=["'][^"']*(?:icon|shortcut icon|apple-touch-icon)[^"']*["'][^>]*href=["']([^"']+)["']`)
	if favicon == "" {
		favicon = extractFirst(html, `(?is)<link[^>]+href=["']([^"']+)["'][^>]*rel=["'][^"']*(?:icon|shortcut icon|apple-touch-icon)[^"']*["']`)
	}
	if favicon != "" {
		if ref, err := url.Parse(favicon); err == nil {
			favicon = u.ResolveReference(ref).String()
		}
	} else {
		favicon = u.Scheme + "://" + u.Host + "/favicon.ico"
	}
	writeJSON(w, 200, map[string]string{"title": strings.TrimSpace(htmlUnescape(title)), "favicon": favicon})
}

func extractFirst(input, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(input)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func htmlUnescape(v string) string {
	replacer := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"", "&#39;", "'")
	return replacer.Replace(v)
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
