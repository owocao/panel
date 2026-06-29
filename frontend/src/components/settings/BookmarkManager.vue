<script setup>
defineProps({
  active: {
    type: Boolean,
    required: true,
  },
  folders: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits([
  'create-folder',
  'open-drawer',
  'edit-folder',
  'request-move-folder',
  'reorder-folder',
  'remove-folder',
])

function folderRowStyle(folder) {
  return { paddingLeft: `${10 + folder.depth * 14}px` }
}
</script>

<template>
  <section v-if="active" class="setting-card manager-card">
    <header class="manager-head">
      <h3>收藏夹管理</h3>
      <div class="inline-actions">
        <button type="button" @click="emit('create-folder')">新增收藏夹</button>
        <button type="button" @click="emit('open-drawer')">打开收藏夹</button>
      </div>
    </header>

    <article
      v-for="folder in folders"
      :key="`manage-folder-${folder.id}`"
      class="manager-row"
      :style="folderRowStyle(folder)"
    >
      <strong>{{ folder.name }}</strong>
      <div class="inline-actions">
        <button type="button" @click="emit('create-folder', folder)">新增收藏夹</button>
        <button type="button" @click="emit('edit-folder', folder)">编辑</button>
        <button type="button" @click="emit('request-move-folder', folder)">移动</button>
        <button type="button" @click="emit('reorder-folder', folder, -1)">上移</button>
        <button type="button" @click="emit('reorder-folder', folder, 1)">下移</button>
        <button type="button" @click="emit('remove-folder', folder)">删除</button>
      </div>
    </article>

    <div v-if="!folders.length" class="empty-state">暂无收藏夹，点击新增收藏夹创建。</div>
  </section>
</template>
