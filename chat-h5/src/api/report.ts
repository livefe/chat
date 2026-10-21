import { post } from './request'

export {
  ReportTargetUser,
  ReportTargetMessage,
  ReportTargetPost,
  reportReasonOptions,
  reportTargetTypeText,
} from '@/constants/report'

export interface SubmitReportRequest {
  targetType: number
  targetId: number
  reason: number
  content?: string
  images?: number[]
}

export function submitReport(data: SubmitReportRequest) {
  return post('/api/report/submit', data)
}
