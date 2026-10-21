import type { RouteLocationRaw, Router } from 'vue-router'

export function goBackOr(router: Router, fallback: RouteLocationRaw = { name: 'PostFeed' }) {
  const state = window.history.state as { back?: unknown } | null
  if (state?.back != null) {
    router.back()
    return
  }
  router.replace(fallback)
}
