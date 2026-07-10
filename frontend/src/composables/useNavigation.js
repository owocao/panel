import { computed, ref } from 'vue'
import { navUrlCandidates, normalizeNetworkMode, resolveNavUrl } from '../utils/navigation'

export function useNavigation({ getNavigation, uploadAsset, settingsForm, settingsDraft, settingsOpen, openEditDialog, assetUploading, isImageValue, onStatus, onToast }) {
  const navGroups = ref([])
  const networkMode = ref('lan')
  const webSearch = ref({ q: '', engine: 'google' })
  const searchPickerOpen = ref(false)

  const displayGroups = computed(() => navGroups.value)
  const searchEngines = computed(() => parseSearchEngines(settingsForm.value.searchEngines))
  const settingsSearchEngines = computed(() => parseSearchEngines(settingsDraft.value.searchEngines))
  const activeSearchEngine = computed(() => searchEngines.value.find((engine) => engine.key === webSearch.value.engine) || searchEngines.value[0])

  function parseSearchEngines(value) {
    try {
      const engines = JSON.parse(value || '[]')
      return Array.isArray(engines) && engines.length ? engines : []
    } catch {
      return []
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

  function cycleNetworkMode() {
    networkMode.value = networkMode.value === 'lan' ? 'wan' : 'lan'
    localStorage.setItem('biu-network-mode', networkMode.value)
    const message = networkMode.value === 'lan' ? '已经切换到优先内网' : '已经切换到优先公网'
    onStatus?.(message)
    onToast?.(message)
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
    openEditDialog?.({ open: true, type: 'searchEngineCreate', title: '增加搜索引擎', form: { key: `custom-${Date.now()}`, title: '', url: '', icon: '', iconMode: 'text' } })
  }

  function editSearchEngine(engine) {
    openEditDialog?.({ open: true, type: 'searchEngine', title: '编辑搜索引擎', form: { ...engine, iconMode: isImageValue(engine.icon) ? 'image' : 'text' } })
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
      onStatus?.(error.message)
    } finally {
      assetUploading.value = false
      event.target.value = ''
    }
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
    const url = resolveNavUrl(item, networkMode.value)
    if (!url || url === '#') return
    window.open(url, target, features)
  }

  async function probeUrl(url, timeoutValue) {
    if (!url || url === '#') return false
    const timeout = Math.max(200, Number(timeoutValue || 800) || 800)
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
    const { primary, fallback } = navUrlCandidates(item, networkMode.value)
    const firstUrl = primary || fallback
    if (!firstUrl) return
    let openedWindow = null
    if (target !== '_self') openedWindow = window.open('about:blank', target, features)
    if (!primary || !fallback) {
      openResolvedUrl(firstUrl, target, openedWindow)
      return
    }
    if (await probeUrl(primary, settingsForm.value.lanDetectTimeout)) {
      openResolvedUrl(primary, target, openedWindow)
      return
    }
    openResolvedUrl(fallback, target, openedWindow)
  }

  return {
    navGroups,
    networkMode,
    webSearch,
    searchPickerOpen,
    displayGroups,
    searchEngines,
    settingsSearchEngines,
    activeSearchEngine,
    loadNavigation,
    cycleNetworkMode,
    runWebSearch,
    selectSearchEngine,
    writeSearchEngines,
    addSearchEngine,
    editSearchEngine,
    removeSearchEngine,
    moveSearchEngine,
    uploadSearchEngineIcon,
    openNavItemFromMenu,
    probeUrl,
    openNavItem,
    normalizeNetworkMode,
  }
}
