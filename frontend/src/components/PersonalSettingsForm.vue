<script setup>
defineProps({
  active: {
    type: Boolean,
    required: true,
  },
  draft: {
    type: Object,
    required: true,
  },
})
</script>

<template>
  <section v-show="active" class="setting-card settings-card-wide">
    <h3>个性化</h3>
    <label class="check-row">
      <input v-model="draft.showTitle" true-value="true" false-value="false" type="checkbox" /> 显示标题
    </label>
    <label>
      首页文本
      <input v-model="draft.siteTitle" />
    </label>
    <label class="check-row">
      <input v-model="draft.showClock" true-value="true" false-value="false" type="checkbox" /> 显示时钟
    </label>
    <label class="check-row">
      <input v-model="draft.showSeconds" true-value="true" false-value="false" type="checkbox" /> 显示秒
    </label>
    <label class="check-row">
      <input v-model="draft.showSearch" true-value="true" false-value="false" type="checkbox" /> 显示搜索栏
    </label>
    <label>
      内外网超时时间 ms
      <input v-model="draft.lanDetectTimeout" />
    </label>
  </section>
</template>
