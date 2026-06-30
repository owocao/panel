<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import BookmarkDrawer from './components/BookmarkDrawer.vue'
import ContextMenu from './components/ContextMenu.vue'
import EditDialog from './components/EditDialog.vue'
import FloatingActions from './components/FloatingActions.vue'
import HomeHero from './components/HomeHero.vue'
import MoveDialog from './components/MoveDialog.vue'
import NavDragFloat from './components/NavDragFloat.vue'
import SettingsPanel from './components/settings/SettingsPanel.vue'
import { useBackupRestore } from './composables/useBackupRestore'
import { useBookmarkActions } from './composables/useBookmarkActions'
import { useBookmarks } from './composables/useBookmarks'
import { useContextMenu } from './composables/useContextMenu'
import { useDragSort } from './composables/useDragSort'
import { useEditDialog } from './composables/useEditDialog'
import { useEditSave } from './composables/useEditSave'
import { useFolderDrafts } from './composables/useFolderDrafts'
import { useNavigation } from './composables/useNavigation'
import { useSettings } from './composables/useSettings'
import { findFolderById, normalizeFolder } from './utils/bookmarkTree'
import { cardTextClass, formatDisplayDate, formatDisplayTime, getNetworkIcon, getNetworkTip, iconUrl, isImageValue, limitText } from './utils/display'
import { ensureHttp, normalizeNetworkMode, resolveNavUrl } from './utils/navigation'
import {
  createBookmark,
  createBookmarkFolder,
  createNavGroup,
  createNavItem,
  deleteBookmark,
  deleteBookmarkFolder,
  deleteNavGroup,
  deleteNavItem,
  downloadFile,
  fetchMetadata,
  getBookmarkFolders,
  getBookmarks,
  getMe,
  getNavigation,
  getSettings,
  importBookmarkHTML,
  login,
  restoreBackup,
  restoreNavigationBackup,
  setup,
  searchBookmarks,
  saveSettings,
  setupStatus,
  testS3,
  updateBookmark,
  updateBookmarkFolder,
  updateNavGroup,
  updateNavItem,
  uploadAsset,
} from './lib/api'

const activeView = ref('boot')
const drawerOpen = ref(false)
const statusText = ref('正在连接后端...')
const toastText = ref('')
const user = ref(null)
const initialized = ref(false)
const moveDialog = ref({ open: false, title: '', items: [], targetFolderId: null })
const loginForm = ref({ username: '', password: '', remember: false })
const setupForm = ref({ username: 'admin', password: '', confirm: '' })
const quickNav = ref({ groupName: '', cardName: '', url: '' })
const quickBookmark = ref({ folderName: '', title: '', url: '', note: '', favicon: '' })
const editingNavGroupId = ref(null)
const metadataLoading = ref(false)
const assetUploading = ref(false)
const now = ref(new Date())
const dateMode = ref('solar')
let clockTimer
let toastTimer
let draftIdSeed = 0
const navGroupsDraft = ref([])
const navDraftDirty = ref(false)

const showNetworkSwitcher = computed(() => true)
const networkTip = computed(() => getNetworkTip(networkMode.value))
const networkIcon = computed(() => getNetworkIcon(networkMode.value))
const displayTime = computed(() => formatDisplayTime(now.value, settingsForm.value.showSeconds === 'true'))
const displayDate = computed(() => formatDisplayDate(now.value, dateMode.value))
function toggleDateMode() {
  dateMode.value = dateMode.value === 'solar' ? 'lunar' : 'solar'
}
const {
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
} = useBookmarks({
  getBookmarkFolders,
  getBookmarks,
  searchBookmarks,
  onError: (error) => { statusText.value = error.message },
})

