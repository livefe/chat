import type { RouteRecordRaw } from 'vue-router'

type ViewLoader = () => Promise<unknown>

export const componentMap: Record<string, ViewLoader> = {
  UserList: () => import('@/views/user/UserList.vue'),
  EmployeeList: () => import('@/views/employee/EmployeeList.vue'),
  DepartmentList: () => import('@/views/department/DepartmentList.vue'),
  RoleList: () => import('@/views/role/RoleList.vue'),
  PermissionList: () => import('@/views/permission/PermissionList.vue'),
  RoomList: () => import('@/views/room/RoomList.vue'),
  MessageList: () => import('@/views/message/MessageList.vue'),
  ConversationList: () => import('@/views/conversation/ConversationList.vue'),
  PostList: () => import('@/views/post/PostList.vue'),
  AdCampaignList: () => import('@/views/ad/AdCampaignList.vue'),
  AdSetting: () => import('@/views/ad/AdSetting.vue'),
  FeedbackList: () => import('@/views/feedback/FeedbackList.vue'),
  ReportList: () => import('@/views/report/ReportList.vue'),
  SensitiveWordList: () => import('@/views/sensitiveWord/SensitiveWordList.vue'),
  StaffChat: () => import('@/views/im/StaffChat.vue'),
  SupportDesk: () => import('@/views/support/SupportDesk.vue'),
  SupportOverview: () => import('@/views/support/SupportOverview.vue'),
  RegionTree: () => import('@/views/region/RegionTree.vue'),
  AuditLogList: () => import('@/views/auditLog/AuditLogList.vue'),
  AppReleaseList: () => import('@/views/appRelease/AppReleaseList.vue'),
}

export const baseRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/shell/LoginPage.vue'),
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/shell/AdminLayout.vue'),
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/shell/DashboardHome.vue'),
        meta: { title: '首页' },
      },
      {
        path: '403',
        name: 'AccessDenied',
        component: () => import('@/views/shell/AccessDenied.vue'),
        meta: { title: '无访问权限' },
      },
    ],
  },
]
