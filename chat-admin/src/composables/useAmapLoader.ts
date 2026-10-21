import AMapLoader from '@amap/amap-jsapi-loader'

export type AMapNamespace = typeof AMap

let loaderPromise: Promise<AMapNamespace> | null = null
let securityConfigured = false

const DEFAULT_PLUGINS = ['AMap.MouseTool', 'AMap.PlaceSearch']

function ensureSecurityConfig() {
  if (securityConfigured) return
  const code = import.meta.env.VITE_AMAP_SECURITY || ''
  if (code) {
    ;(
      window as unknown as { _AMapSecurityConfig?: { securityJsCode: string } }
    )._AMapSecurityConfig = { securityJsCode: code }
  }
  securityConfigured = true
}

export function hasAmapKey() {
  return !!import.meta.env.VITE_AMAP_KEY
}

function getAmapKey(): string {
  return import.meta.env.VITE_AMAP_KEY || ''
}

export function loadAmap(plugins: string[] = []): Promise<AMapNamespace> {
  ensureSecurityConfig()
  if (loaderPromise) return loaderPromise

  const key = getAmapKey()
  if (!key) {
    return Promise.reject(new Error('未配置 VITE_AMAP_KEY'))
  }

  const merged = Array.from(new Set([...DEFAULT_PLUGINS, ...plugins]))
  loaderPromise = AMapLoader.load({
    key,
    version: '2.0',
    plugins: merged,
  }).catch((err) => {
    loaderPromise = null
    throw err
  })
  return loaderPromise
}
