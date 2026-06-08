<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
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
const activeSettings = ref('首页样式')
const menu = ref({ open: false, x: 0, y: 0, title: '', actions: [] })
const statusText = ref('正在连接后端...')
const user = ref(null)
const initialized = ref(false)
const navGroups = ref([])
const folders = ref([])
const bookmarks = ref([])
const activeFolderId = ref(null)
const loginForm = ref({ username: 'admin', password: '', remember: false })
const setupForm = ref({ username: 'admin', password: '', confirm: '' })
const quickNav = ref({ groupName: '', cardName: '', url: '' })
const quickBookmark = ref({ folderName: '', title: '', url: '', note: '', favicon: '' })
const bookmarkSearch = ref({ q: '', loading: false, results: [] })
const editDialog = ref({ open: false, type: '', title: '', form: {} })
const metadataLoading = ref(false)
const assetUploading = ref(false)
const dragState = ref({ type: '', groupId: null, item: null })
const networkMode = ref('auto')
const now = ref(new Date())
let clockTimer
const settingsForm = ref({ siteTitle: 'biu-panel', logoUrl: '', backgroundUrl: '', backgroundColor: '#02030a', lanDetectUrl: 'http://192.168.1.1', lanDetectTimeout: '800', autoDetectLan: 'true', s3Endpoint: '', s3Region: 'auto', s3Bucket: '', s3AccessKey: '', s3SecretKey: '', s3Prefix: 'biu-panel/', s3PathStyle: 'true', s3Enabled: 'false', s3PublicBase: '' })

const fallbackGroups = [{ name: '常用服务', items: [{ name: 'NAS', icon: 'N', wanUrl: '#' }, { name: 'Home Assistant', icon: 'H', wanUrl: '#' }, { name: '思源笔记', icon: 'S', wanUrl: '#' }, { name: '文件管理', icon: 'F', wanUrl: '#' }] }]
const displayGroups = computed(() => (navGroups.value.length ? navGroups.value : fallbackGroups))
const menuStyle = computed(() => ({ left: `${menu.value.x}px`, top: `${menu.value.y}px` }))
const activeFolder = computed(() => folders.value.find((folder) => folder.id === activeFolderId.value))
const networkLabel = computed(() => ({ auto: '自动内网', lan: '内网优先', wan: '外网优先' }[networkMode.value]))
const displayTime = computed(() => now.value.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false }))
const displayDate = computed(() => now.value.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric', weekday: 'long' }).replace('月', '-').replace('日', ''))
const shellStyle = computed(() => ({
  '--runtime-bg': settingsForm.value.backgroundColor || '#02030a',
}))

onMounted(async () => {
  clockTimer = window.setInterval(() => { now.value = new Date() }, 30000)
  networkMode.value = localStorage.getItem('biu-network-mode') || 'auto'
  await refreshBootstrap()
  await loadNavigation()
})

onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
})


function isImageValue(value) {
  return typeof value === 'string' && (value.startsWith('/uploads/') || value.startsWith('http://') || value.startsWith('https://') || value.startsWith('data:image/'))
}

function resolveNavUrl(item) {
  const mode = item.urlMode || networkMode.value
  if (mode === 'lan') return item.lanUrl || item.wanUrl || '#'
  if (mode === 'wan') return item.wanUrl || item.lanUrl || '#'
  if (networkMode.value === 'lan') return item.lanUrl || item.wanUrl || '#'
  if (networkMode.value === 'wan') return item.wanUrl || item.lanUrl || '#'
  return item.lanUrl || item.wanUrl || '#'
}

