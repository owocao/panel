import { computed, ref } from 'vue'

export function useContextMenu({ onError, onStatus, actions }) {
  const menu = ref({ open: false, x: 0, y: 0, title: '', actions: [], compact: false })
  const menuStyle = computed(() => ({ left: `${menu.value.x}px`, top: `${menu.value.y}px`, width: menu.value.width ? `${menu.value.width}px` : undefined }))

  function showMenu(event, title, menuActions, options = {}) {
    event.preventDefault()
    const point = event.touches?.[0] || event
    const menuWidth = options.width || (options.compact ? 128 : 220)
    const x = Math.min(point.clientX, window.innerWidth - menuWidth - 8)
    const menuHeight = options.height || (menuActions.length * 38 + (title ? 36 : 0) + 16)
    const y = Math.min(point.clientY, window.innerHeight - menuHeight - 8)
    menu.value = { open: true, x: Math.max(8, x), y: Math.max(8, y), title, actions: menuActions, compact: Boolean(options.compact), width: options.width || null }
  }

  function closeMenu() {
    menu.value.open = false
  }

  async function runMenuAction(action) {
    closeMenu()
    try {
      if (action?.run) await action.run()
    } catch (error) {
      onError?.(error)
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
      onStatus?.('链接已复制')
    } catch {
      onStatus?.('复制失败，请手动复制')
    }
  }

  function showCardMenu(event, item, group = null) {
    if (group && actions.isNavGroupEditing(group)) {
      event.preventDefault()
      closeMenu()
      return
    }
    showMenu(event, item.name, [
      { label: '新标签页打开', icon: 'window', variant: 'icon', run: () => actions.openNavItemFromMenu(item, '_blank', 'noopener,noreferrer') },
      { label: '新窗口打开', icon: 'external-link-alt', variant: 'icon', run: () => actions.openNavItemFromMenu(item, `biu-nav-window-${item.id || Date.now()}`, 'popup=yes,width=1200,height=800') },
      { divider: true },
      { label: '编辑', icon: 'edit', run: () => actions.editNavCard(item, group) },
      { label: '删除', icon: 'trash-alt', run: () => actions.removeNavCard(item) },
    ], { compact: true })
  }

  function showGroupMenu(event, group) {
    if (actions.isNavGroupEditing(group)) {
      event.preventDefault()
      closeMenu()
      return
    }
    showMenu(event, group.name, [
      { label: '新增卡片', icon: 'plus', run: () => actions.addCardFromMenu(group) },
      { label: '编辑分组', icon: 'edit', run: () => actions.editNavGroup(group) },
      { label: '删除分组', icon: 'trash-alt', run: () => actions.removeNavGroup(group) },
    ])
  }

  function showFolderMenu(event, folder) {
    event.preventDefault()
    event.stopPropagation()
    const current = actions.getFolderFlatList().find((item) => item.id === folder.id)
    const canCreateChild = !current || current.depth < 3
    showMenu(event, folder.name, [
      ...(canCreateChild ? [{ label: '新增收藏夹', icon: 'plus', run: () => actions.createFolderByPrompt(folder) }] : []),
      { label: '新增书签', icon: 'plus', run: () => actions.createBookmarkByPrompt(folder) },
      { divider: true },
      { label: '移动', icon: 'folder', run: () => actions.openFolderMoveDialog(folder) },
      { label: '编辑', icon: 'edit', run: () => actions.editFolder(folder) },
      { label: '删除', icon: 'trash-alt', run: () => actions.removeFolder(folder) },
    ], { width: 148 })
  }

  function showBookmarkMenu(event, bookmark) {
    event.preventDefault()
    event.stopPropagation()
    showMenu(event, bookmark.title, [
      { label: '新标签页打开', icon: 'window', variant: 'icon', run: () => window.open(actions.ensureHttp(bookmark.url), '_blank', 'noopener,noreferrer') },
      { label: '新窗口打开', icon: 'external-link-alt', variant: 'icon', run: () => window.open(actions.ensureHttp(bookmark.url), `biu-bookmark-window-${bookmark.id || Date.now()}`, 'popup=yes,width=1200,height=800') },
      { divider: true },
      { label: '复制链接', icon: 'link', run: () => copyText(bookmark.url) },
      { label: '移动', icon: 'folder', run: () => actions.openMoveDialog([bookmark], '移动') },
      { label: '首页卡片', icon: 'plus', run: () => actions.convertBookmarkToNavCard(bookmark) },
      { divider: true },
      { label: '编辑', icon: 'edit', run: () => actions.editBookmark(bookmark) },
      { label: '删除', icon: 'trash-alt', run: () => actions.removeBookmark(bookmark) },
    ], { width: 148 })
  }

  return {
    menu,
    menuStyle,
    showMenu,
    closeMenu,
    runMenuAction,
    showCardMenu,
    showGroupMenu,
    showFolderMenu,
    showBookmarkMenu,
  }
}
