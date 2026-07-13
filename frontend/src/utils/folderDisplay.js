export function folderDepth(folder) {
  return Number(folder?.depth || 0)
}

export function folderPrefix(folder) {
  const depth = folderDepth(folder)
  return depth > 0 ? '└' : ''
}

export function folderOptionStyle(folder) {
  const depth = folderDepth(folder)
  return {
    paddingLeft: `${9 + depth * 18}px`,
    '--folder-depth-line': `${12 + Math.max(0, depth - 1) * 18}px`,
  }
}
