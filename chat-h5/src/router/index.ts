import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { useUserStore } from '@/stores/user'
import { safeRedirectPath } from '@/utils/safeRedirect'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginPage.vue'),
    meta: { public: true },
  },
  {
    path: '/agreement/user',
    name: 'UserAgreement',
    component: () => import('@/views/agreement/AgreementPage.vue'),
    props: { doc: 'user' },
    meta: { public: true },
  },
  {
    path: '/agreement/privacy',
    name: 'PrivacyPolicy',
    component: () => import('@/views/agreement/AgreementPage.vue'),
    props: { doc: 'privacy' },
    meta: { public: true },
  },
  {
    path: '/download',
    name: 'Download',
    component: () => import('@/views/download/DownloadPage.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    component: () => import('@/views/MainLayout.vue'),
    redirect: '/chat',
    children: [
      {
        path: 'chat',
        name: 'Chat',
        component: () => import('@/views/chat/ChatPage.vue'),
      },
      {
        path: 'discover',
        name: 'PostFeed',
        component: () => import('@/views/post/PostFeedPage.vue'),
      },
      {
        path: 'messages',
        name: 'ConversationList',
        component: () => import('@/views/conversation/ConversationListPage.vue'),
      },
      {
        path: 'me',
        name: 'Me',
        component: () => import('@/views/account/MePage.vue'),
      },
    ],
  },
  {
    path: '/room/:id',
    name: 'Room',
    component: () => import('@/views/chat/RoomPage.vue'),
    props: true,
  },
  {
    path: '/post/publish',
    name: 'PostPublish',
    component: () => import('@/views/post/PostPublishPage.vue'),
  },
  {
    path: '/location/pick',
    name: 'LocationPick',
    component: () => import('@/views/location/LocationPickPage.vue'),
  },
  {
    path: '/post/:id/edit',
    name: 'PostEdit',
    component: () => import('@/views/post/PostPublishPage.vue'),
  },
  {
    path: '/post/my',
    name: 'PostMy',
    component: () => import('@/views/post/PostMyPage.vue'),
  },
  {
    path: '/post/:id/signups',
    name: 'PostSignupList',
    component: () => import('@/views/post/PostSignupListPage.vue'),
    props: true,
  },
  {
    path: '/post/:id/check-in-code',
    name: 'PostCheckInCode',
    component: () => import('@/views/post/PostCheckInCodePage.vue'),
    props: true,
  },
  {
    path: '/scan',
    name: 'CheckInScan',
    component: () => import('@/views/post/PostCheckInScanPage.vue'),
  },
  {
    path: '/check-in-result',
    name: 'PostCheckInResult',
    component: () => import('@/views/post/PostCheckInResultPage.vue'),
  },
  {
    path: '/post/:id',
    name: 'PostDetail',
    component: () => import('@/views/post/PostDetailPage.vue'),
    props: true,
  },
  {
    path: '/user/:id',
    name: 'UserProfile',
    component: () => import('@/views/user/UserProfilePage.vue'),
    props: true,
  },
  {
    path: '/messages/support',
    name: 'SupportChat',
    component: () => import('@/views/support/SupportChatPage.vue'),
  },
  {
    path: '/messages/:peerId',
    name: 'PrivateChat',
    component: () => import('@/views/privateChat/PrivateChatPage.vue'),
    props: true,
  },
  {
    path: '/me/settings',
    name: 'Settings',
    component: () => import('@/views/account/SettingsPage.vue'),
  },
  {
    path: '/me/settings/blocks',
    name: 'BlockList',
    component: () => import('@/views/account/BlockListPage.vue'),
  },
  {
    path: '/me/edit',
    name: 'ProfileEdit',
    component: () => import('@/views/account/ProfileEditPage.vue'),
  },
  {
    path: '/me/edit/nickname',
    name: 'NicknameEdit',
    component: () => import('@/views/account/NicknameEditPage.vue'),
  },
  {
    path: '/me/edit/mobile',
    name: 'MobileEdit',
    component: () => import('@/views/account/MobileEditPage.vue'),
  },
  {
    path: '/report',
    name: 'ReportSubmit',
    component: () => import('@/views/report/ReportSubmitPage.vue'),
  },
  {
    path: '/me/feedback',
    name: 'FeedbackSubmit',
    component: () => import('@/views/feedback/FeedbackSubmitPage.vue'),
  },
  {
    path: '/me/account/deactivate',
    name: 'AccountDeactivate',
    component: () => import('@/views/account/AccountDeactivatePage.vue'),
  },
  {
    path: '/me/feedback/records',
    name: 'FeedbackList',
    component: () => import('@/views/feedback/FeedbackListPage.vue'),
  },
  {
    path: '/me/feedback/records/:id',
    name: 'FeedbackDetail',
    component: () => import('@/views/feedback/FeedbackDetailPage.vue'),
    props: true,
  },
  {
    path: '/forward',
    name: 'Forward',
    component: () => import('@/views/forward/ForwardPage.vue'),
  },
  {
    path: '/chat/rooms',
    name: 'NearbyRooms',
    component: () => import('@/views/chat/NearbyRoomsPage.vue'),
  },
  {
    path: '/chat/room/create',
    name: 'RoomCreate',
    component: () => import('@/views/chat/RoomCreatePage.vue'),
  },
  {
    path: '/chat/room/applications',
    name: 'RoomApplications',
    component: () => import('@/views/chat/RoomApplicationsPage.vue'),
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const userStore = useUserStore()
  if (!to.meta.public && !userStore.accessToken) {
    if (userStore.refreshToken) {
      const ok = await userStore.ensureSession()
      if (!ok) {
        userStore.logout()
        return { path: '/login', query: { redirect: to.fullPath } }
      }
    } else {
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }
  if (to.name === 'Login' && userStore.accessToken) {
    return { path: safeRedirectPath(to.query.redirect) }
  }
  if (!to.meta.public && userStore.accessToken && !userStore.info) {
    const ok = await userStore.ensureSession()
    if (!ok) {
      userStore.logout()
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }
})

export default router