function cycleNetworkMode() {
  networkMode.value = networkMode.value === 'auto' ? 'lan' : networkMode.value === 'lan' ? 'wan' : 'auto'
  localStorage.setItem('biu-network-mode', networkMode.value)
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
  } catch (error) {
    statusText.value = error.message
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

async function loadFolders(parentId) {
  try {
    const data = await getBookmarkFolders(parentId)
    folders.value = data.folders || []
    if (folders.value.length && !activeFolderId.value) await selectFolder(folders.value[0])
  } catch (error) {
    statusText.value = error.message
  }
}

async function selectFolder(folder) {
  activeFolderId.value = folder.id
  bookmarkSearch.value.q = ''
  bookmarkSearch.value.results = []
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


function startDrag(type, item, groupId = null) {
  dragState.value = { type, item, groupId }
}

async function dropNavGroup(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'navGroup' || !source || source.id === target.id) return
  await swapSort(source, target, updateNavGroup, loadNavigation)
  dragState.value = { type: '', groupId: null, item: null }
}

async function dropNavCard(group, target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'navItem' || dragState.value.groupId !== group.id || !source || source.id === target.id) return
  await swapSort(source, target, updateNavItem, loadNavigation)
  dragState.value = { type: '', groupId: null, item: null }
}

async function dropFolder(target) {
  const source = dragState.value.item
  if (!source) return
  if (dragState.value.type === 'bookmark') {
    await updateBookmark({ ...source, folderId: target.id })
    if (activeFolder.value) await selectFolder(activeFolder.value)
    statusText.value = `已移动到「${target.name}」`
    dragState.value = { type: '', groupId: null, item: null }
    return
  }
  if (dragState.value.type !== 'folder' || source.id === target.id) return
  await swapSort(source, target, updateBookmarkFolder, () => loadFolders(source.parentId))
  dragState.value = { type: '', groupId: null, item: null }
}

async function dropBookmark(target) {
  const source = dragState.value.item
  if (dragState.value.type !== 'bookmark' || !source || source.id === target.id) return
  await swapSort(source, target, updateBookmark, () => selectFolder(activeFolder.value))
  dragState.value = { type: '', groupId: null, item: null }
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
  editDialog.value = { open: true, type: 'navItem', title: '编辑导航卡片', form: { ...item } }
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
  await createBookmarkFolder({ name: quickBookmark.value.folderName.trim(), sort: folders.value.length + 1 })
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
  const list = [...folders.value]
  const index = list.findIndex((item) => item.id === folder.id)
  const targetIndex = index + offset
  if (index < 0 || targetIndex < 0 || targetIndex >= list.length) return
  const target = list[targetIndex]
  const folderSort = folder.sort || index + 1
  const targetSort = target.sort || targetIndex + 1
  await updateBookmarkFolder({ ...folder, sort: targetSort })
  await updateBookmarkFolder({ ...target, sort: folderSort })
  await loadFolders()
}

