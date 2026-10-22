export const UserStatus = {
  Disabled: 0,
  Enabled: 1,
} as const

/** 禁言时长（小时），-1 为永久。禁言入口与举报处置共用。 */
export const MUTE_PRESETS = [
  { label: '1 小时', value: 1 },
  { label: '24 小时', value: 24 },
  { label: '7 天', value: 168 },
  { label: '30 天', value: 720 },
  { label: '永久', value: -1 },
] as const

export const MUTE_MAX_HOURS = 87600

export function muteDurationLabel(hours: number) {
  const preset = MUTE_PRESETS.find((p) => p.value === hours)
  if (preset) return preset.label
  if (hours % 24 === 0) return `${hours / 24} 天`
  return `${hours} 小时`
}

export const PunishAction = {
  Warn: 1,
  Mute: 2,
  Unmute: 3,
  Disable: 4,
  Enable: 5,
  Deactivate: 6,
  DeleteMessage: 7,
  DisablePost: 8,
} as const

export function punishActionTagType(action: number): 'danger' | 'warning' | 'success' | 'info' {
  if (action === PunishAction.Unmute || action === PunishAction.Enable) return 'success'
  if (action === PunishAction.Warn || action === PunishAction.Mute) return 'warning'
  if (action === PunishAction.Deactivate) return 'info'
  return 'danger'
}
