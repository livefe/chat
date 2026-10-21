import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { setToastDefaultOptions, showToast } from 'vant'
import 'vant/es/toast/style'
import 'vant/es/dialog/style'
import 'vant/es/image-preview/style'

import App from './App.vue'
import { setAuthLostHandler, setAuthTokenProvider, setRequestErrorNotifier } from './api/request'
import router from './router'
import { useUserStore } from './stores/user'
import { wireStores } from './stores/wireStores'
import './styles/global.css'

setToastDefaultOptions({
  position: 'middle',
  duration: 2200,
})
setRequestErrorNotifier((message) => showToast(message))

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
wireStores(pinia)
app.use(router)

const userStore = useUserStore()
setAuthTokenProvider(
  () => userStore.accessToken,
  () => userStore.tryRefresh(),
)
setAuthLostHandler(() => {
  userStore.logout()
  void router.replace('/login')
  showToast('登录已过期，请重新登录')
})

app.mount('#app')
