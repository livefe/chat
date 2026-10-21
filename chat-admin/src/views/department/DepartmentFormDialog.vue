<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑部门' : '新增部门'"
    width="520px"
    destroy-on-close
    :close-on-click-modal="false"
    @closed="Object.assign(form, defaultForm())"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
      <el-form-item label="上级部门">
        <el-tree-select
          v-model="form.parentId"
          :data="parentTree"
          check-strictly
          clearable
          filterable
          :render-after-expand="false"
          placeholder="不选则为顶级部门"
          style="width: 100%"
          node-key="id"
          :props="{ label: 'name', children: 'children', disabled: 'disabled' }"
        />
      </el-form-item>
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入部门名称" maxlength="64" />
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort" :min="0" :max="9999" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" placeholder="备注" maxlength="255" />
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio :label="1">启用</el-radio>
          <el-radio :label="0">禁用</el-radio>
        </el-radio-group>
        <div
          v-if="form.id && form.status === 0 && form.employeeCount > 0"
          class="form-tip is-warning"
        >
          该部门下有 {{ form.employeeCount }} 名员工，无法禁用
        </div>
        <div
          v-else-if="form.id && form.status === 0 && form.hasChildDepts"
          class="form-tip is-warning"
        >
          请先处理下级部门后再禁用
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { createDepartment, updateDepartment, type DepartmentItem } from '@/api/department'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{ tree: DepartmentItem[] }>()
const emit = defineEmits<{ success: [id?: number] }>()

type ParentNode = DepartmentItem & { disabled?: boolean }

const defaultForm = () => ({
  id: 0,
  parentId: undefined as number | undefined,
  name: '',
  sort: 0,
  remark: '',
  status: 1,
  employeeCount: 0,
  hasChildDepts: false,
})

const visible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive(defaultForm())
const rules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
}

function markDisabled(
  nodes: DepartmentItem[],
  excludeId: number,
  underExclude = false,
): ParentNode[] {
  return nodes.map((n) => {
    const hit = underExclude || n.id === excludeId
    return {
      ...n,
      disabled: hit,
      children: n.children?.length ? markDisabled(n.children, excludeId, hit) : [],
    }
  })
}

const parentTree = computed(() => {
  if (!form.id) return props.tree
  return markDisabled(props.tree, form.id)
})

async function open(data: ReturnType<typeof defaultForm>) {
  Object.assign(form, data)
  visible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

function openCreate(parent?: DepartmentItem | null) {
  open({ ...defaultForm(), parentId: parent?.id })
}

function openEdit(row: DepartmentItem) {
  open({
    id: row.id,
    parentId: row.parentId || undefined,
    name: row.name,
    sort: row.sort,
    remark: row.remark,
    status: row.status,
    employeeCount: row.employeeCount ?? 0,
    hasChildDepts: !!row.children?.length,
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  if (form.id && form.status === 0 && form.employeeCount > 0) {
    ElMessage.warning(`该部门下仍有 ${form.employeeCount} 名员工，无法禁用`)
    return
  }
  if (form.id && form.status === 0 && form.hasChildDepts) {
    ElMessage.warning('请先处理下级部门后再禁用')
    return
  }
  await formRef.value.validate()
  saving.value = true
  try {
    const payload = {
      parentId: form.parentId || 0,
      name: form.name,
      sort: form.sort,
      remark: form.remark,
      status: form.status,
    }
    if (form.id) {
      await updateDepartment({ id: form.id, ...payload })
      visible.value = false
      emit('success', form.id)
    } else {
      await createDepartment(payload)
      visible.value = false
      emit('success')
    }
    ElMessage.success('操作成功')
  } catch (e) {
    showApiError(e, '操作失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ openCreate, openEdit })
</script>
