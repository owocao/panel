import { computed, onMounted, onUnmounted, ref } from 'vue'
import { formatDisplayDate, formatDisplayTime } from '../utils/display'

export function useHomeClock({ settingsForm }) {
  const now = ref(new Date())
  const dateMode = ref('solar')
  let clockTimer

  const displayTime = computed(() => formatDisplayTime(now.value, settingsForm.value.showSeconds === 'true'))
  const displayDate = computed(() => formatDisplayDate(now.value, dateMode.value))

  function toggleDateMode() {
    dateMode.value = dateMode.value === 'solar' ? 'lunar' : 'solar'
  }

  onMounted(() => {
    clockTimer = window.setInterval(() => { now.value = new Date() }, 1000)
  })

  onUnmounted(() => {
    if (clockTimer) window.clearInterval(clockTimer)
  })

  return {
    dateMode,
    displayTime,
    displayDate,
    toggleDateMode,
  }
}
