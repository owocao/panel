<script setup>
defineProps({
  item: { type: Object, default: null },
  dragStyle: { type: Object, default: () => ({}) },
  isImageValue: { type: Function, required: true },
  cardTextClass: { type: Function, required: true },
  limitText: { type: Function, required: true },
})
</script>

<template>
  <div v-if="item" class="nav-drag-float" :style="dragStyle">
    <span class="nav-card"><span v-if="isImageValue(item.icon)" class="card-icon image-icon"><img :src="item.icon" alt="" /></span><span v-else class="card-text-icon" :class="cardTextClass(item.icon || item.name)">{{ limitText(item.icon || item.name, 10) }}</span></span>
    <span class="card-title">{{ limitText(item.name, 10) }}</span>
  </div>
</template>
