<script setup>
defineProps({
  showNetworkSwitcher: { type: Boolean, default: true },
  networkTip: { type: String, default: '' },
  networkIcon: { type: String, default: '' },
  toastText: { type: String, default: '' },
  iconUrl: { type: Function, required: true },
})

const emit = defineEmits(['open-drawer', 'cycle-network-mode', 'open-settings'])
</script>

<template>
  <button class="bookmark-tab" type="button" @click.stop="emit('open-drawer')">收藏夹</button>
  <div class="floating-actions">
    <button v-if="showNetworkSwitcher" type="button" :title="networkTip" @click.stop="emit('cycle-network-mode')"><img :src="iconUrl(networkIcon)" alt="" /></button>
    <button type="button" title="设置" @click.stop="emit('open-settings')"><img :src="iconUrl('setting')" alt="" /></button>
  </div>
  <div v-if="toastText" class="toast-message">{{ toastText }}</div>
</template>
