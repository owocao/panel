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
import { useEditDialog } from './composables/useEditDialog'
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
const settingsOpen = ref(false)
const activeSettings = ref('个性化')
const settingsMenuCollapsed = ref(false)
const settingsMessage = ref('')
const settingsSaving = ref(false)
const menu = ref({ open: false, x: 0, y: 0, title: '', actions: [], compact: false })
const statusText = ref('正在连接后端...')
const toastText = ref('')
const user = ref(null)
const initialized = ref(false)
const navGroups = ref([])
const folders = ref([])
const bookmarks = ref([])
const bookmarkCache = ref({})
const activeFolderId = ref(null)
const bookmarkSelectionMode = ref(false)
const selectedBookmarkIds = ref([])
const moveDialog = ref({ open: false, title: '', items: [], targetFolderId: null })
const loginForm = ref({ username: '', password: '', remember: false })
const setupForm = ref({ username: 'admin', password: '', confirm: '' })
const quickNav = ref({ groupName: '', cardName: '', url: '' })
const quickBookmark = ref({ folderName: '', title: '', url: '', note: '', favicon: '' })
const webSearch = ref({ q: '', engine: 'google' })
const searchPickerOpen = ref(false)
const bookmarkSearch = ref({ q: '', loading: false, results: [] })
const editingNavGroupId = ref(null)
const metadataLoading = ref(false)
const assetUploading = ref(false)
const dragState = ref({ type: '', groupId: null, item: null, overId: null, saving: false, lastMoveAt: 0, settling: false })
const navPointerDrag = ref({ active: false, moved: false, groupId: null, item: null, pointerId: null, startX: 0, startY: 0, x: 0, y: 0, offsetX: 0, offsetY: 0, lastMoveAt: 0, lastTargetId: '' })
const suppressNextNavCardClick = ref(false)
const networkMode = ref('lan')
const now = ref(new Date())
const dateMode = ref('solar')
let clockTimer
let toastTimer
let navLongPressTimer
let draftIdSeed = 0
const settingsForm = ref({ siteTitle: 'biu-panel', showTitle: 'true', showClock: 'true', showSeconds: 'false', showSearch: 'true', searchEngines: JSON.stringify([{ key: 'google', title: 'Google', icon: 'G', url: 'https://www.google.com/search?q={q}' }, { key: 'baidu', title: '百度', icon: '百', url: 'https://www.baidu.com/s?wd={q}' }, { key: 'bing', title: 'Bing', icon: 'B', url: 'https://www.bing.com/search?q={q}' }]), backgroundUrl: '', backgroundColor: '#02030a', lanDetectTimeout: '800', s3Endpoint: '', s3Region: 'auto', s3Bucket: '', s3AccessKey: '', s3SecretKey: '', s3Prefix: 'biu-panel/', s3PathStyle: 'true', s3Enabled: 'false', s3PublicBase: '' })
const settingsDraft = ref({ ...settingsForm.value })
const navGroupsDraft = ref([])
const navDraftDirty = ref(false)
const foldersDraft = ref([])
const foldersDraftDirty = ref(false)

