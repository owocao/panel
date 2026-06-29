export function useBackupRestore({
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
}) {
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

  return {
    restoreBackupFile,
    downloadNavigationBackup,
    restoreNavigationBackupFile,
    importBookmarksFile,
    exportBookmarks,
  }
}
