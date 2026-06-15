<script setup>
defineProps({
  bookmark: { type: Object, required: true },
  selectionMode: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  draggable: { type: Boolean, default: false },
  showActions: { type: Boolean, default: false },
  pathFallback: { type: String, default: '' },
  isImageValue: { type: Function, required: true },
})

const emit = defineEmits(['toggle-selection', 'context-menu', 'drag-start', 'drop', 'edit', 'move-up', 'move-down', 'remove'])
</script>

<template>
  <article class="bookmark-row" :draggable="draggable" @dragstart="emit('drag-start', bookmark)" @dragover.prevent @drop="emit('drop', bookmark)" @contextmenu="emit('context-menu', $event, bookmark)">
    <label v-if="selectionMode" class="bookmark-select"><input type="checkbox" :checked="selected" @change="emit('toggle-selection', bookmark)" /></label>
    <span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span>
    <div>
      <h3>{{ bookmark.title }}</h3>
      <p>{{ bookmark.url }}</p>
      <small>{{ bookmark.path || pathFallback }}</small>
    </div>
    <a v-if="!showActions" class="open-link" :href="bookmark.url">打开</a>
    <div v-else class="row-actions">
      <button type="button" @click="emit('edit', bookmark)">编辑</button>
      <button type="button" @click="emit('move-up', bookmark)">上移</button>
      <button type="button" @click="emit('move-down', bookmark)">下移</button>
      <button type="button" @click="emit('remove', bookmark)">删除</button>
    </div>
  </article>
</template>
