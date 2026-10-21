import { reactive, ref } from 'vue'

import { ROOM_RADIUS_DEFAULT_METERS } from '@/constants/nearby'
import { hasCoordinates } from '@/utils/map/coord'
import type { LocationPickPurpose } from '@/utils/map/locationPick'
import { realPlaceNameOrEmpty } from '@/utils/map/placeName'

import type { GpsReading, LocationPoint, LocationShareTarget, ManualLocationSource } from './types'

const DEFAULT_RADIUS = ROOM_RADIUS_DEFAULT_METERS

export interface LocationDraftsDeps {
  gps: GpsReading
  chatView: {
    committedPoint: () => (LocationPoint & { browsing: boolean }) | null
    browsePoint: () => LocationPoint | null
    setBrowseLocation: (next: LocationPoint) => void
  }
  onBrowseCommitted: () => void
}

export function createLocationDrafts(deps: LocationDraftsDeps) {
  function readChatViewLocation(): (LocationPoint & { source: ManualLocationSource }) | null {
    const point = deps.chatView.committedPoint()
    if (!point || !hasCoordinates(point.latitude, point.longitude)) return null
    return {
      placeName: realPlaceNameOrEmpty(point.placeName),
      latitude: point.latitude,
      longitude: point.longitude,
      source: point.browsing ? 'manual' : 'gps',
    }
  }

  function createManualLocationState() {
    const placeName = ref('')
    const latitude = ref(0)
    const longitude = ref(0)
    const ready = ref(false)
    const error = ref('')
    const source = ref<ManualLocationSource | null>(null)

    function setLocation(next: LocationPoint, nextSource: ManualLocationSource = 'manual') {
      placeName.value = realPlaceNameOrEmpty(next.placeName)
      latitude.value = next.latitude
      longitude.value = next.longitude
      ready.value = hasCoordinates(next.latitude, next.longitude)
      error.value = ''
      source.value = nextSource
    }

    function clear() {
      placeName.value = ''
      latitude.value = 0
      longitude.value = 0
      ready.value = false
      error.value = ''
      source.value = null
    }

    function applyChatView(): boolean {
      const chat = readChatViewLocation()
      if (!chat) return false
      setLocation(chat, chat.source)
      return ready.value
    }

    async function initFromGps(): Promise<boolean> {
      await deps.gps.ensureFreshLocation()

      if (ready.value) return true

      if (!deps.gps.valid) {
        const chat = readChatViewLocation()
        if (chat) {
          setLocation(chat, chat.source)
          return ready.value
        }
        ready.value = false
        error.value = '定位失败，请手动选择位置'
        return false
      }

      const lat = deps.gps.lat
      const lng = deps.gps.lng
      const chat = readChatViewLocation()
      const reuseName =
        chat && Math.abs(chat.latitude - lat) < 1e-6 && Math.abs(chat.longitude - lng) < 1e-6
          ? chat.placeName
          : ''
      setLocation({ placeName: reuseName, latitude: lat, longitude: lng }, 'gps')
      return true
    }

    async function initDefault() {
      if (ready.value) return true
      if (applyChatView()) return true
      return initFromGps()
    }

    function snapshot(): (LocationPoint & { source: ManualLocationSource }) | null {
      if (!ready.value) return null
      return {
        placeName: placeName.value,
        latitude: latitude.value,
        longitude: longitude.value,
        source: source.value ?? 'manual',
      }
    }

    return {
      placeName,
      latitude,
      longitude,
      ready,
      error,
      source,
      setLocation,
      clear,
      initFromGps,
      initDefault,
      snapshot,
    }
  }

  const publishCore = createManualLocationState()

  const publish = reactive({
    placeName: publishCore.placeName,
    latitude: publishCore.latitude,
    longitude: publishCore.longitude,
    ready: publishCore.ready,
    error: publishCore.error,
    source: publishCore.source,
    setLocation: publishCore.setLocation,
    clear: publishCore.clear,
    initDefault: publishCore.initDefault,
    initFromPost(post: LocationPoint) {
      publishCore.setLocation(post, 'manual')
    },
  })

  const roomCore = createManualLocationState()
  const name = ref('')
  const description = ref('')
  const placeType = ref<number | null>(null)
  const radius = ref(DEFAULT_RADIUS)

  function clearRoom() {
    roomCore.clear()
    name.value = ''
    description.value = ''
    placeType.value = null
    radius.value = DEFAULT_RADIUS
  }

  const room = reactive({
    placeName: roomCore.placeName,
    latitude: roomCore.latitude,
    longitude: roomCore.longitude,
    ready: roomCore.ready,
    error: roomCore.error,
    source: roomCore.source,
    name,
    description,
    placeType,
    radius,
    clear: clearRoom,
    initDefault: roomCore.initDefault,
  })

  const pickerCore = createManualLocationState()
  const purpose = ref<LocationPickPurpose | null>(null)
  const shareTarget = ref<LocationShareTarget | null>(null)
  const sharePending = ref<{ loc: LocationPoint; target: LocationShareTarget } | null>(null)

  function clearPicker() {
    purpose.value = null
    pickerCore.clear()
  }

  function beginShare(target: LocationShareTarget) {
    shareTarget.value = target
  }

  function consumeSharePending(): { loc: LocationPoint; target: LocationShareTarget } | null {
    const next = sharePending.value
    sharePending.value = null
    return next
  }

  function seedPicker(nextPurpose: LocationPickPurpose): boolean {
    if (nextPurpose === 'chat') {
      const browse = deps.chatView.browsePoint()
      if (!browse) return false
      pickerCore.setLocation(browse, 'manual')
      return true
    }
    if (nextPurpose !== 'publish' && nextPurpose !== 'room') return false
    const draft = (nextPurpose === 'publish' ? publishCore : roomCore).snapshot()
    const seed = draft ?? readChatViewLocation()
    if (!seed) return false
    pickerCore.setLocation(seed, seed.source)
    return true
  }

  function initPicker(nextPurpose: LocationPickPurpose) {
    purpose.value = nextPurpose
    pickerCore.clear()
    if (seedPicker(nextPurpose)) return
    void pickerCore.initFromGps()
  }

  function commitPicker() {
    if (!purpose.value || !pickerCore.ready.value) return
    const loc = {
      placeName: realPlaceNameOrEmpty(pickerCore.placeName.value),
      latitude: pickerCore.latitude.value,
      longitude: pickerCore.longitude.value,
    }
    if (!hasCoordinates(loc.latitude, loc.longitude)) return
    const src = pickerCore.source.value ?? 'manual'

    if (purpose.value === 'chat') {
      deps.chatView.setBrowseLocation(loc)
      deps.onBrowseCommitted()
      return
    }
    if (purpose.value === 'share') {
      sharePending.value = {
        loc,
        target: shareTarget.value ?? { type: 'nearby' },
      }
      shareTarget.value = null
      return
    }
    if (purpose.value === 'publish') {
      publishCore.setLocation(loc, src)
      return
    }
    roomCore.setLocation(loc, src)
  }

  const picker = reactive({
    placeName: pickerCore.placeName,
    latitude: pickerCore.latitude,
    longitude: pickerCore.longitude,
    ready: pickerCore.ready,
    error: pickerCore.error,
    source: pickerCore.source,
    purpose,
    setLocation: pickerCore.setLocation,
    initFromGps: pickerCore.initFromGps,
    clear: clearPicker,
    init: initPicker,
    commit: commitPicker,
  })

  function clearSession() {
    publish.clear()
    room.clear()
    picker.clear()
    shareTarget.value = null
    sharePending.value = null
  }

  return { publish, room, picker, beginShare, consumeSharePending, clearSession }
}
