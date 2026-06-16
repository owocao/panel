<script setup>
defineProps({
  active: {
    type: Boolean,
    required: true,
  },
  engines: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['add', 'edit', 'move', 'remove'])

function isImageValue(val) {
  return val && (val.startsWith('http') || val.startsWith('/'))
}
</script>

<template>
  <section v-if="active" class="setting-card manager-card">
    <header class="manager-head">
      <h3>搜索引擎</h3>
      <button type="button" @click="emit('add')">增加</button>
    </header>
    <article v-for="engine in engines" :key="`search-manage-${engine.key}`" class="manager-row search-engine-row">
      <span class="engine-mark">
        <img v-if="isImageValue(engine.icon)" :src="engine.icon" alt="" />
        <span v-else>{{ engine.icon || engine.title.slice(0, 1) }}</span>
      </span>
      <div>
        <strong>{{ engine.title }}</strong>
        <p>{{ engine.url }}</p>
      </div>
      <div class="inline-actions">
        <button type="button" @click="emit('edit', engine)">编辑</button>
        <button type="button" @click="emit('move', engine, -1)">上移</button>
        <button type="button" @click="emit('move', engine, 1)">下移</button>
        <button type="button" @click="emit('remove', engine)">删除</button>
      </div>
    </article>
    <div v-if="!engines.length" class="empty-state">
      暂无搜索引擎，点击增加创建。
    </div>
  </section>
</template>