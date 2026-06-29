package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"biu-panel/backend/internal/store"
)

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
