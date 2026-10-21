import { post } from './request'

export type AdEventType = 'impression' | 'click'

export interface AdEventItem {
  requestId: string
  campaignId: number
  eventType: AdEventType
  impressionToken: string
  slotIndex?: number
  latitude?: number
  longitude?: number
}

export function reportAdEvents(events: AdEventItem[]) {
  return post<null>('/api/ad/events', { events }, { silent: true })
}
