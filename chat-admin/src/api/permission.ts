import { request } from './http'

export interface PermissionItem {
  id: number
  name: string
  code: string
  description: string
  type: number
  typeText: string
  parentId: number
  path: string
  method: string
  sort: number
  status: number
  statusText: string
  icon?: string
  component?: string
  hideInMenu?: boolean
  createdAt: string
  updatedAt: string
}

export interface PermissionTreeItem {
  id: number
  name: string
  code: string
  type: number
  parentId: number
  path?: string
  method?: string
  sort: number
  children?: PermissionTreeItem[]
}

export interface PermissionListRequest {
  page: number
  pageSize: number
  name?: string
  code?: string
  type?: number
  status?: number
}

export interface PermissionListResponse {
  total: number
  list: PermissionItem[]
}

export interface CreatePermissionRequest {
  name: string
  code: string
  description: string
  type: number
  parentId: number
  path: string
  method: string
  sort: number
  status: number
  icon?: string
  component?: string
  hideInMenu?: boolean
}

export interface UpdatePermissionRequest extends CreatePermissionRequest {
  id: number
}

export const getPermissionList = (params: PermissionListRequest) =>
  request.get<PermissionListResponse>('/api/admin/permission/list', params)

export const getPermissionTree = (params?: { assignable?: boolean }) =>
  request.get<PermissionTreeItem[]>(
    '/api/admin/permission/tree',
    params?.assignable ? { assignable: '1' } : undefined,
  )

export const createPermission = (data: CreatePermissionRequest) =>
  request.post('/api/admin/permission/create', data)

export const updatePermission = (data: UpdatePermissionRequest) =>
  request.post('/api/admin/permission/update', data)

export const deletePermission = (id: number) => request.post('/api/admin/permission/delete', { id })

export const getEmployeePermissions = () =>
  request.get<PermissionItem[]>('/api/admin/permission/me')
