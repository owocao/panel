package httpx

import (
	"net/http"
	"strings"
)

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	settings, err := s.store.ListSettings()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, settings)
}

func (s *Server) saveSettings(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	var values map[string]string
	if !decodeJSON(w, r, &values) {
		return
	}
	allowed := map[string]bool{"siteTitle": true, "showTitle": true, "showClock": true, "showSeconds": true, "showSearch": true, "searchEngines": true, "backgroundUrl": true, "backgroundColor": true, "lanDetectTimeout": true, "s3Endpoint": true, "s3Region": true, "s3Bucket": true, "s3AccessKey": true, "s3SecretKey": true, "s3Prefix": true, "s3PathStyle": true, "s3Enabled": true, "s3PublicBase": true}
	clean := map[string]string{}
	for k, v := range values {
		if allowed[k] {
			clean[k] = strings.TrimSpace(v)
		}
	}
	if err := s.store.SaveSettings(clean); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, clean)
}
