import type { PostExtras } from '@/api/post'
import type { PostDraftForm, PostDraftImage } from '@/stores/postDraft'
import { parseDateTimeToUnix } from '@/utils/datetime'
import { displayPlaceName } from '@/utils/map/placeName'
import { collectMediaIds } from '@/utils/media'
import { PostCategory, type PostCategoryValue } from '@/utils/post/postCategory'

export type PostPublishLocation = {
  ready: boolean
  placeName: string
  latitude: number
  longitude: number
  source: string | null
}

export type PostPublishPayloadInput = {
  form: PostDraftForm
  category: PostCategoryValue
  imageList: PostDraftImage[]
  isRoomPost: boolean
  roomId: number
  postLocation: PostPublishLocation
}

export function formatOptionalNumber(value: number): string {
  return value > 0 ? String(value) : ''
}

export function parseFeeAmount(form: PostDraftForm, value: string) {
  if (!value.trim()) {
    form.feeAmount = 0
    return
  }
  const n = Number(value)
  form.feeAmount = Number.isFinite(n) && n > 0 ? n : 0
}

export function parseIdlePrice(form: PostDraftForm, value: string) {
  if (!value.trim()) {
    form.idlePrice = 0
    return
  }
  const n = Number(value)
  form.idlePrice = Number.isFinite(n) && n >= 0 ? n : 0
}

export function parseHelpReward(form: PostDraftForm, value: string) {
  if (!value.trim()) {
    form.helpRewardAmount = 0
    return
  }
  const n = Number(value)
  form.helpRewardAmount = Number.isFinite(n) && n > 0 ? n : 0
}

export function parseMaxParticipants(form: PostDraftForm, value: string) {
  if (!value.trim()) {
    form.maxParticipants = 0
    return
  }
  const n = Number.parseInt(value, 10)
  form.maxParticipants = Number.isFinite(n) && n > 0 ? n : 0
}

export function buildPostPublishPayload(input: PostPublishPayloadInput) {
  const { form, category, imageList, isRoomPost, roomId, postLocation } = input
  const images = collectMediaIds(imageList)
  const base = {
    category: form.category,
    pinned: isRoomPost ? form.pinned : false,
    title: form.title,
    content: form.content,
    images,
    startAt: 0,
    endAt: 0,
    extras: undefined as PostExtras | undefined,
    roomId: roomId > 0 ? roomId : 0,
    placeName: displayPlaceName(postLocation.placeName, {
      isCurrentGps: postLocation.source === 'gps',
    }),
    latitude: postLocation.latitude,
    longitude: postLocation.longitude,
  }

  if (category === PostCategory.Event) {
    base.startAt = form.enableSchedule ? parseDateTimeToUnix(form.startAt) : 0
    base.endAt = form.enableSchedule ? parseDateTimeToUnix(form.endAt) : 0
    const signupDeadlineAt = parseDateTimeToUnix(form.signupDeadlineAt)
    base.extras = {
      needSignup: form.needSignup,
      needCheckIn: form.needSignup && form.needCheckIn,
      maxParticipants: form.needSignup && form.limitParticipants ? form.maxParticipants : 0,
      allowWaitlist: form.needSignup && form.limitParticipants ? form.allowWaitlist : false,
      signupDeadlineAt,
      feeType: form.enableFee ? (form.feeType === 'reward' ? 2 : 1) : 0,
      amount: form.enableFee ? form.feeAmount : 0,
      feeDescription: form.enableFee ? form.feeDescription : '',
    }
  } else if (category === PostCategory.Idle) {
    if (form.enableIdlePrice) {
      base.extras = {
        amount: form.idlePrice,
        negotiable: form.idleNegotiable,
      }
    }
  } else if (category === PostCategory.Help) {
    if (form.helpReward) {
      base.extras = {
        feeType: 2,
        amount: form.helpRewardAmount,
        feeDescription: form.helpRewardDesc,
      }
    }
  }

  return base
}

export type PostPublishPayload = ReturnType<typeof buildPostPublishPayload>

export type PostPublishValidateInput = PostPublishPayloadInput & {
  uploading: boolean
  hasPendingImages: boolean
}

export function validatePostPublishForm(
  payload: PostPublishPayload,
  input: PostPublishValidateInput,
): string | null {
  const { form, category, uploading, hasPendingImages } = input

  if (!form.title.trim()) {
    return '请填写标题'
  }
  if (!form.category) {
    return '请选择类型'
  }
  if (category !== PostCategory.Idle && !form.content.trim()) {
    return '请填写内容'
  }
  if (!input.postLocation.ready) {
    return '选择位置'
  }
  if (uploading || hasPendingImages) {
    return '请等待图片上传完成'
  }
  if (category === PostCategory.Event) {
    const extras = payload.extras
    if (payload.startAt > 0 && payload.endAt > 0 && payload.endAt < payload.startAt) {
      return '结束时间不能早于开始时间'
    }
    if ((extras?.needSignup || extras?.needCheckIn) && payload.endAt <= 0) {
      return '开启报名或核验时请设置结束时间'
    }
    const signupDeadlineAt = extras?.signupDeadlineAt || 0
    if (signupDeadlineAt > 0 && payload.startAt > 0 && signupDeadlineAt > payload.startAt) {
      return '报名截止不能晚于开始时间'
    }
    if (extras?.needSignup && signupDeadlineAt > 0) {
      const nowSec = Math.floor(Date.now() / 1000)
      if (signupDeadlineAt + 59 < nowSec) {
        return '报名截止时间不能早于当前时间'
      }
    }
  }
  if (category === PostCategory.Idle && form.enableIdlePrice && form.idlePrice < 0) {
    return '价格无效'
  }
  return null
}
