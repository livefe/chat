import { computed, onScopeDispose } from 'vue'

import { NEARBY_COPY } from '@/constants/nearby'
import { useChatStore } from '@/stores/chat'
import { useLocationStore } from '@/stores/location'

const NEARBY_PRESENCE_INTERVAL_MS = 30_000

export function useNearbyChatBootstrap() {
  const locationStore = useLocationStore()
  const chatStore = useChatStore()
  const gps = locationStore.gps
  const chatView = locationStore.chatView

  let lastViewKey = chatView.hasCommittedLocation ? chatView.viewKey : ''
  let presenceTimer: ReturnType<typeof setInterval> | null = null

  const hasLocation = computed(() => chatView.hasCommittedLocation)
  const isBrowsing = computed(() => chatView.isBrowsing)

  const locationLabel = computed(() => {
    if (gps.locating || chatView.gpsResolving) return NEARBY_COPY.locating
    if (!chatView.hasCommittedLocation) {
      return gps.error ? NEARBY_COPY.locateFailed : NEARBY_COPY.chooseLocation
    }
    return chatView.headerPlaceLabel
  })

  async function ensureReady(): Promise<boolean> {
    if (!chatView.isBrowsing) {
      await gps.ensureFreshLocation()
    }
    if (!chatView.hasCommittedLocation && !chatView.canCommitLocation) return false
    chatStore.connect()
    return true
  }

  async function activate(): Promise<void> {
    const ready = await ensureReady()
    if (!ready) return

    const locationChanged = chatView.viewKey !== lastViewKey
    const needCommit = locationChanged || !chatView.hasCommittedLocation

    if (needCommit && !chatView.commitActiveLocation()) return

    lastViewKey = chatView.viewKey
    if (locationChanged) chatStore.clearNearbyMessages()
    if (chatStore.activeChannel === 'room') chatStore.leaveRoom()

    chatView.ensureGpsPlaceResolved()

    const keepMessages =
      !locationChanged &&
      chatStore.activeChannel === 'nearby' &&
      chatStore.currentMessages.length > 0
    await chatStore.refreshNearby({
      includeMessages: !keepMessages,
      replaceMessages: locationChanged,
      syncServerLocation: needCommit,
    })
  }

  async function relocateToGps(): Promise<boolean> {
    chatStore.clearNearbyMessages()
    await gps.refresh()
    if (!gps.valid) return false
    if (!(await chatView.relocateToGps())) return false
    lastViewKey = chatView.viewKey
    await chatStore.refreshNearby({
      includeMessages: true,
      replaceMessages: true,
      syncServerLocation: true,
    })
    return true
  }

  function refreshRooms() {
    return chatStore.refreshNearby({
      includeMessages: false,
      replaceMessages: false,
      syncServerLocation: false,
    })
  }

  function stopPresencePolling() {
    if (!presenceTimer) return
    clearInterval(presenceTimer)
    presenceTimer = null
  }

  function startPresencePolling() {
    stopPresencePolling()
    presenceTimer = setInterval(() => {
      if (document.visibilityState === 'hidden') return
      if (!chatView.hasCommittedLocation) return
      void chatStore.refreshNearby({ includeMessages: false })
    }, NEARBY_PRESENCE_INTERVAL_MS)
  }

  onScopeDispose(stopPresencePolling)

  return {
    hasLocation,
    isBrowsing,
    locationLabel,
    activate,
    relocateToGps,
    refreshRooms,
    startPresencePolling,
    stopPresencePolling,
  }
}
