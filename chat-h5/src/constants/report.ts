export const ReportTargetUser = 1
export const ReportTargetMessage = 2
export const ReportTargetPost = 3

export const reportReasonOptions = [
  { text: '垃圾广告', value: 1 },
  { text: '色情低俗', value: 2 },
  { text: '辱骂骚扰', value: 3 },
  { text: '违法违规', value: 4 },
  { text: '虚假信息', value: 5 },
  { text: '侵权', value: 6 },
  { text: '其他', value: 7 },
]

export const reportTargetTypeText: Record<number, string> = {
  [ReportTargetUser]: '用户',
  [ReportTargetMessage]: '消息',
  [ReportTargetPost]: '内容',
}
