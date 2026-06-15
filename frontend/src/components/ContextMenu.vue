<script setup>
defineProps({
  open: { type: Boolean, default: false },
  compact: { type: Boolean, default: false },
  actions: { type: Array, default: () => [] },
  menuStyle: { type: Object, default: () => ({}) },
  iconUrl: { type: Function, required: true },
})

const emit = defineEmits(['run'])
</script>

<template>
  <div v-if="open" class="context-menu" :class="{ compact }" :style="menuStyle" @click.stop>
    <div v-if="actions.some((action) => action.variant === 'icon')" class="menu-icon-row">
      <button v-for="action in actions.filter((item) => item.variant === 'icon')" :key="action.label" class="icon-only" type="button" :title="action.label" @click="emit('run', action)"><img :src="iconUrl(action.icon)" alt="" /><span class="visually-hidden">{{ action.label }}</span></button>
    </div>
    <template v-for="(action, index) in actions" :key="action.label || `divider-${index}`">
      <div v-if="action.divider" class="menu-divider"></div>
      <button v-else-if="action.variant !== 'icon'" type="button" @click="emit('run', action)"><img v-if="action.icon" :src="iconUrl(action.icon)" alt="" /><span>{{ action.label }}</span></button>
    </template>
  </div>
</template>
