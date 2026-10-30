import { request, uploadForm } from './http'

export interface EmployeeInfo {
  id: number
  mobile: string
  name: string
  employeeNo: string
  avatarMediaId: number
  avatarUrl: string
  department: string
  status: number
}

export interface EmployeeRoleBrief {
  id: number
  name: string
  code: string
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
  roles: EmployeeRoleBrief[]
  status: number
  statusText: string
  remark: string
  lastLoginAt: string
  lastLoginIp: string
  createdAt: string
  updatedAt: string
}

export interface EmployeeListRequest {
  page: number
  pageSize: number
  id?: number
  mobile?: string
  name?: string
  employeeNo?: string
  departmentId?: number
  roleId?: number
  status?: number
  startTime?: string
  endTime?: string
}

export interface EmployeeListResponse {
  total: number
  list: EmployeeItem[]
}

export interface EmployeeRoleOptions {
  options: EmployeeRoleBrief[]
  assignedIds: number[]
  operatorIsSuper: boolean
  targetIsSuper: boolean
}

export interface EmployeeLoginRecord {
  id: number
  loginType: string
  ip: string
  userAgent: string
  status: number
  statusText: string
  remark: string
  createdAt: string
}

export const getEmployeeList = (params: EmployeeListRequest) =>
  request.get<EmployeeListResponse>('/api/admin/employee/list', params)

export const getEmployeeMe = () => request.get<EmployeeInfo>('/api/admin/employee/me')

export const updateMyProfile = (data: { avatarMediaId: number }) =>
  request.post<EmployeeInfo>('/api/admin/employee/me/update', data)

export const uploadMyAvatar = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<{ id: number; url: string }>('/api/admin/employee/me/avatar', form)
}

export const createEmployee = (data: {
  mobile: string
  name: string
  employeeNo?: string
  departmentId: number
  status: number
  remark?: string
}) => request.post('/api/admin/employee/create', data)

export const updateEmployee = (data: {
  id: number
  mobile: string
  name: string
  employeeNo?: string
  departmentId: number
  remark?: string
}) => request.post('/api/admin/employee/update', data)

export const updateEmployeeStatus = (id: number, status: number) =>
  request.post('/api/admin/employee/update/status', { id, status })

export const getEmployeeRoleOptions = (employeeId?: number) =>
  request.get<EmployeeRoleOptions>('/api/admin/employee/role-options', { employeeId })

export const assignEmployeeRoles = (employeeId: number, roleIds: number[]) =>
  request.post('/api/admin/employee/assign/roles', { employeeId, roleIds })

export const getEmployeeLoginRecords = (params: {
  employeeId: number
  page: number
  pageSize: number
}) =>
  request.get<{ total: number; list: EmployeeLoginRecord[] }>(
    '/api/admin/employee/login-records',
    params,
  )
