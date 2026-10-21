<template>
  <div>
    <div class="section-head">
      <div class="section-title">更新策略</div>
      <span class="hint">低于「最低支持版本」的客户端将被强制更新；低于当前包仅提示</span>
    </div>

    <el-table :data="policies" size="small" class="policy-table" empty-text="暂无更新策略">
      <el-table-column prop="platformText" label="平台" width="90" />
      <el-table-column prop="latestVersion" label="当前对外" width="110">
        <template #default="{ row }">{{ row.latestVersion || '—' }}</template>
      </el-table-column>
      <el-table-column prop="minSupportedVersion" label="最低支持" width="110" />
      <el-table-column prop="title" label="弹窗标题" min-width="120" show-overflow-tooltip />
      <el-table-column prop="updatedAt" label="更新时间" width="170" />
    </el-table>

    <el-form label-width="110px" class="policy-form" @submit.prevent>
      <el-form-item label="平台">
        <el-select v-model="form.platform" style="width: 220px" @change="onSelectPlatform">
          <el-option
            v-for="opt in APP_PLATFORM_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="最低支持版本" required>
        <el-input
          v-model="form.minSupportedVersion"
          placeholder="如 1.0.0"
          maxlength="32"
          style="width: 220px"
        />
      </el-form-item>
      <el-form-item label="弹窗标题">
        <el-input v-model="form.title" maxlength="80" style="width: 360px" />
      </el-form-item>
      <el-form-item label="更新说明">
        <el-input
          v-model="form.message"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          style="width: 480px"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="handleSave">保存策略</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { updateAppUpdatePolicy, type AppUpdatePolicyItem } from '@/api/appRelease'
import { APP_PLATFORM_OPTIONS, type AppPlatform } from '@/constants/appRelease'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{ policies: AppUpdatePolicyItem[] }>()
const emit = defineEmits<{ saved: [] }>()

const saving = ref(false)
const form = reactive({
  platform: 'android' as AppPlatform,
  minSupportedVersion: '',
  title: '',
  message: '',
})

function fillForm(row: AppUpdatePolicyItem) {
  form.minSupportedVersion = row.minSupportedVersion
  form.title = row.title
  form.message = row.message
}

watch(
  () => props.policies,
  (list) => {
    const android = list.find((p) => p.platform === 'android')
    if (!android) return
    form.platform = 'android'
    fillForm(android)
  },
)

function onSelectPlatform(platform: AppPlatform) {
  form.platform = platform
  const row = props.policies.find((p) => p.platform === platform)
  if (row) fillForm(row)
}

async function handleSave() {
  if (!/^\d+\.\d+\.\d+$/.test(form.minSupportedVersion.trim())) {
    ElMessage.warning('最低版本需为 x.y.z')
    return
  }
  const minVer = form.minSupportedVersion.trim()
  const policyRow = props.policies.find((p) => p.platform === form.platform)
  const latest = (policyRow?.latestVersion || '').trim()
  if (latest && minVer === latest) {
    try {
      await ElMessageBox.confirm(
        `最低支持版本与当前对外版本相同（${latest}）。保存后，所有更低版本客户端将被强制更新，无法继续使用旧版。确认保存？`,
        '强制更新确认',
        { type: 'warning', confirmButtonText: '确认保存', cancelButtonText: '取消' },
      )
    } catch {
      return
    }
  }
  saving.value = true
  try {
    await updateAppUpdatePolicy({
      platform: form.platform,
      minSupportedVersion: minVer,
      title: form.title.trim(),
      message: form.message.trim(),
    })
    ElMessage.success('策略已保存')
    emit('saved')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.section-head {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin-bottom: 12px;
}

.hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-weight: normal;
}

.policy-table {
  margin-bottom: 16px;
}

.policy-form {
  max-width: 640px;
}
</style>
