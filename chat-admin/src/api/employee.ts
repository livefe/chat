import { request, uploadForm } from './http'
import type { RoleItem } from './role'

export interface EmployeeInfo {
  id: number
  mobile: string
  name: string
  employeeNo: string
  avatarUrl: string
  department: string
  status: number
}

export interface EmployeeItem {
  id: number
  mobile: string
  name: string
  employeeNo: string
  avatarUrl: string
  departmentId: number
  department: string
  roles: { id: number; name: string; code: string }[]
  status: number
  statusText: string
  remark: string
  createdAt: string
  updatedAt: string
}

export interface EmployeeListRequest {
  page: number
  pageSize: number
  mobile?: string
  name?: string
  departmentId?: number
  status?: number
}

export interface EmployeeListResponse {
  total: number
  list: EmployeeItem[]
}

export const getEmployeeList = (params: EmployeeListRequest) =>
  request.get<EmployeeListResponse>('/api/admin/employee/list', params)

export const getEmployeeMe = () => request.get<EmployeeInfo>('/api/admin/employee/me')

export const updateMyProfile = (data: { avatarUrl?: string }) =>
  request.post<EmployeeInfo>('/api/admin/employee/me/update', data)

export const uploadMyAvatar = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<{ id: number; url: string }>('/api/admin/employee/me/avatar', form)
}

export const createEmployee = (data: {
  mobile: string
  name?: string
  employeeNo?: string
  departmentId: number
  status: number
  remark?: string
}) => request.post('/api/admin/employee/create', data)

export const updateEmployee = (data: {
  id: number
  mobile: string
  name?: string
  employeeNo?: string
  departmentId: number
  status: number
  remark?: string
}) => request.post('/api/admin/employee/update', data)

export const updateEmployeeStatus = (id: number, status: number) =>
  request.post('/api/admin/employee/update/status', { id, status })

export const getEmployeeRoles = (id: number) =>
  request.get<RoleItem[]>('/api/admin/employee/roles', { id })

export const assignEmployeeRoles = (employeeId: number, roleIds: number[]) =>
  request.post('/api/admin/employee/assign/roles', { employeeId, roleIds })
