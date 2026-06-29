package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
