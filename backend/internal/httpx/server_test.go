package httpx

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"biu-panel/backend/internal/config"
	"biu-panel/backend/internal/store"
)

func plainRequest(method, path string, cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	return req
}

func newTestServer(t *testing.T) (*Server, *store.Store, *http.Cookie) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "biu-panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := CreateInitialUser(st, "admin", "password123"); err != nil {
		t.Fatal(err)
	}
	user, err := st.FindUser("admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SaveSession("test-session", user.ID, nil, false); err != nil {
		t.Fatal(err)
	}
	return New(config.Config{DataDir: dir, StaticDir: dir}, st), st, &http.Cookie{Name: "biu_session", Value: "test-session"}
}

func jsonRequest(t *testing.T, method, path string, body any, cookie *http.Cookie) *http.Request {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	return req
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestPrivateReadAPIsRequireLogin(t *testing.T) {
	srv, _, _ := newTestServer(t)
	cases := []struct {
		name string
		path string
		run  http.HandlerFunc
	}{
		{name: "navigation", path: "/api/navigation", run: srv.navigation},
		{name: "bookmark folders", path: "/api/bookmark/folders", run: srv.bookmarkFolders},
		{name: "bookmarks", path: "/api/bookmarks?folderId=1", run: srv.bookmarks},
		{name: "bookmark search", path: "/api/bookmark/search?q=test", run: srv.bookmarkSearch},
		{name: "metadata", path: "/api/metadata?url=https://example.com", run: srv.metadata},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tc.run(rr, plainRequest(http.MethodGet, tc.path, nil))
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("want 401, got %d body=%s", rr.Code, rr.Body.String())
			}
			if got := decodeResponse(t, rr)["error"]; got != "未登录" {
				t.Fatalf("unexpected error: %v", got)
			}
		})
	}
}

func TestPrivateReadAPIsAllowLoggedInNavigationAndFolders(t *testing.T) {
	srv, st, cookie := newTestServer(t)
	if _, err := st.CreateNavGroup(store.NavGroup{Name: "服务", Sort: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateFolder(store.Folder{Name: "资料", Sort: 1}); err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.navigation(rr, plainRequest(http.MethodGet, "/api/navigation", cookie))
	if rr.Code != http.StatusOK {
		t.Fatalf("navigation want 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.bookmarkFolders(rr, plainRequest(http.MethodGet, "/api/bookmark/folders", cookie))
	if rr.Code != http.StatusOK {
		t.Fatalf("bookmarkFolders want 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateNavItemValidation(t *testing.T) {
	srv, st, cookie := newTestServer(t)
	groupID, err := st.CreateNavGroup(store.NavGroup{Name: "服务", Sort: 1})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.createNavItem(rr, jsonRequest(t, http.MethodPost, "/api/navigation/items", store.NavItem{Name: "百度", GroupID: groupID}, cookie))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	if got := decodeResponse(t, rr)["error"]; got != "公网地址不能为空" {
		t.Fatalf("unexpected error: %v", got)
	}

	rr = httptest.NewRecorder()
	srv.createNavItem(rr, jsonRequest(t, http.MethodPost, "/api/navigation/items", store.NavItem{Name: "百度", GroupID: groupID, WANURL: "www.baidu.com", URLMode: "bad"}, cookie))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	if got := decodeResponse(t, rr)["error"]; got != "打开方式不支持" {
		t.Fatalf("unexpected error: %v", got)
	}

	rr = httptest.NewRecorder()
	srv.createNavItem(rr, jsonRequest(t, http.MethodPost, "/api/navigation/items", store.NavItem{Name: " 百度 ", GroupID: groupID, WANURL: " www.baidu.com "}, cookie))
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	if data["name"] != "百度" || data["wanUrl"] != "www.baidu.com" || data["urlMode"] != "wan" {
		t.Fatalf("unexpected data: %#v", data)
	}
}

func TestUpdateAndDeleteNavItemNotFound(t *testing.T) {
	srv, st, cookie := newTestServer(t)
	groupID, err := st.CreateNavGroup(store.NavGroup{Name: "服务", Sort: 1})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.updateNavItem(rr, jsonRequest(t, http.MethodPut, "/api/navigation/items", store.NavItem{ID: 999, Name: "百度", GroupID: groupID, WANURL: "www.baidu.com"}, cookie))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rr.Code)
	}
	if got := decodeResponse(t, rr)["error"]; got != "卡片不存在" {
		t.Fatalf("unexpected error: %v", got)
	}

	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/navigation/items?id=999", nil)
	req.AddCookie(cookie)
	srv.deleteNavItem(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rr.Code)
	}
}

func TestRestoreNavigationBackupValidation(t *testing.T) {
	srv, _, cookie := newTestServer(t)
	backup := `{"version":1,"groups":[{"id":1,"name":"服务","sort":1}],"items":[{"id":1,"groupId":1,"name":"百度","urlMode":"wan"}]}`
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "navigation.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(backup)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/navigation/restore", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	srv.restoreNavigationBackup(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "导航卡片公网地址不能为空") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
