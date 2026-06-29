<script setup>
import BackupRestoreSection from '../BackupRestoreSection.vue'
import PersonalSettingsForm from '../PersonalSettingsForm.vue'
import SearchEngineManagerSection from '../SearchEngineManagerSection.vue'
import SettingsMenu from '../SettingsMenu.vue'
import BookmarkManager from './BookmarkManager.vue'

defineProps({
  open: { type: Boolean, default: false },
  activeSettings: { type: String, default: '个性化' },
  menuCollapsed: { type: Boolean, default: false },
  message: { type: String, default: '' },
  saving: { type: Boolean, default: false },
  settingsDraft: { type: Object, required: true },
  navGroupsDraft: { type: Array, default: () => [] },
  folderManagementFlatList: { type: Array, default: () => [] },
  settingsSearchEngines: { type: Array, default: () => [] },
})

const emit = defineEmits([
  'close',
  'save',
  'toggle-menu',
  'select-menu',
  'panel-wheel',
  'create-group',
  'create-card',
  'edit-group',
  'reorder-group',
  'remove-group',
  'create-folder',
  'open-drawer',
  'edit-folder',
  'request-move-folder',
  'reorder-folder',
  'remove-folder',
  'add-search-engine',
  'edit-search-engine',
  'reorder-search-engine',
  'remove-search-engine',
  'test-s3',
  'restore-global',
  'download-nav',
  'restore-nav',
  'export-bookmarks',
  'import-bookmarks',
])
</script>

<template>
  <section v-if="open" class="settings-mask" @mousedown.self.stop="emit('close')" @wheel.stop.prevent>
    <section class="settings-panel settings-center" @click.stop @wheel.stop="emit('panel-wheel', $event)">
      <header class="settings-head">
        <div><h2>系统设置</h2></div>
        <div class="inline-actions">
          <button type="button" :disabled="saving" @click="emit('save')">{{ saving ? '保存中...' : '保存' }}</button>
          <button type="button" :disabled="saving" @click="emit('close')">关闭</button>
        </div>
      </header>
      <p v-if="message" class="settings-message">{{ message }}</p>
      <div class="settings-layout" :class="{ collapsed: menuCollapsed }">
        <SettingsMenu :collapsed="menuCollapsed" :active="activeSettings" @toggle-collapse="emit('toggle-menu')" @select="emit('select-menu', $event)" />
        <div class="settings-content">
          <section v-if="activeSettings === '分组管理'" class="setting-card manager-card">
            <header class="manager-head">
              <h3>分组管理</h3>
              <button type="button" @click="emit('create-group')">新增分组</button>
            </header>
            <article v-for="group in navGroupsDraft" :key="`manage-${group.id}`" class="manager-row">
              <div>
                <strong>{{ group.name }}</strong>
                <p>{{ group.items?.length || 0 }} 张卡片</p>
              </div>
              <div class="inline-actions">
                <button type="button" @click="emit('create-card', group)">新增卡片</button>
                <button type="button" @click="emit('edit-group', group)">编辑</button>
                <button type="button" @click="emit('reorder-group', group, -1)">上移</button>
                <button type="button" @click="emit('reorder-group', group, 1)">下移</button>
                <button type="button" @click="emit('remove-group', group, true)">删除</button>
              </div>
            </article>
            <div v-if="!navGroupsDraft.length" class="empty-state">暂无导航分组</div>
          </section>

          <BookmarkManager
            :active="activeSettings === '收藏夹'"
            :folders="folderManagementFlatList"
            @create-folder="emit('create-folder', $event)"
            @open-drawer="emit('open-drawer')"
            @edit-folder="emit('edit-folder', $event)"
            @request-move-folder="emit('request-move-folder', $event)"
            @reorder-folder="(folder, offset) => emit('reorder-folder', folder, offset)"
            @remove-folder="emit('remove-folder', $event)"
          />

          <SearchEngineManagerSection
            :active="activeSettings === '搜索引擎'"
            :engines="settingsSearchEngines"
            @add="emit('add-search-engine')"
            @edit="emit('edit-search-engine', $event)"
            @move="(engine, offset) => emit('reorder-search-engine', engine, offset)"
            @remove="emit('remove-search-engine', $event)"
          />

          <section v-if="activeSettings === '关于'" class="setting-card">
            <h3>关于</h3>
            <p>个人自用导航站和网页收藏夹。</p>
          </section>

          <div class="settings-grid">
            <PersonalSettingsForm :active="activeSettings === '个性化'" :draft="settingsDraft" />

            <section v-show="activeSettings === 'S3 存储'" class="setting-card settings-card-wide">
              <h3>S3 存储</h3>
              <label class="check-row"><input v-model="settingsDraft.s3Enabled" true-value="true" false-value="false" type="checkbox" /> 启用 S3 配置</label>
              <label>Endpoint<input v-model="settingsDraft.s3Endpoint" placeholder="https://s3.example.com" /></label>
              <label>Region<input v-model="settingsDraft.s3Region" placeholder="auto" /></label>
              <label>Bucket<input v-model="settingsDraft.s3Bucket" placeholder="biu-panel" /></label>
              <label>Access Key<input v-model="settingsDraft.s3AccessKey" /></label>
              <label>Secret Key<input v-model="settingsDraft.s3SecretKey" type="password" /></label>
              <label>上传前缀<input v-model="settingsDraft.s3Prefix" placeholder="biu-panel/" /></label>
              <label>公开访问域名<input v-model="settingsDraft.s3PublicBase" placeholder="https://cdn.example.com/biu-panel" /></label>
              <label class="check-row"><input v-model="settingsDraft.s3PathStyle" true-value="true" false-value="false" type="checkbox" /> Path-style 兼容模式</label>
              <div class="inline-actions">
                <button type="button" :disabled="saving" @click="emit('save')">{{ saving ? '保存中...' : '保存 S3 配置' }}</button>
                <button type="button" :disabled="saving" @click="emit('test-s3')">测试 S3</button>
              </div>
            </section>

            <BackupRestoreSection
              :active="activeSettings === '备份恢复'"
              @restore-global="emit('restore-global', $event)"
              @download-nav="emit('download-nav')"
              @restore-nav="emit('restore-nav', $event)"
              @export-bookmarks="emit('export-bookmarks')"
              @import-bookmarks="emit('import-bookmarks', $event)"
            />
          </div>
        </div>
      </div>
    </section>
  </section>
</template>
