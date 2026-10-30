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

export interface PermissionTreeItem extends PermissionItem {
  children?: PermissionTreeItem[]
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

export const getPermissionTree = (params?: { assignable?: boolean }) =>
  request.get<PermissionTreeItem[]>(
    '/api/admin/permission/tree',
    params?.assignable ? { assignable: true } : undefined,
  )

export const createPermission = (data: CreatePermissionRequest) =>
  request.post('/api/admin/permission/create', data)

export const updatePermission = (data: UpdatePermissionRequest) =>
  request.post('/api/admin/permission/update', data)

export const deletePermission = (id: number) => request.post('/api/admin/permission/delete', { id })

export const getEmployeePermissions = () =>
  request.get<PermissionItem[]>('/api/admin/permission/me')
