<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  items: { type: Array, default: () => [] },
  targetFolderId: { type: [Number, String, null], default: null },
  folderFlatList: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'confirm', 'update:targetFolderId'])
const openSelect = ref(false)
const selectedFolder = computed(() => props.folderFlatList.find((folder) => folder.id === Number(props.targetFolderId)))

function selectFolder(folder) {
  emit('update:targetFolderId', Number(folder.id))
  openSelect.value = false
}
</script>

<template>
  <section v-if="open" class="modal-mask" @mousedown.self.stop="emit('close')">
    <div class="edit-modal move-modal" @click.stop>
      <header class="modal-head">
        <h2>{{ title }}</h2>
      </header>
      <p class="move-hint">将 {{ items.length }} 条收藏移动到以下收藏夹。</p>
      <label>
        目标收藏夹
        <div class="select-popover" :class="{ open: openSelect }">
          <button type="button" class="select-trigger" @click="openSelect = !openSelect"><span>{{ selectedFolder ? `${'　'.repeat(selectedFolder.depth)}${selectedFolder.name}` : '请选择收藏夹' }}</span><span class="select-arrow">⌄</span></button>
          <div v-if="openSelect" class="select-options">
            <button v-for="folder in folderFlatList" :key="`move-folder-${folder.id}`" type="button" :class="{ active: folder.id === Number(targetFolderId) }" @pointerdown.stop.prevent="selectFolder(folder)">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</button>
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
