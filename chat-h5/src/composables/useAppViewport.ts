import { onMounted, onUnmounted } from 'vue'

export function useAppViewport() {
  function update() {
    const vv = window.visualViewport
    const innerH = window.innerHeight
    if (!vv) {
      document.documentElement.style.setProperty('--app-vh', `${innerH}px`)
      document.documentElement.style.setProperty('--vv-offset-top', '0px')
      document.documentElement.style.setProperty('--keyboard-inset', '0px')
      delete document.documentElement.dataset.keyboardOpen
      return
    }

    const height = Math.round(vv.height)
    const offsetTop = Math.round(vv.offsetTop)
    const inset = Math.max(0, Math.round(innerH - vv.height - vv.offsetTop))
    const keyboardOpen = inset > 48

    document.documentElement.style.setProperty('--keyboard-inset', `${inset}px`)
    document.documentElement.dataset.keyboardOpen = keyboardOpen ? '1' : ''

    if (keyboardOpen) {
      document.documentElement.style.setProperty('--app-vh', `${height}px`)
      document.documentElement.style.setProperty('--vv-offset-top', `${offsetTop}px`)
      return
    }

    document.documentElement.style.setProperty('--app-vh', `${innerH}px`)
    document.documentElement.style.setProperty('--vv-offset-top', '0px')
  }

  onMounted(() => {
    update()
    window.visualViewport?.addEventListener('resize', update)
    window.visualViewport?.addEventListener('scroll', update)
    window.addEventListener('resize', update)
  })

  onUnmounted(() => {
    window.visualViewport?.removeEventListener('resize', update)
    window.visualViewport?.removeEventListener('scroll', update)
    window.removeEventListener('resize', update)
    document.documentElement.style.removeProperty('--app-vh')
    document.documentElement.style.removeProperty('--vv-offset-top')
    document.documentElement.style.removeProperty('--keyboard-inset')
    delete document.documentElement.dataset.keyboardOpen
  })
}
