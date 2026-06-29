export function normalizeFolder(folder, parentId = null) {
  return {
    ...folder,
    parentId: folder.parentId ?? parentId ?? null,
    children: Array.isArray(folder.children) ? folder.children : [],
    childrenLoaded: Boolean(folder.childrenLoaded),
    expanded: Boolean(folder.expanded),
    loading: Boolean(folder.loading),
  }
}

export function flattenFolders(nodes, depth = 0, out = []) {
  for (const folder of nodes || []) {
    out.push({ ...folder, depth })
    if (folder.childrenLoaded && Array.isArray(folder.children) && folder.children.length) {
      flattenFolders(folder.children, depth + 1, out)
    }
  }
  return out
}

export function findFolderById(nodes, id) {
  if (id == null) return null
  for (const folder of nodes || []) {
    if (folder.id === id) return folder
    const nested = findFolderById(folder.children, id)
    if (nested) return nested
  }
  return null
}

export function cloneFolderTree(nodes = []) {
  return (nodes || []).map((folder) => ({
    ...folder,
    children: cloneFolderTree(folder.children || []),
    childrenLoaded: true,
    hasChildren: Boolean(folder.children?.length),
    expanded: Boolean(folder.expanded),
    loading: false,
  }))
}
