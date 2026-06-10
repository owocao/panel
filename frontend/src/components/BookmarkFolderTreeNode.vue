<script setup>
import { computed } from 'vue'

const props = defineProps({
  folder: { type: Object, required: true },
  depth: { type: Number, default: 0 },
  activeFolderId: { type: [Number, String, null], default: null },
})

const emit = defineEmits(['select', 'toggle', 'edit', 'remove', 'move', 'create-child'])

const isActive = computed(() => props.folder.id === props.activeFolderId)
const hasChildren = computed(() => props.folder.children?.length > 0 || props.folder.hasChildren)

function handleSelect() {
  emit('select', props.folder)
}
function handleToggle() {
  emit('toggle', props.folder)
}
function handleEdit() {
  emit('edit', props.folder)
}
function handleRemove() {
  emit('remove', props.folder)
}
function handleMove(offset) {
  emit('move', { folder: props.folder, offset })
}
function handleCreateChild() {
  emit('create-child', props.folder)
}
</script>

<template>
  <div class="tree-node" :style="{ '--tree-depth': depth }">
    <button
      class="folder tree-folder"
      :class="{ active: isActive }"
      type="button"
      @click="handleSelect"
    >
      <span class="folder-toggle" :class="{ expanded: folder.expanded }" @click.stop="handleToggle">{{ hasChildren ? (folder.expanded ? '▾' : '▸') : '•' }}</span>
      <strong>{{ folder.name }}</strong>
      <small>
        <template v-if="folder.loading">加载中…</template>
        <template v-else-if="folder.hasChildren">{{ folder.expanded ? '已展开' : '可展开子目录' }}</template>
        <template v-else>当前目录</template>
      </small>
      <span class="mini-actions">
        <em @click.stop="handleCreateChild">新增子目录</em>
        <em @click.stop="handleEdit">编辑</em>
        <em @click.stop="handleMove(-1)">上移</em>
        <em @click.stop="handleMove(1)">下移</em>
        <em @click.stop="handleRemove">删除</em>
      </span>
    </button>

    <div v-if="folder.expanded" class="tree-children">
      <BookmarkFolderTreeNode
        v-for="child in folder.children || []"
        :key="child.id"
        :folder="child"
        :depth="depth + 1"
        :active-folder-id="activeFolderId"
        @select="$emit('select', $event)"
        @toggle="$emit('toggle', $event)"
        @edit="$emit('edit', $event)"
        @remove="$emit('remove', $event)"
        @move="$emit('move', $event)"
        @create-child="$emit('create-child', $event)"
      />
    </div>
  </div>
</template>