const displayGroups = computed(() => navGroups.value)
const navGroupOptions = computed(() => (settingsOpen.value ? navGroupsDraft.value : navGroups.value))
const menuStyle = computed(() => ({ left: `${menu.value.x}px`, top: `${menu.value.y}px`, width: menu.value.width ? `${menu.value.width}px` : undefined }))
const activeFolder = computed(() => findFolderById(folders.value, activeFolderId.value))
const bookmarkCount = computed(() => bookmarks.value.length)
const folderFlatList = computed(() => flattenFolders(folders.value))
const folderManagementTree = computed(() => (settingsOpen.value && activeSettings.value === '收藏夹' ? foldersDraft.value : folders.value))
const folderManagementFlatList = computed(() => flattenFolders(folderManagementTree.value))
const folderCount = computed(() => folderFlatList.value.length)
const navItemCount = computed(() => navGroups.value.reduce((total, group) => total + (group.items?.length || 0), 0))
const showNetworkSwitcher = computed(() => true)
const networkTip = computed(() => (networkMode.value === 'lan' ? '优先内网，超时后打开公网' : '优先公网，超时后打开内网'))
const networkIcon = computed(() => (networkMode.value === 'lan' ? 'wifi-router' : 'globe'))
const displayTime = computed(() => now.value.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: settingsForm.value.showSeconds === 'true' ? '2-digit' : undefined, hour12: false }))
const lunarDays = ['', '初一', '初二', '初三', '初四', '初五', '初六', '初七', '初八', '初九', '初十', '十一', '十二', '十三', '十四', '十五', '十六', '十七', '十八', '十九', '二十', '廿一', '廿二', '廿三', '廿四', '廿五', '廿六', '廿七', '廿八', '廿九', '三十']
const displayDate = computed(() => {
  const date = now.value
  const weekday = date.toLocaleDateString('zh-CN', { weekday: 'long' })
  if (dateMode.value === 'lunar') {
    const parts = new Intl.DateTimeFormat('zh-CN-u-ca-chinese', { month: 'long', day: 'numeric' }).formatToParts(date)
    const month = parts.find((part) => part.type === 'month')?.value || ''
    const day = Number(parts.find((part) => part.type === 'day')?.value || 0)
    return `${month}${lunarDays[day] || ''}  ${weekday}`
  }
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}-${day}  ${weekday}`
})
function toggleDateMode() {
  dateMode.value = dateMode.value === 'solar' ? 'lunar' : 'solar'
}
const searchEngines = computed(() => {
  try {
    const engines = JSON.parse(settingsForm.value.searchEngines || '[]')
    return Array.isArray(engines) && engines.length ? engines : []
  } catch {
    return []
  }
})
const settingsSearchEngines = computed(() => {
  try {
    const engines = JSON.parse(settingsDraft.value.searchEngines || '[]')
    return Array.isArray(engines) && engines.length ? engines : []
  } catch {
    return []
  }
})
const activeSearchEngine = computed(() => searchEngines.value.find((engine) => engine.key === webSearch.value.engine) || searchEngines.value[0])
const iconUrl = (name) => `https://api.iconify.design/uil/${name}.svg?color=%2368707a`

const {
  editDialog,
  groupSelectOpen,
  closeEditDialog,
  clampEditField,
  getEditGroupName,
  selectEditGroup,
  setNavIconMode,
} = useEditDialog({ navGroupOptions, isImageValue })

function limitText(value, size) {
  return String(value || '').trim().slice(0, size)
}
function cardTextClass(value) {
  const len = limitText(value, 5).length
  if (len <= 2) return 'text-xl'
  if (len <= 4) return 'text-md'
  return 'text-sm'
}

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

function isImageValue(value) {
  return typeof value === 'string' && (value.startsWith('/uploads/') || value.startsWith('http://') || value.startsWith('https://') || value.startsWith('data:image/'))
}

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

function normalizeFolder(folder, parentId = null) {
  return {
    ...folder,
    parentId: folder.parentId ?? parentId ?? null,
    children: Array.isArray(folder.children) ? folder.children : [],
    childrenLoaded: Boolean(folder.childrenLoaded),
    expanded: Boolean(folder.expanded),
    loading: Boolean(folder.loading),
  }
}

function flattenFolders(nodes, depth = 0, out = []) {
  for (const folder of nodes || []) {
    out.push({ ...folder, depth })
    if (folder.childrenLoaded && Array.isArray(folder.children) && folder.children.length) {
      flattenFolders(folder.children, depth + 1, out)
    }
  }
  return out
}

function cloneFolderTree(nodes = []) {
  return (nodes || []).map((folder) => ({
    ...folder,
    children: cloneFolderTree(folder.children || []),
    childrenLoaded: true,
    hasChildren: Boolean(folder.children?.length),
    expanded: Boolean(folder.expanded),
    loading: false,
  }))
}

function syncFoldersDraftFromFolders() {
  foldersDraft.value = cloneFolderTree(folders.value)
  foldersDraftDirty.value = false
}

function findFolderById(nodes, id) {
  if (id == null) return null
  for (const folder of nodes || []) {
    if (folder.id === id) return folder
    const nested = findFolderById(folder.children, id)
    if (nested) return nested
  }
  return null
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

function settingsDraftDirty() {
  const keys = new Set([...Object.keys(settingsForm.value), ...Object.keys(settingsDraft.value)])
  for (const key of keys) {
    if (String(settingsForm.value[key] ?? '') !== String(settingsDraft.value[key] ?? '')) return true
  }
  return false
}

function normalizeParentId(value) {
  return value == null || value === '' ? null : value
}

function folderPayloadChanged(original, payload) {
  if (!original) return true
  return original.name !== payload.name || normalizeParentId(original.parentId) !== normalizeParentId(payload.parentId) || Number(original.sort || 0) !== Number(payload.sort || 0)
}

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

async function moveBookmarkItems(items, folderId) {
  const targetFolderId = Number(folderId) || 0
  if (!targetFolderId) {
    statusText.value = '请选择目标文件夹'
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
    statusText.value = '不能移动到自身'
    return
  }
  const tree = getFolderTreeForAction()
  if (targetParentId && isFolderDescendant(targetParentId, folder.id, tree)) {
    statusText.value = '不能移动到自己的子收藏夹内'
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
    statusText.value = `已移动收藏夹草稿：${folder.name}`
    return
  }
  const siblings = targetParentId ? (findFolderById(folders.value, targetParentId)?.children || []) : folders.value
  await updateBookmarkFolder({ ...folder, parentId: targetParentId, sort: siblings.length + 1 })
  moveDialog.value = { open: false, title: '', items: [], kind: '', itemLabel: '收藏', allowRoot: false, selectableFolders: [], targetFolderId: null }
  await loadFolders()
  await loadAllFolderChildren(folders.value)
  statusText.value = `已移动收藏夹：${folder.name}`
}

async function confirmMoveDialog() {
  if (moveDialog.value.kind === 'folder') {
    await moveFolderToParent(moveDialog.value.items?.[0], moveDialog.value.targetFolderId)
    return
  }
  await moveBookmarkItems(moveDialog.value.items || [], moveDialog.value.targetFolderId)
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

function normalizeNetworkMode(value) {
  return value === 'wan' ? 'wan' : 'lan'
}

function ensureHttp(url) {
  url = String(url || '').trim()
  if (!url || url === '#') return url
  if (!/^https?:\/\//i.test(url) && !url.startsWith('/')) {
    return 'http://' + url
  }
  return url
}

function navUrlCandidates(item) {
  const lanUrl = ensureHttp(item?.lanUrl)
  const wanUrl = ensureHttp(item?.wanUrl)
  if (networkMode.value === 'lan') return { primary: lanUrl, fallback: wanUrl }
  return { primary: wanUrl, fallback: lanUrl }
}

function resolveNavUrl(item) {
  const { primary, fallback } = navUrlCandidates(item)
  return primary || fallback || '#'
}

function openResolvedUrl(url, target = '_self', openedWindow = null) {
  if (!url || url === '#') return
  if (target === '_self') {
    window.location.href = url
    return
  }
  if (openedWindow) {
    openedWindow.location.href = url
    return
  }
  window.open(url, target, 'noopener,noreferrer')
}

function openNavItemFromMenu(item, target = '_blank', features = 'noopener,noreferrer') {
  const url = resolveNavUrl(item)
  if (!url || url === '#') return
  window.open(url, target, features)
}

async function probeUrl(url) {
  if (!url || url === '#') return false
  const timeout = Math.max(200, Number(settingsForm.value.lanDetectTimeout || 800) || 800)
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), timeout)
  try {
    await fetch(url, { mode: 'no-cors', cache: 'no-store', signal: controller.signal })
    return true
  } catch {
    return false
  } finally {
    window.clearTimeout(timer)
  }
}

function openBookmarkUrl(bookmark) {
  const url = ensureHttp(bookmark?.url || '')
  if (!url) return
  window.location.href = url
}

async function openNavItem(item, target = '_self', features = 'noopener,noreferrer') {
  const { primary, fallback } = navUrlCandidates(item)
  const firstUrl = primary || fallback
  if (!firstUrl) return
  let openedWindow = null
  if (target !== '_self') openedWindow = window.open('about:blank', target, features)
  if (!primary || !fallback) {
    openResolvedUrl(firstUrl, target, openedWindow)
    return
  }
  if (await probeUrl(primary)) {
    openResolvedUrl(primary, target, openedWindow)
    return
  }
  openResolvedUrl(fallback, target, openedWindow)
}

function openSettings() {
  closeMenu()
  settingsMessage.value = ''
  settingsSaving.value = false
  activeSettings.value = '个性化'
  settingsDraft.value = { ...settingsForm.value }
  navGroupsDraft.value = navGroups.value.map((group) => ({ ...group, items: [...(group.items || [])] }))
  navDraftDirty.value = false
  foldersDraft.value = []
  foldersDraftDirty.value = false
  settingsOpen.value = true
}
function closeSettings() {
  if (settingsSaving.value) return
  closeMenu()
  settingsMessage.value = ''
  settingsOpen.value = false
  foldersDraft.value = []
  foldersDraftDirty.value = false
}

async function selectSettingsMenu(item) {
  closeMenu()
  activeSettings.value = item
  if (item === '收藏夹') {
    await loadFolders()
    await loadAllFolderChildren(folders.value)
    syncFoldersDraftFromFolders()
  }
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

function cycleNetworkMode() {
  networkMode.value = networkMode.value === 'lan' ? 'wan' : 'lan'
  localStorage.setItem('biu-network-mode', networkMode.value)
  const message = networkMode.value === 'lan' ? '已经切换到优先内网' : '已经切换到优先公网'
  statusText.value = message
  showToast(message)
}

function runWebSearch() {
  const q = webSearch.value.q.trim()
  const engine = activeSearchEngine.value
  if (!q || !engine) return
  const url = (engine.url || '').replace('{q}', encodeURIComponent(q))
  if (url) window.open(url, '_blank')
}

function selectSearchEngine(engine) {
  webSearch.value.engine = engine.key
  searchPickerOpen.value = false
}

function writeSearchEngines(engines) {
  if (settingsOpen.value) settingsDraft.value.searchEngines = JSON.stringify(engines)
  else settingsForm.value.searchEngines = JSON.stringify(engines)
}

function addSearchEngine() {
  editDialog.value = { open: true, type: 'searchEngineCreate', title: '增加搜索引擎', form: { key: `custom-${Date.now()}`, title: '', url: '', icon: '', iconMode: 'text' } }
}

function editSearchEngine(engine) {
  editDialog.value = { open: true, type: 'searchEngine', title: '编辑搜索引擎', form: { ...engine, iconMode: isImageValue(engine.icon) ? 'image' : 'text' } }
}

function removeSearchEngine(engine) {
  if (!confirm(`确认删除搜索引擎「${engine.title}」？`)) return
  const engines = settingsSearchEngines.value.filter((item) => item.key !== engine.key)
  writeSearchEngines(engines)
  if (webSearch.value.engine === engine.key) webSearch.value.engine = engines[0]?.key || ''
}

function moveSearchEngine(engine, offset) {
  const engines = [...settingsSearchEngines.value]
  const index = engines.findIndex((item) => item.key === engine.key)
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= engines.length) return
  const target = engines[targetIndex]
  engines[targetIndex] = engine
  engines[index] = target
  writeSearchEngines(engines)
}

async function uploadSearchEngineIcon(event, engine) {
  const file = event.target.files?.[0]
  if (!file) return
  assetUploading.value = true
  try {
    const result = await uploadAsset(file)
    writeSearchEngines(settingsSearchEngines.value.map((item) => item.key === engine.key ? { ...item, icon: result.url } : item))
  } catch (error) {
    statusText.value = error.message
  } finally {
    assetUploading.value = false
    event.target.value = ''
  }
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


async function loadSettings() {
  try {
    const data = await getSettings()
    settingsForm.value = { ...settingsForm.value, ...data }
    settingsDraft.value = { ...settingsForm.value }
    networkMode.value = normalizeNetworkMode(networkMode.value)
    localStorage.setItem('biu-network-mode', networkMode.value)
    if (settingsForm.value.siteTitle) document.title = settingsForm.value.siteTitle
  } catch {
    // Settings require login; keep defaults for public views.
  }
}




async function submitTestS3() {
  try {
    const data = await testS3()
    statusText.value = `S3 测试成功：${data.key}`
  } catch (error) {
    statusText.value = error.message
  }
}

async function restoreBackupFile(event) {
  const file = event.target.files?.[0]
  if (!file) return
  if (!confirm('恢复备份会覆盖当前数据目录中的同名文件，确认继续？')) {
    event.target.value = ''
    return
  }
  try {
    const data = await restoreBackup(file)
    statusText.value = `恢复完成：${data.files} 个文件，请重启容器后确认数据`
    await loadNavigation()
    if (drawerOpen.value) await loadFolders()
  } catch (error) {
    statusText.value = error.message
  } finally {
    event.target.value = ''
  }
}

async function downloadNavigationBackup() {
  try {
    await downloadFile('/api/navigation/backup')
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = error.message
  }
}

async function restoreNavigationBackupFile(event) {
  const file = event.target.files?.[0]
  if (!file) return
  if (!confirm('恢复导航页备份会替换当前全部导航分组和卡片，确认继续？')) {
    event.target.value = ''
    return
  }
  try {
    const data = await restoreNavigationBackup(file)
    statusText.value = `导航恢复完成：${data.groups} 个分组，${data.items} 张卡片`
    settingsMessage.value = statusText.value
    await loadNavigation()
    if (settingsOpen.value) {
      navGroupsDraft.value = navGroups.value.map((group) => ({ ...group, items: [...(group.items || [])] }))
      navDraftDirty.value = false
    }
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = error.message
  } finally {
    event.target.value = ''
  }
}

async function submitSettings() {
  if (settingsSaving.value) return
  settingsSaving.value = true
  settingsMessage.value = '正在保存，请稍候...'
  try {
    const source = settingsOpen.value ? settingsDraft.value : settingsForm.value
    const shouldSaveSettings = !settingsOpen.value || settingsDraftDirty()
    const data = shouldSaveSettings ? await saveSettings(source) : settingsForm.value
    if (settingsOpen.value && navDraftDirty.value) await saveNavGroupDraftOrder()
    if (settingsOpen.value && foldersDraftDirty.value) await saveFolderDraftOrder()
    settingsForm.value = { ...settingsForm.value, ...data }
    settingsDraft.value = { ...settingsForm.value }
    navDraftDirty.value = false
    foldersDraftDirty.value = false
    if (settingsForm.value.siteTitle) document.title = settingsForm.value.siteTitle
    statusText.value = '设置已保存'
    settingsMessage.value = `设置已保存：${new Date().toLocaleTimeString('zh-CN', { hour12: false })}`
    settingsOpen.value = false
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = `保存失败：${error.message}`
  } finally {
    settingsSaving.value = false
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

async function loadNavigation() {
  try {
    const data = await getNavigation()
    navGroups.value = data.groups || []
  } catch {
    navGroups.value = []
  }
}


async function importBookmarksFile(event) {
  const file = event.target.files?.[0]
  if (!file) return
  try {
    const data = await importBookmarkHTML(file)
    statusText.value = `导入完成：${data.folders} 个文件夹，${data.bookmarks} 条收藏`
    folders.value = []
    activeFolderId.value = null
    bookmarks.value = []
    await loadFolders()
  } catch (error) {
    statusText.value = error.message
  } finally {
    event.target.value = ''
  }
}

function exportBookmarks() {
  window.location.href = '/api/bookmark/export'
}

async function openDrawer() {
  drawerOpen.value = true
  if (!folders.value.length) await loadFolders()
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
    statusText.value = error.message
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
    statusText.value = error.message
  }
}

let bookmarkSearchTimer

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
    statusText.value = error.message
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



async function uploadIconFile(event, target, field = 'icon') {
  const file = event.target.files?.[0]
  if (!file) return
  assetUploading.value = true
  try {
    const data = await uploadAsset(file)
    target[field] = data.url
    statusText.value = '图片已上传到本地数据目录'
  } catch (error) {
    statusText.value = error.message
  } finally {
    assetUploading.value = false
    event.target.value = ''
  }
}

async function fillMetadata(target) {
  const url = target.url || target.wanUrl || target.lanUrl || ''
  if (!url.trim()) {
    statusText.value = '请先填写网址'
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
    statusText.value = '已自动抓取标题和图标'
  } catch (error) {
    statusText.value = error.message || '抓取失败'
    // 弹窗内的提示直接使用浏览器的 alert 进行兜底反馈，确保用户能看到
    alert(statusText.value)
    // 强制显示提示信息以便看到失败反馈
    setTimeout(() => { if (statusText.value === error.message || statusText.value === '抓取失败') statusText.value = '' }, 3000)
  } finally {
    metadataLoading.value = false
  }
}

async function fillMetadataFromField(target, field) {
  const url = target[field] || ''
  if (!url.trim()) {
    statusText.value = '请先填写对应网址'
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

async function addNavGroup() {
  if (!quickNav.value.groupName.trim()) return
  await createNavGroup({ name: quickNav.value.groupName.trim(), sort: navGroups.value.length + 1 })
  quickNav.value.groupName = ''
  await loadNavigation()
}

async function addNavCard() {
  const group = navGroups.value[0]
  if (!group || !quickNav.value.cardName.trim() || !quickNav.value.url.trim()) {
    statusText.value = '请先创建分组，并填写卡片名称和网址'
    return
  }
  await createNavItem({ groupId: group.id, name: quickNav.value.cardName.trim(), icon: quickNav.value.cardName.trim().slice(0, 1), lanUrl: quickNav.value.url.trim(), wanUrl: quickNav.value.url.trim(), urlMode: 'wan', sort: (group.items || []).length + 1 })
  quickNav.value.cardName = ''
  quickNav.value.url = ''
  await loadNavigation()
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

function startDrag(type, item, groupId = null, event = null) {
  dragState.value = { type, item, groupId, overId: null, saving: false, lastMoveAt: 0, settling: false }
  if (type === 'folder' && item?.expanded) {
    item.expanded = false
  }
  if (event?.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', String(item?.id || ''))
    if (type !== 'navItem' && type !== 'navGroup') {
      const canvas = document.createElement('canvas')
      canvas.width = 1
      canvas.height = 1
      canvas.style.position = 'fixed'
      canvas.style.left = '-1000px'
      canvas.style.top = '-1000px'
      document.body.appendChild(canvas)
      event.dataTransfer.setDragImage(canvas, 0, 0)
      window.setTimeout(() => canvas.remove(), 0)
      return
    }
    try {
      const dragNode = event.currentTarget?.cloneNode(true)
      if (dragNode) {
        dragNode.classList.add('drag-preview')
        dragNode.style.position = 'fixed'
        dragNode.style.left = '-1000px'
        dragNode.style.top = '-1000px'
        dragNode.style.pointerEvents = 'none'
        document.body.appendChild(dragNode)
        event.dataTransfer.setDragImage(dragNode, 30, 38)
        window.setTimeout(() => dragNode.remove(), 0)
      }
    } catch {}
  }
}

function reorderListByTarget(list, source, target) {
  const next = [...list]
  const sourceIndex = next.findIndex((item) => item.id === source.id)
  const targetIndex = next.findIndex((item) => item.id === target.id)
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return null
  const [moved] = next.splice(sourceIndex, 1)
  next.splice(targetIndex, 0, moved)
  return next.map((item, index) => ({ ...item, sort: index + 1 }))
}

function hoverBookmark(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'bookmark' || !source || source.id === target.id) return
  const next = reorderListByTarget(bookmarks.value, source, target)
  if (!next) return
  bookmarks.value = next
  dragState.value.item = next.find((item) => item.id === source.id) || source
  dragState.value.overId = target.id
}

function folderSiblings(folder) {
  if (!folder.parentId) return folders.value
  const parent = findFolderById(folders.value, folder.parentId)
  return parent?.children || []
}

function replaceFolderSiblings(folder, next) {
  if (!folder.parentId) {
    folders.value = next
    return
  }
  const parent = findFolderById(folders.value, folder.parentId)
  if (parent) parent.children = next
}

function hoverFolder(target) {
  const source = dragState.value.item
  if (dragState.value.settling || dragState.value.type !== 'folder' || !source || source.id === target.id || source.parentId !== target.parentId) return
  const now = Date.now()
  if (dragState.value.overId === target.id || now - (dragState.value.lastMoveAt || 0) < 220) return
  const next = reorderListByTarget(folderSiblings(target), source, target)
  if (!next) return
  replaceFolderSiblings(target, next)
  dragState.value.item = next.find((item) => item.id === source.id) || source
  dragState.value.overId = target.id
  dragState.value.lastMoveAt = now
  dragState.value.settling = true
  window.setTimeout(() => {
    if (dragState.value.type === 'folder') dragState.value.settling = false
  }, 260)
}

function clearDragState() {
  dragState.value = { type: '', groupId: null, item: null, overId: null, saving: false, lastMoveAt: 0, settling: false }
}

function hoverNavCard(group, target, event = null) {
  const source = dragState.value.item
  if (dragState.value.settling || dragState.value.type !== 'navItem' || dragState.value.groupId !== group.id || !source || source.id === target.id) return
  const list = [...(group.items || [])]
  const sourceIndex = list.findIndex((item) => item.id === source.id)
  const targetIndex = list.findIndex((item) => item.id === target.id)
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return
  const now = Date.now()
  if (now - (dragState.value.lastMoveAt || 0) < 220) return
  const [moved] = list.splice(sourceIndex, 1)
  list.splice(targetIndex, 0, moved)
  group.items = list.map((item, index) => ({ ...item, sort: index + 1 }))
  dragState.value.item = moved
  dragState.value.overId = target.id
  dragState.value.lastMoveAt = now
  dragState.value.settling = true
  window.setTimeout(() => {
    if (dragState.value.type === 'navItem') dragState.value.settling = false
  }, 260)
}

function navDragFloatStyle() {
  if (!navPointerDrag.value.active) return {}
  return {
    left: `${navPointerDrag.value.x}px`,
    top: `${navPointerDrag.value.y}px`,
  }
}

function stopNavPointerListeners() {
  window.removeEventListener('pointermove', handleNavPointerMove)
  window.removeEventListener('pointerup', handleNavPointerUp)
  window.removeEventListener('pointercancel', handleNavPointerCancel)
}

function clearNavLongPressTimer() {
  if (navLongPressTimer) {
    window.clearTimeout(navLongPressTimer)
    navLongPressTimer = null
  }
}

function startNavPointerSort(event, group, item) {
  if (editingNavGroupId.value !== group.id || !item.id) return
  if (event.button != null && event.button !== 0) return
  event.preventDefault()
  closeMenu()
  const rect = event.currentTarget.getBoundingClientRect()
  clearNavLongPressTimer()
  event.currentTarget.setPointerCapture?.(event.pointerId)
  navPointerDrag.value = {
    active: true,
    moved: false,
    groupId: group.id,
    item,
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    x: rect.left,
    y: rect.top,
    offsetX: event.clientX - rect.left,
    offsetY: event.clientY - rect.top,
    lastMoveAt: 0,
    lastTargetId: '',
  }
  window.addEventListener('pointermove', handleNavPointerMove, { passive: false })
  window.addEventListener('pointerup', handleNavPointerUp)
  window.addEventListener('pointercancel', handleNavPointerCancel)
}

function handleNavPointerMove(event) {
  const state = navPointerDrag.value
  if (!state.active || event.pointerId !== state.pointerId) return
  event.preventDefault()
  const dx = event.clientX - state.startX
  const dy = event.clientY - state.startY
  if (!state.moved && Math.hypot(dx, dy) > 4) state.moved = true
  state.x = event.clientX - state.offsetX
  state.y = event.clientY - state.offsetY
  if (!state.moved) return
  const group = displayGroups.value.find((entry) => entry.id === state.groupId)
  if (!group) return
  const groupNode = document.querySelector(`[data-nav-group-id="${CSS.escape(String(state.groupId))}"]`)
  const targetTile = document.elementFromPoint(event.clientX, event.clientY)?.closest?.('[data-nav-item-id]')
  if (!targetTile || !groupNode?.contains(targetTile)) {
    state.lastTargetId = ''
    return
  }
  const targetId = targetTile?.dataset?.navItemId || ''
  if (!targetId || targetId === String(state.item?.id)) return
  if (state.lastTargetId === targetId) return
  const now = Date.now()
  if (now - (state.lastMoveAt || 0) < 110) return
  const list = [...(group.items || [])]
  const sourceIndex = list.findIndex((entry) => String(entry.id) === String(state.item.id))
  const targetIndex = list.findIndex((entry) => String(entry.id) === targetId)
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return
  const [moved] = list.splice(sourceIndex, 1)
  list.splice(targetIndex, 0, moved)
  group.items = list.map((entry, index) => ({ ...entry, sort: index + 1 }))
  state.item = moved
  state.lastTargetId = targetId
  state.lastMoveAt = now
}

function resetNavPointerDrag() {
  navPointerDrag.value = { active: false, moved: false, groupId: null, item: null, pointerId: null, startX: 0, startY: 0, x: 0, y: 0, offsetX: 0, offsetY: 0, lastMoveAt: 0, lastTargetId: '' }
}

function handleNavPointerCancel() {
  clearNavLongPressTimer()
  stopNavPointerListeners()
  resetNavPointerDrag()
}

async function handleNavPointerUp(event) {
  const state = navPointerDrag.value
  if (!state.active || event.pointerId !== state.pointerId) return
  stopNavPointerListeners()
  const group = displayGroups.value.find((entry) => entry.id === state.groupId)
  const moved = state.moved
  const item = state.item
  resetNavPointerDrag()
  if (!group || !item) return
  if (!moved) {
    suppressNextNavCardClick.value = true
    window.setTimeout(() => { suppressNextNavCardClick.value = false }, 0)
    editNavCard(item)
    return
  }
  suppressNextNavCardClick.value = true
  window.setTimeout(() => { suppressNextNavCardClick.value = false }, 0)
  const reordered = [...(group.items || [])].map((entry, index) => ({ ...entry, sort: index + 1 }))
  group.items = reordered
  if (typeof group.id === 'number') {
    try {
      await Promise.all(reordered.map((entry) => updateNavItem(entry)))
      statusText.value = '卡片排序已保存'
    } catch (error) {
      statusText.value = `排序保存失败：${error.message}`
      await loadNavigation()
    }
  }
}


async function dropNavGroup(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'navGroup' || !source || source.id === target.id) return
  await swapSort(source, target, updateNavGroup, loadNavigation)
  clearDragState()
}

async function dropNavCard(group) {
  if (dragState.value.type !== 'navItem' || dragState.value.groupId !== group.id) {
    clearDragState()
    return
  }
  const reordered = [...(group.items || [])].map((item, index) => ({ ...item, sort: index + 1 }))
  group.items = reordered
  dragState.value = { ...dragState.value, overId: null, saving: true }
  try {
    await Promise.all(reordered.map((item) => updateNavItem(item)))
    statusText.value = '卡片排序已保存'
  } catch (error) {
    statusText.value = `排序保存失败：${error.message}`
    await loadNavigation()
  } finally {
    clearDragState()
  }
}

async function dropFolder(target) {
  const source = dragState.value.item
  if (!source) return
  if (dragState.value.type === 'bookmark') {
    await updateBookmark({ ...source, folderId: target.id })
    if (activeFolder.value) await selectFolder(activeFolder.value)
    statusText.value = `已移动到「${target.name}」`
    clearDragState()
    return
  }
  if (dragState.value.type !== 'folder') return
  const siblings = folderSiblings(source)
  const reordered = siblings.map((folder, index) => ({ ...folder, sort: index + 1 }))
  clearDragState()
  try {
    await Promise.all(reordered.map((folder) => updateBookmarkFolder(folder)))
    statusText.value = '收藏夹排序已保存'
  } catch (error) {
    statusText.value = `收藏夹排序保存失败：${error.message}`
    await loadFolders(source.parentId)
  }
}

async function dropBookmark(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'bookmark' || !source) return
  await Promise.all(bookmarks.value.map((bookmark, index) => updateBookmark({ ...bookmark, sort: index + 1 })))
  if (activeFolder.value) await selectFolder(activeFolder.value)
  clearDragState()
}

async function swapSort(source, target, updater, refresh) {
  const sourceSort = source.sort || 1
  const targetSort = target.sort || 1
  await updater({ ...source, sort: targetSort })
  await updater({ ...target, sort: sourceSort })
  await refresh()
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
    statusText.value = error.message
  }
}

async function addFolder() {
  if (!quickBookmark.value.folderName.trim()) return
  await createBookmarkFolder({ name: quickBookmark.value.folderName.trim(), sort: folderFlatList.value.length + 1 })
  quickBookmark.value.folderName = ''
  await loadFolders()
}

async function addBookmark() {
  if (!activeFolderId.value || !quickBookmark.value.title.trim() || !quickBookmark.value.url.trim()) {
    statusText.value = '请选择文件夹，并填写标题和网址'
    return
  }
  await createBookmark({ folderId: activeFolderId.value, title: quickBookmark.value.title.trim(), url: quickBookmark.value.url.trim(), note: quickBookmark.value.note.trim(), favicon: quickBookmark.value.favicon || '', sort: bookmarks.value.length + 1 })
  quickBookmark.value.title = ''
  quickBookmark.value.url = ''
  quickBookmark.value.note = ''
  quickBookmark.value.favicon = ''
  await selectFolder(activeFolder.value)
}

async function editFolder(folder) {
  if (!isSettingsBookmarkManager()) {
    if (!folders.value.length) await loadFolders()
    await loadAllFolderChildren(folders.value)
  }
  editDialog.value = { open: true, type: 'folder', title: '编辑收藏夹', form: { ...folder } }
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
    statusText.value = `排序保存失败：${error.message}`
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
    statusText.value = `已删除收藏夹草稿：${folder.name}`
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
  if (!confirm('删除后无法恢复，确认永久删除？')) return
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
    statusText.value = '请先选择要删除的收藏'
    return
  }
  if (!confirm(`确认删除选中的 ${items.length} 条收藏？`)) return
  if (!confirm('删除后无法恢复，确认永久删除？')) return
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
    statusText.value = '请先选择要移动的收藏'
    return
  }
  openMoveDialog(items, '批量移动收藏')
}

