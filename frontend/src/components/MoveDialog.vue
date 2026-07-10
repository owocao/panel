<script setup>
import { computed, ref } from 'vue'
import { folderOptionStyle, folderPrefix } from '../utils/folderDisplay'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  items: { type: Array, default: () => [] },
  itemLabel: { type: String, default: '收藏' },
  allowRoot: { type: Boolean, default: false },
  targetFolderId: { type: [Number, String, null], default: null },
  folderFlatList: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'confirm', 'update:targetFolderId'])
const openSelect = ref(false)
const selectedFolder = computed(() => props.folderFlatList.find((folder) => folder.id === props.targetFolderId || folder.id === Number(props.targetFolderId)))
const selectedName = computed(() => {
  if (props.allowRoot && (props.targetFolderId == null || props.targetFolderId === '')) return '根目录'
  return selectedFolder.value ? selectedFolder.value.name : '请选择收藏夹'
})

function selectFolder(folder) {
  emit('update:targetFolderId', folder.id)
  openSelect.value = false
}

function selectRoot() {
  emit('update:targetFolderId', null)
  openSelect.value = false
}
</script>

<template>
  <section v-if="open" class="modal-mask" @mousedown.self.stop="emit('close')" @wheel.stop.prevent>
    <div class="edit-modal move-modal" @click.stop @wheel.stop>
      <header class="modal-head">
        <h2>{{ title }}</h2>
      </header>
      <p class="move-hint">将 {{ items.length }} 条{{ itemLabel }}移动到以下收藏夹。</p>
      <label>
        目标收藏夹
        <div class="select-popover" :class="{ open: openSelect }">
          <button type="button" class="select-trigger" @click="openSelect = !openSelect"><span>{{ selectedName }}</span><span class="select-arrow">⌄</span></button>
          <div v-if="openSelect" class="select-options">
            <button v-if="allowRoot" type="button" class="folder-option depth-0" :class="{ active: targetFolderId == null || targetFolderId === '' }" @pointerdown.stop.prevent="selectRoot">
              <span class="folder-option-prefix"></span>
              <span class="folder-option-name">根目录</span>
            </button>
            <button
              v-for="folder in folderFlatList"
              :key="`move-folder-${folder.id}`"
              type="button"
              class="folder-option"
              :class="{ active: folder.id === targetFolderId || folder.id === Number(targetFolderId) }"
              :style="folderOptionStyle(folder)"
              @pointerdown.stop.prevent="selectFolder(folder)"
            >
              <span class="folder-option-prefix">{{ folderPrefix(folder) }}</span>
              <span class="folder-option-name">{{ folder.name }}</span>
            </button>
          </div>
        </div>
      </label>
      <footer class="modal-actions">
        <button type="button" @click="emit('confirm')">确认</button>
        <button type="button" @click="emit('close')">取消</button>
      </footer>
    </div>
  </section>
</template>
