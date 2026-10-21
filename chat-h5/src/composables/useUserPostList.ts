import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import { postUserList, type PostItem } from '@/api/post'
import { useLocationStore } from '@/stores/location'
import { showCatchError } from '@/utils/toast'

const PAGE_SIZE = 20

export function useUserPostList(userId: MaybeRefOrGetter<number>) {
  const gps = useLocationStore().gps

  const listLoading = ref(false)
  const refreshing = ref(false)
  const finished = ref(false)
  const page = ref(1)
  const posts = ref<PostItem[]>([])
  let listEpoch = 0

  function listParams() {
    return {
      page: page.value,
      pageSize: PAGE_SIZE,
      latitude: gps.valid ? gps.lat : undefined,
      longitude: gps.valid ? gps.lng : undefined,
    }
  }

  function reset() {
    listEpoch += 1
    page.value = 1
    finished.value = false
    refreshing.value = false
    listLoading.value = false
    posts.value = []
  }

  function setListLoading(value: boolean) {
    if (!refreshing.value) listLoading.value = value
  }

  async function loadMorePosts() {
    const id = toValue(userId)
    if (finished.value || refreshing.value || !id) return
    const epoch = listEpoch
    listLoading.value = true
    try {
      const res = await postUserList(id, listParams())
      if (epoch !== listEpoch) return
      const items = res.data.list || []
      const seen = new Set(posts.value.map((p) => p.id))
      const fresh = items.filter((p) => !seen.has(p.id))
      posts.value.push(...fresh)
      if (posts.value.length >= res.data.total || items.length < PAGE_SIZE) {
        finished.value = true
      } else {
        page.value += 1
      }
    } catch (e: unknown) {
      if (epoch !== listEpoch) return
      showCatchError(e, '加载失败', true)
      finished.value = true
    } finally {
      if (epoch === listEpoch) listLoading.value = false
    }
  }

  async function onRefresh() {
    const id = toValue(userId)
    if (!id) {
      refreshing.value = false
      return
    }
    listEpoch += 1
    const epoch = listEpoch
    page.value = 1
    finished.value = false
    try {
      listLoading.value = true
      const res = await postUserList(id, listParams())
      if (epoch !== listEpoch) return
      const items = res.data.list || []
      posts.value = items
      if (posts.value.length >= res.data.total || items.length < PAGE_SIZE) {
        finished.value = true
      } else {
        page.value = 2
      }
    } catch (e: unknown) {
      if (epoch !== listEpoch) return
      showCatchError(e, '加载失败', true)
      finished.value = true
    } finally {
      if (epoch === listEpoch) {
        listLoading.value = false
        refreshing.value = false
      }
    }
  }

  return {
    posts,
    listLoading,
    refreshing,
    finished,
    reset,
    setListLoading,
    loadMorePosts,
    onRefresh,
  }
}
