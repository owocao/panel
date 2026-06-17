<script setup>
defineProps({
  bookmark: { type: Object, required: true },
  selectionMode: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  draggable: { type: Boolean, default: false },
  showActions: { type: Boolean, default: false },
  compact: { type: Boolean, default: false },
  pathFallback: { type: String, default: '' },
  isImageValue: { type: Function, required: true },
})

const emit = defineEmits(['toggle-selection', 'context-menu', 'drag-start', 'drag-over', 'drop', 'edit', 'move-up', 'move-down', 'remove', 'open'])
</script>

<template>
  <article class="bookmark-row" :class="{ compact }" :data-url="bookmark.url" :draggable="draggable" @click="emit('open', bookmark)" @dragstart="emit('drag-start', bookmark, $event)" @dragover.prevent="emit('drag-over', bookmark)" @drop="emit('drop', bookmark)" @contextmenu.prevent.stop="emit('context-menu', $event, bookmark)">
    <label v-if="selectionMode" class="bookmark-select" @click.stop><input type="checkbox" :checked="selected" @change="emit('toggle-selection', bookmark)" /></label>
    <span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span>
    <div>
      <h3>{{ bookmark.title }}</h3>
      <p v-if="!compact">{{ bookmark.url }}</p>
      <small v-if="bookmark.path || pathFallback" class="bookmark-path">{{ bookmark.path || pathFallback }}</small>
    </div>
    <div v-if="showActions" class="row-actions">
      <button type="button" @click="emit('edit', bookmark)">编辑</button>
      <button type="button" @click="emit('move-up', bookmark)">上移</button>
      <button type="button" @click="emit('move-down', bookmark)">下移</button>
      <button type="button" @click="emit('remove', bookmark)">删除</button>
    </div>
  </article>
</template>
