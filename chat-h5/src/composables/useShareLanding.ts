import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { useRoute } from 'vue-router'

import { isShareTraffic, type ShareResource, shareResourcePath } from '@/constants/shareLinks'
import { isInAppShell } from '@/utils/appShell'
import { buildDownloadUrl, tryOpenApp } from '@/utils/share/openApp'

export function useShareLanding(resource: MaybeRefOrGetter<ShareResource>) {
  const route = useRoute()
  const resolved = computed(() => toValue(resource))
  const path = computed(() => shareResourcePath(resolved.value))
  const isShare = computed(() => isShareTraffic(route.query))
  const inApp = computed(() => isInAppShell(route.query))
  const showOpenApp = computed(() => isShare.value && !inApp.value)
  const downloadHref = computed(() => buildDownloadUrl(path.value.split('?')[0] ?? path.value))

  function openApp() {
    tryOpenApp(resolved.value)
  }

  return {
    inApp,
    showOpenApp,
    downloadHref,
    openApp,
  }
}