const {
  settingsOpen,
  activeSettings,
  settingsMenuCollapsed,
  settingsMessage,
  settingsSaving,
  settingsForm,
  settingsDraft,
  openSettings,
  closeSettings,
  selectSettingsMenu,
  loadSettings,
  submitTestS3,
  submitSettings,
  settingsDraftDirty,
} = useSettings({
  getSettings,
  saveSettings,
  testS3,
  getNetworkMode: () => networkMode.value,
  setNetworkMode: (mode) => { networkMode.value = mode },
  navGroupsDraft,
  navDraftDirty,
  folders,
  getFoldersDraftDirty: () => foldersDraftDirty.value,
  setFoldersDraft: (value) => { foldersDraft.value = value },
  setFoldersDraftDirty: (value) => { foldersDraftDirty.value = value },
  getNavGroups: () => navGroups.value,
  closeMenu: () => closeMenu(),
  loadFolders,
  loadAllFolderChildren,
  syncFoldersDraftFromFolders: () => syncFoldersDraftFromFolders(),
  saveNavGroupDraftOrder,
  saveFolderDraftOrder: () => saveFolderDraftOrder(),
  onStatus: (message) => { statusText.value = message },
})

const {
  navGroups,
  networkMode,
  webSearch,
  searchPickerOpen,
  displayGroups,
  navItemCount,
  searchEngines,
  settingsSearchEngines,
  activeSearchEngine,
  loadNavigation,
  cycleNetworkMode,
  runWebSearch,
  selectSearchEngine,
  addSearchEngine,
  editSearchEngine,
  removeSearchEngine,
  moveSearchEngine,
  uploadSearchEngineIcon,
  openNavItemFromMenu,
  openNavItem,
  writeSearchEngines,
} = useNavigation({
  getNavigation,
  uploadAsset,
  settingsForm,
  settingsDraft,
  settingsOpen,
  openEditDialog: (dialog) => { editDialog.value = dialog },
  assetUploading,
  isImageValue,
  onStatus: (message) => { statusText.value = message },
  onToast: (message) => showToast(message),
})
const navGroupOptions = computed(() => (settingsOpen.value ? navGroupsDraft.value : navGroups.value))

const {
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
  saveFolderDraftOrder,
  moveFolder,
  removeFolder,
} = useFolderDrafts({
  settingsOpen,
  activeSettings,
  folders,
  bookmarks,
  activeFolderId,
  loadFolders,
  createBookmarkFolder,
  updateBookmarkFolder,
  deleteBookmarkFolder,
  onStatus: (message) => { statusText.value = message },
})

const {
  editDialog,
  groupSelectOpen,
  closeEditDialog,
  clampEditField,
  getEditGroupName,
  selectEditGroup,
  setNavIconMode,
} = useEditDialog({ navGroupOptions, isImageValue })

const {
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
} = useEditSave({
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
  onStatus: (message) => { statusText.value = message },
  getStatus: () => statusText.value,
})

const {
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
} = useBookmarkActions({
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
  onStatus: (message) => { statusText.value = message },
})

const {
  menu,
  menuStyle,
  closeMenu,
  runMenuAction,
  showCardMenu,
  showGroupMenu,
  showFolderMenu,
  showBookmarkMenu,
} = useContextMenu({
  onError: (error) => { statusText.value = error.message },
  onStatus: (message) => { statusText.value = message },
  actions: {
    isNavGroupEditing: (group) => editingNavGroupId.value === group.id,
    openNavItemFromMenu,
    editNavCard,
    removeNavCard,
    addCardFromMenu,
    editNavGroup,
    removeNavGroup,
    getFolderFlatList: () => folderFlatList.value,
    createFolderByPrompt,
    createBookmarkByPrompt,
    openFolderMoveDialog,
    editFolder,
    removeFolder,
    ensureHttp,
    openMoveDialog,
    convertBookmarkToNavCard,
    editBookmark,
    removeBookmark,
  },
})

const {
  dragState,
  navPointerDrag,
  suppressNextNavCardClick,
  startDrag,
  hoverBookmark,
  hoverFolder,
  clearDragState,
  hoverNavCard,
  navDragFloatStyle,
  stopNavPointerListeners,
  clearNavLongPressTimer,
  startNavPointerSort,
  dropNavGroup,
  dropNavCard,
  dropFolder,
  dropBookmark,
} = useDragSort({
  folders,
  bookmarks,
  activeFolder,
  displayGroups,
  editingNavGroupId,
  updateBookmark,
  updateBookmarkFolder,
  updateNavGroup,
  updateNavItem,
  loadFolders,
  loadNavigation,
  selectFolder,
  closeMenu,
  editNavCard,
  onStatus: (message) => { statusText.value = message },
})

