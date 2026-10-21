import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import 'element-plus/es/components/loading/style/css'

import App from './App.vue'
import { setAuthLostHandler, setAuthTokenProvider } from '@/api/http'
import router from './router'
import { useEmployeeStore } from '@/stores/employee'
import { ADMIN_ICONS } from '@/constants/adminIcons'
import '@/styles/theme.css'
import '@/styles/layout.css'
import '@/styles/page.css'
import '@/styles/tree-manage.css'

const app = createApp(App)

for (const [key, component] of Object.entries(ADMIN_ICONS)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)

const employeeStore = useEmployeeStore()
setAuthTokenProvider(
  () => employeeStore.accessToken,
  () => employeeStore.tryRefresh(),
)
setAuthLostHandler(() => {
  employeeStore.logout()
  void router.replace('/login')
})

app.mount('#app')
