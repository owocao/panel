<script setup>
const transparentDragImage = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw=='
let dragImage

defineProps({
  bookmark: { type: Object, required: true },
  selectionMode: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
  draggable: { type: Boolean, default: false },
  showActions: { type: Boolean, default: false },
  compact: { type: Boolean, default: false },
  pathFallback: { type: String, default: '' },
  isImageValue: { type: Function, required: true },
  dragOver: { type: Boolean, default: false },
  dragInsertPosition: { type: String, default: '' },
})

const emit = defineEmits(['toggle-selection', 'context-menu', 'drag-start', 'drag-over', 'drag-end', 'drop', 'edit', 'move-up', 'move-down', 'remove', 'open'])

function handleDragStart(bookmark, event) {
  if (event.dataTransfer) {
    if (!dragImage) {
      dragImage = new Image()
      dragImage.src = transparentDragImage
    }
    event.dataTransfer.setDragImage(dragImage, 0, 0)
  }
  emit('drag-start', bookmark, event)
}
</script>

<template>
  <article class="bookmark-row" :class="{ compact, 'drag-over': dragOver, 'insert-before': dragOver && dragInsertPosition === 'before', 'insert-after': dragOver && dragInsertPosition === 'after' }" :data-url="bookmark.url" :draggable="draggable" @click="emit('open', bookmark)" @dragstart="handleDragStart(bookmark, $event)" @dragend="emit('drag-end')" @dragover.prevent="emit('drag-over', bookmark, $event)" @drop="emit('drop', bookmark, $event)" @contextmenu.prevent.stop="emit('context-menu', $event, bookmark)">
    <label v-if="selectionMode" class="bookmark-select" @click.stop><input type="checkbox" :checked="selected" @change="emit('toggle-selection', bookmark)" /></label>
    <span class="favicon"><img v-if="isImageValue(bookmark.favicon)" :src="bookmark.favicon" alt="" /><span v-else>{{ bookmark.title.slice(0, 1) }}</span></span>
    <div>
      <h3>{{ bookmark.title }}</h3>
      <p v-if="!compact">{{ bookmark.url }}</p>
      <small v-if="bookmark.path || pathFallback" class="bookmark-path">{{ bookmark.path || pathFallback }}</small>
    </div>
    <div v-if="showActions" class="row-actions">
      <button type="button" @click.stop="emit('edit', bookmark)">编辑</button>
      <button type="button" @click.stop="emit('move-up', bookmark)">上移</button>
      <button type="button" @click.stop="emit('move-down', bookmark)">下移</button>
      <button type="button" @click.stop="emit('remove', bookmark)">删除</button>
    </div>
  </article>
</template>
