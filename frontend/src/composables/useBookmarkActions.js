import { ensureHttp } from '../utils/navigation'
import { isImageValue } from '../utils/display'

export function useBookmarkActions({
  folders,
  bookmarks,
  activeFolderId,
  activeFolder,
  bookmarkSearch,
  bookmarkSelectionMode,
  selectedBookmarkIds,
  moveDialog,
  editDialog,
  folderFlatList,
  foldersDraft,
  updateBookmark,
  updateBookmarkFolder,
  deleteBookmark,
  refreshBookmarkFavicon,
  loadFolders,
  loadAllFolderChildren,
  selectFolder,
  runBookmarkSearch,
  clearBookmarkSelection,
  isBookmarkSelected,
  isSettingsBookmarkManager,
  syncFoldersDraftFromFolders,
  getFolderTreeForAction,
  eligibleFolderParents,
  isFolderDescendant,
  findFolderContainerAndIndex,
  findFolderById,
  markFoldersDraftDirty,
  onStatus,
}) {
  const faviconRefreshRequests = new Set()

  function enableBookmarkSelection(bookmark) {
    if (!bookmarkSelectionMode.value) bookmarkSelectionMode.value = true
    if (!isBookmarkSelected(bookmark.id)) selectedBookmarkIds.value = [...selectedBookmarkIds.value, bookmark.id]
  }

  async function openMoveDialog(items, title) {
    if (!folders.value.length) await loadFolders()
    await loadAllFolderChildren(folders.value)
    moveDialog.value = {
      open: true,
      title,
      items,
      kind: 'bookmarks',
      itemLabel: '收藏',
      allowRoot: false,
      selectableFolders: folderFlatList.value,
      targetFolderId: activeFolderId.value || folderFlatList.value[0]?.id || null,
    }
  }

  async function openFolderMoveDialog(folder) {
    if (isSettingsBookmarkManager()) {
      if (!foldersDraft.value.length && folders.value.length) syncFoldersDraftFromFolders()
    } else {
      if (!folders.value.length) await loadFolders()
      await loadAllFolderChildren(folders.value)
    }
    const tree = getFolderTreeForAction()
    moveDialog.value = {
      open: true,
      title: `移动收藏夹 · ${folder.name}`,
      items: [folder],
      kind: 'folder',
      itemLabel: '收藏夹',
      allowRoot: true,
      selectableFolders: eligibleFolderParents(folder, tree),
      targetFolderId: folder.parentId ?? null,
    }
  }

  async function moveBookmarkItems(items, folderId) {
    const targetFolderId = Number(folderId) || 0
    if (!targetFolderId) {
      onStatus?.('请选择目标文件夹')
      return
    }
    for (const bookmark of items) {
      await updateBookmark({ ...bookmark, folderId: targetFolderId })
    }
    moveDialog.value = { open: false, title: '', items: [], targetFolderId: null }
    clearBookmarkSelection()
    if (activeFolder.value) await selectFolder(activeFolder.value)
  }

  async function moveFolderToParent(folder, parentId) {
    if (!folder) return
    const targetParentId = parentId == null || parentId === '' ? null : parentId
    if (targetParentId === folder.id) {
      onStatus?.('不能移动到自身')
      return
    }
    const tree = getFolderTreeForAction()
    if (targetParentId && isFolderDescendant(targetParentId, folder.id, tree)) {
      onStatus?.('不能移动到自己的子收藏夹内')
      return
    }
    if (isSettingsBookmarkManager()) {
      const found = findFolderContainerAndIndex(foldersDraft.value, folder.id)
      if (!found) return
      const moved = found.folder
      found.siblings.splice(found.index, 1)
      const siblings = targetParentId ? (findFolderById(foldersDraft.value, targetParentId)?.children || []) : foldersDraft.value
      moved.parentId = targetParentId
      moved.sort = siblings.length + 1
      siblings.push(moved)
      markFoldersDraftDirty()
      moveDialog.value = { open: false, title: '', items: [], kind: '', itemLabel: '收藏', allowRoot: false, selectableFolders: [], targetFolderId: null }
      onStatus?.(`已移动收藏夹草稿：${folder.name}`)
      return
    }
    const siblings = targetParentId ? (findFolderById(folders.value, targetParentId)?.children || []) : folders.value
    await updateBookmarkFolder({ ...folder, parentId: targetParentId, sort: siblings.length + 1 })
    moveDialog.value = { open: false, title: '', items: [], kind: '', itemLabel: '收藏', allowRoot: false, selectableFolders: [], targetFolderId: null }
    await loadFolders()
    await loadAllFolderChildren(folders.value)
    onStatus?.(`已移动收藏夹：${folder.name}`)
  }

  async function confirmMoveDialog() {
    if (moveDialog.value.kind === 'folder') {
      await moveFolderToParent(moveDialog.value.items?.[0], moveDialog.value.targetFolderId)
      return
    }
    await moveBookmarkItems(moveDialog.value.items || [], moveDialog.value.targetFolderId)
  }

  async function editFolder(folder) {
    if (!isSettingsBookmarkManager()) {
      if (!folders.value.length) await loadFolders()
      await loadAllFolderChildren(folders.value)
    }
    editDialog.value = { open: true, type: 'folder', title: '编辑收藏夹', form: { ...folder } }
  }

  function editBookmark(bookmark) {
    editDialog.value = { open: true, type: 'bookmark', title: '编辑书签', form: { ...bookmark } }
  }

  async function moveBookmark(bookmark, offset) {
    const list = [...bookmarks.value]
    const index = list.findIndex((item) => item.id === bookmark.id)
    const targetIndex = index + offset
    if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
    const target = list[targetIndex]
    const bookmarkSort = bookmark.sort || index + 1
    const targetSort = target.sort || targetIndex + 1
    await updateBookmark({ ...bookmark, sort: targetSort })
    await updateBookmark({ ...target, sort: bookmarkSort })
    await selectFolder(activeFolder.value)
  }

  async function removeBookmark(bookmark) {
    if (!confirm(`确认删除收藏「${bookmark.title}」？`)) return
    await deleteBookmark(bookmark.id)
    if (bookmarkSelectionMode.value) {
      selectedBookmarkIds.value = selectedBookmarkIds.value.filter((id) => id !== bookmark.id)
      if (!selectedBookmarkIds.value.length) bookmarkSelectionMode.value = false
    }
    if (activeFolder.value) await selectFolder(activeFolder.value)
  }

  function getVisibleBookmarkList() {
    return bookmarkSearch.value.q.trim() ? bookmarkSearch.value.results : bookmarks.value
  }

  function getSelectedBookmarks() {
    const source = new Map(getVisibleBookmarkList().map((item) => [item.id, item]))
    return selectedBookmarkIds.value.map((id) => source.get(id)).filter(Boolean)
  }

  async function deleteSelectedBookmarks() {
    const items = getSelectedBookmarks()
    if (!items.length) {
      onStatus?.('请先选择要删除的收藏')
      return
    }
    if (!confirm(`确认删除选中的 ${items.length} 条收藏？`)) return
    for (const bookmark of items) {
      await deleteBookmark(bookmark.id)
    }
    clearBookmarkSelection()
    if (activeFolder.value) await selectFolder(activeFolder.value)
    if (bookmarkSearch.value.q.trim()) await runBookmarkSearch()
  }

  function openSelectedMoveDialog() {
    const items = getSelectedBookmarks()
    if (!items.length) {
      onStatus?.('请先选择要移动的收藏')
      return
    }
    openMoveDialog(items, '批量移动收藏')
  }

  async function batchSelectBookmark(bookmark) {
    enableBookmarkSelection(bookmark)
    onStatus?.(`已加入批量选择：${bookmark.title}`)
  }

  async function openBookmarkUrl(bookmark) {
    const url = ensureHttp(bookmark?.url || '')
    if (!url) return
    const refreshRequest = requestBookmarkFaviconRefresh(bookmark)
    if (refreshRequest) {
      await Promise.race([refreshRequest, new Promise((resolve) => setTimeout(resolve, 250))]).catch(() => {})
    }
    window.location.href = url
  }

  function requestBookmarkFaviconRefresh(bookmark) {
    if (!refreshBookmarkFavicon || !bookmark?.id || isImageValue(bookmark.favicon)) return null
    if (faviconRefreshRequests.has(bookmark.id)) return null
    faviconRefreshRequests.add(bookmark.id)
    return refreshBookmarkFavicon(bookmark.id).catch(() => {
      faviconRefreshRequests.delete(bookmark.id)
    })
  }

  return {
    openMoveDialog,
    openFolderMoveDialog,
    confirmMoveDialog,
    moveBookmarkItems,
    moveFolderToParent,
    editFolder,
    editBookmark,
    moveBookmark,
    removeBookmark,
    getVisibleBookmarkList,
    getSelectedBookmarks,
    deleteSelectedBookmarks,
    openSelectedMoveDialog,
    batchSelectBookmark,
    openBookmarkUrl,
  }
}
