import { onMounted, reactive, ref, watch, type Ref } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'

import { showApiError } from '@/utils/apiError'

export type ListFieldType = 'string' | 'number' | 'boolean'

interface PageResult<T> {
  data: { list: T[]; total: number }
}

interface BaseQuery {
  page: number
  pageSize: number
}

export interface ListQueryOptions<T, Q extends BaseQuery> {
  fetch: (query: Q) => Promise<PageResult<T>>
  defaults: Q
  schema?: Partial<Record<Exclude<keyof Q, keyof BaseQuery>, ListFieldType>>
  errorMessage?: string
  syncUrl?: boolean
  immediate?: boolean
}

function firstValue(raw: LocationQuery[string] | undefined): string | undefined {
  const v = Array.isArray(raw) ? raw[0] : raw
  if (v === null || v === undefined || v === '') return undefined
  return String(v)
}

function parseValue(raw: string, type: ListFieldType): unknown {
  if (type === 'number') {
    const n = Number(raw)
    return Number.isFinite(n) ? n : undefined
  }
  if (type === 'boolean') return raw === '1' || raw === 'true'
  return raw
}

function isEmpty(v: unknown) {
  return v === undefined || v === null || (typeof v === 'string' && v.trim() === '')
}

export function useListQuery<T, Q extends BaseQuery>(options: ListQueryOptions<T, Q>) {
  const route = useRoute()
  const router = useRouter()
  const syncUrl = options.syncUrl ?? true
  const defaults = { ...options.defaults }
  const schema = (options.schema ?? {}) as Record<string, ListFieldType>
  const managedKeys = [...Object.keys(schema), 'page', 'pageSize']
  const fieldType = (key: string): ListFieldType =>
    key === 'page' || key === 'pageSize' ? 'number' : (schema[key] ?? 'string')

  const query = reactive({ ...defaults }) as Q
  const loading = ref(false)
  const error = ref(false)
  const list = ref([]) as Ref<T[]>
  const total = ref(0)

  let requestSeq = 0
  let mountedPath = ''
  let lastWritten = ''

  function serialize(): Record<string, string> {
    const out: Record<string, string> = {}
    const q = query as Record<string, unknown>
    const d = defaults as Record<string, unknown>
    for (const key of managedKeys) {
      const v = q[key]
      if (isEmpty(v) || v === d[key]) continue
      out[key] = typeof v === 'boolean' ? (v ? '1' : '0') : String(v)
    }
    return out
  }

  function managedFromRoute(): Record<string, string> {
    const out: Record<string, string> = {}
    for (const key of managedKeys) {
      const v = firstValue(route.query[key])
      if (v !== undefined) out[key] = v
    }
    return out
  }

  function applyFromRoute() {
    const q = query as Record<string, unknown>
    Object.assign(q, { ...defaults })
    const managed = managedFromRoute()
    for (const [key, raw] of Object.entries(managed)) {
      const v = parseValue(raw, fieldType(key))
      if (v !== undefined) q[key] = v
    }
    if (!(query.page >= 1)) query.page = defaults.page
    if (!(query.pageSize >= 1)) query.pageSize = defaults.pageSize
    lastWritten = JSON.stringify(serialize())
  }

  function writeUrl() {
    if (!syncUrl) return
    const managed = serialize()
    const key = JSON.stringify(managed)
    if (key === lastWritten && key === JSON.stringify(managedFromRoute())) return
    lastWritten = key
    const rest: LocationQuery = {}
    for (const [k, v] of Object.entries(route.query)) {
      if (!managedKeys.includes(k)) rest[k] = v
    }
    void router.replace({ path: route.path, query: { ...rest, ...managed } })
  }

  async function load() {
    const seq = ++requestSeq
    loading.value = true
    error.value = false
    try {
      const { data } = await options.fetch(query)
      if (seq !== requestSeq) return
      list.value = data.list
      total.value = data.total
    } catch (e) {
      if (seq !== requestSeq) return
      list.value = []
      total.value = 0
      error.value = true
      showApiError(e, options.errorMessage ?? '获取列表失败')
    } finally {
      if (seq === requestSeq) loading.value = false
    }
  }

  function fetchList() {
    writeUrl()
    return load()
  }

  function handleSearch() {
    query.page = 1
    return fetchList()
  }

  function handleReset() {
    Object.assign(query as Record<string, unknown>, { ...defaults })
    return fetchList()
  }

  function afterRemove(count = 1) {
    if (query.page > 1 && list.value.length <= count) query.page -= 1
    return fetchList()
  }

  if (syncUrl) {
    watch(
      () => route.query,
      () => {
        if (route.path !== mountedPath) return
        const incoming = JSON.stringify(managedFromRoute())
        if (incoming === lastWritten) return
        applyFromRoute()
        void load()
      },
    )
  }

  onMounted(() => {
    mountedPath = route.path
    if (syncUrl) applyFromRoute()
    if (options.immediate ?? true) void load()
  })

  return {
    query,
    list,
    total,
    loading,
    error,
    fetchList,
    handleSearch,
    handleReset,
    afterRemove,
  }
}
