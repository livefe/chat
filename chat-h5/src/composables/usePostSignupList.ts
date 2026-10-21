import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import {
  emptySignupListCounts,
  postCheckIn,
  postDetail,
  postSignupList,
  postSignupRemark,
  type PostSignupItem,
  type PostSignupListCounts,
} from '@/api/post'
import { SignupCheckIn, SignupStatus } from '@/constants/post'
import { formatDateTimeChinese } from '@/utils/datetime'
import { downloadExcelTable } from '@/utils/post/csv'
import {
  hasCheckInTime,
  isSignupCheckInWindowOpen,
  signupCheckInText,
  signupListStatusText,
} from '@/utils/post/post'

const PAGE_SIZE = 30
const EXPORT_PAGE_SIZE = 100

function formatNowDateTime() {
  const d = new Date()
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

function fileStamp(raw: string) {
  return raw.replace(/[-: ]/g, '').replace(/^(\d{8})(\d{6})$/, '$1-$2')
}

function sanitizeExportFilenamePart(raw: string) {
  return raw.replace(/[\\/:*?"<>|]/g, '_').trim()
}

export function usePostSignupList(
  postIdSource: MaybeRefOrGetter<number>,
  opts: { onLoadError: (error: unknown) => void },
) {
  const postTitle = ref('')
  const list = ref<PostSignupItem[]>([])
  const loading = ref(false)
  const finished = ref(false)
  const refreshing = ref(false)
  const page = ref(1)
  const statusFilter = ref<number | undefined>(undefined)
  const checkInFilter = ref<number | undefined>(undefined)
  const needCheckIn = ref(false)
  const checkInWindowOpen = ref(true)
  const counts = ref<PostSignupListCounts>(emptySignupListCounts())

  const postId = () => toValue(postIdSource)

  async function loadPostTitle() {
    if (!postId()) return
    try {
      const res = await postDetail(postId())
      postTitle.value = res.data.title
      needCheckIn.value = !!res.data.extras?.needSignup && !!res.data.extras?.needCheckIn
      checkInWindowOpen.value = isSignupCheckInWindowOpen(
        res.data.startAt || 0,
        res.data.endAt || 0,
      )
    } catch {
      postTitle.value = ''
    }
  }

  async function fetchPage(pageNo: number, pageSize = PAGE_SIZE) {
    const res = await postSignupList({
      postId: postId(),
      page: pageNo,
      pageSize,
      status: statusFilter.value,
      checkIn: checkInFilter.value,
    })
    needCheckIn.value = !!res.data.needCheckIn
    counts.value = res.data.counts
    return res.data
  }

  function applyPage(items: PostSignupItem[], total: number, pageNo: number) {
    if (list.value.length >= total || items.length < PAGE_SIZE) {
      finished.value = true
    } else {
      page.value = pageNo + 1
    }
  }

  async function loadMore() {
    if (!postId() || finished.value) {
      loading.value = false
      return
    }
    loading.value = true
    try {
      const pageNo = page.value
      const data = await fetchPage(pageNo)
      const items = data.list || []
      list.value.push(...items)
      applyPage(items, data.total, pageNo)
    } catch (e: unknown) {
      opts.onLoadError(e)
      finished.value = true
    } finally {
      loading.value = false
    }
  }

  async function reloadList() {
    page.value = 1
    finished.value = false
    list.value = []
    await loadMore()
  }

  async function refresh() {
    page.value = 1
    finished.value = false
    try {
      loading.value = true
      const data = await fetchPage(1)
      const items = data.list || []
      list.value = items
      applyPage(items, data.total, 1)
    } catch (e: unknown) {
      opts.onLoadError(e)
      finished.value = true
    } finally {
      loading.value = false
      refreshing.value = false
    }
  }

  function setFilters(status: number | undefined, checkIn: number | undefined) {
    statusFilter.value = status
    checkInFilter.value = checkIn
    void reloadList()
  }

  function canManualCheckIn(item: PostSignupItem) {
    return (
      needCheckIn.value &&
      checkInWindowOpen.value &&
      item.signupStatus === SignupStatus.Confirmed &&
      !hasCheckInTime(item.checkedInAt)
    )
  }

  async function manualCheckIn(item: PostSignupItem) {
    const res = await postCheckIn({ postId: postId(), userId: item.userId })
    const checkedInAt = formatNowDateTime()
    const leaveUnchecked = checkInFilter.value === SignupCheckIn.NotChecked
    list.value = leaveUnchecked
      ? list.value.filter((row) => row.id !== item.id)
      : list.value.map((row) => (row.id === item.id ? { ...row, checkedInAt } : row))
    counts.value = {
      ...counts.value,
      notCheckedIn: Math.max(0, counts.value.notCheckedIn - 1),
      checkedIn: counts.value.checkedIn + (res.data.alreadyCheckedIn ? 0 : 1),
    }
    return !!res.data.alreadyCheckedIn
  }

  async function saveRemark(item: PostSignupItem, remark: string) {
    await postSignupRemark({ postId: postId(), userId: item.userId, remark })
    list.value = list.value.map((row) => (row.id === item.id ? { ...row, remark } : row))
  }

  async function fetchAllSignups() {
    const all: PostSignupItem[] = []
    let pageNo = 1
    while (true) {
      const data = await fetchPage(pageNo, EXPORT_PAGE_SIZE)
      const items = data.list || []
      all.push(...items)
      if (all.length >= data.total || items.length < EXPORT_PAGE_SIZE) break
      pageNo += 1
    }
    return all
  }

  function downloadSignups(rows: PostSignupItem[]) {
    const stamp = fileStamp(formatNowDateTime())
    const titlePart = sanitizeExportFilenamePart(postTitle.value) || String(postId())
    const headers = needCheckIn.value
      ? ['昵称', '状态', '报名时间', '核验', '核验时间', '取消时间', '备注']
      : ['昵称', '状态', '报名时间', '取消时间', '备注']
    downloadExcelTable(
      `报名名单-${titlePart}-${stamp}`,
      headers,
      rows.map((item) => {
        const row = [
          item.userNickname || '',
          signupListStatusText(item),
          formatDateTimeChinese(item.createdAt, true),
        ]
        if (needCheckIn.value) {
          row.push(
            item.signupStatus === SignupStatus.Confirmed ? signupCheckInText(item.checkedInAt) : '',
            hasCheckInTime(item.checkedInAt)
              ? formatDateTimeChinese(item.checkedInAt || '', true)
              : '',
          )
        }
        row.push(
          item.cancelledAt ? formatDateTimeChinese(item.cancelledAt, true) : '',
          (item.remark || '').trim(),
        )
        return row
      }),
      { sheetName: '报名名单' },
    )
  }

  return {
    postTitle,
    list,
    loading,
    finished,
    refreshing,
    statusFilter,
    checkInFilter,
    needCheckIn,
    counts,
    loadPostTitle,
    loadMore,
    reloadList,
    refresh,
    setFilters,
    canManualCheckIn,
    manualCheckIn,
    saveRemark,
    fetchAllSignups,
    downloadSignups,
  }
}
