import { defineStore } from 'pinia'
import { ref } from 'vue'

export type UserDrawerTab =
  | 'overview'
  | 'punishments'
  | 'posts'
  | 'signups'
  | 'conversations'
  | 'reports'
  | 'feedback'
  | 'devices'
  | 'support'

export const useUserDrawerStore = defineStore('userDrawer', () => {
  const visible = ref(false)
  const userId = ref(0)
  const version = ref(0)
  const changed = ref(0)
  const tab = ref<UserDrawerTab>('overview')

  function open(id: number, initialTab: UserDrawerTab = 'overview') {
    if (!id || id <= 0) return
    userId.value = id
    tab.value = initialTab
    version.value += 1
    visible.value = true
  }

  function close() {
    visible.value = false
  }

  function refresh() {
    version.value += 1
    changed.value += 1
  }

  return { visible, userId, version, changed, tab, open, close, refresh }
})
