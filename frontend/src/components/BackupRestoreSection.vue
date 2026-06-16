<script setup>
defineProps({
  active: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['restore-global', 'download-nav', 'restore-nav', 'export-bookmarks', 'import-bookmarks'])
</script>

<template>
  <section v-show="active" class="setting-card backup-card settings-card-wide">
    <h3>备份恢复</h3>
    <div class="backup-zone">
      <h4>全部备份</h4>
      <p>全局备份包含导航页和收藏夹；恢复时按备份包内容恢复对应数据。</p>
      <div class="inline-actions backup-actions">
        <button type="button" @click="window.location.href = '/api/backup/download'">全局备份</button>
        <label class="file-button">全局恢复<input type="file" accept=".gz,.tgz,application/gzip" @change="emit('restore-global', $event)" /></label>
      </div>
    </div>
    <div class="backup-zone">
      <h4>导航页</h4>
      <p>导航页数据包含分组、卡片和排序。</p>
      <div class="inline-actions backup-actions">
        <button type="button" @click="emit('download-nav')">备份</button>
        <label class="file-button">恢复<input type="file" accept=".json,application/json" @change="emit('restore-nav', $event)" /></label>
      </div>
    </div>
    <div class="backup-zone">
      <h4>收藏夹</h4>
      <p>收藏夹导入导出集中在这里。</p>
      <div class="inline-actions backup-actions">
        <button type="button" @click="emit('export-bookmarks')">导出</button>
        <label class="file-button">导入<input type="file" accept=".html,.htm,text/html" @change="emit('import-bookmarks', $event)" /></label>
      </div>
    </div>
  </section>
</template>