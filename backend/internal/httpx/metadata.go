package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type metadataResult struct {
	Title   string
	Favicon string
}

type metadataFetchError struct {
	status  int
	message string
}

func (e metadataFetchError) Error() string {
	return e.message
}

func (s *Server) metadata(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	data, err := fetchMetadata(r.Context(), raw)
	if err != nil {
		if fetchErr, ok := err.(metadataFetchError); ok {
			writeError(w, fetchErr.status, fetchErr.message)
			return
		}
		writeError(w, 502, "抓取网页失败")
		return
	}
	writeJSON(w, 200, map[string]string{"title": data.Title, "favicon": data.Favicon})
}

func fetchMetadata(ctx context.Context, raw string) (metadataResult, error) {
	var result metadataResult
	if raw == "" {
		return result, metadataFetchError{status: 400, message: "url 必填"}
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return result, metadataFetchError{status: 400, message: "仅支持 http/https 地址"}
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return result, metadataFetchError{status: 400, message: "网址格式错误"}
	}
	req.Header.Set("User-Agent", "biu-panel/0.1 metadata fetcher")
	resp, err := client.Do(req)
	if err != nil {
		return result, metadataFetchError{status: 502, message: "抓取网页失败"}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return result, metadataFetchError{status: 502, message: "网页返回错误状态"}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return result, metadataFetchError{status: 502, message: "读取网页失败"}
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
	return metadataResult{Title: strings.TrimSpace(htmlUnescape(title)), Favicon: favicon}, nil
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
