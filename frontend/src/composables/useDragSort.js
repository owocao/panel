import { ref } from 'vue'
import { findFolderById } from '../utils/bookmarkTree'

const emptyDragState = () => ({ type: '', groupId: null, item: null, overId: null, saving: false, lastMoveAt: 0, settling: false })
const emptyNavPointerDrag = () => ({ active: false, moved: false, groupId: null, item: null, pointerId: null, startX: 0, startY: 0, x: 0, y: 0, offsetX: 0, offsetY: 0, lastMoveAt: 0, lastTargetId: '' })

export function useDragSort({
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
  onStatus,
}) {
  const dragState = ref(emptyDragState())
  const navPointerDrag = ref(emptyNavPointerDrag())
  const suppressNextNavCardClick = ref(false)
  let navLongPressTimer

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
    dragState.value = emptyDragState()
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
    navPointerDrag.value = emptyNavPointerDrag()
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
        onStatus?.('卡片排序已保存')
      } catch (error) {
        onStatus?.(`排序保存失败：${error.message}`)
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
      onStatus?.('卡片排序已保存')
    } catch (error) {
      onStatus?.(`排序保存失败：${error.message}`)
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
      onStatus?.(`已移动到「${target.name}」`)
      clearDragState()
      return
    }
    if (dragState.value.type !== 'folder') return
    const siblings = folderSiblings(source)
    const reordered = siblings.map((folder, index) => ({ ...folder, sort: index + 1 }))
    clearDragState()
    try {
      await Promise.all(reordered.map((folder) => updateBookmarkFolder(folder)))
      onStatus?.('收藏夹排序已保存')
    } catch (error) {
      onStatus?.(`收藏夹排序保存失败：${error.message}`)
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

  return {
    dragState,
    navPointerDrag,
    suppressNextNavCardClick,
    startDrag,
    reorderListByTarget,
    hoverBookmark,
    folderSiblings,
    replaceFolderSiblings,
    hoverFolder,
    clearDragState,
    hoverNavCard,
    navDragFloatStyle,
    stopNavPointerListeners,
    clearNavLongPressTimer,
    startNavPointerSort,
    handleNavPointerMove,
    resetNavPointerDrag,
    handleNavPointerCancel,
    handleNavPointerUp,
    dropNavGroup,
    dropNavCard,
    dropFolder,
    dropBookmark,
    swapSort,
  }
}
