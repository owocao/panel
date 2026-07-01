<script setup>
defineProps({
  collapsed: { type: Boolean, default: false },
  active: { type: String, default: '个性化' },
})

const emit = defineEmits(['toggle-collapse', 'select'])

const items = ['个性化', '搜索引擎', '分组管理', '收藏夹', '备份恢复', 'S3 存储', '关于']
</script>

<template>
  <button class="menu-collapse" type="button" :aria-expanded="!collapsed" :title="collapsed ? '展开菜单' : '收起菜单'" @click="emit('toggle-collapse')">{{ collapsed ? '›' : '‹' }}</button>
  <aside v-if="!collapsed" class="settings-menu" aria-label="设置菜单">
    <button v-for="item in items" :key="item" type="button" :aria-current="active === item ? 'page' : undefined" :class="{ active: active === item }" @click="emit('select', item)">{{ item }}</button>
  </aside>
</template>
