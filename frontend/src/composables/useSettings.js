import { ref } from 'vue'
import { normalizeNetworkMode } from '../utils/navigation'

const defaultSettings = {
  siteTitle: 'biu-panel',
  showTitle: 'true',
  showClock: 'true',
  showSeconds: 'false',
  showSearch: 'true',
  searchEngines: JSON.stringify([
    { key: 'google', title: 'Google', icon: 'G', url: 'https://www.google.com/search?q={q}' },
    { key: 'baidu', title: '百度', icon: '百', url: 'https://www.baidu.com/s?wd={q}' },
    { key: 'bing', title: 'Bing', icon: 'B', url: 'https://www.bing.com/search?q={q}' },
  ]),
  backgroundUrl: '',
  backgroundColor: '#02030a',
  lanDetectTimeout: '800',
  s3Endpoint: '',
  s3Region: 'auto',
  s3Bucket: '',
  s3AccessKey: '',
  s3SecretKey: '',
  s3Prefix: 'biu-panel/',
  s3PathStyle: 'true',
  s3Enabled: 'false',
  s3PublicBase: '',
}

export function useSettings({
  getSettings,
  saveSettings,
  testS3,
  getNetworkMode,
  setNetworkMode,
  navGroupsDraft,
  navDraftDirty,
  foldersDraft,
  foldersDraftDirty,
  folders,
  getNavGroups,
  closeMenu,
  loadFolders,
  loadAllFolderChildren,
  syncFoldersDraftFromFolders,
  saveNavGroupDraftOrder,
  saveFolderDraftOrder,
  onStatus,
}) {
  const settingsOpen = ref(false)
  const activeSettings = ref('个性化')
  const settingsMenuCollapsed = ref(false)
  const settingsMessage = ref('')
  const settingsSaving = ref(false)
  const settingsForm = ref({ ...defaultSettings })
  const settingsDraft = ref({ ...settingsForm.value })

  function openSettings() {
    closeMenu()
    settingsMessage.value = ''
    settingsSaving.value = false
    activeSettings.value = '个性化'
    settingsDraft.value = { ...settingsForm.value }
    navGroupsDraft.value = getNavGroups().map((group) => ({ ...group, items: [...(group.items || [])] }))
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

  function settingsDraftDirty() {
    const keys = new Set([...Object.keys(settingsForm.value), ...Object.keys(settingsDraft.value)])
    for (const key of keys) {
      if (String(settingsForm.value[key] ?? '') !== String(settingsDraft.value[key] ?? '')) return true
    }
    return false
  }

  async function loadSettings() {
    try {
      const data = await getSettings()
      settingsForm.value = { ...settingsForm.value, ...data }
      settingsDraft.value = { ...settingsForm.value }
      const nextNetworkMode = normalizeNetworkMode(getNetworkMode())
      setNetworkMode(nextNetworkMode)
      localStorage.setItem('biu-network-mode', nextNetworkMode)
      if (settingsForm.value.siteTitle) document.title = settingsForm.value.siteTitle
    } catch {
      // Settings require login; keep defaults for public views.
    }
  }

  async function submitTestS3() {
    try {
      const data = await testS3()
      onStatus?.(`S3 测试成功：${data.key}`)
    } catch (error) {
      onStatus?.(error.message)
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
      onStatus?.('设置已保存')
      settingsMessage.value = `设置已保存：${new Date().toLocaleTimeString('zh-CN', { hour12: false })}`
      settingsOpen.value = false
    } catch (error) {
      onStatus?.(error.message)
      settingsMessage.value = `保存失败：${error.message}`
    } finally {
      settingsSaving.value = false
    }
  }

  return {
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
  }
}
