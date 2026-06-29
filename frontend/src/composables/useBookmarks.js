import { computed, ref } from 'vue'
import { findFolderById, flattenFolders } from '../utils/bookmarkTree'

export function useBookmarks() {
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
  }
}
