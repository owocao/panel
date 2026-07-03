<script setup>
import { computed } from 'vue'
import BookmarkFolderTreeNode from './BookmarkFolderTreeNode.vue'
import BookmarkRow from './BookmarkRow.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  folders: { type: Array, default: () => [] },
  bookmarks: { type: Array, default: () => [] },
  bookmarkSearch: { type: Object, required: true },
  activeFolder: { type: Object, default: null },
  activeFolderId: { type: [Number, String, null], default: null },
  folderCount: { type: Number, default: 0 },
  bookmarkCount: { type: Number, default: 0 },
  selectionMode: { type: Boolean, default: false },
  selectedBookmarkIds: { type: Array, default: () => [] },
  isImageValue: { type: Function, required: true },
  dragging: { type: Boolean, default: false },
  dragType: { type: String, default: '' },
  dragOverId: { type: [Number, String, null], default: null },
  dragInsertPosition: { type: String, default: '' },
  dragSourceId: { type: [Number, String, null], default: null },
})

const emit = defineEmits([
  'close-menu',
  'panel-wheel',
  'create-folder',
  'create-bookmark',
  'toggle-selection-mode',
  'clear-selection',
  'batch-move',
  'batch-delete',
  'search-input',
  'toggle-all-folders',
  'select-folder',
  'toggle-folder',
  'folder-context-menu',
  'folder-drag-start',
  'folder-drag-over',
  'drag-end',
  'folder-drop',
  'toggle-bookmark-selection',
  'bookmark-context-menu',
  'bookmark-drag-start',
  'bookmark-drag-over',
  'bookmark-drop',
  'open-bookmark',
])

const hasSearchQuery = computed(() => props.bookmarkSearch.q.trim())
const contentTitle = computed(() => (hasSearchQuery.value ? '搜索结果' : (props.activeFolder?.name || '选择文件夹')))
const contentCount = computed(() => (hasSearchQuery.value ? `${props.bookmarkSearch.results.length} 条匹配` : `${props.bookmarks.length} 条收藏`))

function isBookmarkSelected(bookmark) {
  return props.selectedBookmarkIds.includes(bookmark.id)
}
</script>

<template>
  <aside v-if="open" class="bookmark-drawer" :class="{ 'is-dragging': dragging }" aria-label="收藏夹" @click.stop="emit('close-menu')" @wheel.stop="emit('panel-wheel', $event)">
    <div class="drawer-head">
      <div>
        <span>收藏夹</span>
        <small>{{ folderCount }} 个文件夹 · {{ bookmarkCount }} 条收藏</small>
      </div>
    </div>

    <div class="bookmark-toolbar">
      <div class="bookmark-search-row">
        <label class="bookmark-search">
          <input v-model="bookmarkSearch.q" placeholder="输入标题、网址或备注搜索" @input="emit('search-input')" />
        </label>
        <div class="inline-actions search-actions">
          <span v-if="bookmarkSearch.loading">搜索中...</span>
          <span v-else-if="bookmarkSearch.results.length">找到 {{ bookmarkSearch.results.length }} 条</span>
        </div>
      </div>
      <div class="inline-actions bookmark-primary-actions">
        <button type="button" @click="emit('create-folder')">新增收藏夹</button>
        <button type="button" @click="emit('create-bookmark')">新增书签</button>
        <button type="button" :class="{ 'active': selectionMode }" @click="selectionMode ? emit('clear-selection') : emit('toggle-selection-mode')">
          {{ selectionMode ? '退出批量' : '批量操作' }}
        </button>
        <template v-if="selectionMode">
          <button type="button" @click="emit('batch-move')">批量移动</button>
          <button class="bookmark-action-danger" type="button" @click="emit('batch-delete')">批量删除</button>
        </template>
      </div>
    </div>

    <section class="bookmark-body explorer-layout">
      <nav class="folder-tree explorer-sidebar">
        <div class="folder-tree-head">
          <button type="button" class="folder-tree-title" @click="emit('toggle-all-folders')">收藏夹</button>
          <small>{{ folderCount }}</small>
        </div>
        <TransitionGroup tag="div" class="tree-root" name="tree-list">
          <BookmarkFolderTreeNode
            v-for="folder in folders"
            :key="folder.id"
            :folder="folder"
            :active-folder-id="activeFolderId"
            :selection-mode="selectionMode"
            :selected-ids="selectedBookmarkIds"
            :is-image-value="isImageValue"
            :drag-over-id="dragType === 'folder' ? dragOverId : null"
            :drag-insert-position="dragType === 'folder' ? dragInsertPosition : ''"
            :drag-source-id="dragType === 'folder' ? dragSourceId : null"
            @select="emit('select-folder', $event)"
            @toggle="emit('toggle-folder', $event)"
            @context-menu="(event, item) => emit('folder-context-menu', event, item)"
            @drag-start="(item, event) => emit('folder-drag-start', item, event)"
            @drag-over="emit('folder-drag-over', $event)"
            @drag-end="emit('drag-end')"
            @drop="emit('folder-drop', $event)"
          />
        </TransitionGroup>
        <div v-if="!folders.length" class="empty-state compact-empty">暂无文件夹。</div>
      </nav>

      <main class="explorer-content bookmark-list">
        <header class="explorer-section-head">
          <div>
            <strong>{{ contentTitle }}</strong>
          </div>
          <small>{{ contentCount }}</small>
        </header>
        <template v-if="hasSearchQuery">
          <BookmarkRow
            v-for="bookmark in bookmarkSearch.results"
            :key="`search-${bookmark.id}`"
            :bookmark="bookmark"
            :selection-mode="selectionMode"
            :selected="isBookmarkSelected(bookmark)"
            path-fallback="搜索结果"
            :is-image-value="isImageValue"
            :drag-over="false"
            compact
            @toggle-selection="emit('toggle-bookmark-selection', $event)"
            @context-menu="(event, item) => emit('bookmark-context-menu', event, item)"
            @open="emit('open-bookmark', $event)"
          />
          <div v-if="!bookmarkSearch.loading && !bookmarkSearch.results.length" class="empty-state compact-empty">没有匹配的收藏。</div>
        </template>
        <template v-else>
          <BookmarkRow
            v-for="bookmark in bookmarks"
            :key="`${bookmark.id}-${bookmark.sort || 0}`"
            :bookmark="bookmark"
            :selection-mode="selectionMode"
            :selected="isBookmarkSelected(bookmark)"
            draggable
            :is-image-value="isImageValue"
            :drag-over="dragging && dragType === 'bookmark' && dragOverId === bookmark.id"
            :drag-insert-position="dragType === 'bookmark' ? dragInsertPosition : ''"
            compact
            @toggle-selection="emit('toggle-bookmark-selection', $event)"
            @context-menu="(event, item) => emit('bookmark-context-menu', event, item)"
            @drag-start="(item, event) => emit('bookmark-drag-start', item, event)"
            @drag-over="(item, event) => emit('bookmark-drag-over', item, event)"
            @drag-end="emit('drag-end')"
            @drop="(item, event) => emit('bookmark-drop', item, event)"
            @open="emit('open-bookmark', $event)"
          />
          <div v-if="activeFolderId && !bookmarks.length" class="empty-state compact-empty">这个文件夹还没有收藏。</div>
          <div v-if="!activeFolderId" class="empty-state compact-empty">选择左侧文件夹查看收藏。</div>
        </template>
      </main>
    </section>
  </aside>
</template>
