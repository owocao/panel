import { ref, unref } from 'vue'

export function createEmptyEditDialog() {
  return { open: false, type: '', title: '', form: {} }
}

export function useEditDialog({ navGroupOptions, isImageValue }) {
  const editDialog = ref(createEmptyEditDialog())
  const groupSelectOpen = ref(false)

  function closeEditDialog() {
    groupSelectOpen.value = false
    editDialog.value = createEmptyEditDialog()
  }

  function clampEditField(field, max) {
    const value = String(editDialog.value.form[field] || '')
    if (value.length > max) editDialog.value.form[field] = value.slice(0, max)
  }

  function getEditGroupName() {
    const groups = unref(navGroupOptions) || []
    return groups.find((group) => group.id === editDialog.value.form.groupId)?.name || '请选择分组'
  }

  function selectEditGroup(group) {
    editDialog.value.form.groupId = group.id
    groupSelectOpen.value = false
    window.setTimeout(() => { groupSelectOpen.value = false }, 0)
  }

  function setNavIconMode(form, mode) {
    if (form.iconMode === mode) return

    if (form.iconMode === 'text') {
      form.__textIcon = form.icon
    } else if (form.iconMode === 'image') {
      form.__imageIcon = form.icon
    }

    form.iconMode = mode

    if (mode === 'text') {
      form.icon = form.__textIcon !== undefined ? form.__textIcon : (isImageValue(form.icon) ? '' : form.icon)
    } else if (mode === 'image') {
      form.icon = form.__imageIcon !== undefined ? form.__imageIcon : (!isImageValue(form.icon) ? '' : form.icon)
    }
  }

  return {
    editDialog,
    groupSelectOpen,
    closeEditDialog,
    clampEditField,
    getEditGroupName,
    selectEditGroup,
    setNavIconMode,
  }
}
