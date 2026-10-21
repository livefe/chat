import { defineStore } from 'pinia'
import { reactive } from 'vue'

import { createChatViewLocation } from './chatView'
import { createLocationDrafts } from './drafts'
import { createGpsLocation } from './gps'

export type { ChatViewMode } from './chatView'
export type { LocationPoint, LocationShareTarget, ManualLocationSource } from './types'

export const useLocationStore = defineStore('location', () => {
  const gps = reactive(createGpsLocation())
  const chatView = reactive(createChatViewLocation(gps))

  let browseCommittedHandler: () => void = () => {}

  function setBrowseCommittedHandler(handler: () => void) {
    browseCommittedHandler = handler
  }

  const drafts = createLocationDrafts({
    gps,
    chatView,
    onBrowseCommitted: () => browseCommittedHandler(),
  })

  function clearSession() {
    gps.clearSession()
    chatView.clearSession()
    drafts.clearSession()
  }

  return {
    gps,
    chatView,
    publish: drafts.publish,
    room: drafts.room,
    picker: drafts.picker,
    beginShare: drafts.beginShare,
    consumeSharePending: drafts.consumeSharePending,
    setBrowseCommittedHandler,
    clearSession,
  }
})
