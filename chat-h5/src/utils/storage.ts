export const H5StorageKey = {
  AccessToken: 'h5_access_token',
  RefreshToken: 'h5_refresh_token',
  Location: 'h5_location',
  ChatViewLocation: 'h5_chat_view_location',
} as const

export function readH5AuthStorage() {
  return {
    accessToken: localStorage.getItem(H5StorageKey.AccessToken) || '',
    refreshToken: localStorage.getItem(H5StorageKey.RefreshToken) || '',
  }
}
