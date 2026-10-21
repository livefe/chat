import { ReportStatus } from '@/constants/report'

export function reportStatusTagType(status: number): 'success' | 'info' | 'danger' {
  if (status === ReportStatus.Resolved) return 'success'
  if (status === ReportStatus.Rejected) return 'info'
  return 'danger'
}