const {
  restoreBackupFile,
  downloadNavigationBackup,
  restoreNavigationBackupFile,
  importBookmarksFile,
  exportBookmarks,
} = useBackupRestore({
  downloadFile,
  restoreBackup,
  restoreNavigationBackup,
  importBookmarkHTML,
  loadNavigation,
  loadFolders,
  drawerOpen,
  settingsOpen,
  navGroups,
  navGroupsDraft,
  navDraftDirty,
  folders,
  activeFolderId,
  bookmarks,
  statusText,
  settingsMessage,
})

function handleOverlayWheel(event) {
  const target = event.target instanceof Element ? event.target : event.target?.parentElement
  const scrollable = target?.closest?.('.explorer-sidebar, .explorer-content, .settings-menu, .settings-content, .edit-modal, .context-menu, .select-options')
  if (!scrollable || scrollable.scrollHeight <= scrollable.clientHeight) {
    event.preventDefault()
    return
  }
  if (!event.deltaY) return
  const atTop = scrollable.scrollTop <= 0
  const atBottom = Math.ceil(scrollable.scrollTop + scrollable.clientHeight) >= scrollable.scrollHeight
  if ((event.deltaY < 0 && atTop) || (event.deltaY > 0 && atBottom)) event.preventDefault()
}

const shellStyle = computed(() => ({
  '--runtime-bg': settingsForm.value.backgroundColor || '#02030a',
}))

onMounted(async () => {
  clockTimer = window.setInterval(() => { now.value = new Date() }, 1000)
  window.addEventListener('keydown', handleGlobalKeydown)
  window.addEventListener('biu-auth-expired', handleAuthExpired)
  networkMode.value = normalizeNetworkMode(localStorage.getItem('biu-network-mode'))
  await refreshBootstrap()
  if (activeView.value === 'home') await loadNavigation()
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
  window.removeEventListener('biu-auth-expired', handleAuthExpired)
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
  clearNavLongPressTimer()
  stopNavPointerListeners()
})

function handleGlobalKeydown(event) {
  if (event.key !== 'Escape') return
  if (editDialog.value.open) {
    closeEditDialog()
    return
  }
  if (bookmarkSelectionMode.value) {
    clearBookmarkSelection()
    return
  }
  closeMenu()
}

function handleAuthExpired() {
  user.value = null
  settingsOpen.value = false
  drawerOpen.value = false
  editDialog.value.open = false
  moveDialog.value.open = false
  closeMenu()
  clearBookmarkSelection()
  activeView.value = 'login'
  statusText.value = '登录已失效，请重新登录'
  settingsMessage.value = ''
  settingsSaving.value = false
}

async function loadFolderNodeChildren(folder, options = {}) {
  folder.loading = true
  try {
    const data = await getBookmarkFolders(folder.id)
    const children = (data.folders || []).map((child) => normalizeFolder(child, folder.id))
    folder.children = children
    folder.childrenLoaded = true
    folder.hasChildren = children.length > 0
    if (options.expand !== false) folder.expanded = true
  } catch (error) {
    statusText.value = error.message
  } finally {
    folder.loading = false
  }
}

async function loadAllFolderChildren(nodes = folders.value) {
  for (const folder of nodes || []) {
    if (folder.hasChildren && !folder.childrenLoaded) await loadFolderNodeChildren(folder, { expand: false })
    if (Array.isArray(folder.children) && folder.children.length) await loadAllFolderChildren(folder.children)
  }
}

async function toggleFolderExpanded(folder) {
  if (folder.expanded) {
    folder.expanded = false
    return
  }
  if (!folder.childrenLoaded) {
    folder.expanded = true
    loadFolderNodeChildren(folder)
  } else {
    folder.expanded = true
  }
}

function collapseFolderTree(nodes = folders.value) {
  nodes.forEach((folder) => {
    folder.expanded = false
    if (Array.isArray(folder.children)) collapseFolderTree(folder.children)
  })
}

