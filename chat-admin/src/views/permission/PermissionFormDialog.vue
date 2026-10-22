<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑权限' : '新增权限'"
    width="640px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="perm-hint"
      :title="PERMISSION_MODEL_HINT"
    />
    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="类型" prop="type">
        <el-radio-group v-model="form.type" :disabled="apiPathLocked">
          <el-radio-button v-for="t in permissionTypeOptions" :key="t.value" :value="t.value">
            {{ t.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入名称" />
      </el-form-item>
      <el-form-item label="编码" prop="code">
        <el-input v-model="form.code" :placeholder="codePlaceholder" />
      </el-form-item>
      <el-form-item label="父级" prop="parentId">
        <el-tree-select
          v-model="form.parentId"
          :data="parentTreeData"
          :props="{ label: 'name', children: 'children' }"
          node-key="id"
          check-strictly
          default-expand-all
          filterable
          :render-after-expand="false"
          placeholder="请选择父级（空为顶级）"
          style="width: 100%"
        />
        <div class="form-tip">
          {{
            form.type === PermissionType.Menu
              ? '菜单可选顶级或其他菜单作为父级'
              : '按钮、接口挂到所属菜单下；只读接口用菜单编码，写接口用按钮编码'
          }}
        </div>
      </el-form-item>
      <el-form-item
        v-if="form.type !== PermissionType.Button"
        label="路径"
        :prop="form.type === PermissionType.API ? 'path' : undefined"
      >
        <el-input
          v-model="form.path"
          :placeholder="form.type === PermissionType.API ? '/api/admin/user/list' : '/org/employee'"
          :disabled="apiPathLocked"
        />
        <div v-if="apiPathLocked" class="form-tip">接口路径创建后不可修改</div>
      </el-form-item>
      <el-form-item v-if="form.type === PermissionType.API" label="方法" prop="method">
        <el-select
          v-model="form.method"
          placeholder="请选择"
          style="width: 160px"
          :disabled="apiPathLocked"
        >
          <el-option v-for="m in HTTP_METHODS" :key="m" :label="m" :value="m" />
        </el-select>
        <div v-if="apiPathLocked" class="form-tip">请求方法创建后不可修改</div>
      </el-form-item>
      <template v-if="form.type === PermissionType.Menu">
        <el-form-item label="图标">
          <el-select
            v-model="form.icon"
            filterable
            clearable
            allow-create
            default-first-option
            placeholder="选择或输入图标名"
            style="width: 100%"
          >
            <el-option v-for="name in ADMIN_ICON_NAMES" :key="name" :label="name" :value="name">
              <span class="perm-icon-option">
                <el-icon><component :is="name" /></el-icon>
                <span>{{ name }}</span>
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="组件">
          <el-select
            v-model="form.component"
            filterable
            clearable
            allow-create
            default-first-option
            placeholder="选择路由组件（须在 componentMap 中）"
            style="width: 100%"
          >
            <el-option v-for="name in componentOptions" :key="name" :label="name" :value="name" />
          </el-select>
          <div class="form-tip">仅叶子菜单需要组件；分组菜单可留空</div>
        </el-form-item>
        <el-form-item label="隐藏菜单">
          <el-switch v-model="form.hideInMenu" />
        </el-form-item>
      </template>
      <el-form-item label="排序">
        <el-input-number v-model="form.sort" :min="0" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" />
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">禁用</el-radio>
        </el-radio-group>
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
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  createPermission,
  updatePermission,
  type PermissionItem,
  type PermissionTreeItem,
} from '@/api/permission'
import { ADMIN_ICON_NAMES } from '@/constants/adminIcons'
import { PermissionType } from '@/constants/permission'
import { componentMap } from '@/router/routes'
import { showApiError } from '@/utils/apiError'
import { HTTP_METHODS, PERMISSION_MODEL_HINT, permissionTypeOptions } from '@/utils/permission'

interface ParentOption {
  id: number
  name: string
  children?: ParentOption[]
}

const props = defineProps<{ tree: PermissionTreeItem[] }>()
const emit = defineEmits<{ success: [] }>()

const componentOptions = Object.keys(componentMap).sort()

const visible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

function emptyForm() {
  return {
    id: 0,
    name: '',
    code: '',
    description: '',
    type: PermissionType.Menu as number,
    parentId: 0,
    path: '',
    method: '',
    icon: '',
    component: '',
    hideInMenu: false,
    sort: 0,
    status: 1,
  }
}

const form = reactive(emptyForm())

const rules: FormRules = {
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择权限类型', trigger: 'change' }],
  path: [
    {
      validator: (_rule, v: string, cb) => {
        if (form.type === PermissionType.API && !v) return cb(new Error('接口权限必须填写路径'))
        cb()
      },
      trigger: 'blur',
    },
  ],
  method: [
    {
      validator: (_rule, v: string, cb) => {
        if (form.type === PermissionType.API && !v) return cb(new Error('接口权限必须选择方法'))
        cb()
      },
      trigger: 'change',
    },
  ],
}

const codePlaceholder = computed(() => {
  switch (form.type) {
    case PermissionType.Menu:
      return '如：user:account'
    case PermissionType.Button:
      return '如：user:account:create'
    default:
      return '只读用菜单编码如 user:account；写操作用按钮编码如 user:account:create'
  }
})

const apiPathLocked = computed(() => form.id > 0 && form.type === PermissionType.API)

const parentTreeData = computed<ParentOption[]>(() => [
  { id: 0, name: '顶级（无父级）' },
  ...buildParentTree(props.tree, form.id),
])

function buildParentTree(nodes: PermissionTreeItem[], excludeId: number): ParentOption[] {
  const out: ParentOption[] = []
  for (const n of nodes) {
    if (n.id === excludeId || n.type !== PermissionType.Menu) continue
    out.push({
      id: n.id,
      name: n.name,
      children: n.children ? buildParentTree(n.children, excludeId) : undefined,
    })
  }
  return out
}

watch(
  () => form.type,
  (t, old) => {
    if (t === old) return
    if (t !== PermissionType.API) form.method = ''
    if (t === PermissionType.Button) form.path = ''
    if (t !== PermissionType.Menu) {
      form.icon = ''
      form.component = ''
      form.hideInMenu = false
    }
  },
)

function openCreate(parent?: PermissionItem) {
  Object.assign(form, emptyForm())
  if (parent) {
    form.parentId = parent.id
    form.type = PermissionType.Button
    form.code = `${parent.code}:`
  }
  visible.value = true
}

function openEdit(row: PermissionItem) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    type: row.type,
    parentId: row.parentId,
    path: row.path,
    method: row.method,
    icon: row.icon || '',
    component: row.component || '',
    hideInMenu: !!row.hideInMenu,
    sort: row.sort,
    status: row.status,
  })
  visible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    if (form.id) await updatePermission(form)
    else await createPermission(form)
    ElMessage.success(form.id ? '已保存' : '已新增')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ openCreate, openEdit })
</script>

<style scoped>
.perm-hint {
  margin-bottom: 16px;
}

.perm-icon-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
</style>
