export type AppPlatform = 'android' | 'ios' | 'h5'

export const APP_PLATFORM_OPTIONS: { label: string; value: AppPlatform }[] = [
  { label: 'Android', value: 'android' },
  { label: 'iOS', value: 'ios' },
  { label: 'H5', value: 'h5' },
]
