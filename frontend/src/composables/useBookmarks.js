import { computed, ref } from 'vue'
import { findFolderById, flattenFolders, normalizeFolder } from '../utils/bookmarkTree'

export function useBookmarks({ getBookmarkFolders, getBookmarks, searchBookmarks, onError }) {
  const folders = ref([])
  const bookmarks = ref([])
  const bookmarkCache = ref({})
  const activeFolderId = ref(null)
  const bookmarkSearch = ref({ q: '', loading: false, results: [] })
  const bookmarkSelectionMode = ref(false)
  const selectedBookmarkIds = ref([])

  const activeFolder = computed(() => findFolderById(folders.value, activeFolderId.value))
  const folderFlatList = computed(() => flattenFolders(folders.value))
  const folderCount = computed(() => folderFlatList.value.length)
  const bookmarkCount = computed(() => bookmarks.value.length)
  let bookmarkSearchTimer

  function getBookmarkSelectionIds() {
    return selectedBookmarkIds.value
  }

  function isBookmarkSelected(bookmarkId) {
    return selectedBookmarkIds.value.includes(bookmarkId)
  }

  function clearBookmarkSelection() {
    bookmarkSelectionMode.value = false
    selectedBookmarkIds.value = []
  }

  function toggleBookmarkSelection(bookmark) {
    const ids = new Set(selectedBookmarkIds.value)
    if (ids.has(bookmark.id)) ids.delete(bookmark.id)
    else ids.add(bookmark.id)
    selectedBookmarkIds.value = Array.from(ids)
    bookmarkSelectionMode.value = selectedBookmarkIds.value.length > 0
  }

  async function loadFolders(parentId = null) {
    try {
      const data = await getBookmarkFolders(parentId || undefined)
      const list = (data.folders || []).map((folder) => normalizeFolder(folder, parentId))
      if (parentId == null) {
        folders.value = list
        if (!folders.value.length) {
          activeFolderId.value = null
          bookmarks.value = []
          clearBookmarkSelection()
          return
        }
        if (!activeFolderId.value) await selectFolder(folders.value[0])
      } else {
        const parent = findFolderById(folders.value, parentId)
        if (parent) {
          parent.children = list
          parent.childrenLoaded = true
          parent.hasChildren = list.length > 0
          parent.expanded = true
          parent.loading = false
        }
      }
    } catch (error) {
      onError?.(error)
    }
  }

  async function selectFolder(folder) {
    const cached = bookmarkCache.value[folder.id]
    activeFolderId.value = folder.id
    bookmarkSearch.value.q = ''
    bookmarkSearch.value.results = []
    clearBookmarkSelection()
    bookmarks.value = cached ? [...cached] : []
    try {
      const data = await getBookmarks(folder.id)
      const items = data.items || []
      bookmarkCache.value = { ...bookmarkCache.value, [folder.id]: items }
      if (activeFolderId.value === folder.id) {
        bookmarks.value = items
      }
    } catch (error) {
      onError?.(error)
    }
  }

  async function runBookmarkSearch() {
    const q = bookmarkSearch.value.q.trim()
    if (!q) {
      bookmarkSearch.value.results = []
      return
    }

    // 保存当前搜索词，防止旧请求覆盖新请求
    const currentQ = q

    bookmarkSearch.value.loading = true
    try {
      const data = await searchBookmarks(q)
      if (bookmarkSearch.value.q.trim() === currentQ) {
        bookmarkSearch.value.results = data.items || []
      }
    } catch (error) {
      onError?.(error)
    } finally {
      if (bookmarkSearch.value.q.trim() === currentQ) {
        bookmarkSearch.value.loading = false
      }
    }
  }

  function handleBookmarkSearchInput() {
    clearTimeout(bookmarkSearchTimer)
    const q = bookmarkSearch.value.q.trim()
    if (!q) {
      bookmarkSearch.value.results = []
      return
    }
    bookmarkSearchTimer = setTimeout(() => {
      runBookmarkSearch()
    }, 250)
  }

  function clearBookmarkSearch() {
    bookmarkSearch.value.q = ''
    bookmarkSearch.value.results = []
  }

  return {
    folders,
    bookmarks,
    bookmarkCache,
    activeFolderId,
    bookmarkSearch,
    bookmarkSelectionMode,
    selectedBookmarkIds,
    activeFolder,
    folderFlatList,
    folderCount,
    bookmarkCount,
    getBookmarkSelectionIds,
    isBookmarkSelected,
    clearBookmarkSelection,
    toggleBookmarkSelection,
    loadFolders,
    selectFolder,
    runBookmarkSearch,
    handleBookmarkSearchInput,
    clearBookmarkSearch,
  }
}
