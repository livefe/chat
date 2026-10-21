import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { showFailToast, showSuccessToast } from 'vant'

import { postCreate, postDetail, postUpdate } from '@/api/post'
import { nearbyBootstrap, roomDetail, type NearbyRoomItem } from '@/api/room'
import { NEARBY_RADIUS_METERS } from '@/constants/nearby'
import { useVanImageUpload } from '@/composables/useVanImageUpload'
import { usePostDraftStore, type PostDraftImage } from '@/stores/postDraft'
import { useLocationStore } from '@/stores/location'
import { useUserStore } from '@/stores/user'
import { unixToStorageDateTime } from '@/utils/datetime'
import { displayPlaceName } from '@/utils/map/placeName'
import { resolveMediaUrl } from '@/utils/media'
import { PostStatus } from '@/constants/post'
import { PostCategory, postCategoryOption, type PostCategoryValue } from '@/utils/post/postCategory'
import { buildPostPublishPayload, validatePostPublishForm } from '@/utils/post/postPublishPayload'
import { showCatchError } from '@/utils/toast'

const NEARBY_ATTACH_LIMIT = 5

export function usePostPublishForm() {
  const router = useRouter()
  const route = useRoute()
  const postLocation = useLocationStore().publish
  const postDraft = usePostDraftStore()
  const userStore = useUserStore()
  const form = postDraft.form
  const { imageList } = storeToRefs(postDraft)
  const { uploading, onAfterRead } = useVanImageUpload()

  const loading = ref(false)
  const pageLoading = ref(false)
  const editId = computed(() => (route.name === 'PostEdit' ? Number(route.params.id) : 0))
  const isEdit = computed(() => editId.value > 0)
  const roomId = ref(0)
  const attachedRoomName = ref('')
  const nearbyRooms = ref<NearbyRoomItem[]>([])
  const nearbyLoading = ref(false)
  let nearbySeq = 0
  let skipNearbyReload = false

  const isRoomPost = computed(() => roomId.value > 0)
  const category = computed(() => form.category as PostCategoryValue)
  const categoryMeta = computed(() => postCategoryOption(category.value))
  const hasPendingImages = computed(() => imageList.value.some((x) => x.file && !x.url))

  const postLocationLabel = computed(() => {
    if (postLocation.ready) {
      return displayPlaceName(postLocation.placeName, {
        isCurrentGps: postLocation.source === 'gps',
      })
    }
    if (postLocation.error) return '定位失败，请手动选择位置'
    return '定位中…'
  })

  const displayNearbyRooms = computed(() => {
    const list = nearbyRooms.value
    if (roomId.value <= 0) return list
    if (list.some((r) => r.id === roomId.value)) return list
    if (!attachedRoomName.value) return list
    const fallback: NearbyRoomItem = {
      id: roomId.value,
      name: attachedRoomName.value,
      placeType: 0,
      shapeType: 1,
      centerLat: postLocation.latitude,
      centerLng: postLocation.longitude,
      radius: 0,
      distance: 0,
      requireGeofence: false,
    }
    return [fallback, ...list].slice(0, NEARBY_ATTACH_LIMIT)
  })

  const showNearbyRooms = computed(
    () => postLocation.ready && (displayNearbyRooms.value.length > 0 || nearbyLoading.value),
  )

  const canSubmit = computed(() => {
    if (loading.value || uploading.value || hasPendingImages.value) return false
    return postLocation.ready
  })

  async function loadNearbyRooms() {
    if (!postLocation.ready) {
      nearbyRooms.value = []
      return
    }
    const seq = ++nearbySeq
    nearbyLoading.value = true
    try {
      const res = await nearbyBootstrap(
        postLocation.latitude,
        postLocation.longitude,
        NEARBY_ATTACH_LIMIT,
        false,
        NEARBY_RADIUS_METERS,
      )
      if (seq !== nearbySeq) return
      nearbyRooms.value = (res.data.rooms || []).slice(0, NEARBY_ATTACH_LIMIT)
    } catch {
      if (seq !== nearbySeq) return
      nearbyRooms.value = []
    } finally {
      if (seq === nearbySeq) nearbyLoading.value = false
    }
  }

  function clearAttachRoom() {
    roomId.value = 0
    attachedRoomName.value = ''
    form.pinned = false
  }

  function selectAttachRoom(room: NearbyRoomItem) {
    roomId.value = room.id
    attachedRoomName.value = room.name || '聊天区'
    if (!(room.centerLat || room.centerLng)) return

    if (nearbyRooms.value.length > 0) skipNearbyReload = true
    postLocation.setLocation(
      {
        placeName: room.name || '聊天区',
        latitude: room.centerLat,
        longitude: room.centerLng,
      },
      'manual',
    )
    queueMicrotask(() => {
      skipNearbyReload = false
    })
  }

  watch(
    () => [postLocation.ready, postLocation.latitude, postLocation.longitude] as const,
    () => {
      if (skipNearbyReload) return
      void loadNearbyRooms()
    },
  )

  async function attachRoomById(id: number) {
    const res = await roomDetail(id)
    selectAttachRoom({
      id,
      name: res.data.name || '聊天区',
      placeType: res.data.placeType,
      placeTypeText: res.data.placeTypeText,
      shapeType: res.data.shapeType,
      centerLat: res.data.centerLat,
      centerLng: res.data.centerLng,
      radius: res.data.radius,
      distance: 0,
      requireGeofence: res.data.requireGeofence,
    })
  }

  async function loadForEdit() {
    if (!editId.value) return
    pageLoading.value = true
    try {
      const res = await postDetail(editId.value)
      const d = res.data
      if (d.publisherId !== userStore.info?.id) {
        showFailToast('无权编辑')
        router.back()
        return
      }
      if (d.status === PostStatus.Published) {
        showFailToast('已发布，无法编辑')
        router.back()
        return
      }
      if (d.status === PostStatus.Disabled) {
        showFailToast('已下架，无法编辑')
        router.back()
        return
      }
      roomId.value = d.roomId || 0
      attachedRoomName.value = d.roomId > 0 ? d.placeName || '聊天区' : ''
      form.title = d.title
      form.content = d.content
      form.category = d.category || PostCategory.Idle
      form.pinned = !!d.pinned
      form.enableSchedule = d.startAt > 0 || d.endAt > 0
      form.startAt = unixToStorageDateTime(d.startAt)
      form.endAt = unixToStorageDateTime(d.endAt)
      imageList.value = (d.images || [])
        .map((url, index): PostDraftImage => ({
          url: resolveMediaUrl(url),
          status: 'done',
          isImage: true,
          mediaId: d.imageIds?.[index] || 0,
        }))
        .filter((x) => !!x.url)
      await postLocation.initFromPost({
        placeName: d.placeName,
        latitude: d.latitude,
        longitude: d.longitude,
      })
      if (d.roomId > 0) {
        const room = await roomDetail(d.roomId).catch(() => null)
        if (room) attachedRoomName.value = room.data.name || attachedRoomName.value
      }
      const extras = d.extras || {}
      form.idlePrice = Number(extras.amount) || 0
      form.idleNegotiable = !!extras.negotiable
      form.enableIdlePrice = form.idlePrice > 0 || form.idleNegotiable
      form.needSignup = !!extras.needSignup
      form.needCheckIn = !!extras.needSignup && !!extras.needCheckIn
      form.limitParticipants = (extras.maxParticipants || 0) > 0
      form.maxParticipants = extras.maxParticipants || 0
      form.allowWaitlist = !!extras.allowWaitlist
      form.signupDeadlineAt = unixToStorageDateTime(extras.signupDeadlineAt || 0)
      form.enableFee = extras.feeType === 1 || extras.feeType === 2
      form.feeType = extras.feeType === 2 ? 'reward' : 'charge'
      form.feeAmount = extras.amount || 0
      form.feeDescription = extras.feeDescription || ''
      form.helpReward = extras.feeType === 2
      form.helpRewardAmount = extras.feeType === 2 ? extras.amount || 0 : 0
      form.helpRewardDesc = extras.feeType === 2 ? extras.feeDescription || '' : ''
    } catch (e: unknown) {
      showCatchError(e, '加载失败', true)
      router.back()
    } finally {
      pageLoading.value = false
    }
  }

  function publishInput() {
    return {
      form,
      category: category.value,
      imageList: imageList.value,
      isRoomPost: isRoomPost.value,
      roomId: roomId.value,
      postLocation: {
        ready: postLocation.ready,
        placeName: postLocation.placeName,
        latitude: postLocation.latitude,
        longitude: postLocation.longitude,
        source: postLocation.source,
      },
    }
  }

  async function submit() {
    const payload = buildPostPublishPayload(publishInput())
    const invalid = validatePostPublishForm(payload, {
      ...publishInput(),
      uploading: uploading.value,
      hasPendingImages: hasPendingImages.value,
    })
    if (invalid) {
      showFailToast(invalid)
      return
    }

    loading.value = true
    try {
      if (isEdit.value) {
        await postUpdate({ id: editId.value, ...payload })
        showSuccessToast('已提交，等待审核')
      } else {
        await postCreate(payload)
        showSuccessToast('提交成功，等待审核')
      }
      router.replace('/post/my?mode=published')
    } catch (e: unknown) {
      showCatchError(e, '提交失败', true)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    postLocation.clear()
    postDraft.reset()
  }

  return {
    form,
    imageList,
    uploading,
    onAfterRead,
    loading,
    pageLoading,
    isEdit,
    roomId,
    isRoomPost,
    category,
    categoryMeta,
    hasPendingImages,
    canSubmit,
    postLocation,
    postLocationLabel,
    nearbyLoading,
    displayNearbyRooms,
    showNearbyRooms,
    clearAttachRoom,
    selectAttachRoom,
    attachRoomById,
    loadForEdit,
    submit,
    reset,
  }
}
