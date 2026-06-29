import { findFolderById, normalizeFolder } from '../utils/bookmarkTree'

export function useEditSave({
  editDialog,
  quickNav,
  quickBookmark,
  metadataLoading,
  assetUploading,
  settingsOpen,
  navGroups,
  navGroupsDraft,
  settingsSearchEngines,
  webSearch,
  folders,
  foldersDraft,
  folderFlatList,
  folderManagementFlatList,
  activeFolder,
  activeFolderId,
  bookmarks,
  bookmarkSearch,
  createNavGroup,
  createNavItem,
  createBookmarkFolder,
  createBookmark,
  updateNavGroup,
  updateNavItem,
  updateBookmarkFolder,
  updateBookmark,
  deleteNavItem,
  uploadAsset,
  fetchMetadata,
  loadNavigation,
  loadFolders,
  loadAllFolderChildren,
  selectFolder,
  runBookmarkSearch,
  closeEditDialog,
  isImageValue,
  isSettingsBookmarkManager,
  getFolderTreeForAction,
  isFolderDescendant,
  findFolderContainerAndIndex,
  markFoldersDraftDirty,
  markNavDraftDirty,
  updateNavDraftGroup,
  upsertNavDraftItem,
  removeNavDraftItem,
  createDraftId,
  writeSearchEngines,
  saveBookmarkAsNavCard,
  onStatus,
  getStatus,
}) {
  async function deleteEditingNavCard() {
    const item = editDialog.value.form
    if (!item?.id || !confirm(`确认删除卡片「${item.name}」？`)) return
    try {
      if (settingsOpen.value) {
        removeNavDraftItem(item.id)
        closeEditDialog()
        return
      }
      await deleteNavItem(item.id)
      await loadNavigation()
      closeEditDialog()
    } catch (error) {
      onStatus?.(error.message)
    }
  }

  async function uploadIconFile(event, target, field = 'icon') {
    const file = event.target.files?.[0]
    if (!file) return
    assetUploading.value = true
    try {
      const data = await uploadAsset(file)
      target[field] = data.url
      onStatus?.('图片已上传到本地数据目录')
    } catch (error) {
      onStatus?.(error.message)
    } finally {
      assetUploading.value = false
      event.target.value = ''
    }
  }

  async function fillMetadata(target) {
    const url = target.url || target.wanUrl || target.lanUrl || ''
    if (!url.trim()) {
      onStatus?.('请先填写网址')
      return
    }
    let fetchUrl = url.trim()
    if (!/^https?:\/\//i.test(fetchUrl)) {
      fetchUrl = 'http://' + fetchUrl
    }

    metadataLoading.value = true
    try {
      const data = await fetchMetadata(fetchUrl)
      if ('title' in target && data.title && !target.title) target.title = data.title
      if ('name' in target && data.title && !target.name) target.name = data.title
      if ('favicon' in target && data.favicon) target.favicon = data.favicon
      if ('icon' in target && data.favicon) {
        target.icon = data.favicon
        target.iconMode = 'image'
      }
      onStatus?.('已自动抓取标题和图标')
    } catch (error) {
      const message = error.message || '抓取失败'
      onStatus?.(message)
      alert(message)
      setTimeout(() => { if (getStatus?.() === message) onStatus?.('') }, 3000)
    } finally {
      metadataLoading.value = false
    }
  }

  async function fillMetadataFromField(target, field) {
    const url = target[field] || ''
    if (!url.trim()) {
      onStatus?.('请先填写对应网址')
      return
    }
    const previousUrl = target.url
    target.url = url
    await fillMetadata(target)
    if (previousUrl === undefined) delete target.url
    else target.url = previousUrl
  }

  async function fillQuickBookmarkMetadata() {
    await fillMetadata(quickBookmark.value)
  }

  async function fillQuickNavMetadata() {
    const target = { name: quickNav.value.cardName, wanUrl: quickNav.value.url, lanUrl: quickNav.value.url, icon: '' }
    await fillMetadata(target)
    quickNav.value.cardName = target.name || quickNav.value.cardName
  }

  async function saveEditDialog() {
    const { type, form } = editDialog.value
    try {
      if (type === 'navGroup' || type === 'navGroupCreate') {
        if (!form.name?.trim()) return
        if (form.name.trim().length > 10) {
          onStatus?.('分组名称最多 10 个字')
          return
        }
        if (settingsOpen.value) {
          if (type === 'navGroupCreate') {
            navGroupsDraft.value = [...navGroupsDraft.value, { id: createDraftId('draft-group'), name: form.name.trim(), sort: navGroupsDraft.value.length + 1, items: [] }]
            markNavDraftDirty()
          }
          else updateNavDraftGroup(form.id, (group) => ({ ...group, name: form.name.trim() }))
          closeEditDialog()
          return
        }
        if (type === 'navGroupCreate') await createNavGroup({ name: form.name.trim(), sort: navGroups.value.length + 1 })
        else await updateNavGroup({ ...form, name: form.name.trim() })
        await loadNavigation()
      }
      if (type === 'navItem' || type === 'navItemCreate') {
        if (!form.name?.trim()) {
          onStatus?.('请填写标题')
          return
        }
        if (!form.groupId) {
          onStatus?.('请选择分组')
          return
        }
        if (!form.wanUrl?.trim()) {
          onStatus?.('请填写公网地址')
          return
        }
        const name = form.name.trim()
        if (name.length > 15) {
          onStatus?.('标题最多 15 个字')
          return
        }
        const iconMode = form.iconMode || (isImageValue(form.icon) ? 'image' : 'text')
        const icon = iconMode === 'image' ? (form.icon || '') : (form.icon || name)
        if (iconMode === 'text' && icon.length > 5) {
          onStatus?.('文本内容最多 5 个字')
          return
        }
        const payload = { groupId: form.groupId, name, icon, lanUrl: form.lanUrl || '', wanUrl: form.wanUrl || '', urlMode: 'wan', sort: form.sort || 0 }
        if (settingsOpen.value) {
          upsertNavDraftItem({ id: form.id || createDraftId('draft-card'), ...payload })
          closeEditDialog()
          return
        }
        if (type === 'navItemCreate') await createNavItem(payload)
        else await updateNavItem({ id: form.id, ...payload })
        await loadNavigation()
      }
      if (type === 'folder' || type === 'folderCreate') {
        if (!form.name?.trim()) return
        const parentId = form.parentId == null || form.parentId === '' ? null : form.parentId
        if (type === 'folder' && parentId === form.id) {
          onStatus?.('不能移动到自身')
          return
        }
        const tree = getFolderTreeForAction()
        if (type === 'folder' && parentId && isFolderDescendant(parentId, form.id, tree)) {
          onStatus?.('不能移动到自己的子收藏夹内')
          return
        }
        if (isSettingsBookmarkManager()) {
          if (type === 'folderCreate') {
            const siblings = parentId ? (findFolderById(foldersDraft.value, parentId)?.children || []) : foldersDraft.value
            siblings.push(normalizeFolder({ id: createDraftId('draft-folder'), parentId, name: form.name.trim(), sort: siblings.length + 1, children: [], childrenLoaded: true }, parentId))
          } else {
            const found = findFolderContainerAndIndex(foldersDraft.value, form.id)
            if (found) {
              const moved = found.folder
              if (moved.parentId !== parentId) {
                found.siblings.splice(found.index, 1)
                const siblings = parentId ? (findFolderById(foldersDraft.value, parentId)?.children || []) : foldersDraft.value
                moved.parentId = parentId
                moved.sort = siblings.length + 1
                siblings.push(moved)
              }
              moved.name = form.name.trim()
            }
          }
          markFoldersDraftDirty()
          closeEditDialog()
          return
        }
        if (type === 'folderCreate') await createBookmarkFolder({ parentId, name: form.name.trim(), sort: form.sort || folderFlatList.value.length + 1 })
        else await updateBookmarkFolder({ ...form, parentId, name: form.name.trim() })
        await loadFolders()
        await loadAllFolderChildren(folders.value)
      }
      if (type === 'bookmark' || type === 'bookmarkCreate') {
        if (!form.title?.trim() || !form.url?.trim()) return
        const folderId = form.folderId || activeFolderId.value
        if (!folderId) {
          onStatus?.('请选择要新增到的文件夹')
          return
        }
        const payload = { ...form, title: form.title.trim(), url: form.url.trim(), note: form.note || '', favicon: form.favicon || form.title.trim().slice(0, 1), folderId }
        if (type === 'bookmarkCreate') await createBookmark(payload)
        else await updateBookmark(payload)
        if (activeFolder.value) await selectFolder(activeFolder.value)
        if (bookmarkSearch.value.q.trim()) await runBookmarkSearch()
      }
      if (type === 'bookmarkToNav') {
        if (!form.bookmark) return
        await saveBookmarkAsNavCard(form.bookmark, form.groupId)
      }
      if (type === 'searchEngine' || type === 'searchEngineCreate') {
        if (!form.title?.trim() || !form.url?.trim()) {
          onStatus?.('请填写搜索引擎标题和 URL')
          return
        }
        const next = { key: form.key || `custom-${Date.now()}`, title: form.title.trim(), url: form.url.trim(), icon: form.icon || form.title.trim().slice(0, 1) }
        const engines = type === 'searchEngineCreate' ? [...settingsSearchEngines.value, next] : settingsSearchEngines.value.map((item) => item.key === next.key ? next : item)
        writeSearchEngines(engines)
        if (!webSearch.value.engine) webSearch.value.engine = next.key
      }
      closeEditDialog()
    } catch (error) {
      onStatus?.(error.message)
    }
  }

  async function createGroupByPrompt() {
    editDialog.value = { open: true, type: 'navGroupCreate', title: '新增导航分组', form: { name: '' } }
  }

  async function addCardFromMenu(group) {
    editDialog.value = {
      open: true,
      type: 'navItemCreate',
      title: `新增卡片 · ${group.name}`,
      form: { groupId: group.id, name: '', iconMode: 'text', icon: '', lanUrl: '', wanUrl: '', urlMode: 'wan', sort: (group.items?.length || 0) + 1 },
    }
  }

  async function createFolderByPrompt(parent = null) {
    const list = isSettingsBookmarkManager() ? folderManagementFlatList.value : folderFlatList.value
    editDialog.value = { open: true, type: 'folderCreate', title: parent ? `新增收藏夹 · ${parent.name}` : '新增收藏夹', form: { parentId: parent?.id || null, name: '', sort: parent ? (parent.children?.length || 0) + 1 : list.length + 1 } }
  }

  async function createBookmarkByPrompt(folder = activeFolder.value || folders.value[0]) {
    if (isSettingsBookmarkManager()) {
      onStatus?.('设置页内暂不直接新增书签，请保存后在收藏夹抽屉中新增')
      return
    }
    let targetFolder = folder
    if (!targetFolder) {
      editDialog.value = { open: true, type: 'folderCreate', title: '请先创建收藏夹', form: { parentId: null, name: '默认收藏', sort: folderFlatList.value.length + 1 } }
      return
    }
    if (!targetFolder) {
      onStatus?.('请先创建收藏夹')
      return
    }
    editDialog.value = { open: true, type: 'bookmarkCreate', title: `新增书签 · ${targetFolder.name}`, form: { folderId: targetFolder.id, title: '', url: '', favicon: '', note: '', sort: bookmarks.value.length + 1 } }
  }

  return {
    saveEditDialog,
    deleteEditingNavCard,
    uploadIconFile,
    fillMetadata,
    fillMetadataFromField,
    fillQuickBookmarkMetadata,
    fillQuickNavMetadata,
    createGroupByPrompt,
    addCardFromMenu,
    createFolderByPrompt,
    createBookmarkByPrompt,
  }
}
