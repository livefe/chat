import { request } from './http'

export interface AdminPendingCounts {
  imUnread: number
  supportQueue: number
  supportUnread: number
  reportPending: number
  feedbackPending: number
  roomAuditPending: number
  postAuditPending: number
}

export const getPendingCounts = () =>
  request.get<AdminPendingCounts>('/api/admin/ops/pending-counts')
