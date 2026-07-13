export function isTextInputTarget(target) {
  return target instanceof Element && Boolean(target.closest('input, textarea, select, [contenteditable="true"]'))
}

export function preventHomeSelection(event) {
  if (isTextInputTarget(event.target)) return
  event.preventDefault()
}

export function handleHomeKeydown(event) {
  if (!((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'a')) return
  if (isTextInputTarget(event.target)) return
  event.preventDefault()
}
