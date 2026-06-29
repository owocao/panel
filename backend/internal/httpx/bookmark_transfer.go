package httpx

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"

	"biu-panel/backend/internal/store"
)

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