async function batchSelectBookmark(bookmark) {
  enableBookmarkSelection(bookmark)
  statusText.value = `已加入批量选择：${bookmark.title}`
}


async function saveEditDialog() {
  const { type, form } = editDialog.value
  try {
    if (type === 'navGroup' || type === 'navGroupCreate') {
      if (!form.name?.trim()) return
      if (form.name.trim().length > 10) {
        statusText.value = '分组名称最多 10 个字'
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
        statusText.value = '请填写标题'
        return
      }
      if (!form.groupId) {
        statusText.value = '请选择分组'
        return
      }
      if (!form.wanUrl?.trim()) {
        statusText.value = '请填写公网地址'
        return
      }
      const name = form.name.trim()
      if (name.length > 15) {
        statusText.value = '标题最多 15 个字'
        return
      }
      const iconMode = form.iconMode || (isImageValue(form.icon) ? 'image' : 'text')
      const icon = iconMode === 'image' ? (form.icon || '') : (form.icon || name)
      if (iconMode === 'text' && icon.length > 5) {
        statusText.value = '文本内容最多 5 个字'
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
        statusText.value = '不能移动到自身'
        return
      }
      const tree = getFolderTreeForAction()
      if (type === 'folder' && parentId && isFolderDescendant(parentId, form.id, tree)) {
        statusText.value = '不能移动到自己的子收藏夹内'
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
        statusText.value = '请选择要新增到的文件夹'
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
        statusText.value = '请填写搜索引擎标题和 URL'
        return
      }
      const next = { key: form.key || `custom-${Date.now()}`, title: form.title.trim(), url: form.url.trim(), icon: form.icon || form.title.trim().slice(0, 1) }
      const engines = type === 'searchEngineCreate' ? [...settingsSearchEngines.value, next] : settingsSearchEngines.value.map((item) => item.key === next.key ? next : item)
      writeSearchEngines(engines)
      if (!webSearch.value.engine) webSearch.value.engine = next.key
    }
    closeEditDialog()
  } catch (error) {
    statusText.value = error.message
  }
}

