/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_TIANDITU_TK: string
  readonly VITE_APP_SCHEME: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

interface NearbyAppBridge {
  close: () => void
}

interface Window {
  NearbyApp?: NearbyAppBridge
}
