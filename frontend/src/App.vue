<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import BookmarkFolderTreeNode from './components/BookmarkFolderTreeNode.vue'
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
  fetchMetadata,
  getBookmarkFolders,
  getBookmarks,
  getMe,
  getNavigation,
  getSettings,
  importBookmarkHTML,
  login,
  restoreBackup,
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
const bookmarkSearch = ref({ q: '', loading: false, results: [] })
const editDialog = ref({ open: false, type: '', title: '', form: {} })
const editingNavGroupId = ref(null)
const metadataLoading = ref(false)
const assetUploading = ref(false)
const dragState = ref({ type: '', groupId: null, item: null, overId: null, saving: false, lastMoveAt: 0, settling: false })
const navPointerDrag = ref({ active: false, moved: false, groupId: null, item: null, pointerId: null, startX: 0, startY: 0, x: 0, y: 0, offsetX: 0, offsetY: 0, lastMoveAt: 0 })
const suppressNextNavCardClick = ref(false)
const networkMode = ref('auto')
const now = ref(new Date())
const dateMode = ref('solar')
let clockTimer
let toastTimer
const settingsForm = ref({ siteTitle: 'biu-panel', logoUrl: '', showTitle: 'true', showLogo: 'true', showClock: 'true', showSeconds: 'false', showSearch: 'true', searchEngines: JSON.stringify([{ key: 'google', title: 'Google', icon: 'G', url: 'https://www.google.com/search?q={q}' }, { key: 'baidu', title: '百度', icon: '百', url: 'https://www.baidu.com/s?wd={q}' }, { key: 'bing', title: 'Bing', icon: 'B', url: 'https://www.bing.com/search?q={q}' }]), backgroundUrl: '', backgroundColor: '#02030a', lanDetectUrl: 'http://192.168.1.1', lanDetectTimeout: '800', autoDetectLan: 'true', s3Endpoint: '', s3Region: 'auto', s3Bucket: '', s3AccessKey: '', s3SecretKey: '', s3Prefix: 'biu-panel/', s3PathStyle: 'true', s3Enabled: 'false', s3PublicBase: '' })

const fallbackGroups = [{ name: '常用服务', items: [{ name: 'NAS', icon: 'N', wanUrl: '#' }, { name: 'Home Assistant', icon: 'H', wanUrl: '#' }, { name: '思源笔记', icon: 'S', wanUrl: '#' }, { name: '文件管理', icon: 'F', wanUrl: '#' }] }]
const displayGroups = computed(() => (navGroups.value.length ? navGroups.value : fallbackGroups))
const menuStyle = computed(() => ({ left: `${menu.value.x}px`, top: `${menu.value.y}px` }))
const activeFolder = computed(() => findFolderById(folders.value, activeFolderId.value))
const bookmarkCount = computed(() => bookmarks.value.length)
const folderFlatList = computed(() => flattenFolders(folders.value))
const folderCount = computed(() => folderFlatList.value.length)
const navItemCount = computed(() => navGroups.value.reduce((total, group) => total + (group.items?.length || 0), 0))
const showNetworkSwitcher = computed(() => settingsForm.value.autoDetectLan !== 'true')
const networkTip = computed(() => (networkMode.value === 'lan' ? '内网优先' : '外网优先'))
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
  networkMode.value = localStorage.getItem('biu-network-mode') || 'auto'
  await refreshBootstrap()
  await loadNavigation()
})

onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
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

function resolveNavUrl(item) {
  const mode = item.urlMode || networkMode.value
  if (mode === 'lan') return item.lanUrl || item.wanUrl || '#'
  if (mode === 'wan') return item.wanUrl || item.lanUrl || '#'
  if (networkMode.value === 'lan') return item.lanUrl || item.wanUrl || '#'
  if (networkMode.value === 'wan') return item.wanUrl || item.lanUrl || '#'
  return item.lanUrl || item.wanUrl || '#'
}

function openSettings() {
  settingsMessage.value = ''
  settingsOpen.value = true
}
function closeSettings() {
  settingsMessage.value = ''
  settingsOpen.value = false
}

