<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import BookmarkFolderTreeNode from './components/BookmarkFolderTreeNode.vue'
import BookmarkRow from './components/BookmarkRow.vue'
import ContextMenu from './components/ContextMenu.vue'
import FloatingActions from './components/FloatingActions.vue'
import HomeHero from './components/HomeHero.vue'
import MoveDialog from './components/MoveDialog.vue'
import NavDragFloat from './components/NavDragFloat.vue'
import SettingsMenu from './components/SettingsMenu.vue'
import {
  backupToS3,
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

const activeView = ref('home')
const drawerOpen = ref(false)
const settingsOpen = ref(false)
const activeSettings = ref('个性化')
const settingsMenuCollapsed = ref(false)
const settingsMessage = ref('')
const menu = ref({ open: false, x: 0, y: 0, title: '', actions: [], compact: false })
const statusText = ref('正在连接后端...')
const toastText = ref('')
const user = ref(null)
const initialized = ref(false)
const navGroups = ref([])
const folders = ref([])
const bookmarks = ref([])
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
const editDialog = ref({ open: false, type: '', title: '', form: {} })
const groupSelectOpen = ref(false)
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
const settingsForm = ref({ siteTitle: 'biu-panel', logoUrl: '', showTitle: 'true', showLogo: 'true', showClock: 'true', showSeconds: 'false', showSearch: 'true', searchEngines: JSON.stringify([{ key: 'google', title: 'Google', icon: 'G', url: 'https://www.google.com/search?q={q}' }, { key: 'baidu', title: '百度', icon: '百', url: 'https://www.baidu.com/s?wd={q}' }, { key: 'bing', title: 'Bing', icon: 'B', url: 'https://www.bing.com/search?q={q}' }]), backgroundUrl: '', backgroundColor: '#02030a', lanDetectTimeout: '800', s3Endpoint: '', s3Region: 'auto', s3Bucket: '', s3AccessKey: '', s3SecretKey: '', s3Prefix: 'biu-panel/', s3PathStyle: 'true', s3Enabled: 'false', s3PublicBase: '' })
const settingsDraft = ref({ ...settingsForm.value })
const navGroupsDraft = ref([])

const fallbackGroups = ref([
  { id: 'demo-core', name: '常用服务', items: [{ id: 'demo-nas', name: 'NAS', icon: 'N', wanUrl: '#' }, { id: 'demo-ha', name: 'Home Assistant', icon: 'H', wanUrl: '#' }, { id: 'demo-siyuan', name: '思源笔记', icon: '思', wanUrl: '#' }, { id: 'demo-files', name: '文件管理', icon: '文', wanUrl: '#' }, { id: 'demo-router', name: '路由器', icon: '路', wanUrl: '#' }, { id: 'demo-photo', name: '相册', icon: '相', wanUrl: '#' }, { id: 'demo-download', name: '下载器', icon: '下', wanUrl: '#' }, { id: 'demo-monitor', name: '监控', icon: '监', wanUrl: '#' }, { id: 'demo-note', name: '备忘录', icon: '记', wanUrl: '#' }, { id: 'demo-docs', name: '文档库', icon: '档', wanUrl: '#' }, { id: 'demo-git', name: '代码仓库', icon: 'Git', wanUrl: '#' }, { id: 'demo-media', name: '影音中心', icon: '影', wanUrl: '#' }] },
  { id: 'demo-dev', name: '开发工具', items: [{ id: 'demo-vscode', name: 'VS Code', icon: 'VS', wanUrl: '#' }, { id: 'demo-api', name: 'API 调试', icon: 'API', wanUrl: '#' }, { id: 'demo-db', name: '数据库', icon: 'DB', wanUrl: '#' }, { id: 'demo-ci', name: '构建服务', icon: 'CI', wanUrl: '#' }, { id: 'demo-log', name: '日志平台', icon: 'Log', wanUrl: '#' }, { id: 'demo-wiki', name: '项目 Wiki', icon: 'W', wanUrl: '#' }, { id: 'demo-design', name: '设计稿', icon: '设', wanUrl: '#' }, { id: 'demo-test', name: '测试环境', icon: '测', wanUrl: '#' }] },
  { id: 'demo-life', name: '生活收藏', items: [{ id: 'demo-mail', name: '邮箱', icon: '邮', wanUrl: '#' }, { id: 'demo-calendar', name: '日历', icon: '日', wanUrl: '#' }, { id: 'demo-cloud', name: '网盘', icon: '云', wanUrl: '#' }, { id: 'demo-music', name: '音乐', icon: '音', wanUrl: '#' }, { id: 'demo-read', name: '阅读', icon: '读', wanUrl: '#' }, { id: 'demo-map', name: '地图', icon: '图', wanUrl: '#' }] },
])
const displayGroups = computed(() => (navGroups.value.length ? navGroups.value : fallbackGroups.value))
const navGroupOptions = computed(() => (settingsOpen.value ? navGroupsDraft.value : (navGroups.value.length ? navGroups.value : fallbackGroups.value)))
const menuStyle = computed(() => ({ left: `${menu.value.x}px`, top: `${menu.value.y}px` }))
const activeFolder = computed(() => findFolderById(folders.value, activeFolderId.value))
const bookmarkCount = computed(() => bookmarks.value.length)
const folderFlatList = computed(() => flattenFolders(folders.value))
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

function limitText(value, size) {
  return String(value || '').trim().slice(0, size)
}
function cardTextClass(value) {
  const len = limitText(value, 5).length
  if (len <= 2) return 'text-xl'
  if (len <= 4) return 'text-md'
  return 'text-sm'
}
function clampEditField(field, max) {
  const value = String(editDialog.value.form[field] || '')
  if (value.length > max) editDialog.value.form[field] = value.slice(0, max)
}
const shellStyle = computed(() => ({
  '--runtime-bg': settingsForm.value.backgroundColor || '#02030a',
}))

onMounted(async () => {
  clockTimer = window.setInterval(() => { now.value = new Date() }, 1000)
  networkMode.value = normalizeNetworkMode(localStorage.getItem('biu-network-mode'))
  await refreshBootstrap()
  await loadNavigation()
})

onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
  clearNavLongPressTimer()
  stopNavPointerListeners()
})

