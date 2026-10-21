<template>
  <el-drawer v-model="visible" title="部门详情" size="420px">
    <template v-if="department">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ department.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ department.name }}</el-descriptions-item>
        <el-descriptions-item label="上级">{{
          department.parentName || '顶级'
        }}</el-descriptions-item>
        <el-descriptions-item label="层级">{{ department.level }}</el-descriptions-item>
        <el-descriptions-item label="员工数">{{ department.employeeCount }}</el-descriptions-item>
        <el-descriptions-item label="排序">{{ department.sort }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="department.status === 1 ? 'success' : 'danger'">{{
            department.statusText
          }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="备注">{{ department.remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ department.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ department.updatedAt }}</el-descriptions-item>
      </el-descriptions>
      <div class="detail-actions">
        <el-button
          v-if="department.employeeCount > 0 && can(PERM.OrgEmployee)"
          type="primary"
          @click="emit('employees', department)"
        >
          查看员工
        </el-button>
        <el-button
          v-if="can(PERM.OrgDepartmentCreate) && department.level < 5"
          @click="emit('add-child', department)"
        >
          新增下级
        </el-button>
        <el-button v-if="can(PERM.OrgDepartmentUpdate)" @click="emit('edit', department)">
          编辑
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import type { DepartmentItem } from '@/api/department'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'

const visible = defineModel<boolean>({ required: true })
defineProps<{ department: DepartmentItem | null }>()
const emit = defineEmits<{
  employees: [row: DepartmentItem]
  'add-child': [row: DepartmentItem]
  edit: [row: DepartmentItem]
}>()

const { can } = usePermission()
</script>
