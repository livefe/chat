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
        <el-select v-model="platform" style="width: 220px">
          <el-option
            v-for="opt in APP_PLATFORM_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
        <span v-if="!currentPolicy" class="hint form-hint">该平台尚未配置策略</span>
      </el-form-item>
      <el-form-item label="最低支持版本" required>
        <el-input
          v-model="form.minSupportedVersion"
          placeholder="如 1.0.0"
          maxlength="32"
          style="width: 220px"
        />
        <span v-if="currentPolicy?.latestVersion" class="hint form-hint">
          当前对外 {{ currentPolicy.latestVersion }}，不能高于该版本
        </span>
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
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { updateAppUpdatePolicy, type AppUpdatePolicyItem } from '@/api/appRelease'
import {
  APP_PLATFORM_OPTIONS,
  SEMVER_PATTERN,
  compareSemver,
  type AppPlatform,
} from '@/constants/appRelease'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{ policies: AppUpdatePolicyItem[] }>()
const emit = defineEmits<{ saved: [] }>()

const saving = ref(false)
const platform = ref<AppPlatform>('android')
const form = reactive({ minSupportedVersion: '', title: '', message: '' })

const currentPolicy = computed(() => props.policies.find((p) => p.platform === platform.value))

watch(
  currentPolicy,
  (row) => {
    form.minSupportedVersion = row?.minSupportedVersion || ''
    form.title = row?.title || ''
    form.message = row?.message || ''
  },
  { immediate: true },
)

function confirmText(prev: string, next: string, latest: string) {
  const label = APP_PLATFORM_OPTIONS.find((o) => o.value === platform.value)?.label
  if (!prev) return `为 ${label} 设置最低支持版本 ${next}，低于该版本的客户端将被强制更新。`
  const cmp = compareSemver(next, prev)
  if (cmp > 0) {
    const all = latest && next === latest ? '（等于当前对外版本，所有旧版本都将被强制更新）' : ''
    return `${label} 最低支持版本将从 ${prev} 提高到 ${next}${all}。版本低于 ${next} 的客户端将被强制更新，无法继续使用旧版。`
  }
  if (cmp < 0) {
    return `${label} 最低支持版本将从 ${prev} 降到 ${next}，版本在 ${next} 与 ${prev} 之间的客户端不再被强制更新。`
  }
  return ''
}

async function handleSave() {
  const minVer = form.minSupportedVersion.trim()
  if (!SEMVER_PATTERN.test(minVer)) {
    ElMessage.warning('最低版本需为 x.y.z')
    return
  }
  const latest = currentPolicy.value?.latestVersion || ''
  if (latest && compareSemver(minVer, latest) > 0) {
    ElMessage.warning(`最低支持版本不能高于当前对外版本 ${latest}`)
    return
  }
  const text = confirmText(currentPolicy.value?.minSupportedVersion || '', minVer, latest)
  if (text) {
    try {
      await ElMessageBox.confirm(text, '确认修改最低支持版本', {
        type: 'warning',
        confirmButtonText: '确认保存',
      })
    } catch {
      return
    }
  }
  saving.value = true
  try {
    await updateAppUpdatePolicy({
      platform: platform.value,
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

.form-hint {
  margin-left: 12px;
}

.policy-table {
  margin-bottom: 16px;
}

.policy-form {
  max-width: 720px;
}
</style>