function expandFolderTree(nodes = folders.value) {
  nodes.forEach((folder) => {
    folder.expanded = true
    if (!folder.childrenLoaded) {
      loadFolderNodeChildren(folder).then(() => expandFolderTree(folder.children || []))
      return
    }
    if (Array.isArray(folder.children) && folder.children.length) expandFolderTree(folder.children)
  })
}

function toggleAllBookmarkFolders() {
  const hasCollapsed = folderFlatList.value.some((folder) => !folder.expanded)
  if (hasCollapsed) expandFolderTree()
  else collapseFolderTree()
}

async function convertBookmarkToNavCard(bookmark) {
  const defaultGroup = navGroups.value[0]
  editDialog.value = { open: true, type: 'bookmarkToNav', title: '首页卡片', form: { bookmark, groupId: defaultGroup?.id || null } }
}

async function saveBookmarkAsNavCard(bookmark, groupId) {
  const group = navGroups.value.find((item) => item.id === Number(groupId))
  if (!group) {
    statusText.value = '请选择目标分组'
    return
  }
  await createNavItem({
    groupId: group.id,
    name: bookmark.title,
    icon: bookmark.favicon || bookmark.title.slice(0, 1),
    lanUrl: bookmark.url,
    wanUrl: bookmark.url,
    urlMode: 'wan',
    sort: (navGroups.value.find((item) => item.id === group.id)?.items?.length || 0) + 1,
  })
  statusText.value = `已设为首页卡片：${bookmark.title}`
  await loadNavigation()
}

function showToast(message) {
  toastText.value = message
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => { toastText.value = '' }, 1800)
}

function createDraftId(prefix) {
  draftIdSeed += 1
  return `${prefix}-${Date.now()}-${draftIdSeed}`
}

function markNavDraftDirty() {
  if (settingsOpen.value) navDraftDirty.value = true
}

function updateNavDraftGroup(groupId, updater) {
  navGroupsDraft.value = navGroupsDraft.value.map((group) => group.id === groupId ? updater(group) : group)
  markNavDraftDirty()
}

function removeNavDraftGroup(groupId) {
  navGroupsDraft.value = navGroupsDraft.value.filter((group) => group.id !== groupId)
  markNavDraftDirty()
}

function upsertNavDraftItem(item) {
  navGroupsDraft.value = navGroupsDraft.value.map((group) => {
    const existingItems = [...(group.items || [])]
    const existingIndex = existingItems.findIndex((entry) => entry.id === item.id)
    const items = existingItems.filter((entry) => entry.id !== item.id)
    if (group.id !== item.groupId) return { ...group, items }
    const nextItems = [...items]
    nextItems.splice(existingIndex >= 0 ? existingIndex : nextItems.length, 0, { ...item })
    return { ...group, items: nextItems.map((entry, index) => ({ ...entry, sort: index + 1 })) }
  })
  markNavDraftDirty()
}

function removeNavDraftItem(itemId) {
  navGroupsDraft.value = navGroupsDraft.value.map((group) => ({ ...group, items: (group.items || []).filter((item) => item.id !== itemId) }))
  markNavDraftDirty()
}

async function refreshBootstrap() {
  try {
    const setupInfo = await setupStatus()
    initialized.value = setupInfo.initialized
    if (!initialized.value) {
      activeView.value = 'setup'
      statusText.value = '等待首次初始化'
      return
    }
    try {
      const me = await getMe()
      user.value = me
      statusText.value = `已登录：${me.username}`
      await loadSettings()
      activeView.value = 'home'
    } catch {
      activeView.value = 'login'
      statusText.value = '请登录管理员账号'
    }
  } catch (error) {
    activeView.value = 'login'
    statusText.value = `后端未连接：${error.message}`
  }
}


