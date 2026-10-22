import { request } from './http'

export interface DepartmentItem {
  id: number
  parentId: number
  parentName: string
  name: string
  path: string
  level: number
  status: number
  statusText: string
  sort: number
  remark: string
  /** 直属员工数 */
  employeeCount: number
  /** 含全部下级部门的员工数 */
  totalEmployeeCount: number
  createdAt: string
  updatedAt: string
  children?: DepartmentItem[]
}

export interface DepartmentOption {
  id: number
  name: string
  children?: DepartmentOption[]
}

export interface DepartmentTreeResponse {
  list: DepartmentItem[]
}

export const getDepartmentList = () =>
  request.get<DepartmentTreeResponse>('/api/admin/department/list')

export const getDepartmentOptions = () =>
  request.get<DepartmentOption[]>('/api/admin/department/options')

export const createDepartment = (data: {
  parentId?: number
  name: string
  status: number
  sort?: number
  remark?: string
}) => request.post('/api/admin/department/create', data)

export const updateDepartment = (data: {
  id: number
  parentId?: number
  name: string
  status: number
  sort?: number
  remark?: string
}) => request.post('/api/admin/department/update', data)

export const updateDepartmentStatus = (id: number, status: number) =>
  request.post('/api/admin/department/update/status', { id, status })