function showMenu(event, title, actions, options = {}) {
  event.preventDefault()
  const point = event.touches?.[0] || event
  const menuWidth = options.width || (options.compact ? 128 : 220)
  const x = Math.min(point.clientX, window.innerWidth - menuWidth - 8)
  const menuHeight = options.height || (actions.length * 38 + (title ? 36 : 0) + 16)
  const y = Math.min(point.clientY, window.innerHeight - menuHeight - 8)
  menu.value = { open: true, x: Math.max(8, x), y: Math.max(8, y), title, actions, compact: Boolean(options.compact), width: options.width || null }
}
function closeMenu() { menu.value.open = false }
async function runMenuAction(action) {
  closeMenu()
  try {
    if (action?.run) await action.run()
  } catch (error) {
    statusText.value = error.message
  }
}
async function copyText(value) {
  try {
    if (navigator.clipboard?.writeText) await navigator.clipboard.writeText(value)
    else {
      const textarea = document.createElement('textarea')
      textarea.value = value
      textarea.style.position = 'fixed'
      textarea.style.left = '-1000px'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      textarea.remove()
    }
    statusText.value = '链接已复制'
  } catch {
    statusText.value = '复制失败，请手动复制'
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
    statusText.value = '设置页内暂不直接新增书签，请保存后在收藏夹抽屉中新增'
    return
  }
  let targetFolder = folder
  if (!targetFolder) {
    editDialog.value = { open: true, type: 'folderCreate', title: '请先创建收藏夹', form: { parentId: null, name: '默认收藏', sort: folderFlatList.value.length + 1 } }
    return
  }
  if (!targetFolder) {
    statusText.value = '请先创建收藏夹'
    return
  }
  editDialog.value = { open: true, type: 'bookmarkCreate', title: `新增书签 · ${targetFolder.name}`, form: { folderId: targetFolder.id, title: '', url: '', favicon: '', note: '', sort: bookmarks.value.length + 1 } }
}
function showCardMenu(event, item, group = null) {
  if (group && editingNavGroupId.value === group.id) {
    event.preventDefault()
    closeMenu()
    return
  }
  showMenu(event, item.name, [
    { label: '新标签页打开', icon: 'window', variant: 'icon', run: () => openNavItemFromMenu(item, '_blank', 'noopener,noreferrer') },
    { label: '新窗口打开', icon: 'external-link-alt', variant: 'icon', run: () => openNavItemFromMenu(item, `biu-nav-window-${item.id || Date.now()}`, 'popup=yes,width=1200,height=800') },
    { divider: true },
    { label: '编辑', icon: 'edit', run: () => editNavCard(item, group) },
    { label: '删除', icon: 'trash-alt', run: () => removeNavCard(item) },
  ], { compact: true })
}
function showGroupMenu(event, group) {
  if (editingNavGroupId.value === group.id) {
    event.preventDefault()
    closeMenu()
    return
  }
  showMenu(event, group.name, [
    { label: '新增卡片', icon: 'plus', run: () => addCardFromMenu(group) },
    { label: '编辑分组', icon: 'edit', run: () => editNavGroup(group) },
    { label: '删除分组', icon: 'trash-alt', run: () => removeNavGroup(group) },
  ])
}
function showFolderMenu(event, folder) {
  event.preventDefault()
  event.stopPropagation()
  const current = folderFlatList.value.find((item) => item.id === folder.id)
  const canCreateChild = !current || current.depth < 3
  showMenu(event, folder.name, [
    ...(canCreateChild ? [{ label: '新增收藏夹', icon: 'plus', run: () => createFolderByPrompt(folder) }] : []),
    { label: '新增书签', icon: 'plus', run: () => createBookmarkByPrompt(folder) },
    { divider: true },
    { label: '移动', icon: 'folder', run: () => openFolderMoveDialog(folder) },
    { label: '编辑', icon: 'edit', run: () => editFolder(folder) },
    { label: '删除', icon: 'trash-alt', run: () => removeFolder(folder) },
  ], { width: 148 })
}

