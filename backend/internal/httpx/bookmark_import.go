package httpx

import (
	"bytes"
	"database/sql"
	"strconv"
	"strings"

	"biu-panel/backend/internal/store"
	"golang.org/x/net/html"
)

type ImportResult struct {
	Folders   int `json:"folders"`
	Bookmarks int `json:"bookmarks"`
	Skipped   int `json:"skipped"`
}

func (s *Server) parseBookmarkHTML(input string) (ImportResult, error) {
	var result ImportResult
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return result, err
	}
	tx, err := s.store.DB.Begin()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	existing, err := loadBookmarkImportExistingURLs(tx)
	if err != nil {
		return result, err
	}

	var traverse func(*html.Node, *int64) error

	// Create a default folder id placeholder in case bookmarks are found outside of any folder
	var defaultFolderID *int64

	traverse = func(n *html.Node, currentFolderID *int64) error {
		if n.Type == html.ElementNode && strings.ToUpper(n.Data) == "DT" {
			// A <DT> can contain either an <H3> (folder) or an <A> (bookmark)
			var h3, a, dl, dd *html.Node
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode {
					tagName := strings.ToUpper(c.Data)
					if tagName == "H3" {
						h3 = c
					} else if tagName == "A" {
						a = c
					} else if tagName == "DL" {
						dl = c
					}
				}
			}

			// Some formats have <DD> after <A> for notes. We look at the next sibling of <DT>
			for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
				if sibling.Type == html.ElementNode {
					if strings.ToUpper(sibling.Data) == "DD" {
						dd = sibling
					}
					break // Only check the immediate next element node
				} else if sibling.Type == html.TextNode && strings.TrimSpace(sibling.Data) == "" {
					continue
				} else {
					break
				}
			}

			if h3 != nil {
				// It's a folder
				folderName := extractText(h3)

				// Create the folder
				id, err := createImportFolder(tx, store.Folder{
					ParentID: currentFolderID,
					Name:     folderName,
					Sort:     result.Folders + 1,
				})
				if err != nil {
					return err
				}
				result.Folders++

				// Check if there is a <DL> inside this <DT> following the <H3>
				if dl != nil {
					if err := traverse(dl, &id); err != nil {
						return err
					}
				} else {
					// In most formats, <DL> is a sibling of the <H3>'s <DT>
					for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
						if sibling.Type == html.ElementNode {
							if strings.ToUpper(sibling.Data) == "DL" {
								if err := traverse(sibling, &id); err != nil {
									return err
								}
							}
							break
						}
					}
				}
			} else if a != nil {
				// It's a bookmark
				title := extractText(a)
				var url, icon string
				for _, attr := range a.Attr {
					key := strings.ToUpper(attr.Key)
					if key == "HREF" {
						url = attr.Val
					} else if key == "ICON" {
						icon = attr.Val
					}
				}

				var note string
				if dd != nil {
					note = extractText(dd)
				}

				if url != "" {
					targetFolderID := currentFolderID
					if targetFolderID == nil {
						if defaultFolderID == nil {
							id, err := createImportFolder(tx, store.Folder{
								Name: "导入书签",
								Sort: result.Folders + 1,
							})
							if err != nil {
								return err
							}
							defaultFolderID = &id
							result.Folders++
						}
						targetFolderID = defaultFolderID
					}

					// Check for duplicates in the current folder
					key := bookmarkImportKey(*targetFolderID, url)
					if existing[key] {
						result.Skipped++
					} else {
						_, err := createImportBookmark(tx, store.Bookmark{
							FolderID: *targetFolderID,
							Title:    title,
							URL:      url,
							Favicon:  icon,
							Note:     note,
							Sort:     result.Bookmarks + 1,
						})
						if err != nil {
							return err
						}
						existing[key] = true
						result.Bookmarks++
					}
				}
			}
		} else {
			// Traverse children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := traverse(c, currentFolderID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	err = traverse(doc, nil)
	if err != nil {
		return result, err
	}
	return result, tx.Commit()
}

func loadBookmarkImportExistingURLs(tx *sql.Tx) (map[string]bool, error) {
	rows, err := tx.Query(`SELECT folder_id,url FROM bookmarks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	existing := map[string]bool{}
	for rows.Next() {
		var folderID int64
		var url string
		if err := rows.Scan(&folderID, &url); err != nil {
			return nil, err
		}
		existing[bookmarkImportKey(folderID, url)] = true
	}
	return existing, rows.Err()
}

func bookmarkImportKey(folderID int64, url string) string {
	return strconv.FormatInt(folderID, 10) + "\x00" + url
}

func createImportFolder(tx *sql.Tx, f store.Folder) (int64, error) {
	var parent any
	if f.ParentID != nil {
		parent = *f.ParentID
	}
	res, err := tx.Exec(`INSERT INTO bookmark_folders(parent_id,name,sort) VALUES(?,?,?)`, parent, f.Name, f.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func createImportBookmark(tx *sql.Tx, b store.Bookmark) (int64, error) {
	res, err := tx.Exec(`INSERT INTO bookmarks(folder_id,title,url,favicon,note,sort) VALUES(?,?,?,?,?,?)`, b.FolderID, b.Title, b.URL, b.Favicon, b.Note, b.Sort)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func extractText(n *html.Node) string {
	var buf bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(buf.String())
}
