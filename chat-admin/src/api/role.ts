import { request } from './http'
import type { PermissionItem } from './permission'

export interface RoleItem {
  id: number
  name: string
  code: string
  description: string
  status: number
  statusText: string
  createdAt: string
  updatedAt: string
}

export interface RoleListRequest {
  page: number
  pageSize: number
  name?: string
  code?: string
  status?: number
}

export interface RoleListResponse {
  total: number
  list: RoleItem[]
}

export interface CreateRoleRequest {
  name: string
  code: string
  description: string
  status: number
}

export interface UpdateRoleRequest extends CreateRoleRequest {
  id: number
}

export const getRoleList = (params: RoleListRequest) =>
  request.get<RoleListResponse>('/api/admin/role/list', params)

export const createRole = (data: CreateRoleRequest) => request.post('/api/admin/role/create', data)

export const updateRole = (data: UpdateRoleRequest) => request.post('/api/admin/role/update', data)

export const deleteRole = (id: number) => request.post('/api/admin/role/delete', { id })

export const getRolePermissions = (id: number) =>
  request.get<PermissionItem[]>('/api/admin/role/permissions', { id })

export const assignRolePermissions = (roleId: number, permissionIds: number[]) =>
  request.post('/api/admin/role/assign/permissions', { roleId, permissionIds })
