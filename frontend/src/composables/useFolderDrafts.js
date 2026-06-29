import { computed, ref } from 'vue'
import { cloneFolderTree, findFolderById, flattenFolders, normalizeFolder } from '../utils/bookmarkTree'

export function useFolderDrafts({
  settingsOpen,
  activeSettings,
  folders,
  bookmarks,
  activeFolderId,
  loadFolders,
  createBookmarkFolder,
  updateBookmarkFolder,
  deleteBookmarkFolder,
  onStatus,
}) {
  const foldersDraft = ref([])
  const foldersDraftDirty = ref(false)
  const folderManagementTree = computed(() => (settingsOpen.value && activeSettings.value === '收藏夹' ? foldersDraft.value : folders.value))
  const folderManagementFlatList = computed(() => flattenFolders(folderManagementTree.value))

  function syncFoldersDraftFromFolders() {
    foldersDraft.value = cloneFolderTree(folders.value)
    foldersDraftDirty.value = false
  }

  function findFolderContainerAndIndex(nodes, id, parent = null) {
    for (let i = 0; i < (nodes || []).length; i += 1) {
      const folder = nodes[i]
      if (folder.id === id) return { folder, siblings: nodes, index: i, parent }
      const nested = findFolderContainerAndIndex(folder.children, id, folder)
      if (nested) return nested
    }
    return null
  }

  function isFolderDescendant(folderId, ancestorId, nodes = folders.value) {
    for (const folder of nodes || []) {
      if (folder.id === ancestorId) {
        return Boolean(findFolderById(folder.children || [], folderId))
      }
      if (isFolderDescendant(folderId, ancestorId, folder.children || [])) return true
    }
    return false
  }

  function eligibleFolderParents(folder, nodes = folders.value) {
    return flattenFolders(nodes).filter((item) => item.id !== folder.id && item.depth < 3 && !isFolderDescendant(item.id, folder.id, nodes))
  }

  function isSettingsBookmarkManager() {
    return settingsOpen.value && activeSettings.value === '收藏夹'
  }

  function getFolderTreeForAction() {
    return isSettingsBookmarkManager() ? foldersDraft.value : folders.value
  }

  function markFoldersDraftDirty() {
    if (isSettingsBookmarkManager()) foldersDraftDirty.value = true
  }

  function normalizeParentId(value) {
    return value == null || value === '' ? null : value
  }

  function folderPayloadChanged(original, payload) {
    if (!original) return true
    return original.name !== payload.name || normalizeParentId(original.parentId) !== normalizeParentId(payload.parentId) || Number(original.sort || 0) !== Number(payload.sort || 0)
  }

  async function saveFolderDraftOrder() {
    const originalFlat = flattenFolders(folders.value)
    const draftFlat = flattenFolders(foldersDraft.value)
    const originalIds = originalFlat.map((folder) => folder.id)
    const originalById = new Map(originalFlat.map((folder) => [folder.id, folder]))
    const draftExistingIds = draftFlat.filter((folder) => typeof folder.id === 'number').map((folder) => folder.id)
    const idMap = new Map()

    const removedIds = new Set(originalFlat.filter((folder) => !draftExistingIds.includes(folder.id)).map((folder) => folder.id))
    const removedFolders = originalFlat
      .filter((folder) => removedIds.has(folder.id) && !removedIds.has(folder.parentId))
      .sort((a, b) => a.depth - b.depth)
    for (const folder of removedFolders) {
      await deleteBookmarkFolder(folder.id)
    }

    async function saveFolderNodes(nodes, parentId = null) {
      for (let index = 0; index < (nodes || []).length; index += 1) {
        const folder = nodes[index]
        const targetParentId = parentId == null ? null : idMap.get(parentId) || parentId
        const payload = { name: folder.name, parentId: targetParentId, sort: index + 1 }
        if (typeof folder.id === 'number' && originalIds.includes(folder.id)) {
          if (folderPayloadChanged(originalById.get(folder.id), payload)) {
            await updateBookmarkFolder({ id: folder.id, ...payload })
          }
          idMap.set(folder.id, folder.id)
        } else {
          const created = await createBookmarkFolder(payload)
          idMap.set(folder.id, created.id || created)
        }
        await saveFolderNodes(folder.children || [], folder.id)
      }
    }

    await saveFolderNodes(foldersDraft.value)
    await loadFolders()
    syncFoldersDraftFromFolders()
  }

  async function moveFolder(folder, offset) {
    const tree = getFolderTreeForAction()
    const found = findFolderContainerAndIndex(tree, folder.id)
    if (!found) return
    const { siblings, index } = found
    const targetIndex = index + offset
    if (index < 0 || targetIndex < 0 || targetIndex >= siblings.length) return
    const moved = found.folder
    const target = siblings[targetIndex]
    const folderSort = moved.sort || index + 1
    const targetSort = target.sort || targetIndex + 1
    siblings.splice(index, 1)
    siblings.splice(targetIndex, 0, moved)
    moved.sort = targetSort
    target.sort = folderSort
    if (isSettingsBookmarkManager()) {
      markFoldersDraftDirty()
      return
    }
    try {
      await Promise.all([
        updateBookmarkFolder({ ...moved, sort: moved.sort }),
        updateBookmarkFolder({ ...target, sort: target.sort }),
      ])
    } catch (error) {
      onStatus?.(`排序保存失败：${error.message}`)
      await loadFolders(moved.parentId ?? null)
    }
  }

  async function removeFolder(folder) {
    if (isSettingsBookmarkManager()) {
      if (!confirm(`确认在草稿中删除文件夹「${folder.name}」及其内容？点击保存后才会真正删除。`)) return
      const found = findFolderContainerAndIndex(foldersDraft.value, folder.id)
      if (!found) return
      found.siblings.splice(found.index, 1)
      markFoldersDraftDirty()
      onStatus?.(`已删除收藏夹草稿：${folder.name}`)
      return
    }
    if (!confirm(`确认删除文件夹「${folder.name}」及其内容？`)) return
    if (!confirm('删除后无法恢复，确认永久删除？')) return
    await deleteBookmarkFolder(folder.id)
    if (activeFolderId.value === folder.id) {
      activeFolderId.value = null
      bookmarks.value = []
    }
    await loadFolders(folder.parentId ?? null)
  }

  function createDraftFolder({ id, parentId, name, sort }) {
    return normalizeFolder({ id, parentId, name, sort, children: [], childrenLoaded: true }, parentId)
  }

  return {
    foldersDraft,
    foldersDraftDirty,
    folderManagementTree,
    folderManagementFlatList,
    syncFoldersDraftFromFolders,
    isSettingsBookmarkManager,
    getFolderTreeForAction,
    markFoldersDraftDirty,
    eligibleFolderParents,
    isFolderDescendant,
    findFolderContainerAndIndex,
    folderPayloadChanged,
    normalizeParentId,
    saveFolderDraftOrder,
    moveFolder,
    removeFolder,
    createDraftFolder,
  }
}
