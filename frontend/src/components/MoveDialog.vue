<script setup>
defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  items: { type: Array, default: () => [] },
  targetFolderId: { type: [Number, String, null], default: null },
  folderFlatList: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'confirm', 'update:targetFolderId'])
</script>

<template>
  <section v-if="open" class="modal-mask" @mousedown.self.stop="emit('close')">
    <div class="edit-modal" @click.stop>
      <header class="modal-head">
        <h2>{{ title }}</h2>
        <button type="button" @click="emit('close')">关闭</button>
      </header>
      <p class="move-hint">将 {{ items.length }} 条收藏移动到以下文件夹。</p>
      <label>
        目标文件夹
        <select :value="targetFolderId" @change="emit('update:targetFolderId', Number($event.target.value))">
          <option v-for="folder in folderFlatList" :key="`move-folder-${folder.id}`" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
        </select>
      </label>
      <footer class="modal-actions">
        <button type="button" @click="emit('close')">取消</button>
        <button type="button" @click="emit('confirm')">确认移动</button>
      </footer>
    </div>
  </section>
</template>
