package main

import (
"database/sql"
"fmt"
_ "modernc.org/sqlite"
)

func main() {
db, err := sql.Open("sqlite", "./data/db/biu-panel.db")
if err != nil {
panic(err)
}
defer db.Close()

q := "google"
like := "%" + q + "%"
rows, err := db.Query(`WITH RECURSIVE folder_paths(id, path) AS (
SELECT id, name FROM bookmark_folders WHERE parent_id IS NULL
UNION ALL
SELECT f.id, folder_paths.path || ' / ' || f.name FROM bookmark_folders f JOIN folder_paths ON f.parent_id = folder_paths.id
)
SELECT b.title, COALESCE(folder_paths.path, f.name)
FROM bookmarks b
JOIN bookmark_folders f ON f.id=b.folder_id
LEFT JOIN folder_paths ON folder_paths.id=f.id
WHERE b.title LIKE ? OR b.url LIKE ? OR b.note LIKE ?
ORDER BY b.title LIMIT 100`, like, like, like)

if err != nil {
fmt.Println("Error:", err)
return
}
defer rows.Close()
for rows.Next() {
var title, path string
rows.Scan(&title, &path)
fmt.Println("Title:", title, "Path:", path)
}
}
