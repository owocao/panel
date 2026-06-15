<script setup>
defineProps({
  settingsForm: { type: Object, required: true },
  displayTime: { type: String, default: '' },
  displayDate: { type: String, default: '' },
  dateMode: { type: String, default: 'solar' },
  webSearch: { type: Object, required: true },
  activeSearchEngine: { type: Object, default: null },
  searchEngines: { type: Array, default: () => [] },
  searchPickerOpen: { type: Boolean, default: false },
  isImageValue: { type: Function, required: true },
  iconUrl: { type: Function, required: true },
})

const emit = defineEmits(['toggle-date-mode', 'run-web-search', 'toggle-search-picker', 'select-search-engine', 'update-search-query'])
</script>

<template>
  <section class="hero-card">
    <div class="topbar"><div class="home-title"><h2><span v-if="settingsForm.showTitle === 'true'" class="title-text">{{ settingsForm.siteTitle || 'biu-panel' }}</span><template v-if="settingsForm.showTitle === 'true' && settingsForm.showClock === 'true'"><em>｜</em></template><time v-if="settingsForm.showClock === 'true'" class="time-stack"><strong>{{ displayTime }}</strong><small class="date-toggle" role="button" tabindex="0" :title="dateMode === 'solar' ? '点击切换农历' : '点击切换公历'" @click.stop="emit('toggle-date-mode')" @keyup.enter="emit('toggle-date-mode')">{{ displayDate }}</small></time></h2></div></div>
    <div v-if="settingsForm.showSearch === 'true'" class="search-wrap">
      <form class="sun-search" @submit.prevent="emit('run-web-search')">
        <button class="engine-button" type="button" @click="emit('toggle-search-picker')"><img v-if="isImageValue(activeSearchEngine?.icon)" :src="activeSearchEngine.icon" alt="" /><span v-else>{{ activeSearchEngine?.icon || 'G' }}</span></button>
        <input :value="webSearch.q" placeholder="网页搜索" @input="emit('update-search-query', $event.target.value)" />
        <button type="submit" title="搜索"><img class="button-icon" :src="iconUrl('search')" alt="" /></button>
      </form>
      <div v-if="searchPickerOpen" class="search-engine-picker"><button v-for="engine in searchEngines" :key="`picker-${engine.key}`" type="button" :title="engine.title" :class="{ active: engine.key === webSearch.engine }" @click="emit('select-search-engine', engine)"><img v-if="isImageValue(engine.icon)" :src="engine.icon" alt="" /><span v-else>{{ engine.icon || engine.title.slice(0, 1) }}</span></button></div>
    </div>
  </section>
</template>