function isImageValue(value) {
  return typeof value === 'string' && (value.startsWith('/uploads/') || value.startsWith('http://') || value.startsWith('https://') || value.startsWith('data:image/'))
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

function openMoveDialog(items, title) {
  moveDialog.value = {
    open: true,
    title,
    items,
    targetFolderId: activeFolderId.value || folderFlatList.value[0]?.id || null,
  }
}

async function loadFolderNodeChildren(folder) {
  folder.loading = true
  try {
    const data = await getBookmarkFolders(folder.id)
    const children = (data.folders || []).map((child) => normalizeFolder(child, folder.id))
    folder.children = children
    folder.childrenLoaded = true
    folder.hasChildren = children.length > 0
    folder.expanded = true
  } catch (error) {
    statusText.value = error.message
  } finally {
    folder.loading = false
  }
}

async function toggleFolderExpanded(folder) {
  if (folder.expanded) {
    folder.expanded = false
    return
  }
  if (!folder.childrenLoaded) {
    await loadFolderNodeChildren(folder)
  } else {
    folder.expanded = true
  }
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

async function confirmMoveDialog() {
  await moveBookmarkItems(moveDialog.value.items || [], moveDialog.value.targetFolderId)
}

async function convertBookmarkToNavCard(bookmark) {
  const defaultGroup = navGroups.value[0]
  let group = defaultGroup
  const requested = prompt(`目标分组名称（留空使用「${defaultGroup?.name || '新分组'}」）`, defaultGroup?.name || '')
  const groupName = requested?.trim()
  if (groupName) {
    group = navGroups.value.find((item) => item.name === groupName)
    if (!group) {
      const created = await createNavGroup({ name: groupName, sort: navGroups.value.length + 1 })
      await loadNavigation()
      group = { id: created.id || created, name: groupName }
    }
  } else if (!group) {
    const created = await createNavGroup({ name: '收藏卡片', sort: navGroups.value.length + 1 })
    await loadNavigation()
    group = { id: created.id || created, name: '收藏卡片' }
  }
  await createNavItem({
    groupId: group.id,
    name: bookmark.title,
    icon: bookmark.favicon || bookmark.title.slice(0, 1),
    lanUrl: bookmark.url,
    wanUrl: bookmark.url,
    urlMode: 'auto',
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
  settingsMessage.value = ''
  settingsDraft.value = { ...settingsForm.value }
  navGroupsDraft.value = displayGroups.value.map((group) => ({ ...group, items: [...(group.items || [])] }))
  settingsOpen.value = true
}
function closeSettings() {
  settingsMessage.value = ''
  settingsOpen.value = false
}

function selectSettingsMenu(item) {
  activeSettings.value = item
  if (item === '收藏夹') loadFolders()
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

function updateNavDraftGroup(groupId, updater) {
  navGroupsDraft.value = navGroupsDraft.value.map((group) => group.id === groupId ? updater(group) : group)
}

function removeNavDraftGroup(groupId) {
  navGroupsDraft.value = navGroupsDraft.value.filter((group) => group.id !== groupId)
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
}

function removeNavDraftItem(itemId) {
  navGroupsDraft.value = navGroupsDraft.value.map((group) => ({ ...group, items: (group.items || []).filter((item) => item.id !== itemId) }))
}

function removeFallbackNavItem(itemId) {
  fallbackGroups.value = fallbackGroups.value.map((group) => ({ ...group, items: (group.items || []).filter((item) => item.id !== itemId) }))
}

function upsertFallbackNavItem(item) {
  fallbackGroups.value = fallbackGroups.value.map((group) => {
    const existingItems = [...(group.items || [])]
    const existingIndex = existingItems.findIndex((entry) => entry.id === item.id)
    const items = existingItems.filter((entry) => entry.id !== item.id)
    if (group.id !== item.groupId) return { ...group, items }
    const nextItems = [...items]
    nextItems.splice(existingIndex >= 0 ? existingIndex : nextItems.length, 0, { ...item })
    return { ...group, items: nextItems.map((entry, index) => ({ ...entry, sort: index + 1 })) }
  })
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
    } catch {
      activeView.value = 'login'
      statusText.value = '请登录管理员账号'
    }
  } catch (error) {
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

async function submitBackupToS3() {
  try {
    const data = await backupToS3()
    statusText.value = `S3 备份完成：${data.key}`
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
    if (settingsOpen.value) navGroupsDraft.value = displayGroups.value.map((group) => ({ ...group, items: [...(group.items || [])] }))
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = error.message
  } finally {
    event.target.value = ''
  }
}

async function submitSettings() {
  try {
    const source = settingsOpen.value ? settingsDraft.value : settingsForm.value
    const data = await saveSettings(source)
    if (settingsOpen.value) await saveNavGroupDraftOrder()
    settingsForm.value = { ...settingsForm.value, ...data }
    settingsDraft.value = { ...settingsForm.value }
    if (settingsForm.value.siteTitle) document.title = settingsForm.value.siteTitle
    statusText.value = '设置已保存'
    settingsMessage.value = `设置已保存：${new Date().toLocaleTimeString('zh-CN', { hour12: false })}`
    settingsOpen.value = false
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = `保存失败：${error.message}`
  }
}

async function saveNavGroupDraftOrder() {
  const originalGroups = navGroups.value
  const draftGroups = navGroupsDraft.value
  const originalGroupIds = originalGroups.map((group) => group.id)
  const draftExistingGroupIds = draftGroups.filter((group) => typeof group.id === 'number').map((group) => group.id)
  const createdGroupIds = new Map()
  for (const group of originalGroups) {
    if (!draftExistingGroupIds.includes(group.id)) await deleteNavGroup(group.id)
  }
  for (let groupIndex = 0; groupIndex < draftGroups.length; groupIndex += 1) {
    const group = draftGroups[groupIndex]
    const payload = { name: group.name, sort: groupIndex + 1, collapsed: Boolean(group.collapsed) }
    if (typeof group.id === 'number' && originalGroupIds.includes(group.id)) {
      await updateNavGroup({ id: group.id, ...payload })
      createdGroupIds.set(group.id, group.id)
    } else {
      const created = await createNavGroup(payload)
      createdGroupIds.set(group.id, created.id || created)
    }
  }
  const originalItemIds = originalGroups.flatMap((group) => group.items || []).map((item) => item.id)
  const draftExistingItemIds = draftGroups.flatMap((group) => group.items || []).filter((item) => typeof item.id === 'number').map((item) => item.id)
  for (const id of originalItemIds) {
    if (!draftExistingItemIds.includes(id)) await deleteNavItem(id)
  }
  for (const group of draftGroups) {
    const groupId = createdGroupIds.get(group.id)
    for (let itemIndex = 0; itemIndex < (group.items || []).length; itemIndex += 1) {
      const item = group.items[itemIndex]
      const payload = { groupId, name: item.name, icon: item.icon || '', lanUrl: item.lanUrl || '', wanUrl: item.wanUrl || '', urlMode: item.urlMode || 'auto', sort: itemIndex + 1 }
      if (typeof item.id === 'number' && originalItemIds.includes(item.id)) await updateNavItem({ id: item.id, ...payload })
      else await createNavItem(payload)
    }
  }
  await loadNavigation()
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
  activeFolderId.value = folder.id
  bookmarkSearch.value.q = ''
  bookmarkSearch.value.results = []
  clearBookmarkSelection()
  const data = await getBookmarks(folder.id)
  bookmarks.value = data.items || []
}

async function runBookmarkSearch() {
  const q = bookmarkSearch.value.q.trim()
  if (!q) {
    bookmarkSearch.value.results = []
    return
  }
  bookmarkSearch.value.loading = true
  try {
    const data = await searchBookmarks(q)
    bookmarkSearch.value.results = data.items || []
  } catch (error) {
    statusText.value = error.message
  } finally {
    bookmarkSearch.value.loading = false
  }
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
  metadataLoading.value = true
  try {
    const data = await fetchMetadata(url.trim())
    if ('title' in target && data.title && !target.title) target.title = data.title
    if ('name' in target && data.title && !target.name) target.name = data.title
    if ('favicon' in target && data.favicon) target.favicon = data.favicon
    if ('icon' in target && data.favicon) target.icon = data.favicon
    statusText.value = '已自动抓取标题和图标'
  } catch (error) {
    statusText.value = error.message
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

function setNavIconMode(form, mode) {
  form.iconMode = mode
  if (mode === 'text' && isImageValue(form.icon)) form.icon = ''
  if (mode === 'image' && !isImageValue(form.icon)) form.icon = ''
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
  await createNavItem({ groupId: group.id, name: quickNav.value.cardName.trim(), icon: quickNav.value.cardName.trim().slice(0, 1), lanUrl: quickNav.value.url.trim(), wanUrl: quickNav.value.url.trim(), urlMode: 'auto', sort: (group.items || []).length + 1 })
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
  drawerOpen.value = false
  if (!editingNavGroupId.value) return
  if (event.target.closest?.('.nav-group.editing')) return
  editingNavGroupId.value = null
}

function startDrag(type, item, groupId = null, event = null) {
  dragState.value = { type, item, groupId, overId: null, saving: false, lastMoveAt: 0, settling: false }
  if (event?.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
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
  if (dragState.value.type !== 'folder' || source.id === target.id) return
  await swapSort(source, target, updateBookmarkFolder, () => loadFolders(source.parentId))
  clearDragState()
}

async function dropBookmark(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'bookmark' || !source || source.id === target.id) return
  await swapSort(source, target, updateBookmark, () => selectFolder(activeFolder.value))
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
    return
  }
  const list = [...navGroups.value]
  const index = list.findIndex((item) => item.id === group.id)
  const targetIndex = index + offset
  if (!navGroups.value.length) {
    const demoIndex = fallbackGroups.value.findIndex((item) => item.id === group.id)
    const demoTargetIndex = demoIndex + offset
    if (demoIndex < 0 || demoTargetIndex < 0 || demoTargetIndex >= fallbackGroups.value.length) return
    const next = [...fallbackGroups.value]
    const target = next[demoTargetIndex]
    next[demoTargetIndex] = group
    next[demoIndex] = target
    fallbackGroups.value = next
    return
  }
  if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
  const target = list[targetIndex]
  const groupSort = group.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateNavGroup({ ...group, sort: targetSort })
  await updateNavGroup({ ...target, sort: groupSort })
  await loadNavigation()
}

async function removeNavGroup(group, draftOnly = false) {
  if (!group.id || !confirm(`确认删除分组「${group.name}」？`)) return
  if (draftOnly || settingsOpen.value) {
    removeNavDraftGroup(group.id)
    return
  }
  if (!navGroups.value.length) {
    fallbackGroups.value = fallbackGroups.value.filter((item) => item.id !== group.id)
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
  if (!navGroups.value.length) {
    removeFallbackNavItem(item.id)
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
    if (!navGroups.value.length) {
      removeFallbackNavItem(item.id)
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

function editFolder(folder) {
  editDialog.value = { open: true, type: 'folder', title: '编辑收藏夹文件夹', form: { ...folder } }
}

async function moveFolder(folder, offset) {
  const found = findFolderContainerAndIndex(folders.value, folder.id)
  if (!found) return
  const { siblings, index } = found
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= siblings.length) return
  const target = siblings[targetIndex]
  const folderSort = folder.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateBookmarkFolder({ ...folder, sort: targetSort })
  await updateBookmarkFolder({ ...target, sort: folderSort })
  await loadFolders(folder.parentId ?? null)
}

async function removeFolder(folder) {
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
  editDialog.value = { open: true, type: 'bookmark', title: '编辑网址收藏', form: { ...bookmark } }
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


function closeEditDialog() {
  groupSelectOpen.value = false
  editDialog.value = { open: false, type: '', title: '', form: {} }
}

function getEditGroupName() {
  return navGroupOptions.value.find((group) => group.id === editDialog.value.form.groupId)?.name || '请选择分组'
}

function selectEditGroup(group) {
  editDialog.value.form.groupId = group.id
  groupSelectOpen.value = false
  window.setTimeout(() => { groupSelectOpen.value = false }, 0)
}

async function saveEditDialog() {
  const { type, form } = editDialog.value
  try {
    if (type === 'navGroup' || type === 'navGroupCreate') {
      if (!form.name?.trim()) return
      if (settingsOpen.value) {
        if (type === 'navGroupCreate') navGroupsDraft.value = [...navGroupsDraft.value, { id: createDraftId('draft-group'), name: form.name.trim(), sort: navGroupsDraft.value.length + 1, collapsed: false, items: [] }]
        else updateNavDraftGroup(form.id, (group) => ({ ...group, name: form.name.trim() }))
        closeEditDialog()
        return
      }
      if (!navGroups.value.length) {
        if (type === 'navGroupCreate') fallbackGroups.value = [...fallbackGroups.value, { id: createDraftId('demo-group'), name: form.name.trim(), sort: fallbackGroups.value.length + 1, collapsed: false, items: [] }]
        else fallbackGroups.value = fallbackGroups.value.map((group) => group.id === form.id ? { ...group, name: form.name.trim() } : group)
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
      const payload = { groupId: form.groupId, name, icon, lanUrl: form.lanUrl || '', wanUrl: form.wanUrl || '', urlMode: 'auto', sort: form.sort || 0 }
      if (settingsOpen.value) {
        upsertNavDraftItem({ id: form.id || createDraftId('draft-card'), ...payload })
        closeEditDialog()
        return
      }
      if (!navGroups.value.length) {
        upsertFallbackNavItem({ id: form.id || createDraftId('demo-card'), ...payload })
        closeEditDialog()
        return
      }
      if (type === 'navItemCreate') await createNavItem(payload)
      else await updateNavItem({ id: form.id, ...payload })
      await loadNavigation()
    }
    if (type === 'folder') {
      if (!form.name?.trim()) return
      await updateBookmarkFolder({ ...form, name: form.name.trim() })
      await loadFolders(form.parentId)
    }
    if (type === 'bookmark') {
      if (!form.title?.trim() || !form.url?.trim()) return
      await updateBookmark({ ...form, title: form.title.trim(), url: form.url.trim(), note: form.note || '', favicon: form.favicon || '' })
      if (activeFolder.value) await selectFolder(activeFolder.value)
      if (bookmarkSearch.value.q.trim()) await runBookmarkSearch()
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
  const menuWidth = options.compact ? 128 : 220
  const x = Math.min(point.clientX, window.innerWidth - menuWidth - 8)
  const y = Math.min(point.clientY, window.innerHeight - 360)
  menu.value = { open: true, x: Math.max(8, x), y: Math.max(8, y), title, actions, compact: Boolean(options.compact) }
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
    await navigator.clipboard.writeText(value)
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
    form: { groupId: group.id, name: '', iconMode: 'text', icon: '', lanUrl: '', wanUrl: '', urlMode: 'auto', sort: (group.items?.length || 0) + 1 },
  }
}
async function createFolderByPrompt(parent = null) {
  const name = prompt(parent ? `子目录名称 · ${parent.name}` : '文件夹名称')
  if (!name?.trim()) return
  await createBookmarkFolder({ parentId: parent?.id || null, name: name.trim(), sort: parent ? (parent.children?.length || 0) + 1 : folderFlatList.value.length + 1 })
  await loadFolders(parent?.id ?? null)
}
async function createBookmarkByPrompt(folder = activeFolder.value || folders.value[0]) {
  let targetFolder = folder
  if (!targetFolder) {
    const name = prompt('还没有收藏文件夹，请先创建一个文件夹', '默认收藏')
    if (!name?.trim()) return
    const created = await createBookmarkFolder({ name: name.trim(), sort: folderFlatList.value.length + 1 })
    await loadFolders()
    targetFolder = created
    if (!targetFolder?.id) {
      targetFolder = findFolderById(folders.value, created.id) || folders.value[0]
    }
  }
  if (!targetFolder) {
    statusText.value = '请先创建收藏夹文件夹'
    return
  }
  const title = prompt('收藏标题')
  if (!title?.trim()) return
  const url = prompt('收藏网址')
  if (!url?.trim()) return
  const note = prompt('备注，可不填') || ''
  await createBookmark({ folderId: targetFolder.id, title: title.trim(), url: url.trim(), note: note.trim(), favicon: title.trim().slice(0, 1), sort: bookmarks.value.length + 1 })
  await selectFolder(targetFolder)
}
function showCardMenu(event, item, group = null) {
  if (group && editingNavGroupId.value === group.id) {
    event.preventDefault()
    closeMenu()
    return
  }
  showMenu(event, item.name, [
    { label: '新标签页打开', icon: 'external-link-alt', variant: 'icon', run: () => openNavItemFromMenu(item, '_blank', 'noopener,noreferrer') },
    { label: '新窗口打开', icon: 'window', variant: 'icon', run: () => openNavItemFromMenu(item, `biu-nav-window-${item.id || Date.now()}`, 'popup=yes,width=1200,height=800') },
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
function showBookmarkMenu(event, bookmark) {
  showMenu(event, bookmark.title, [
    { label: '打开', icon: 'external-link-alt', run: () => { window.location.href = bookmark.url } },
    { label: '新标签页打开', icon: 'window', variant: 'icon', run: () => window.open(bookmark.url, '_blank', 'noopener,noreferrer') },
    { divider: true },
    { label: '复制链接', icon: 'link', run: () => copyText(bookmark.url) },
    { label: '移动到文件夹', icon: 'folder', run: () => openMoveDialog([bookmark], '移动到文件夹') },
    { label: '设为首页卡片', icon: 'square', run: () => convertBookmarkToNavCard(bookmark) },
    { label: '批量选择', icon: 'check-square', run: () => batchSelectBookmark(bookmark) },
    { divider: true },
    { label: '编辑', icon: 'edit', run: () => editBookmark(bookmark) },
    { label: '删除', icon: 'trash-alt', run: () => removeBookmark(bookmark) },
  ])
}
</script>

<template>
  <main class="shell sun-shell" @click="handleShellClick">
    <section v-if="activeView === 'login'" class="auth-screen">
      <div class="auth-box"><div class="logo big"><img v-if="settingsForm.logoUrl" :src="settingsForm.logoUrl" alt="Logo" /><span v-else>B</span></div><span class="eyebrow dark">biu-panel</span><h1>欢迎回来</h1><p>{{ statusText }}</p><form class="form-grid" @submit.prevent="submitLogin"><label>账号<input v-model="loginForm.username" /></label><label>密码<input v-model="loginForm.password" type="password" /></label><label class="check-row"><input v-model="loginForm.remember" type="checkbox" /> 记住登录</label><button type="submit">登录</button></form></div>
    </section>

    <section v-else-if="activeView === 'setup'" class="auth-screen">
      <div class="auth-box"><div class="logo big">B</div><span class="eyebrow dark">First run</span><h1>初始化管理员</h1><p>{{ statusText }}</p><form class="form-grid" @submit.prevent="submitSetup"><label>管理员账号<input v-model="setupForm.username" /></label><label>管理员密码<input v-model="setupForm.password" type="password" placeholder="至少 8 位" /></label><label>确认密码<input v-model="setupForm.confirm" type="password" /></label><button type="submit">创建管理员</button></form></div>
    </section>

    <template v-else>
      <FloatingActions :show-network-switcher="showNetworkSwitcher" :network-tip="networkTip" :network-icon="networkIcon" :toast-text="toastText" :icon-url="iconUrl" @open-drawer="openDrawer" @cycle-network-mode="cycleNetworkMode" @open-settings="openSettings" />

      <aside v-if="drawerOpen" class="bookmark-drawer" aria-label="收藏夹" @click.stop>
        <div class="drawer-head">
          <div>
            <span>收藏夹</span>
            <small>{{ folderCount }} 个文件夹 · {{ bookmarkCount }} 条收藏</small>
          </div>
          <div class="inline-actions">
            <button type="button" @click="createBookmarkByPrompt()">新增收藏</button>
            <button type="button" @click="exportBookmarks">导出</button>
            <label class="file-button">导入<input type="file" accept=".html,.htm,text/html" @change="importBookmarksFile" /></label>
          </div>
        </div>

        <div class="bookmark-toolbar">
          <label class="bookmark-search">
            <span>搜索收藏</span>
            <input v-model="bookmarkSearch.q" placeholder="输入标题、网址或备注" @keyup.enter="runBookmarkSearch" />
          </label>
          <div class="inline-actions search-actions">
            <button type="button" @click="runBookmarkSearch">搜索</button>
            <button type="button" @click="clearBookmarkSearch">清空</button>
            <span v-if="bookmarkSearch.loading">搜索中...</span>
            <span v-else-if="bookmarkSearch.results.length">找到 {{ bookmarkSearch.results.length }} 条</span>
          </div>
          <div class="quick-create">
            <input v-model="quickBookmark.folderName" placeholder="新文件夹名称" />
            <button type="button" @click="addFolder">新增文件夹</button>
          </div>
          <div v-if="bookmarkSelectionMode" class="selection-bar">
            <strong>已选择 {{ selectedBookmarkIds.length }} 条收藏</strong>
            <select v-model="moveDialog.targetFolderId">
              <option v-for="folder in folderFlatList" :key="folder.id" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
            </select>
            <div class="inline-actions">
              <button type="button" @click="openSelectedMoveDialog">批量移动</button>
              <button type="button" @click="deleteSelectedBookmarks">批量删除</button>
              <button type="button" @click="clearBookmarkSelection">取消选择</button>
            </div>
          </div>
        </div>

        <section class="bookmark-body">
          <nav class="folder-tree">
            <div class="folder-tree-head">
              <strong>目录树</strong>
              <button type="button" @click="loadFolders()">刷新</button>
            </div>
            <BookmarkFolderTreeNode
              v-for="folder in folders"
              :key="folder.id"
              :folder="folder"
              :active-folder-id="activeFolderId"
              @select="selectFolder"
              @toggle="toggleFolderExpanded"
              @edit="editFolder"
              @remove="removeFolder"
              @move="({ folder, offset }) => moveFolder(folder, offset)"
              @create-child="createFolderByPrompt"
            />
            <div v-if="!folders.length" class="empty-state">暂无文件夹，先创建一个目录。
              <div class="inline-actions empty-actions">
                <button type="button" @click="createFolderByPrompt()">新建文件夹</button>
                <button type="button" @click="createBookmarkByPrompt()">直接新增收藏</button>
              </div>
            </div>
          </nav>

          <div class="bookmark-list">
            <template v-if="bookmarkSearch.q.trim()">
              <BookmarkRow v-for="bookmark in bookmarkSearch.results" :key="`search-${bookmark.id}`" :bookmark="bookmark" :selection-mode="bookmarkSelectionMode" :selected="isBookmarkSelected(bookmark.id)" path-fallback="搜索结果" :is-image-value="isImageValue" @toggle-selection="toggleBookmarkSelection" @context-menu="showBookmarkMenu" />
              <div v-if="!bookmarkSearch.loading && !bookmarkSearch.results.length" class="empty-state">没有匹配的收藏。</div>
            </template>
            <template v-else>
              <div v-if="activeFolderId" class="quick-create bookmark-create">
                <input v-model="quickBookmark.title" placeholder="收藏标题" />
                <input v-model="quickBookmark.url" placeholder="https://example.com" />
                <input v-model="quickBookmark.note" placeholder="备注" />
                <button type="button" @click="fillQuickBookmarkMetadata">{{ metadataLoading ? '抓取中' : '自动抓取' }}</button>
                <button type="button" @click="addBookmark">新增收藏</button>
              </div>
              <BookmarkRow v-for="bookmark in bookmarks" :key="bookmark.id" :bookmark="bookmark" :selection-mode="bookmarkSelectionMode" :selected="isBookmarkSelected(bookmark.id)" draggable show-actions path-fallback="当前文件夹" :is-image-value="isImageValue" @toggle-selection="toggleBookmarkSelection" @context-menu="showBookmarkMenu" @drag-start="(item) => startDrag('bookmark', item)" @drop="dropBookmark" @edit="editBookmark" @move-up="(item) => moveBookmark(item, -1)" @move-down="(item) => moveBookmark(item, 1)" @remove="removeBookmark" />
              <div v-if="activeFolderId && !bookmarks.length" class="empty-state">这个文件夹还没有收藏。</div>
            </template>
          </div>
        </section>
      </aside>

      <section v-if="activeView === 'home'" class="home-panel sun-panel">
        <HomeHero :settings-form="settingsForm" :display-time="displayTime" :display-date="displayDate" :date-mode="dateMode" :web-search="webSearch" :active-search-engine="activeSearchEngine" :search-engines="searchEngines" :search-picker-open="searchPickerOpen" :is-image-value="isImageValue" :icon-url="iconUrl" @toggle-date-mode="toggleDateMode" @run-web-search="runWebSearch" @toggle-search-picker="searchPickerOpen = !searchPickerOpen" @select-search-engine="selectSearchEngine" @update-search-query="webSearch.q = $event" />

        <section v-for="group in displayGroups" :key="group.id || group.name" class="nav-group" :class="{ editing: editingNavGroupId === group.id, 'drag-saving': dragState.saving && dragState.groupId === group.id, 'dragging-card': navPointerDrag.active && navPointerDrag.groupId === group.id }" :data-nav-group-id="group.id" :draggable="typeof group.id === 'number'" @dragstart="typeof group.id === 'number' && startDrag('navGroup', group, null, $event)" @dragend="clearDragState" @dragover.prevent @drop="typeof group.id === 'number' && dropNavGroup(group)"><header class="group-head" @contextmenu="showGroupMenu($event, group)"><h2>{{ group.name }}</h2><div class="group-tools"><button type="button" title="新增卡片" @click="addCardFromMenu(group)"><img :src="iconUrl('plus')" alt="" /></button><button type="button" title="编辑分组" @click="toggleNavGroupEdit(group)"><img :src="iconUrl('edit')" alt="" /></button></div></header><TransitionGroup tag="div" class="card-grid" name="nav-card-list"><a v-for="item in group.items" :key="item.id || item.name" class="app-tile" :class="{ dragging: navPointerDrag.active && navPointerDrag.item?.id === item.id }" :data-nav-item-id="item.id" :href="editingNavGroupId === group.id ? '#' : resolveNavUrl(item)" :draggable="false" @pointerdown.stop="startNavPointerSort($event, group, item)" @click="handleNavCardClick($event, group, item)" @contextmenu="showCardMenu($event, item, group)"><span class="nav-card"><span v-if="isImageValue(item.icon)" class="card-icon image-icon"><img :src="item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(item.icon || item.name)">{{ limitText(item.icon || item.name, 5) }}</span></span><span class="card-title">{{ limitText(item.name, 10) }}</span></a></TransitionGroup></section>
      </section>

      <NavDragFloat v-if="navPointerDrag.active" :item="navPointerDrag.item" :drag-style="navDragFloatStyle()" :is-image-value="isImageValue" :card-text-class="cardTextClass" :limit-text="limitText" />

      <section v-if="settingsOpen" class="settings-mask" @mousedown.self.stop="closeSettings">
        <section class="settings-panel settings-center" @click.stop>
          <header class="settings-head"><div><h2>系统设置</h2></div><div class="inline-actions"><button type="button" @click="submitSettings">保存</button><button type="button" @click="closeSettings">关闭</button></div></header>
          <p v-if="settingsMessage" class="settings-message">{{ settingsMessage }}</p>
          <div class="settings-layout" :class="{ collapsed: settingsMenuCollapsed }">
            <SettingsMenu :collapsed="settingsMenuCollapsed" :active="activeSettings" @toggle-collapse="settingsMenuCollapsed = !settingsMenuCollapsed" @select="selectSettingsMenu" />
            <div class="settings-content">
              <section v-if="activeSettings === '导航管理'" class="setting-card manager-card"><header class="manager-head"><h3>导航管理</h3><button type="button" @click="createGroupByPrompt">新增分组</button></header><article v-for="group in navGroupsDraft" :key="`manage-${group.id}`" class="manager-row"><div><strong>{{ group.name }}</strong><p>{{ group.items?.length || 0 }} 张卡片</p></div><div class="inline-actions"><button type="button" @click="addCardFromMenu(group)">新增卡片</button><button type="button" @click="editNavGroup(group)">编辑</button><button type="button" @click="moveNavGroup(group, -1)">上移</button><button type="button" @click="moveNavGroup(group, 1)">下移</button><button type="button" @click="removeNavGroup(group, true)">删除</button></div></article><div v-if="!navGroupsDraft.length" class="empty-state">暂无导航分组</div></section>
              <section v-if="activeSettings === '收藏夹'" class="setting-card manager-card"><header class="manager-head"><h3>收藏夹管理</h3><div class="inline-actions"><button type="button" @click="createFolderByPrompt()">新增文件夹</button><button type="button" @click="createBookmarkByPrompt()">新增收藏</button><button type="button" @click="openDrawer">打开收藏夹</button></div></header><article v-for="folder in folderFlatList" :key="`manage-folder-${folder.id}`" class="manager-row" :style="{ paddingLeft: `${10 + folder.depth * 14}px` }"><strong>{{ folder.name }}</strong><div class="inline-actions"><button type="button" @click="createFolderByPrompt(folder)">新增子目录</button><button type="button" @click="createBookmarkByPrompt(folder)">新增收藏</button><button type="button" @click="editFolder(folder)">编辑</button><button type="button" @click="removeFolder(folder)">删除</button></div></article><div v-if="!folderFlatList.length" class="empty-state">暂无收藏夹文件夹，点击新增文件夹创建。</div></section>
              <section v-if="activeSettings === '搜索引擎'" class="setting-card manager-card"><header class="manager-head"><h3>搜索引擎</h3><button type="button" @click="addSearchEngine">增加</button></header><article v-for="engine in settingsSearchEngines" :key="`search-manage-${engine.key}`" class="manager-row search-engine-row"><span class="engine-mark"><img v-if="isImageValue(engine.icon)" :src="engine.icon" alt="" /><span v-else>{{ engine.icon || engine.title.slice(0, 1) }}</span></span><div><strong>{{ engine.title }}</strong><p>{{ engine.url }}</p></div><div class="inline-actions"><button type="button" @click="editSearchEngine(engine)">编辑</button><button type="button" @click="moveSearchEngine(engine, -1)">上移</button><button type="button" @click="moveSearchEngine(engine, 1)">下移</button><button type="button" @click="removeSearchEngine(engine)">删除</button></div></article><div v-if="!settingsSearchEngines.length" class="empty-state">暂无搜索引擎，点击增加创建。</div></section>
              <section v-if="activeSettings === '关于'" class="setting-card"><h3>关于</h3><p>这是个人自用导航站和网页收藏夹，当前正在按 Sun-Panel 的交互方式重做。</p></section>
              <div class="settings-grid"><section v-show="activeSettings === '个性化'" class="setting-card settings-card-wide"><h3>个性化</h3><p>{{ statusText }}</p><label class="check-row"><input v-model="settingsDraft.showTitle" true-value="true" false-value="false" type="checkbox" /> 显示标题</label><label>首页文本<input v-model="settingsDraft.siteTitle" /></label><label class="check-row"><input v-model="settingsDraft.showClock" true-value="true" false-value="false" type="checkbox" /> 显示时钟</label><label class="check-row"><input v-model="settingsDraft.showSeconds" true-value="true" false-value="false" type="checkbox" /> 显示秒</label><label class="check-row"><input v-model="settingsDraft.showSearch" true-value="true" false-value="false" type="checkbox" /> 显示搜索栏</label><label>内外网超时时间 ms<input v-model="settingsDraft.lanDetectTimeout" /></label></section><section v-show="activeSettings === 'S3 存储'" class="setting-card settings-card-wide"><h3>S3 存储</h3><label class="check-row"><input v-model="settingsDraft.s3Enabled" true-value="true" false-value="false" type="checkbox" /> 启用 S3 配置</label><label>Endpoint<input v-model="settingsDraft.s3Endpoint" placeholder="https://s3.example.com" /></label><label>Region<input v-model="settingsDraft.s3Region" placeholder="auto" /></label><label>Bucket<input v-model="settingsDraft.s3Bucket" placeholder="biu-panel" /></label><label>Access Key<input v-model="settingsDraft.s3AccessKey" /></label><label>Secret Key<input v-model="settingsDraft.s3SecretKey" type="password" /></label><label>上传前缀<input v-model="settingsDraft.s3Prefix" placeholder="biu-panel/" /></label><label>公开访问域名<input v-model="settingsDraft.s3PublicBase" placeholder="https://cdn.example.com/biu-panel" /></label><label class="check-row"><input v-model="settingsDraft.s3PathStyle" true-value="true" false-value="false" type="checkbox" /> Path-style 兼容模式</label><div class="inline-actions"><button type="button" @click="submitSettings">保存 S3 配置</button><button type="button" @click="submitTestS3">测试 S3</button></div></section><section v-show="activeSettings === '备份恢复'" class="setting-card backup-card settings-card-wide"><h3>备份恢复</h3><div class="backup-zone"><h4>全部备份</h4><p>全局备份包含导航页和收藏夹；恢复时按备份包内容恢复对应数据。</p><div class="inline-actions backup-actions"><button type="button" @click="window.location.href = '/api/backup/download'">全局备份</button><label class="file-button">全局恢复<input type="file" accept=".gz,.tgz,application/gzip" @change="restoreBackupFile" /></label></div></div><div class="backup-zone"><h4>导航页</h4><p>导航页数据包含分组、卡片和排序。</p><div class="inline-actions backup-actions"><button type="button" @click="downloadNavigationBackup">备份</button><label class="file-button">恢复<input type="file" accept=".json,application/json" @change="restoreNavigationBackupFile" /></label></div></div><div class="backup-zone"><h4>收藏夹</h4><p>收藏夹导入导出集中在这里。</p><div class="inline-actions backup-actions"><button type="button" @click="exportBookmarks">导出</button><label class="file-button">导入<input type="file" accept=".html,.htm,text/html" @change="importBookmarksFile" /></label></div></div></section></div>
            </div>
          </div>
        </section>
      </section>
    </template>

    <section v-if="editDialog.open" class="modal-mask" @mousedown.self.stop="closeEditDialog">
      <form class="edit-modal" @click.stop @submit.prevent="saveEditDialog">
        <header class="modal-head"><h2>{{ editDialog.title }}</h2><button type="button" @click="closeEditDialog">关闭</button></header>
        <label v-if="editDialog.type === 'navGroup' || editDialog.type === 'navGroupCreate' || editDialog.type === 'folder'">名称<input v-model="editDialog.form.name" placeholder="请输入分组名称" /></label>
        <template v-if="editDialog.type === 'navItem' || editDialog.type === 'navItemCreate'">
          <label>
            <span class="label-line"><span>标题 <em class="required">*</em></span><small>{{ String(editDialog.form.name || '').length }}/15</small></span>
            <input v-model="editDialog.form.name" maxlength="15" required placeholder="请输入标题" @input="clampEditField('name', 15)" />
          </label>
          <div class="icon-mode-block">
            <span class="icon-mode-title">图标风格</span>
            <div class="segmented">
              <button type="button" :class="{ active: editDialog.form.iconMode !== 'image' }" @click="setNavIconMode(editDialog.form, 'text')">文字</button>
              <button type="button" :class="{ active: editDialog.form.iconMode === 'image' }" @click="setNavIconMode(editDialog.form, 'image')">图片</button>
            </div>
            <label>
              <span class="label-line"><span>{{ editDialog.form.iconMode === 'image' ? '图片地址' : '文本内容' }}</span><small v-if="editDialog.form.iconMode !== 'image'">{{ String(editDialog.form.icon || '').length }}/5</small></span>
              <span class="input-with-button">
                <input v-model="editDialog.form.icon" :maxlength="editDialog.form.iconMode === 'image' ? undefined : 5" :placeholder="editDialog.form.iconMode === 'image' ? '输入图标地址或上传' : '请输入文本内容'" @input="editDialog.form.iconMode !== 'image' && clampEditField('icon', 5)" />
                <label v-if="editDialog.form.iconMode === 'image'" class="upload-inline">上传<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'icon')" /></label>
              </span>
            </label>
          </div>
          <label>
            <span class="label-line"><span>分组 <em class="required">*</em></span></span>
            <div class="select-popover" :class="{ open: groupSelectOpen }">
              <button type="button" class="select-trigger" @click="groupSelectOpen = !groupSelectOpen">
                <span>{{ getEditGroupName() }}</span>
                <span class="select-arrow">⌄</span>
              </button>
              <div v-if="groupSelectOpen" class="select-options">
                <button v-for="group in navGroupOptions" :key="`edit-group-${group.id}`" type="button" :class="{ active: group.id === editDialog.form.groupId }" @click.stop="selectEditGroup(group)">{{ group.name }}</button>
              </div>
            </div>
          </label>
          <label>
            <span class="label-line"><span>公网地址 <em class="required">*</em></span></span>
            <span class="input-with-button metadata-inline">
              <input v-model="editDialog.form.wanUrl" required placeholder="https://example.com" />
              <button class="field-action" type="button" @click="fillMetadataFromField(editDialog.form, 'wanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button>
            </span>
          </label>
          <label>
            <span class="label-line"><span>内网地址</span></span>
            <span class="input-with-button metadata-inline">
              <input v-model="editDialog.form.lanUrl" placeholder="http://192.168.x.x" />
              <button class="field-action" type="button" @click="fillMetadataFromField(editDialog.form, 'lanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button>
            </span>
          </label>
        </template>
        <template v-if="editDialog.type === 'bookmark'"><label>标题<input v-model="editDialog.form.title" /></label><label>网址<input v-model="editDialog.form.url" /></label><label>图标<input v-model="editDialog.form.favicon" /></label><label>上传图标图片<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'favicon')" /></label><label>备注<input v-model="editDialog.form.note" /></label><button type="button" @click="fillMetadata(editDialog.form)">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></template>
        <template v-if="editDialog.type === 'searchEngine' || editDialog.type === 'searchEngineCreate'"><label>标题<input v-model="editDialog.form.title" placeholder="例如 Google" /></label><label>URL<input v-model="editDialog.form.url" placeholder="https://example.com/search?q={q}" /></label><div class="icon-mode-block"><span class="icon-mode-title">图标风格</span><div class="segmented"><button type="button" :class="{ active: editDialog.form.iconMode !== 'image' }" @click="setNavIconMode(editDialog.form, 'text')">文字</button><button type="button" :class="{ active: editDialog.form.iconMode === 'image' }" @click="setNavIconMode(editDialog.form, 'image')">图片</button></div><label><span class="label-line"><span>{{ editDialog.form.iconMode === 'image' ? '图片地址' : '文本内容' }}</span></span><span class="input-with-button"><input v-model="editDialog.form.icon" :placeholder="editDialog.form.iconMode === 'image' ? '输入图标地址或上传' : '请输入文本内容'" /><label v-if="editDialog.form.iconMode === 'image'" class="upload-inline">上传<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'icon')" /></label></span></label></div></template>
        <footer class="modal-actions">
          <button type="submit">保存</button>
          <button v-if="editDialog.type !== 'navItemCreate' && editDialog.type !== 'navGroupCreate' && editDialog.type !== 'searchEngineCreate'" type="button" @click="closeEditDialog">取消</button>
          <button v-if="editDialog.type === 'navItem'" class="danger" type="button" @click="deleteEditingNavCard">删除</button>
        </footer>
      </form>
    </section>


    <MoveDialog v-model:target-folder-id="moveDialog.targetFolderId" :open="moveDialog.open" :title="moveDialog.title" :items="moveDialog.items" :folder-flat-list="folderFlatList" @close="moveDialog.open = false" @confirm="confirmMoveDialog" />

    <ContextMenu :open="menu.open" :compact="menu.compact" :actions="menu.actions" :menu-style="menuStyle" :icon-url="iconUrl" @run="runMenuAction" />
  </main>
</template>
