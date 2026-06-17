package httpx

import (
	"bytes"
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
				id, err := s.store.CreateFolder(store.Folder{
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
							id, err := s.store.CreateFolder(store.Folder{
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
					exists, err := s.store.BookmarkExistsInFolder(*targetFolderID, url)
					if err != nil {
						return err
					}

					if exists {
						result.Skipped++
					} else {
						_, err := s.store.CreateBookmark(store.Bookmark{
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
	return result, err
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