async function saveNavGroupDraftOrder() {
  const originalGroups = navGroups.value
  const draftGroups = navGroupsDraft.value
  const originalGroupIds = originalGroups.map((group) => group.id)
  const draftExistingGroupIds = draftGroups.filter((group) => typeof group.id === 'number').map((group) => group.id)
  const createdGroupIds = new Map()

  const originalItemIds = originalGroups.flatMap((group) => group.items || []).map((item) => item.id)
  const draftExistingItemIds = draftGroups.flatMap((group) => group.items || []).filter((item) => typeof item.id === 'number').map((item) => item.id)
  
  const deleteItemPromises = []
  for (const id of originalItemIds) {
    if (!draftExistingItemIds.includes(id)) deleteItemPromises.push(deleteNavItem(id))
  }
  await Promise.all(deleteItemPromises)

  const deleteGroupPromises = []
  for (const group of originalGroups) {
    if (!draftExistingGroupIds.includes(group.id)) deleteGroupPromises.push(deleteNavGroup(group.id))
  }
  await Promise.all(deleteGroupPromises)

  const groupPromises = draftGroups.map(async (group, groupIndex) => {
    const payload = { name: group.name, sort: groupIndex + 1 }
    if (typeof group.id === 'number' && originalGroupIds.includes(group.id)) {
      await updateNavGroup({ id: group.id, ...payload })
      createdGroupIds.set(group.id, group.id)
    } else {
      const created = await createNavGroup(payload)
      createdGroupIds.set(group.id, created.id || created)
    }
  })
  await Promise.all(groupPromises)

  const itemPromises = []
  for (const group of draftGroups) {
    const groupId = createdGroupIds.get(group.id)
    for (let itemIndex = 0; itemIndex < (group.items || []).length; itemIndex += 1) {
      const item = group.items[itemIndex]
      const payload = { groupId, name: item.name, icon: item.icon || '', lanUrl: item.lanUrl || '', wanUrl: item.wanUrl || '', urlMode: item.urlMode === 'lan' ? 'lan' : 'wan', sort: itemIndex + 1 }
      if (typeof item.id === 'number' && originalItemIds.includes(item.id)) {
        itemPromises.push(updateNavItem({ id: item.id, ...payload }))
      } else {
        itemPromises.push(createNavItem(payload))
      }
    }
  }
  await Promise.all(itemPromises)

  await loadNavigation()
}

async function openDrawer() {
  drawerOpen.value = true
  if (!folders.value.length) await loadFolders()
}

async function submitLogin() {
  try {
    const data = await login(loginForm.value)
    user.value = data
    activeView.value = 'home'
    statusText.value = `已登录：${data.username}`
    await loadSettings()
    await loadNavigation()
  } catch (error) {
    statusText.value = error.message
  }
}

async function submitSetup() {
  if (setupForm.value.password !== setupForm.value.confirm) {
    statusText.value = '两次密码不一致'
    return
  }
  try {
    await setup({ username: setupForm.value.username, password: setupForm.value.password })
    initialized.value = true
    activeView.value = 'login'
    statusText.value = '初始化完成，请登录'
  } catch (error) {
    statusText.value = error.message
  }
}



function editNavGroup(group) {
  editDialog.value = { open: true, type: 'navGroup', title: '编辑导航分组', form: { ...group } }
}
function toggleNavGroupEdit(group) {
  editingNavGroupId.value = editingNavGroupId.value === group.id ? null : group.id
}
function openCardEditorInGroup(event, group, item) {
  if (editingNavGroupId.value !== group.id) return false
  event.preventDefault()
  if (suppressNextNavCardClick.value) {
    suppressNextNavCardClick.value = false
    return true
  }
  editNavCard(item, group)
  return true
}

function handleNavCardClick(event, group, item) {
  if (openCardEditorInGroup(event, group, item)) return
  event.preventDefault()
  openNavItem(item)
}

function handleShellClick(event) {
  closeMenu()
  if (editDialog.value.open) return
  drawerOpen.value = false
  if (!editingNavGroupId.value) return
  if (event.target.closest?.('.nav-group.editing')) return
  editingNavGroupId.value = null
}

