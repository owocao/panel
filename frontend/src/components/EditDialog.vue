<script setup>
const props = defineProps({
  dialog: {
    type: Object,
    required: true,
  },
  groupSelectOpen: {
    type: Boolean,
    default: false,
  },
  navGroupOptions: {
    type: Array,
    default: () => [],
  },
  folderManagementFlatList: {
    type: Array,
    default: () => [],
  },
  folderFlatList: {
    type: Array,
    default: () => [],
  },
  eligibleFolderParents: {
    type: Array,
    default: () => [],
  },
  metadataLoading: {
    type: Boolean,
    default: false,
  },
  editGroupName: {
    type: String,
    default: '请选择分组',
  },
})

const emit = defineEmits([
  'close',
  'save',
  'delete-nav-card',
  'clamp-field',
  'select-group',
  'set-icon-mode',
  'upload-icon',
  'fill-metadata',
  'fill-metadata-from-field',
  'create-group',
  'add-card',
  'create-folder',
  'create-bookmark',
  'update:groupSelectOpen',
  'panel-wheel',
])

function toggleGroupSelect() {
  emit('update:groupSelectOpen', !props.groupSelectOpen)
}
</script>

<template>
  <section v-if="dialog.open" class="modal-mask" @mousedown.self.stop="emit('close')" @wheel.stop.prevent>
    <form class="edit-modal" :class="{ 'bookmark-to-nav-modal': dialog.type === 'bookmarkToNav' }" @click.stop @wheel.stop="emit('panel-wheel', $event)" @submit.prevent="emit('save')">
      <header class="modal-head"><h2>{{ dialog.title }}</h2></header>

      <label v-if="dialog.type === 'navGroup' || dialog.type === 'navGroupCreate'">
        名称
        <span class="label-line"><small>{{ String(dialog.form.name || '').length }}/10</small></span>
        <input v-model="dialog.form.name" maxlength="10" placeholder="请输入分组名称" @input="emit('clamp-field', 'name', 10)" />
      </label>

      <template v-if="dialog.type === 'folder' || dialog.type === 'folderCreate'">
        <label>名称<input v-model="dialog.form.name" placeholder="请输入收藏夹名称" /></label>
        <label v-if="dialog.type === 'folderCreate'">上级收藏夹
          <select v-model="dialog.form.parentId">
            <option :value="null">根目录</option>
            <option v-for="folder in folderManagementFlatList.filter((item) => item.depth < 3)" :key="`folder-parent-${folder.id}`" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
          </select>
        </label>
        <label v-if="dialog.type === 'folder'">移动
          <select v-model="dialog.form.parentId">
            <option :value="null">根目录</option>
            <option v-for="folder in eligibleFolderParents" :key="`folder-move-parent-${folder.id}`" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
          </select>
        </label>
      </template>

      <template v-if="dialog.type === 'navItem' || dialog.type === 'navItemCreate'">
        <label>
          <span class="label-line"><span>标题 <em class="required">*</em></span><small>{{ String(dialog.form.name || '').length }}/15</small></span>
          <input v-model="dialog.form.name" maxlength="15" required placeholder="请输入标题" @input="emit('clamp-field', 'name', 15)" />
        </label>
        <div class="icon-mode-block">
          <span class="icon-mode-title">图标风格</span>
          <div class="segmented">
            <button type="button" :class="{ active: dialog.form.iconMode !== 'image' }" @click="emit('set-icon-mode', dialog.form, 'text')">文字</button>
            <button type="button" :class="{ active: dialog.form.iconMode === 'image' }" @click="emit('set-icon-mode', dialog.form, 'image')">图片</button>
          </div>
          <label>
            <span class="label-line"><span>{{ dialog.form.iconMode === 'image' ? '图片地址' : '文本内容' }}</span><small v-if="dialog.form.iconMode !== 'image'">{{ String(dialog.form.icon || '').length }}/5</small></span>
            <span class="input-with-button">
              <input v-model="dialog.form.icon" :maxlength="dialog.form.iconMode === 'image' ? undefined : 5" :placeholder="dialog.form.iconMode === 'image' ? '输入图标地址或上传' : '请输入文本内容'" @input="dialog.form.iconMode !== 'image' && emit('clamp-field', 'icon', 5)" />
              <label v-if="dialog.form.iconMode === 'image'" class="upload-inline">上传<input type="file" accept="image/*" @change="emit('upload-icon', $event, dialog.form, 'icon')" /></label>
            </span>
          </label>
        </div>
        <label>
          <span class="label-line"><span>分组 <em class="required">*</em></span></span>
          <div class="select-popover" :class="{ open: groupSelectOpen }">
            <button type="button" class="select-trigger" @click="toggleGroupSelect">
              <span>{{ editGroupName }}</span>
              <span class="select-arrow">⌄</span>
            </button>
            <div v-if="groupSelectOpen" class="select-options">
              <button v-for="group in navGroupOptions" :key="`edit-group-${group.id}`" type="button" :class="{ active: group.id === dialog.form.groupId }" @click.stop="emit('select-group', group)">{{ group.name }}</button>
            </div>
          </div>
        </label>
        <label>
          <span class="label-line"><span>公网地址 <em class="required">*</em></span></span>
          <span class="input-with-button metadata-inline">
            <input v-model="dialog.form.wanUrl" required placeholder="https://example.com" />
            <button class="field-action" type="button" @click="emit('fill-metadata-from-field', dialog.form, 'wanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button>
          </span>
        </label>
        <label>
          <span class="label-line"><span>内网地址</span></span>
          <span class="input-with-button metadata-inline">
            <input v-model="dialog.form.lanUrl" placeholder="http://192.168.x.x" />
            <button class="field-action" type="button" @click="emit('fill-metadata-from-field', dialog.form, 'lanUrl')">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button>
          </span>
        </label>
      </template>

      <template v-if="dialog.type === 'bookmark' || dialog.type === 'bookmarkCreate'">
        <label>标题<input v-model="dialog.form.title" /></label>
        <label><span class="label-line"><span>网址 <b class="required">*</b></span></span><input v-model="dialog.form.url" required /></label>
        <label>备注<input v-model="dialog.form.note" /></label>
        <label>所属文件夹
          <select v-model="dialog.form.folderId">
            <option v-for="folder in folderFlatList" :key="`edit-bookmark-folder-${folder.id}`" :value="folder.id">{{ '　'.repeat(folder.depth) }}{{ folder.name }}</option>
          </select>
        </label>
        <button type="button" @click="emit('fill-metadata', dialog.form)">{{ metadataLoading ? '抓取中' : '自动抓取标题/图标' }}</button>
      </template>

      <template v-if="dialog.type === 'bookmarkToNav'">
        <label>
          目标分组
          <div class="select-popover" :class="{ open: groupSelectOpen }">
            <button type="button" class="select-trigger" @click="toggleGroupSelect"><span>{{ editGroupName }}</span><span class="select-arrow">⌄</span></button>
            <div v-if="groupSelectOpen" class="select-options">
              <button v-for="group in navGroupOptions" :key="`bookmark-nav-group-${group.id}`" type="button" :class="{ active: group.id === dialog.form.groupId }" @click.stop="emit('select-group', group)">{{ group.name }}</button>
            </div>
          </div>
        </label>
        <p class="muted">将收藏「{{ dialog.form.bookmark?.title }}」复制为首页导航卡片。</p>
      </template>

      <template v-if="dialog.type === 'searchEngine' || dialog.type === 'searchEngineCreate'">
        <label>标题<input v-model="dialog.form.title" placeholder="例如 Google" /></label>
        <label>URL<input v-model="dialog.form.url" placeholder="https://example.com/search?q={q}" /></label>
        <div class="icon-mode-block">
          <span class="icon-mode-title">图标风格</span>
          <div class="segmented">
            <button type="button" :class="{ active: dialog.form.iconMode !== 'image' }" @click="emit('set-icon-mode', dialog.form, 'text')">文字</button>
            <button type="button" :class="{ active: dialog.form.iconMode === 'image' }" @click="emit('set-icon-mode', dialog.form, 'image')">图片</button>
          </div>
          <label>
            <span class="label-line"><span>{{ dialog.form.iconMode === 'image' ? '图片地址' : '文本内容' }}</span></span>
            <span class="input-with-button">
              <input v-model="dialog.form.icon" :placeholder="dialog.form.iconMode === 'image' ? '输入图标地址或上传' : '请输入文本内容'" />
              <label v-if="dialog.form.iconMode === 'image'" class="upload-inline">上传<input type="file" accept="image/*" @change="emit('upload-icon', $event, dialog.form, 'icon')" /></label>
            </span>
          </label>
        </div>
      </template>

      <footer class="modal-actions">
        <button type="submit">保存</button>
        <button v-if="dialog.type !== 'navItemCreate' && dialog.type !== 'navGroupCreate' && dialog.type !== 'searchEngineCreate' && dialog.type !== 'folderCreate' && dialog.type !== 'bookmarkCreate'" type="button" @click="emit('close')">取消</button>
        <button v-if="dialog.type === 'navItem'" class="danger" type="button" @click="emit('delete-nav-card')">删除</button>
      </footer>
    </form>
  </section>
</template>
