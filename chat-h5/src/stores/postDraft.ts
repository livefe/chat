import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'

export interface PostDraftImage {
  url?: string
  file?: File
  isImage?: boolean
  status?: '' | 'uploading' | 'done' | 'failed'
  message?: string
  mediaId?: number
}

export interface PostDraftForm {
  title: string
  content: string
  category: string
  pinned: boolean

  enableSchedule: boolean
  startAt: string
  endAt: string
  needSignup: boolean
  needCheckIn: boolean
  limitParticipants: boolean
  maxParticipants: number
  allowWaitlist: boolean
  signupDeadlineAt: string
  enableFee: boolean
  feeType: 'charge' | 'reward'
  feeAmount: number
  feeDescription: string

  enableIdlePrice: boolean
  idlePrice: number
  idleNegotiable: boolean

  helpReward: boolean
  helpRewardAmount: number
  helpRewardDesc: string
}

function emptyForm(): PostDraftForm {
  return {
    title: '',
    content: '',
    category: '',
    pinned: false,
    enableSchedule: false,
    startAt: '',
    endAt: '',
    needSignup: false,
    needCheckIn: false,
    limitParticipants: false,
    maxParticipants: 0,
    allowWaitlist: false,
    signupDeadlineAt: '',
    enableFee: false,
    feeType: 'charge',
    feeAmount: 0,
    feeDescription: '',
    enableIdlePrice: false,
    idlePrice: 0,
    idleNegotiable: true,
    helpReward: false,
    helpRewardAmount: 0,
    helpRewardDesc: '',
  }
}

export const usePostDraftStore = defineStore('postDraft', () => {
  const form = reactive<PostDraftForm>(emptyForm())
  const imageList = ref<PostDraftImage[]>([])

  function reset() {
    Object.assign(form, emptyForm())
    imageList.value = []
  }

  return {
    form,
    imageList,
    reset,
  }
})