async function moveNavGroup(group, offset) {
  if (settingsOpen.value) {
    const list = [...navGroupsDraft.value]
    const index = list.findIndex((item) => item.id === group.id)
    const targetIndex = index + offset
    if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
    const target = list[targetIndex]
    list[targetIndex] = group
    list[index] = target
    navGroupsDraft.value = list
    markNavDraftDirty()
    return
  }
  const list = [...navGroups.value]
  const index = list.findIndex((item) => item.id === group.id)
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
  const target = list[targetIndex]
  const groupSort = group.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateNavGroup({ ...group, sort: targetSort })
  await updateNavGroup({ ...target, sort: groupSort })
  await loadNavigation()
}

async function removeNavGroup(group, draftOnly = false) {
  if (!group.id) return
  if (draftOnly || settingsOpen.value) {
    if (group.items && group.items.length > 0) {
      if (confirm(`分组「${group.name}」内存在卡片，点击“确定”将删除该分组及内部所有卡片；点击“取消”放弃删除。`)) {
        removeNavDraftGroup(group.id)
      }
      return
    }
    if (confirm(`确认删除分组「${group.name}」？`)) {
      removeNavDraftGroup(group.id)
    }
    return
  }
  if (group.items && group.items.length > 0) {
    if (!confirm(`分组「${group.name}」内存在卡片，点击“确定”将删除该分组及内部所有卡片；点击“取消”放弃删除。`)) return
    await Promise.all((group.items || []).filter((item) => typeof item.id === 'number').map((item) => deleteNavItem(item.id)))
  } else if (!confirm(`确认删除分组「${group.name}」？`)) {
    return
  }
  await deleteNavGroup(group.id)
  await loadNavigation()
}

function editNavCard(item, group = null) {
  const groupId = item.groupId ?? group?.id ?? navGroupOptions.value[0]?.id ?? null
  editDialog.value = { open: true, type: 'navItem', title: '编辑导航卡片', form: { ...item, groupId, iconMode: isImageValue(item.icon) ? 'image' : 'text', __originalName: item.name, __originalIcon: item.icon } }
}

async function moveNavCard(group, item, offset) {
  const list = [...(group.items || [])]
  const index = list.findIndex((entry) => entry.id === item.id)
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
  const target = list[targetIndex]
  if (settingsOpen.value) {
    list[targetIndex] = item
    list[index] = target
    updateNavDraftGroup(group.id, (entry) => ({ ...entry, items: list.map((card, cardIndex) => ({ ...card, sort: cardIndex + 1 })) }))
    return
  }
  const itemSort = item.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateNavItem({ ...item, sort: targetSort })
  await updateNavItem({ ...target, sort: itemSort })
  await loadNavigation()
}

async function removeNavCard(item) {
  if (!item.id || !confirm(`确认删除卡片「${item.name}」？`)) return
  if (settingsOpen.value) {
    removeNavDraftItem(item.id)
    return
  }
  await deleteNavItem(item.id)
  await loadNavigation()
}

</script>

