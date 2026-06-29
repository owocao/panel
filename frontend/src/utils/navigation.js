export function normalizeNetworkMode(value) {
  return value === 'wan' ? 'wan' : 'lan'
}

export function ensureHttp(url) {
  url = String(url || '').trim()
  if (!url || url === '#') return url
  if (!/^https?:\/\//i.test(url) && !url.startsWith('/')) {
    return 'http://' + url
  }
  return url
}

export function navUrlCandidates(item, networkMode = 'lan') {
  const lanUrl = ensureHttp(item?.lanUrl)
  const wanUrl = ensureHttp(item?.wanUrl)
  if (normalizeNetworkMode(networkMode) === 'lan') return { primary: lanUrl, fallback: wanUrl }
  return { primary: wanUrl, fallback: lanUrl }
}

export function resolveNavUrl(item, networkMode = 'lan') {
  const { primary, fallback } = navUrlCandidates(item, networkMode)
  return primary || fallback || '#'
}
