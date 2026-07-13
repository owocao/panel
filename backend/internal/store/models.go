package store

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type NavGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

type NavItem struct {
	ID      int64  `json:"id"`
	GroupID int64  `json:"groupId"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	LANURL  string `json:"lanUrl"`
	WANURL  string `json:"wanUrl"`
	URLMode string `json:"urlMode"`
	Sort    int    `json:"sort"`
}

type Folder struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parentId"`
	Name        string `json:"name"`
	Sort        int    `json:"sort"`
	HasChildren bool   `json:"hasChildren"`
}

type Bookmark struct {
	ID       int64  `json:"id"`
	FolderID int64  `json:"folderId"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Favicon  string `json:"favicon"`
	Note     string `json:"note"`
	Sort     int    `json:"sort"`
	Path     string `json:"path,omitempty"`
}