function showBookmarkMenu(event, bookmark) {
  event.preventDefault()
  event.stopPropagation()
  showMenu(event, bookmark.title, [
    { label: '新标签页打开', icon: 'window', variant: 'icon', run: () => window.open(ensureHttp(bookmark.url), '_blank', 'noopener,noreferrer') },
    { label: '新窗口打开', icon: 'external-link-alt', variant: 'icon', run: () => window.open(ensureHttp(bookmark.url), `biu-bookmark-window-${bookmark.id || Date.now()}`, 'popup=yes,width=1200,height=800') },
    { divider: true },
    { label: '复制链接', icon: 'link', run: () => copyText(bookmark.url) },
    { label: '移动', icon: 'folder', run: () => openMoveDialog([bookmark], '移动') },
    { label: '首页卡片', icon: 'plus', run: () => convertBookmarkToNavCard(bookmark) },
    { divider: true },
    { label: '编辑', icon: 'edit', run: () => editBookmark(bookmark) },
    { label: '删除', icon: 'trash-alt', run: () => removeBookmark(bookmark) },
  ], { width: 148 })
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

        <section v-for="group in displayGroups" :key="group.id || group.name" class="nav-group" :class="{ editing: editingNavGroupId === group.id, 'drag-saving': dragState.saving && dragState.groupId === group.id, 'dragging-card': navPointerDrag.active && navPointerDrag.groupId === group.id }" :data-nav-group-id="group.id" :draggable="typeof group.id === 'number'" @dragstart="typeof group.id === 'number' && startDrag('navGroup', group, null, $event)" @dragend="clearDragState" @dragover.prevent @drop="typeof group.id === 'number' && dropNavGroup(group)"><header class="group-head" @contextmenu="showGroupMenu($event, group)"><h2>{{ group.name }}</h2><div class="group-tools"><button type="button" title="新增卡片" @click="addCardFromMenu(group)"><img :src="iconUrl('plus')" alt="" /></button><button type="button" title="编辑分组" @click="toggleNavGroupEdit(group)"><img :src="iconUrl('edit')" alt="" /></button></div></header><TransitionGroup tag="div" class="card-grid" name="nav-card-list"><a v-for="item in group.items" :key="item.id || item.name" class="app-tile" :class="{ dragging: navPointerDrag.active && navPointerDrag.item?.id === item.id }" :data-nav-item-id="item.id" :href="editingNavGroupId === group.id ? '#' : resolveNavUrl(item)" :draggable="false" @pointerdown.stop="startNavPointerSort($event, group, item)" @click="handleNavCardClick($event, group, item)" @contextmenu="showCardMenu($event, item, group)"><span class="nav-card"><span v-if="isImageValue(item.icon)" class="card-icon image-icon"><img :src="item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(item.icon || item.name)">{{ limitText(item.icon || item.name, 5) }}</span></span><span class="card-title">{{ limitText(item.name, 10) }}</span></a></TransitionGroup></section>
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
