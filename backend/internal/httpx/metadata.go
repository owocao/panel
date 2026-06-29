package httpx

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (s *Server) metadata(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
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