async function removeFolder(folder) {
  if (!confirm(`确认删除文件夹「${folder.name}」及其内容？`)) return
  await deleteBookmarkFolder(folder.id)
  activeFolderId.value = null
  bookmarks.value = []
  await loadFolders()
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
  await deleteBookmark(bookmark.id)
  await selectFolder(activeFolder.value)
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
    if (type === 'navItem') {
      if (!form.name?.trim()) return
      await updateNavItem({ ...form, name: form.name.trim(), icon: form.icon || form.name.trim().slice(0, 1), lanUrl: form.lanUrl || '', wanUrl: form.wanUrl || '', urlMode: form.urlMode || 'auto' })
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

function showMenu(event, title, actions) {
  event.preventDefault()
  const point = event.touches?.[0] || event
  menu.value = { open: true, x: point.clientX, y: point.clientY, title, actions }
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
async function addCardFromMenu(group) {
  const name = prompt('卡片名称')
  if (!name?.trim()) return
  const url = prompt('卡片网址')
  if (!url?.trim()) return
  await createNavItem({ groupId: group.id, name: name.trim(), icon: name.trim().slice(0, 1), lanUrl: url.trim(), wanUrl: url.trim(), urlMode: 'auto', sort: (group.items?.length || 0) + 1 })
  await loadNavigation()
}
function showCardMenu(event, item) {
  const url = resolveNavUrl(item)
  showMenu(event, item.name, [
    { label: '打开', run: () => { window.location.href = url } },
    { label: '新标签页打开', run: () => window.open(url, '_blank') },
    { label: '编辑卡片', run: () => editNavCard(item) },
    { label: '复制链接', run: () => copyText(url) },
    { label: '删除', run: () => removeNavCard(item) },
  ])
}
function showGroupMenu(event, group) {
  showMenu(event, group.name, [
    { label: '新增卡片', run: () => addCardFromMenu(group) },
    { label: '编辑分组', run: () => editNavGroup(group) },
    { label: '上移分组', run: () => moveNavGroup(group, -1) },
    { label: '下移分组', run: () => moveNavGroup(group, 1) },
    { label: '删除分组', run: () => removeNavGroup(group) },
  ])
}
function showBookmarkMenu(event, bookmark) {
  showMenu(event, bookmark.title, [
    { label: '打开', run: () => { window.location.href = bookmark.url } },
    { label: '新标签页打开', run: () => window.open(bookmark.url, '_blank') },
    { label: '编辑', run: () => editBookmark(bookmark) },
    { label: '复制链接', run: () => copyText(bookmark.url) },
    { label: '删除', run: () => removeBookmark(bookmark) },
  ])
}
</script>

<template>
  <main class="shell sun-shell" @click="closeMenu(); drawerOpen = false">
    <section v-if="activeView === 'login'" class="auth-screen">
      <div class="auth-box"><div class="logo big"><img v-if="settingsForm.logoUrl" :src="settingsForm.logoUrl" alt="Logo" /><span v-else>B</span></div><span class="eyebrow dark">biu-panel</span><h1>欢迎回来</h1><p>{{ statusText }}</p><form class="form-grid" @submit.prevent="submitLogin"><label>账号<input v-model="loginForm.username" /></label><label>密码<input v-model="loginForm.password" type="password" /></label><label class="check-row"><input v-model="loginForm.remember" type="checkbox" /> 记住登录</label><button type="submit">登录</button></form></div>
    </section>

    <section v-else-if="activeView === 'setup'" class="auth-screen">
      <div class="auth-box"><div class="logo big">B</div><span class="eyebrow dark">First run</span><h1>初始化管理员</h1><p>{{ statusText }}</p><form class="form-grid" @submit.prevent="submitSetup"><label>管理员账号<input v-model="setupForm.username" /></label><label>管理员密码<input v-model="setupForm.password" type="password" placeholder="至少 8 位" /></label><label>确认密码<input v-model="setupForm.confirm" type="password" /></label><button type="submit">创建管理员</button></form></div>
    </section>

    <template v-else>
      <button class="bookmark-tab" type="button" @click.stop="openDrawer">收藏夹</button>
      <div class="floating-actions"><button type="button" @click.stop="cycleNetworkMode">{{ networkLabel }}</button><button type="button" @click.stop="settingsOpen = true">设置</button></div>

      <aside v-if="drawerOpen" class="bookmark-drawer" aria-label="收藏夹" @click.stop>
        <div class="drawer-head"><span>收藏夹</span><div class="inline-actions"><button type="button" @click="exportBookmarks">导出</button><label class="file-button">导入<input type="file" accept=".html,.htm,text/html" @change="importBookmarksFile" /></label></div></div>
        <label class="bookmark-search"><span>搜索收藏</span><input v-model="bookmarkSearch.q" placeholder="输入标题、网址或备注" @keyup.enter="runBookmarkSearch" /></label><div class="inline-actions search-actions"><button type="button" @click="runBookmarkSearch">搜索</button><button type="button" @click="clearBookmarkSearch">清空</button><span v-if="bookmarkSearch.loading">搜索中...</span><span v-else-if="bookmarkSearch.results.length">找到 {{ bookmarkSearch.results.length }} 条</span></div>
        <div class="quick-create"><input v-model="quickBookmark.folderName" placeholder="新文件夹名称" /><button type="button" @click="addFolder">新增文件夹</button></div>
        <section class="bookmark-body"><nav class="folder-tree"><button v-for="folder in folders" :key="folder.id" class="folder" :class="{ active: folder.id === activeFolderId }" draggable="true" @dragstart="startDrag('folder', folder)" @dragover.prevent @drop="dropFolder(folder)" type="button" @click="selectFolder(folder)"><strong>{{ folder.name }}</strong><span>{{ folder.hasChildren ? '可展开子目录' : '当前目录' }}</span><span class="mini-actions"><em @click.stop="editFolder(folder)">编辑</em><em @click.stop="moveFolder(folder, -1)">上移</em><em @click.stop="moveFolder(folder, 1)">下移</em><em @click.stop="removeFolder(folder)">删除</em></span></button><div v-if="!folders.length" class="empty-state">暂无文件夹，先创建一个目录。</div></nav>
          <div class="bookmark-list"><template v-if="bookmarkSearch.q.trim()"><article v-for="bookmark in bookmarkSearch.results" :key="`search-${bookmark.id}`" class="bookmark-row" @contextmenu="showBookmarkMenu($event, bookmark)" @touchstart.passive="showBookmarkMenu($event, bookmark)"><span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span><div><h3>{{ bookmark.title }}</h3><p>{{ bookmark.url }}</p><small>{{ bookmark.path || '搜索结果' }}</small></div><a class="open-link" :href="bookmark.url">打开</a></article><div v-if="!bookmarkSearch.loading && !bookmarkSearch.results.length" class="empty-state">没有匹配的收藏。</div></template><template v-else><div v-if="activeFolderId" class="quick-create bookmark-create"><input v-model="quickBookmark.title" placeholder="收藏标题" /><input v-model="quickBookmark.url" placeholder="https://example.com" /><input v-model="quickBookmark.note" placeholder="备注" /><button type="button" @click="fillQuickBookmarkMetadata">{{ metadataLoading ? '抓取中' : '自动抓取' }}</button><button type="button" @click="addBookmark">新增收藏</button></div><article v-for="bookmark in bookmarks" :key="bookmark.id" class="bookmark-row" draggable="true" @dragstart="startDrag('bookmark', bookmark)" @dragover.prevent @drop="dropBookmark(bookmark)" @contextmenu="showBookmarkMenu($event, bookmark)" @touchstart.passive="showBookmarkMenu($event, bookmark)"><span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span><div><h3>{{ bookmark.title }}</h3><p>{{ bookmark.url }}</p><small>{{ bookmark.path || '当前文件夹' }}</small></div><div class="row-actions"><button type="button" @click="editBookmark(bookmark)">编辑</button><button type="button" @click="moveBookmark(bookmark, -1)">上移</button><button type="button" @click="moveBookmark(bookmark, 1)">下移</button><button type="button" @click="removeBookmark(bookmark)">删除</button></div></article><div v-if="activeFolderId && !bookmarks.length" class="empty-state">这个文件夹还没有收藏。</div></template></div>
        </section>
      </aside>

      <section v-if="activeView === 'home'" class="home-panel sun-panel">
        <section class="hero-card"><div class="sun-title"><h2>{{ settingsForm.siteTitle || 'biu-panel' }}</h2><b>|</b><div class="clock"><strong>{{ displayTime }}</strong><span>{{ displayDate }}</span></div></div><div class="sun-search"><span>B</span><input placeholder="Enter search content" @keyup.enter="runBookmarkSearch" /><em>⌕</em></div><p>{{ statusText }}</p></section>

        <section v-for="group in displayGroups" :key="group.id || group.name" class="nav-group" :draggable="!!group.id" @dragstart="group.id && startDrag('navGroup', group)" @dragover.prevent @drop="group.id && dropNavGroup(group)"><header class="group-head" @contextmenu="showGroupMenu($event, group)"><h2>{{ group.name }}</h2></header><div class="card-grid"><a v-for="item in group.items" :key="item.id || item.name" class="nav-card" :href="resolveNavUrl(item)" :draggable="!!item.id" @dragstart="item.id && startDrag('navItem', item, group.id)" @dragover.prevent @drop.prevent="item.id && dropNavCard(group, item)" @contextmenu="showCardMenu($event, item)" @touchstart.passive="showCardMenu($event, item)"><span class="card-icon"><img v-if="isImageValue(item.icon)" :src="item.icon" alt="" /><span v-else>{{ (item.icon || item.name).slice(0, 1) }}</span></span><span>{{ item.name }}</span></a></div></section>
      </section>

      <section v-if="settingsOpen" class="settings-mask" @click.stop="settingsOpen = false"><section class="settings-panel settings-center" @click.stop><header class="settings-head"><div><span class="eyebrow dark">设置中心</span><h2>系统设置</h2></div><div class="inline-actions"><button type="button" @click="submitSettings">保存设置</button><button type="button" @click="settingsOpen = false">关闭</button></div></header><div class="settings-layout"><aside class="settings-menu"><button type="button" :class="{ active: activeSettings === '个人信息' }" @click="activeSettings = '个人信息'">个人信息</button><button type="button" :class="{ active: activeSettings === '首页样式' }" @click="activeSettings = '首页样式'">首页样式</button><button type="button" :class="{ active: activeSettings === '导航管理' }" @click="activeSettings = '导航管理'">导航管理</button><button type="button" :class="{ active: activeSettings === '收藏夹' }" @click="activeSettings = '收藏夹'">收藏夹</button><button type="button" :class="{ active: activeSettings === '导入导出' }" @click="activeSettings = '导入导出'">导入导出</button><button type="button" :class="{ active: activeSettings === '备份恢复' }" @click="activeSettings = '备份恢复'">备份恢复</button><button type="button" :class="{ active: activeSettings === 'S3 存储' }" @click="activeSettings = 'S3 存储'">S3 存储</button><button type="button" :class="{ active: activeSettings === '关于' }" @click="activeSettings = '关于'">关于</button></aside><div class="settings-content"><section v-if="activeSettings === '个人信息'" class="setting-card"><h3>个人信息</h3><p>{{ statusText }}</p></section><section v-if="activeSettings === '导航管理'" class="setting-card"><h3>导航管理</h3><p>导航新增、编辑、删除暂时通过首页右键菜单操作；后续会在这里做完整管理列表。</p></section><section v-if="activeSettings === '收藏夹'" class="setting-card"><h3>收藏夹</h3><p>收藏夹请从页面左侧按钮打开；后续会在这里加入目录管理、批量导入和批量整理。</p></section><section v-if="activeSettings === '导入导出'" class="setting-card"><h3>导入导出</h3><p>收藏夹导入导出现在在左侧收藏夹抽屉顶部；后续会集中到这里。</p></section><section v-if="activeSettings === '关于'" class="setting-card"><h3>关于</h3><p>这是个人自用导航站和网页收藏夹，当前正在按 Sun-Panel 的交互方式重做。</p></section><div class="settings-grid"><section v-show="activeSettings === '首页样式'" class="setting-card"><h3>站点个性化</h3><label>站点标题<input v-model="settingsForm.siteTitle" /></label><label>Logo 图片<input v-model="settingsForm.logoUrl" placeholder="本地上传地址或图片链接" /></label><label>上传 Logo<input type="file" accept="image/*" @change="uploadIconFile($event, settingsForm, 'logoUrl')" /></label><label>背景图<input v-model="settingsForm.backgroundUrl" placeholder="图片链接" /></label><label>上传背景图<input type="file" accept="image/*" @change="uploadIconFile($event, settingsForm, 'backgroundUrl')" /></label><label>背景色<input v-model="settingsForm.backgroundColor" /></label></section><section v-show="activeSettings === 'S3 存储'" class="setting-card"><h3>S3 存储</h3><label class="check-row"><input v-model="settingsForm.s3Enabled" true-value="true" false-value="false" type="checkbox" /> 启用 S3 配置</label><label>Endpoint<input v-model="settingsForm.s3Endpoint" placeholder="https://s3.example.com" /></label><label>Region<input v-model="settingsForm.s3Region" placeholder="auto" /></label><label>Bucket<input v-model="settingsForm.s3Bucket" placeholder="biu-panel" /></label><label>Access Key<input v-model="settingsForm.s3AccessKey" /></label><label>Secret Key<input v-model="settingsForm.s3SecretKey" type="password" /></label><label>上传前缀<input v-model="settingsForm.s3Prefix" placeholder="biu-panel/" /></label><label>公开访问域名<input v-model="settingsForm.s3PublicBase" placeholder="https://cdn.example.com/biu-panel" /></label><label class="check-row"><input v-model="settingsForm.s3PathStyle" true-value="true" false-value="false" type="checkbox" /> Path-style 兼容模式</label><div class="inline-actions"><button type="button" @click="submitSettings">保存 S3 配置</button><button type="button" @click="submitTestS3">测试 S3</button></div></section><section v-show="activeSettings === '首页样式'" class="setting-card"><h3>内外网判断</h3><label>统一检测地址<input v-model="settingsForm.lanDetectUrl" /></label><label>超时时间 ms<input v-model="settingsForm.lanDetectTimeout" /></label><label class="check-row"><input v-model="settingsForm.autoDetectLan" true-value="true" false-value="false" type="checkbox" /> 启用自动判断</label></section><section v-show="activeSettings === '备份恢复'" class="setting-card"><h3>备份恢复</h3><p>系统备份为 .tar.gz，包含数据库、本地上传文件和版本信息。</p><div class="inline-actions"><button type="button" @click="window.location.href = '/api/backup/download'">下载备份</button><button type="button" @click="submitBackupToS3">备份到 S3</button><label class="file-button">恢复备份<input type="file" accept=".gz,.tgz,application/gzip" @change="restoreBackupFile" /></label></div></section></div></div></div></section></section>
    </template>

    <section v-if="editDialog.open" class="modal-mask" @click.stop="closeEditDialog"><form class="edit-modal" @click.stop @submit.prevent="saveEditDialog"><header class="modal-head"><h2>{{ editDialog.title }}</h2><button type="button" @click="closeEditDialog">关闭</button></header><label v-if="editDialog.type === 'navGroup' || editDialog.type === 'folder'">名称<input v-model="editDialog.form.name" /></label><template v-if="editDialog.type === 'navItem'"><label>名称<input v-model="editDialog.form.name" /></label><label>图标文字 / 图片链接<input v-model="editDialog.form.icon" /></label><label>上传图标图片<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'icon')" /></label><label>内网地址<input v-model="editDialog.form.lanUrl" /></label><label>外网地址<input v-model="editDialog.form.wanUrl" /></label><label>打开模式<select v-model="editDialog.form.urlMode"><option value="auto">自动</option><option value="lan">强制内网</option><option value="wan">强制外网</option></select></label><button type="button" @click="fillMetadata(editDialog.form)">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></template><template v-if="editDialog.type === 'bookmark'"><label>标题<input v-model="editDialog.form.title" /></label><label>网址<input v-model="editDialog.form.url" /></label><label>图标<input v-model="editDialog.form.favicon" /></label><label>上传图标图片<input type="file" accept="image/*" @change="uploadIconFile($event, editDialog.form, 'favicon')" /></label><label>备注<input v-model="editDialog.form.note" /></label><button type="button" @click="fillMetadata(editDialog.form)">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button></template><footer class="modal-actions"><button type="button" @click="closeEditDialog">取消</button><button type="submit">保存</button></footer></form></section>
    <div v-if="menu.open" class="context-menu" :style="menuStyle" @click.stop><strong>{{ menu.title }}</strong><button v-for="action in menu.actions" :key="action.label" type="button" @click="runMenuAction(action)">{{ action.label }}</button></div>
  </main>
</template>