<template>
  <main class="shell sun-shell" @click="handleShellClick">
    <section v-if="activeView === 'boot'" class="auth-screen">
      <div class="auth-box"><p class="auth-status">{{ statusText || '正在连接后端...' }}</p></div>
    </section>

    <section v-else-if="activeView === 'login'" class="auth-screen">
      <div class="auth-box"><p v-if="statusText" class="auth-status">{{ statusText }}</p><form class="form-grid" @submit.prevent="submitLogin"><label>账号<input v-model="loginForm.username" /></label><label>密码<input v-model="loginForm.password" type="password" /></label><label class="check-row"><input v-model="loginForm.remember" type="checkbox" /> 记住登录</label><button type="submit">登录</button></form></div>
    </section>

    <section v-else-if="activeView === 'setup'" class="auth-screen">
      <div class="auth-box"><p v-if="statusText" class="auth-status">{{ statusText }}</p><form class="form-grid" @submit.prevent="submitSetup"><h2 style="text-align: center; margin-bottom: 8px; font-size: 20px;">初始化管理员</h2><label>管理员账号<input v-model="setupForm.username" /></label><label>管理员密码<input v-model="setupForm.password" type="password" placeholder="至少 8 位" /></label><label>确认密码<input v-model="setupForm.confirm" type="password" /></label><button type="submit">创建管理员</button></form></div>
    </section>

    <template v-else>
      <FloatingActions :drawer-open="drawerOpen" :show-network-switcher="showNetworkSwitcher" :network-tip="networkTip" :network-icon="networkIcon" :toast-text="toastText" :icon-url="iconUrl" @open-drawer="openDrawer" @cycle-network-mode="cycleNetworkMode" @open-settings="openSettings" />

      <BookmarkDrawer
        :open="drawerOpen"
        :folders="folders"
        :bookmarks="bookmarks"
        :bookmark-search="bookmarkSearch"
        :active-folder="activeFolder"
        :active-folder-id="activeFolderId"
        :folder-count="folderCount"
        :bookmark-count="bookmarkCount"
        :selection-mode="bookmarkSelectionMode"
        :selected-bookmark-ids="selectedBookmarkIds"
        :is-image-value="isImageValue"
        @close-menu="closeMenu"
        @panel-wheel="handleOverlayWheel"
        @create-folder="createFolderByPrompt"
        @create-bookmark="createBookmarkByPrompt(activeFolder)"
        @toggle-selection-mode="bookmarkSelectionMode = true"
        @clear-selection="clearBookmarkSelection"
        @batch-move="openSelectedMoveDialog"
        @batch-delete="deleteSelectedBookmarks"
        @search-input="handleBookmarkSearchInput"
        @toggle-all-folders="toggleAllBookmarkFolders"
        @select-folder="selectFolder"
        @toggle-folder="toggleFolderExpanded"
        @folder-context-menu="showFolderMenu"
        @folder-drag-start="(item, event) => startDrag('folder', item, null, event)"
        @folder-drag-over="hoverFolder"
        @folder-drop="dropFolder"
        @toggle-bookmark-selection="toggleBookmarkSelection"
        @bookmark-context-menu="showBookmarkMenu"
        @bookmark-drag-start="(item, event) => startDrag('bookmark', item, null, event)"
        @bookmark-drag-over="hoverBookmark"
        @bookmark-drop="dropBookmark"
        @open-bookmark="openBookmarkUrl"
      />

      <section v-if="activeView === 'home'" class="home-panel sun-panel">
        <HomeHero :settings-form="settingsForm" :display-time="displayTime" :display-date="displayDate" :date-mode="dateMode" :web-search="webSearch" :active-search-engine="activeSearchEngine" :search-engines="searchEngines" :search-picker-open="searchPickerOpen" :is-image-value="isImageValue" :icon-url="iconUrl" @toggle-date-mode="toggleDateMode" @run-web-search="runWebSearch" @toggle-search-picker="searchPickerOpen = !searchPickerOpen" @select-search-engine="selectSearchEngine" @update-search-query="webSearch.q = $event" />

        <section v-for="group in displayGroups" :key="group.id || group.name" class="nav-group" :class="{ editing: editingNavGroupId === group.id, 'drag-saving': dragState.saving && dragState.groupId === group.id, 'dragging-card': navPointerDrag.active && navPointerDrag.groupId === group.id }" :data-nav-group-id="group.id" :draggable="typeof group.id === 'number'" @dragstart="typeof group.id === 'number' && startDrag('navGroup', group, null, $event)" @dragend="clearDragState" @dragover.prevent @drop="typeof group.id === 'number' && dropNavGroup(group)"><header class="group-head" @contextmenu="showGroupMenu($event, group)"><h2>{{ group.name }}</h2><div class="group-tools"><button type="button" title="新增卡片" @click="addCardFromMenu(group)"><img :src="iconUrl('plus')" alt="" /></button><button type="button" title="编辑分组" @click="toggleNavGroupEdit(group)"><img :src="iconUrl('edit')" alt="" /></button></div></header><TransitionGroup tag="div" class="card-grid" name="nav-card-list"><a v-for="item in group.items" :key="item.id || item.name" class="app-tile" :class="{ dragging: navPointerDrag.active && navPointerDrag.item?.id === item.id }" :data-nav-item-id="item.id" :href="editingNavGroupId === group.id ? '#' : resolveNavUrl(item, networkMode)" :draggable="false" @pointerdown.stop="startNavPointerSort($event, group, item)" @click="handleNavCardClick($event, group, item)" @contextmenu="showCardMenu($event, item, group)"><span class="nav-card"><span v-if="isImageValue(item.icon)" class="card-icon image-icon"><img :src="item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(item.icon || item.name)">{{ limitText(item.icon || item.name, 5) }}</span></span><span class="card-title">{{ limitText(item.name, 10) }}</span></a></TransitionGroup></section>
        <div v-if="!displayGroups.length" class="empty-state nav-empty">暂无导航分组，请在系统设置的分组管理中新增。</div>
      </section>

      <NavDragFloat v-if="navPointerDrag.active" :item="navPointerDrag.item" :drag-style="navDragFloatStyle()" :is-image-value="isImageValue" :card-text-class="cardTextClass" :limit-text="limitText" />

      <SettingsPanel
        :open="settingsOpen"
        :active-settings="activeSettings"
        :menu-collapsed="settingsMenuCollapsed"
        :message="settingsMessage"
        :saving="settingsSaving"
        :settings-draft="settingsDraft"
        :nav-groups-draft="navGroupsDraft"
        :folder-management-flat-list="folderManagementFlatList"
        :settings-search-engines="settingsSearchEngines"
        @close="closeSettings"
        @save="submitSettings"
        @toggle-menu="settingsMenuCollapsed = !settingsMenuCollapsed"
        @select-menu="selectSettingsMenu"
        @panel-wheel="handleOverlayWheel"
        @create-group="createGroupByPrompt"
        @create-card="addCardFromMenu"
        @edit-group="editNavGroup"
        @reorder-group="moveNavGroup"
        @remove-group="removeNavGroup"
        @create-folder="createFolderByPrompt"
        @open-drawer="openDrawer"
        @edit-folder="editFolder"
        @request-move-folder="openFolderMoveDialog"
        @reorder-folder="moveFolder"
        @remove-folder="removeFolder"
        @add-search-engine="addSearchEngine"
        @edit-search-engine="editSearchEngine"
        @reorder-search-engine="moveSearchEngine"
        @remove-search-engine="removeSearchEngine"
        @test-s3="submitTestS3"
        @restore-global="restoreBackupFile"
        @download-nav="downloadNavigationBackup"
        @restore-nav="restoreNavigationBackupFile"
        @export-bookmarks="exportBookmarks"
        @import-bookmarks="importBookmarksFile"
      />
    </template>

    <EditDialog
      v-model:group-select-open="groupSelectOpen"
      :dialog="editDialog"
      :nav-group-options="navGroupOptions"
      :folder-management-flat-list="folderManagementFlatList"
      :folder-flat-list="folderFlatList"
      :eligible-folder-parents="eligibleFolderParents(editDialog.form, getFolderTreeForAction())"
      :metadata-loading="metadataLoading"
      :edit-group-name="getEditGroupName()"
      @close="closeEditDialog"
      @save="saveEditDialog"
      @delete-nav-card="deleteEditingNavCard"
      @clamp-field="clampEditField"
      @select-group="selectEditGroup"
      @set-icon-mode="setNavIconMode"
      @upload-icon="uploadIconFile"
      @fill-metadata="fillMetadata"
      @fill-metadata-from-field="fillMetadataFromField"
      @panel-wheel="handleOverlayWheel"
    />


    <MoveDialog v-model:target-folder-id="moveDialog.targetFolderId" :open="moveDialog.open" :title="moveDialog.title" :items="moveDialog.items" :folder-flat-list="moveDialog.selectableFolders || folderFlatList" :item-label="moveDialog.itemLabel || '收藏'" :allow-root="Boolean(moveDialog.allowRoot)" @close="moveDialog.open = false" @confirm="confirmMoveDialog" />

    <ContextMenu :open="menu.open" :compact="menu.compact" :actions="menu.actions" :menu-style="menuStyle" :icon-url="iconUrl" @run="runMenuAction" />
  </main>
</template>
