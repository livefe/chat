import type { Pinia } from 'pinia'

import { useChatStore } from './chat'
import { useLocationStore } from './location'
import { usePostDraftStore } from './postDraft'
import { useSupportStore } from './support'
import { useUserStore } from './user'

export function wireStores(pinia: Pinia) {
  const userStore = useUserStore(pinia)
  const chatStore = useChatStore(pinia)
  const locationStore = useLocationStore(pinia)
  const supportStore = useSupportStore(pinia)
  const postDraftStore = usePostDraftStore(pinia)

  chatStore.setLocationSource(() => locationStore.chatView.committedCoords())
  chatStore.setSupportEventHandler((env) => supportStore.handleWs(env))
  locationStore.setBrowseCommittedHandler(() => chatStore.clearNearbyMessages())
  userStore.setLogoutHandler(() => {
    chatStore.disconnect({ clearSession: true })
    locationStore.clearSession()
    postDraftStore.reset()
  })
}
