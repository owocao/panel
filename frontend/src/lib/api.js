const API_BASE = import.meta.env.VITE_API_BASE || ''

function notifyAuthExpired() {
  if (typeof window !== 'undefined') window.dispatchEvent(new CustomEvent('biu-auth-expired'))
}

function ensureResponseOK(response, body, fallback) {
  if (response.status === 401) notifyAuthExpired()
  if (!response.ok || body.success === false) {
    throw new Error(body.error || fallback || `请求失败：${response.status}`)
  }
}

export async function api(path, options = {}) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), options.timeout || 10000)
  
  try {
    const response = await fetch(`${API_BASE}${path}`, {
      credentials: 'include',
      headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
      signal: controller.signal,
      ...options,
    })
    clearTimeout(timeoutId)
    const body = await response.json().catch(() => ({}))
    ensureResponseOK(response, body)
    return body.data
  } catch (error) {
    clearTimeout(timeoutId)
    if (error.name === 'AbortError') {
      throw new Error('请求超时，请重试')
    }
    throw error
  }
}

const jsonRequest = (method, body) => ({ method, body: JSON.stringify(body) })

export const setupStatus = () => api('/api/setup/status')
export const login = (payload) => api('/api/auth/login', jsonRequest('POST', payload))
export const setup = (payload) => api('/api/setup', jsonRequest('POST', payload))
export const getMe = () => api('/api/auth/me')
export const logout = () => api('/api/auth/logout', { method: 'POST' })

export const getNavigation = () => api('/api/navigation')
export const createNavGroup = (payload) => api('/api/navigation/groups', jsonRequest('POST', payload))
export const updateNavGroup = (payload) => api('/api/navigation/groups', jsonRequest('PUT', payload))
export const deleteNavGroup = (id) => api(`/api/navigation/groups?id=${id}`, { method: 'DELETE' })
export const createNavItem = (payload) => api('/api/navigation/items', jsonRequest('POST', payload))
export const updateNavItem = (payload) => api('/api/navigation/items', jsonRequest('PUT', payload))
export const deleteNavItem = (id) => api(`/api/navigation/items?id=${id}`, { method: 'DELETE' })

export const getBookmarkFolders = (parentId) => api(`/api/bookmark/folders${parentId ? `?parentId=${parentId}` : ''}`)
export const createBookmarkFolder = (payload) => api('/api/bookmark/folders', { ...jsonRequest('POST', payload), timeout: 30000 })
export const updateBookmarkFolder = (payload) => api('/api/bookmark/folders', { ...jsonRequest('PUT', payload), timeout: 30000 })
export const deleteBookmarkFolder = (id) => api(`/api/bookmark/folders?id=${id}`, { method: 'DELETE', timeout: 30000 })
export const getBookmarks = (folderId) => api(`/api/bookmarks?folderId=${folderId}`)
export const createBookmark = (payload) => api('/api/bookmarks', jsonRequest('POST', payload))
export const updateBookmark = (payload) => api('/api/bookmarks', jsonRequest('PUT', payload))
export const deleteBookmark = (id) => api(`/api/bookmarks?id=${id}`, { method: 'DELETE' })
export const searchBookmarks = (q) => api(`/api/bookmark/search?q=${encodeURIComponent(q)}`)
export const fetchMetadata = (url) => api(`/api/metadata?url=${encodeURIComponent(url)}`, { timeout: 3000 })
export const getSettings = () => api('/api/settings')
export const saveSettings = (payload) => api('/api/settings', { ...jsonRequest('PUT', payload), timeout: 30000 })
export const testS3 = () => api('/api/s3/test', { method: 'POST' })
export async function downloadFile(path) {
  const response = await fetch(`${API_BASE}${path}`, { credentials: 'include' })
  if (response.status === 401) notifyAuthExpired()
  if (!response.ok) throw new Error(`下载失败：${response.status}`)
  const blob = await response.blob()
  const disposition = response.headers.get('Content-Disposition') || ''
  const match = disposition.match(/filename="?([^";]+)"?/i)
  const fileName = match?.[1] || 'download'
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
export async function uploadAsset(file) {
  const form = new FormData()
  form.append('file', file)
  const response = await fetch(`${API_BASE}/api/assets/upload`, { method: 'POST', credentials: 'include', body: form })
  const body = await response.json().catch(() => ({}))
  ensureResponseOK(response, body, `上传失败：${response.status}`)
  return body.data
}
export async function importBookmarkHTML(file) {
  const form = new FormData()
  form.append('file', file)
  const response = await fetch(`${API_BASE}/api/bookmark/import`, { method: 'POST', credentials: 'include', body: form })
  const body = await response.json().catch(() => ({}))
  ensureResponseOK(response, body, `导入失败：${response.status}`)
  return body.data
}
export async function restoreBackup(file) {
  const form = new FormData()
  form.append('file', file)
  const response = await fetch(`${API_BASE}/api/backup/restore`, { method: 'POST', credentials: 'include', body: form })
  const body = await response.json().catch(() => ({}))
  ensureResponseOK(response, body, `恢复失败：${response.status}`)
  return body.data
}
export async function restoreNavigationBackup(file) {
  const form = new FormData()
  form.append('file', file)
  const response = await fetch(`${API_BASE}/api/navigation/restore`, { method: 'POST', credentials: 'include', body: form })
  const body = await response.json().catch(() => ({}))
  ensureResponseOK(response, body, `导航恢复失败：${response.status}`)
  return body.data
}
