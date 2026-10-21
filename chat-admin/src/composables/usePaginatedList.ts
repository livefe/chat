import { reactive, ref } from 'vue'

import { showApiError } from '@/utils/apiError'

interface PageResult<T> {
  data: { list: T[]; total: number }
}

export function usePaginatedList<T, Q extends { page: number; pageSize: number }>(
  fetchFn: (query: Q) => Promise<PageResult<T>>,
  initialQuery: Q,
  errorMessage = '获取列表失败',
) {
  const loading = ref(false)
  const list = ref<T[]>([])
  const total = ref(0)
  const initial = { ...initialQuery }
  const query = reactive({ ...initial }) as Q
  let requestSeq = 0

  async function fetchList() {
    const seq = ++requestSeq
    loading.value = true
    try {
      const { data } = await fetchFn(query)
      if (seq !== requestSeq) return
      list.value = data.list
      total.value = data.total
    } catch (e) {
      if (seq !== requestSeq) return
      list.value = []
      total.value = 0
      showApiError(e, errorMessage)
    } finally {
      if (seq === requestSeq) loading.value = false
    }
  }

  function handleSearch() {
    query.page = 1
    fetchList()
  }

  function handleReset() {
    Object.assign(query, { ...initial })
    handleSearch()
  }

  return { loading, list, total, query, fetchList, handleSearch, handleReset }
}
