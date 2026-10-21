import { onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

interface RouteQueryField<Q extends Record<string, unknown>> {
  key: string
  apply: (value: string, query: Q) => void
  clear: (query: Q) => void
}

function readQueryParam(query: Record<string, unknown>, key: string): string | undefined {
  const raw = query[key]
  if (raw === undefined || raw === null || raw === '') return undefined
  if (Array.isArray(raw)) return raw[0] ? String(raw[0]) : undefined
  return String(raw)
}

export function useRouteQuery<Q extends Record<string, unknown>>(
  query: Q,
  fields: RouteQueryField<Q>[],
  options?: {
    onApplied?: () => void
  },
) {
  const route = useRoute()
  const router = useRouter()
  const routeKeys = fields.map((f) => f.key)
  let skipApply = false

  function applyFromRoute() {
    if (skipApply) {
      skipApply = false
      return
    }
    for (const field of fields) {
      const value = readQueryParam(route.query as Record<string, unknown>, field.key)
      if (value === undefined) {
        field.clear(query)
      } else {
        field.apply(value, query)
      }
    }
    options?.onApplied?.()
  }

  function clearRouteQuery() {
    const hasAny = routeKeys.some((key) => route.query[key] !== undefined)
    if (!hasAny) return
    const nextQuery = { ...route.query }
    for (const key of routeKeys) {
      delete nextQuery[key]
    }
    router.replace({ path: route.path, query: nextQuery })
  }

  function wrapReset(resetFn: () => void) {
    return () => {
      const hasAny = routeKeys.some((key) => route.query[key] !== undefined)
      if (hasAny) skipApply = true
      clearRouteQuery()
      resetFn()
    }
  }

  onMounted(applyFromRoute)
  watch(() => route.query, applyFromRoute)

  return { wrapReset }
}
