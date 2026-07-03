<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  folder: { type: Object, required: true },
  depth: { type: Number, default: 0 },
  activeFolderId: { type: [Number, String, null], default: null },
  selectionMode: { type: Boolean, default: false },
  selectedIds: { type: Array, default: () => [] },
  isImageValue: { type: Function, required: true },
  dragOverId: { type: [Number, String, null], default: null },
  dragInsertPosition: { type: String, default: '' },
  dragSourceId: { type: [Number, String, null], default: null },
})

const emit = defineEmits(['select', 'toggle', 'context-menu', 'drag-start', 'drag-over', 'drag-end', 'drop'])

const isActive = computed(() => props.folder.id === props.activeFolderId)
const hasChildren = computed(() => props.folder.children?.length > 0 || props.folder.hasChildren)
const isDragOver = computed(() => props.folder.id === props.dragOverId)
const isDragSource = computed(() => props.folder.id === props.dragSourceId)
const dragging = ref(false)

function handleSelect() {
  if (dragging.value) return
  emit('select', props.folder)
}
function handleToggle() {
  emit('toggle', props.folder)
}
function handleContextMenu(event) {
  emit('context-menu', event, props.folder)
}
function forwardContextMenu(event, folder) {
  emit('context-menu', event, folder)
}
function forwardDragStart(folder, event) {
  emit('drag-start', folder, event)
}
function forwardDrop(folder) {
  emit('drop', folder)
}
function forwardDragOver(folder) {
  emit('drag-over', folder)
}
function forwardDragEnd() {
  emit('drag-end')
}
function handleDragStart(event) {
  dragging.value = true
  emit('drag-start', props.folder, event)
}
function handleDragEnd() {
  window.setTimeout(() => { dragging.value = false }, 0)
  emit('drag-end')
}
</script>

<template>
  <div class="tree-node" :style="{ '--tree-depth': depth }">
    <div
      class="folder tree-folder"
      :class="{ active: isActive, 'drag-over': isDragOver, 'drag-source': isDragSource, 'insert-before': isDragOver && dragInsertPosition === 'before', 'insert-after': isDragOver && dragInsertPosition === 'after' }"
      draggable="true"
      @dragstart="handleDragStart"
      @dragend="handleDragEnd"
      @dragover.prevent="emit('drag-over', folder)"
      @drop="emit('drop', folder)"
      @contextmenu.prevent="handleContextMenu"
    >
      <button type="button" class="folder-toggle" :class="{ expanded: folder.expanded }" draggable="false" @click.stop="handleToggle">{{ hasChildren ? (folder.expanded ? '▾' : '▸') : '📁' }}</button>
      <button type="button" class="folder-name" draggable="false" @click="handleSelect"><strong>{{ folder.name }}</strong></button>
    </div>

    <TransitionGroup v-if="folder.expanded" tag="div" class="tree-children" name="tree-list">
      <BookmarkFolderTreeNode
        v-for="child in folder.children || []"
        :key="child.id"
        :folder="child"
        :depth="depth + 1"
        :active-folder-id="activeFolderId"
        :selection-mode="selectionMode"
        :selected-ids="selectedIds"
        :is-image-value="isImageValue"
        :drag-over-id="dragOverId"
        :drag-insert-position="dragInsertPosition"
        :drag-source-id="dragSourceId"
        @select="$emit('select', $event)"
        @toggle="$emit('toggle', $event)"
        @context-menu="forwardContextMenu"
        @drag-start="forwardDragStart"
        @drag-over="forwardDragOver"
        @drag-end="forwardDragEnd"
        @drop="forwardDrop"
      />
    </TransitionGroup>
  </div>
</template>
