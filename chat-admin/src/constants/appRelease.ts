export type AppPlatform = 'android' | 'ios' | 'h5'

export const APP_PLATFORM_OPTIONS: { label: string; value: AppPlatform }[] = [
  { label: 'Android', value: 'android' },
  { label: 'iOS', value: 'ios' },
  { label: 'H5', value: 'h5' },
]

/** 安装包平台对应的文件后缀；H5 通过「登记网页版」发布，不上传文件。 */
export const APP_PACKAGE_EXT: Record<Exclude<AppPlatform, 'h5'>, string> = {
  android: '.apk',
  ios: '.ipa',
}

export const APP_PACKAGE_MAX_MB = 100

export const SEMVER_PATTERN = /^\d+\.\d+\.\d+$/

/** 比较 x.y.z 版本号：a < b 返回负数，相等返回 0，a > b 返回正数。 */
export function compareSemver(a: string, b: string) {
  const pa = a.split('.').map(Number)
  const pb = b.split('.').map(Number)
  for (let i = 0; i < 3; i++) {
    const diff = (pa[i] ?? 0) - (pb[i] ?? 0)
    if (diff !== 0) return diff
  }
  return 0
}
