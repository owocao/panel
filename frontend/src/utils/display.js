const lunarDays = ['', '初一', '初二', '初三', '初四', '初五', '初六', '初七', '初八', '初九', '初十', '十一', '十二', '十三', '十四', '十五', '十六', '十七', '十八', '十九', '二十', '廿一', '廿二', '廿三', '廿四', '廿五', '廿六', '廿七', '廿八', '廿九', '三十']

export function isImageValue(value) {
  return typeof value === 'string' && (value.startsWith('/uploads/') || value.startsWith('http://') || value.startsWith('https://') || value.startsWith('data:image/'))
}

export function limitText(value, size) {
  return String(value || '').trim().slice(0, size)
}

export function cardTextClass(value) {
  const len = limitText(value, 5).length
  if (len <= 2) return 'text-xl'
  if (len <= 4) return 'text-md'
  return 'text-sm'
}

export function iconUrl(name) {
  return `https://api.iconify.design/uil/${name}.svg?color=%2368707a`
}

export function formatDisplayTime(date, showSeconds) {
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: showSeconds ? '2-digit' : undefined, hour12: false })
}

export function formatDisplayDate(date, dateMode) {
  const weekday = date.toLocaleDateString('zh-CN', { weekday: 'long' })
  if (dateMode === 'lunar') {
    const parts = new Intl.DateTimeFormat('zh-CN-u-ca-chinese', { month: 'long', day: 'numeric' }).formatToParts(date)
    const month = parts.find((part) => part.type === 'month')?.value || ''
    const day = Number(parts.find((part) => part.type === 'day')?.value || 0)
    return `${month}${lunarDays[day] || ''}  ${weekday}`
  }
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}-${day}  ${weekday}`
}

export function getNetworkTip(mode) {
  return mode === 'lan' ? '优先内网，超时后打开公网' : '优先公网，超时后打开内网'
}

export function getNetworkIcon(mode) {
  return mode === 'lan' ? 'wifi-router' : 'globe'
}