function showToast(message) {
  toastText.value = message
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => { toastText.value = '' }, 1800)
}

function cycleNetworkMode() {
  networkMode.value = networkMode.value === 'lan' ? 'wan' : 'lan'
  localStorage.setItem('biu-network-mode', networkMode.value)
  const message = networkMode.value === 'lan' ? '已经切换到内网环境' : '已经切换到公网环境'
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

function addSearchEngine() {
  const title = prompt('搜索引擎名称')
  if (!title?.trim()) return
  const url = prompt('搜索 URL，用 {q} 代表搜索词')
  if (!url?.trim()) return
  const icon = prompt('图标文字或图片 URL，可不填') || title.trim().slice(0, 1)
  const engines = [...searchEngines.value, { key: `custom-${Date.now()}`, title: title.trim(), icon: icon.trim(), url: url.trim() }]
  settingsForm.value.searchEngines = JSON.stringify(engines)
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
    if (settingsForm.value.autoDetectLan !== 'true' && !['lan', 'wan'].includes(networkMode.value)) {
      networkMode.value = 'lan'
      localStorage.setItem('biu-network-mode', networkMode.value)
    }
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

async function submitSettings() {
  try {
    const data = await saveSettings(settingsForm.value)
    settingsForm.value = { ...settingsForm.value, ...data }
    if (settingsForm.value.siteTitle) document.title = settingsForm.value.siteTitle
    statusText.value = '设置已保存'
    settingsMessage.value = `设置已保存：${new Date().toLocaleTimeString('zh-CN', { hour12: false })}`
  } catch (error) {
    statusText.value = error.message
    settingsMessage.value = `保存失败：${error.message}`
  }
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
  if (editingNavGroupId.value !== group.id) return
  event.preventDefault()
  if (suppressNextNavCardClick.value) {
    suppressNextNavCardClick.value = false
    return
  }
  editNavCard(item)
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

function startNavPointerSort(event, group, item) {
  if (editingNavGroupId.value !== group.id || !item.id) return
  if (event.button != null && event.button !== 0) return
  event.preventDefault()
  closeMenu()
  const rect = event.currentTarget.getBoundingClientRect()
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
  const group = navGroups.value.find((entry) => entry.id === state.groupId)
  if (!group) return
  const targetTile = document.elementFromPoint(event.clientX, event.clientY)?.closest?.('[data-nav-item-id]')
  const targetId = Number(targetTile?.dataset?.navItemId || 0)
  if (!targetId || targetId === state.item?.id) return
  const now = Date.now()
  if (now - (state.lastMoveAt || 0) < 120) return
  const list = [...(group.items || [])]
  const sourceIndex = list.findIndex((entry) => entry.id === state.item.id)
  const targetIndex = list.findIndex((entry) => entry.id === targetId)
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return
  const [moved] = list.splice(sourceIndex, 1)
  list.splice(targetIndex, 0, moved)
  group.items = list.map((entry, index) => ({ ...entry, sort: index + 1 }))
  state.item = moved
  state.lastMoveAt = now
}

function resetNavPointerDrag() {
  navPointerDrag.value = { active: false, moved: false, groupId: null, item: null, pointerId: null, startX: 0, startY: 0, x: 0, y: 0, offsetX: 0, offsetY: 0, lastMoveAt: 0 }
}

function handleNavPointerCancel() {
  stopNavPointerListeners()
  resetNavPointerDrag()
}

async function handleNavPointerUp(event) {
  const state = navPointerDrag.value
  if (!state.active || event.pointerId !== state.pointerId) return
  stopNavPointerListeners()
  const group = navGroups.value.find((entry) => entry.id === state.groupId)
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
  try {
    await Promise.all(reordered.map((entry) => updateNavItem(entry)))
    statusText.value = '卡片排序已保存'
  } catch (error) {
    statusText.value = `排序保存失败：${error.message}`
    await loadNavigation()
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

async function removeNavGroup(group) {
  if (!group.id || !confirm(`确认删除分组「${group.name}」？`)) return
  await deleteNavGroup(group.id)
  await loadNavigation()
}

function editNavCard(item) {
  editDialog.value = { open: true, type: 'navItem', title: '编辑导航卡片', form: { ...item, iconMode: isImageValue(item.icon) ? 'image' : 'text', __originalName: item.name, __originalIcon: item.icon } }
}

async function moveNavCard(group, item, offset) {
  const list = [...(group.items || [])]
  const index = list.findIndex((entry) => entry.id === item.id)
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
  const target = list[targetIndex]
  const itemSort = item.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateNavItem({ ...item, sort: targetSort })
  await updateNavItem({ ...target, sort: itemSort })
  await loadNavigation()
}

async function removeNavCard(item) {
  if (!item.id || !confirm(`确认删除卡片「${item.name}」？`)) return
  await deleteNavItem(item.id)
  await loadNavigation()
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
  editDialog.value = { open: false, type: '', title: '', form: {} }
}

async function saveEditDialog() {
  const { type, form } = editDialog.value
  try {
    if (type === 'navGroup') {
      if (!form.name?.trim()) return
      await updateNavGroup({ ...form, name: form.name.trim() })
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
      if (name.length > 10) {
        statusText.value = '标题最多 10 个字'
        return
      }
      const iconMode = form.iconMode || (isImageValue(form.icon) ? 'image' : 'text')
      const icon = iconMode === 'image' ? (form.icon || '') : (form.icon || name)
      if (iconMode === 'text' && icon.length > 5) {
        statusText.value = '文本内容最多 5 个字'
        return
      }
      const payload = { groupId: form.groupId, name, icon, lanUrl: form.lanUrl || '', wanUrl: form.wanUrl || '', urlMode: form.urlMode || 'auto', sort: form.sort || 0 }
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
  const name = prompt('分组名称')
  if (!name?.trim()) return
  await createNavGroup({ name: name.trim(), sort: navGroups.value.length + 1 })
  await loadNavigation()
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
  const url = resolveNavUrl(item)
  showMenu(event, item.name, [
    { label: '新标签页打开', icon: 'external-link-alt', variant: 'icon', run: () => window.open(url, '_blank', 'noopener,noreferrer') },
    { label: '新窗口打开', icon: 'window', variant: 'icon', run: () => window.open(url, '_blank', 'noopener,noreferrer,width=1200,height=800') },
    { divider: true },
    { label: '编辑', icon: 'edit', run: () => editNavCard(item) },
    { label: '删除', icon: 'trash-alt', run: () => removeNavCard(item) },
  ], { compact: true })
}
function showGroupMenu(event, group) {
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
      <button class="bookmark-tab" type="button" @click.stop="openDrawer">收藏夹</button>
      <div class="floating-actions"><button v-if="showNetworkSwitcher" type="button" :title="networkTip" @click.stop="cycleNetworkMode"><img :src="iconUrl(networkIcon)" alt="" /></button><button type="button" title="设置" @click.stop="openSettings"><img :src="iconUrl('setting')" alt="" /></button></div>
      <div v-if="toastText" class="toast-message">{{ toastText }}</div>

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
              <article v-for="bookmark in bookmarkSearch.results" :key="`search-${bookmark.id}`" class="bookmark-row" @contextmenu="showBookmarkMenu($event, bookmark)">
                <label v-if="bookmarkSelectionMode" class="bookmark-select"><input type="checkbox" :checked="isBookmarkSelected(bookmark.id)" @change="toggleBookmarkSelection(bookmark)" /></label>
                <span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span>
                <div>
                  <h3>{{ bookmark.title }}</h3>
                  <p>{{ bookmark.url }}</p>
                  <small>{{ bookmark.path || '搜索结果' }}</small>
                </div>
                <a class="open-link" :href="bookmark.url">打开</a>
              </article>
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
              <article v-for="bookmark in bookmarks" :key="bookmark.id" class="bookmark-row" draggable="true" @dragstart="startDrag('bookmark', bookmark)" @dragover.prevent @drop="dropBookmark(bookmark)" @contextmenu="showBookmarkMenu($event, bookmark)">
                <label v-if="bookmarkSelectionMode" class="bookmark-select"><input type="checkbox" :checked="isBookmarkSelected(bookmark.id)" @change="toggleBookmarkSelection(bookmark)" /></label>
                <span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span>
                <div>
                  <h3>{{ bookmark.title }}</h3>
                  <p>{{ bookmark.url }}</p>
                  <small>{{ bookmark.path || '当前文件夹' }}</small>
                </div>
                <div class="row-actions">
                  <button type="button" @click="editBookmark(bookmark)">编辑</button>
                  <button type="button" @click="moveBookmark(bookmark, -1)">上移</button>
                  <button type="button" @click="moveBookmark(bookmark, 1)">下移</button>
                  <button type="button" @click="removeBookmark(bookmark)">删除</button>
                </div>
              </article>
              <div v-if="activeFolderId && !bookmarks.length" class="empty-state">这个文件夹还没有收藏。</div>
            </template>
          </div>
        </section>
      </aside>

      <section v-if="activeView === 'home'" class="home-panel sun-panel">
        <section class="hero-card"><div class="topbar"><div class="home-title"><span v-if="settingsForm.showLogo === 'true'" class="home-logo"><img v-if="settingsForm.logoUrl" :src="settingsForm.logoUrl" alt="" /><span v-else>B</span></span><h2><span v-if="settingsForm.showTitle === 'true'" class="title-text">{{ settingsForm.siteTitle || 'biu-panel' }}</span><template v-if="settingsForm.showTitle === 'true' && settingsForm.showClock === 'true'"><em>｜</em></template><time v-if="settingsForm.showClock === 'true'" class="time-stack"><strong>{{ displayTime }}</strong><small class="date-toggle" role="button" tabindex="0" :title="dateMode === 'solar' ? '点击切换农历' : '点击切换公历'" @click.stop="toggleDateMode" @keyup.enter="toggleDateMode">{{ displayDate }}</small></time></h2></div></div><form v-if="settingsForm.showSearch === 'true'" class="sun-search" @submit.prevent="runWebSearch"><button class="engine-button" type="button" @click="webSearch.engine = searchEngines[(searchEngines.findIndex((engine) => engine.key === webSearch.engine) + 1) % searchEngines.length]?.key || webSearch.engine"><img v-if="isImageValue(activeSearchEngine?.icon)" :src="activeSearchEngine.icon" alt="" /><span v-else>{{ activeSearchEngine?.icon || 'G' }}</span></button><input v-model="webSearch.q" placeholder="网页搜索" /><button type="submit" title="搜索"><img class="button-icon" :src="iconUrl('search')" alt="" /></button></form></section>

        <section v-for="group in displayGroups" :key="group.id || group.name" class="nav-group" :class="{ editing: editingNavGroupId === group.id, 'drag-saving': dragState.saving && dragState.groupId === group.id, 'dragging-card': navPointerDrag.active && navPointerDrag.groupId === group.id }" :draggable="!!group.id" @dragstart="group.id && startDrag('navGroup', group, null, $event)" @dragend="clearDragState" @dragover.prevent @drop="group.id && dropNavGroup(group)"><header class="group-head" @contextmenu="showGroupMenu($event, group)"><h2>{{ group.name }}<small v-if="editingNavGroupId === group.id">排序模式</small></h2><div class="group-tools"><button type="button" title="新增卡片" @click="addCardFromMenu(group)"><img :src="iconUrl('plus')" alt="" /></button><button type="button" title="编辑分组" @click="toggleNavGroupEdit(group)"><img :src="iconUrl('edit')" alt="" /></button></div></header><TransitionGroup tag="div" class="card-grid" name="nav-card-list"><a v-for="item in group.items" :key="item.id || item.name" class="app-tile" :class="{ dragging: navPointerDrag.active && navPointerDrag.item?.id === item.id }" :data-nav-item-id="item.id" :href="editingNavGroupId === group.id ? '#' : resolveNavUrl(item)" :draggable="false" @pointerdown.stop="startNavPointerSort($event, group, item)" @click="openCardEditorInGroup($event, group, item)" @contextmenu="showCardMenu($event, item, group)"><span class="nav-card"><span v-if="isImageValue(item.icon)" class="card-icon image-icon"><img :src="item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(item.icon || item.name)">{{ limitText(item.icon || item.name, 5) }}</span></span><span class="card-title">{{ limitText(item.name, 10) }}</span></a></TransitionGroup></section>
      </section>

      <div v-if="navPointerDrag.active && navPointerDrag.item" class="nav-drag-float" :style="navDragFloatStyle()">
        <span class="nav-card"><span v-if="isImageValue(navPointerDrag.item.icon)" class="card-icon image-icon"><img :src="navPointerDrag.item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(navPointerDrag.item.icon || navPointerDrag.item.name)">{{ limitText(navPointerDrag.item.icon || navPointerDrag.item.name, 5) }}</span></span>
        <span class="card-title">{{ limitText(navPointerDrag.item.name, 10) }}</span>
      </div>

      <section v-if="settingsOpen" class="settings-mask" @mousedown.self.stop="closeSettings"><section class="settings-panel settings-center" @click.stop><header class="settings-head"><div><span class="eyebrow dark">设置中心</span><h2>系统设置</h2></div><div class="inline-actions"><button type="button" @click="submitSettings">保存设置</button><button type="button" @click="closeSettings">关闭</button></div></header><p v-if="settingsMessage" class="settings-message">{{ settingsMessage }}</p><div class="settings-layout"><aside class="settings-menu"><button type="button" :class="{ active: activeSettings === '个人信息' }" @click="activeSettings = '个人信息'">个人信息</button><button type="button" :class="{ active: activeSettings === '个性化' }" @click="activeSettings = '个性化'">个性化</button><button type="button" :class="{ active: activeSettings === '导航管理' }" @click="activeSettings = '导航管理'">导航管理</button><button type="button" :class="{ active: activeSettings === '收藏夹' }" @click="activeSettings = '收藏夹'; loadFolders()">收藏夹</button><button type="button" :class="{ active: activeSettings === '导入导出' }" @click="activeSettings = '导入导出'">导入导出</button><button type="button" :class="{ active: activeSettings === '备份恢复' }" @click="activeSettings = '备份恢复'">备份恢复</button><button type="button" :class="{ active: activeSettings === 'S3 存储' }" @click="activeSettings = 'S3 存储'">S3 存储</button><button type="button" :class="{ active: activeSettings === '关于' }" @click="activeSettings = '关于'">关于</button></aside><div class="settings-content"><section v-if="activeSettings === '个人信息'" class="setting-card"><h3>个人信息</h3><p>{{ statusText }}</p></section><section v-if="activeSettings === '导航管理'" class="setting-card manager-card"><header class="manager-head"><h3>导航管理</h3><button type="button" @click="createGroupByPrompt">新增分组</button></header><article v-for="group in navGroups" :key="`manage-${group.id}`" class="manager-row"><strong>{{ group.name }}</strong><div class="inline-actions"><button type="button" @click="addCardFromMenu(group)">新增卡片</button><button type="button" @click="editNavGroup(group)">编辑</button><button type="button" @click="removeNavGroup(group)">删除</button></div></article><div v-if="!navGroups.length" class="empty-state">暂无导航分组</div></section><section v-if="activeSettings === '收藏夹'" class="setting-card manager-card"><header class="manager-head"><h3>收藏夹管理</h3><div class="inline-actions"><button type="button" @click="createFolderByPrompt()">新增文件夹</button><button type="button" @click="createBookmarkByPrompt()">新增收藏</button><button type="button" @click="openDrawer">打开收藏夹</button></div></header><article v-for="folder in folderFlatList" :key="`manage-folder-${folder.id}`" class="manager-row" :style="{ paddingLeft: `${10 + folder.depth * 14}px` }"><strong>{{ folder.name }}</strong><div class="inline-actions"><button type="button" @click="createFolderByPrompt(folder)">新增子目录</button><button type="button" @click="createBookmarkByPrompt(folder)">新增收藏</button><button type="button" @click="editFolder(folder)">编辑</button><button type="button" @click="removeFolder(folder)">删除</button></div></article><div v-if="!folderFlatList.length" class="empty-state">暂无收藏夹文件夹，点击新增文件夹创建。</div></section><section v-if="activeSettings === '导入导出'" class="setting-card"><h3>导入导出</h3><p>收藏夹导入导出现在在左侧收藏夹抽屉顶部；后续会集中到这里。</p></section><section v-if="activeSettings === '关于'" class="setting-card"><h3>关于</h3><p>这是个人自用导航站和网页收藏夹，当前正在按 Sun-Panel 的交互方式重做。</p></section><div class="settings-grid"><section v-show="activeSettings === '个性化'" class="setting-card"><h3>个性化</h3><label class="check-row"><input v-model="settingsForm.showLogo" true-value="true" false-value="false" type="checkbox" /> 显示 Logo</label><label class="check-row"><input v-model="settingsForm.showTitle" true-value="true" false-value="false" type="checkbox" /> 显示标题</label><label>首页文本<input v-model="settingsForm.siteTitle" /></label><label>Logo 图片<input v-model="settingsForm.logoUrl" placeholder="本地上传地址或图片链接" /></label><label>上传 Logo<input type="file" accept="image/*" @change="uploadIconFile($event, settingsForm, 'logoUrl')" /></label><label class="check-row"><input v-model="settingsForm.showClock" true-value="true" false-value="false" type="checkbox" /> 显示时钟</label><label class="check-row"><input v-model="settingsForm.showSeconds" true-value="true" false-value="false" type="checkbox" /> 显示秒</label><label class="check-row"><input v-model="settingsForm.showSearch" true-value="true" false-value="false" type="checkbox" /> 显示搜索栏</label><div class="search-engine-list"><strong>搜索引擎</strong><button type="button" @click="addSearchEngine">添加搜索引擎</button><p v-for="engine in searchEngines" :key="engine.key">{{ engine.icon }} {{ engine.title }} · {{ engine.url }}</p></div></section><section v-show="activeSettings === 'S3 存储'" class="setting-card"><h3>S3 存储</h3><label class="check-row"><input v-model="settingsForm.s3Enabled" true-value="true" false-value="false" type="checkbox" /> 启用 S3 配置</label><label>Endpoint<input v-model="settingsForm.s3Endpoint" placeholder="https://s3.example.com" /></label><label>Region<input v-model="settingsForm.s3Region" placeholder="auto" /></label><label>Bucket<input v-model="settingsForm.s3Bucket" placeholder="biu-panel" /></label><label>Access Key<input v-model="settingsForm.s3AccessKey" /></label><label>Secret Key<input v-model="settingsForm.s3SecretKey" type="password" /></label><label>上传前缀<input v-model="settingsForm.s3Prefix" placeholder="biu-panel/" /></label><label>公开访问域名<input v-model="settingsForm.s3PublicBase" placeholder="https://cdn.example.com/biu-panel" /></label><label class="check-row"><input v-model="settingsForm.s3PathStyle" true-value="true" false-value="false" type="checkbox" /> Path-style 兼容模式</label><div class="inline-actions"><button type="button" @click="submitSettings">保存 S3 配置</button><button type="button" @click="submitTestS3">测试 S3</button></div></section><section v-show="activeSettings === '个性化'" class="setting-card"><h3>内外网判断</h3><label>统一检测地址<input v-model="settingsForm.lanDetectUrl" /></label><label>超时时间 ms<input v-model="settingsForm.lanDetectTimeout" /></label><label class="check-row"><input v-model="settingsForm.autoDetectLan" true-value="true" false-value="false" type="checkbox" /> 启用自动判断</label></section><section v-show="activeSettings === '备份恢复'" class="setting-card"><h3>备份恢复</h3><p>系统备份为 .tar.gz，包含数据库、本地上传文件和版本信息。</p><div class="inline-actions"><button type="button" @click="window.location.href = '/api/backup/download'">下载备份</button><button type="button" @click="submitBackupToS3">备份到 S3</button><label class="file-button">恢复备份<input type="file" accept=".gz,.tgz,application/gzip" @change="restoreBackupFile" /></label></div></section></div></div></div></section></section>
    </template>

    <section v-if="editDialog.open" class="modal-mask" @mousedown.self.stop="closeEditDialog">
      <form class="edit-modal" @click.stop @submit.prevent="saveEditDialog">
        <header class="modal-head"><h2>{{ editDialog.title }}</h2><button type="button" @click="closeEditDialog">关闭</button></header>
        <label v-if="editDialog.type === 'navGroup' || editDialog.type === 'folder'">名称<input v-model="editDialog.form.name" /></label>
        <template v-if="editDialog.type === 'navItem' || editDialog.type === 'navItemCreate'">
          <label>
            <span class="label-line"><span>标题 <em class="required">*</em></span><small>{{ String(editDialog.form.name || '').length }}/10</small></span>
            <input v-model="editDialog.form.name" maxlength="10" required placeholder="请输入标题" @input="clampEditField('name', 10)" />
          </label>
          <label>
            <span class="label-line"><span>分组 <em class="required">*</em></span></span>
            <select v-model="editDialog.form.groupId" required>
              <option v-for="group in navGroups" :key="`edit-group-${group.id}`" :value="group.id">{{ group.name }}</option>
            </select>
          </label>
          <label>
            <span class="label-line"><span>内网地址</span><button class="field-action" type="button" @click="fillMetadataFromField(editDialog.form, 'lanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></span>
            <input v-model="editDialog.form.lanUrl" placeholder="http://192.168.x.x" />
          </label>
          <label>
            <span class="label-line"><span>公网地址 <em class="required">*</em></span><button class="field-action" type="button" @click="fillMetadataFromField(editDialog.form, 'wanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></span>
            <input v-model="editDialog.form.wanUrl" required placeholder="https://example.com" />
          </label>
          <label>打开模式<select v-model="editDialog.form.urlMode"><option value="auto">自动</option><option value="lan">强制内网</option><option value="wan">强制公网</option></select></label>
          <div class="icon-mode-block">
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
        </template>
        <template v-if="editDialog.type === 'bookmark'"><label>标题<input v-model="editDialog.form.title" /></label><label>网址<input v-model="editDialog.form.url" /></label><label>图标<input v-model="editDialog.form.favicon" /></label><label>上传图标图片<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'favicon')" /></label><label>备注<input v-model="editDialog.form.note" /></label><button type="button" @click="fillMetadata(editDialog.form)">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></template>
        <footer class="modal-actions"><button type="button" @click="closeEditDialog">取消</button><button type="submit">保存</button></footer>
      </form>
    </section>


    <section v-if="moveDialog.open" class="modal-mask" @mousedown.self.stop="moveDialog.open = false">
      <div class="edit-modal" @click.stop>
        <header class="modal-head">
          <h2>{{ moveDialog.title }}</h2>
          <button type="button" @click="moveDialog.open = false">关闭</button>
        </header>
        <p class="move-hint">将 {{ moveDialog.items.length }} 条收藏移动到以下文件夹。</p>
        <label>
          目标文件夹
          <select v-model="moveDialog.targetFolderId">
            <option v-for="folder in folderFlatList" :key="`move-folder-${folder.id}`" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
          </select>
        </label>
        <footer class="modal-actions">
          <button type="button" @click="moveDialog.open = false">取消</button>
          <button type="button" @click="confirmMoveDialog">确认移动</button>
        </footer>
      </div>
    </section>

    <div v-if="menu.open" class="context-menu" :class="{ compact: menu.compact }" :style="menuStyle" @click.stop>
      <div v-if="menu.actions.some((action) => action.variant === 'icon')" class="menu-icon-row">
        <button v-for="action in menu.actions.filter((item) => item.variant === 'icon')" :key="action.label" class="icon-only" type="button" :title="action.label" @click="runMenuAction(action)"><img :src="iconUrl(action.icon)" alt="" /><span class="visually-hidden">{{ action.label }}</span></button>
      </div>
      <template v-for="(action, index) in menu.actions" :key="action.label || `divider-${index}`">
        <div v-if="action.divider" class="menu-divider"></div>
        <button v-else-if="action.variant !== 'icon'" type="button" @click="runMenuAction(action)"><img v-if="action.icon" :src="iconUrl(action.icon)" alt="" /><span>{{ action.label }}</span></button>
      </template>
    </div>
  </main>
</template>
