import { onBeforeUnmount, ref } from 'vue'
import { showFailToast } from 'vant'

import {
  hasTiandituKey,
  searchPoiInRegion,
  searchPoiNearby,
  type TiandituPoi,
} from '@/utils/map/tianditu'

export function useMapPoiSearch(opts: {
  center: () => { lat: number; lng: number }
  regionCode: () => string | undefined
}) {
  const searchKey = ref('')
  const searchLoading = ref(false)
  const searchResults = ref<TiandituPoi[]>([])
  const showSearchList = ref(false)
  let searchTimer: ReturnType<typeof setTimeout> | null = null

  function onSearchInput() {
    if (searchTimer) clearTimeout(searchTimer)
    const keyword = searchKey.value.trim()
    if (!keyword) {
      searchResults.value = []
      showSearchList.value = false
      return
    }
    searchTimer = setTimeout(() => {
      void runSearch(keyword)
    }, 400)
  }

  async function runSearch(keyword: string) {
    if (!hasTiandituKey()) return
    const text = keyword.trim()
    if (!text) return

    const { lat, lng } = opts.center()
    const regionCode = opts.regionCode()
    searchLoading.value = true
    showSearchList.value = true
    try {
      searchResults.value = regionCode
        ? await searchPoiInRegion(text, regionCode)
        : await searchPoiNearby(text, lat, lng, 20)
    } catch {
      searchResults.value = []
      showFailToast('定位失败')
    } finally {
      searchLoading.value = false
    }
  }

  function clearSearch() {
    showSearchList.value = false
    searchKey.value = ''
    searchResults.value = []
  }

  onBeforeUnmount(() => {
    if (searchTimer) clearTimeout(searchTimer)
  })

  return {
    searchKey,
    searchLoading,
    searchResults,
    showSearchList,
    onSearchInput,
    runSearch,
    clearSearch,
  }
}
