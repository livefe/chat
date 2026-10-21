<template>
  <div v-loading="loading" class="support-config">
    <el-alert
      v-if="!canSave"
      type="info"
      :closable="false"
      show-icon
      title="当前为只读模式，如需修改请联系管理员分配客服配置权限"
      class="support-config__alert"
    />
    <el-form :model="form" label-width="120px" class="support-config__form" @submit.prevent>
      <el-form-item label="品牌名称">
        <el-input
          v-model="form.brandName"
          placeholder="客服品牌名称"
          maxlength="32"
          show-word-limit
          :disabled="!canSave"
        />
      </el-form-item>
      <el-form-item label="品牌头像">
        <div class="support-config__avatar">
          <el-avatar :size="56" :src="form.brandAvatarUrl || undefined">
            {{ (form.brandName || '客').slice(0, 1) }}
          </el-avatar>
          <div class="support-config__avatar-ops">
            <el-upload
              :show-file-list="false"
              accept="image/*"
              :http-request="onUploadAvatar"
              :disabled="!canUploadAvatar || uploading"
            >
              <el-button :loading="uploading" :disabled="!canUploadAvatar">上传头像</el-button>
            </el-upload>
            <el-input
              v-model="form.brandAvatarUrl"
              placeholder="或粘贴图片 URL"
              clearable
              :disabled="!canSave"
            />
          </div>
        </div>
      </el-form-item>
      <el-form-item label="欢迎语">
        <el-input
          v-model="form.welcomeText"
          type="textarea"
          :rows="3"
          placeholder="用户进入客服时的欢迎文案"
          maxlength="500"
          show-word-limit
          :disabled="!canSave"
        />
      </el-form-item>
      <el-form-item label="离线提示">
        <el-input
          v-model="form.offlineText"
          type="textarea"
          :rows="3"
          placeholder="客服离线时的提示文案"
          maxlength="500"
          show-word-limit
          :disabled="!canSave"
        />
      </el-form-item>
      <el-form-item label="启用客服">
        <el-switch v-model="form.enabled" :disabled="!canSave" />
      </el-form-item>
      <el-form-item>
        <el-button v-if="canSave" type="primary" :loading="saving" @click="handleSave">
          保存
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { ElMessage, type UploadRequestHandler } from 'element-plus'
import { showApiError } from '@/utils/apiError'
import { getSupportConfig, updateSupportConfig } from '@/api/support'
import { uploadImImage } from '@/api/im'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const { can } = usePermission()

const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)

const form = reactive({
  brandName: '',
  brandAvatarUrl: '',
  welcomeText: '',
  offlineText: '',
  enabled: true,
})

const canSave = computed(() => can(PERM.ImSupportOverviewConfig))
const canUploadAvatar = computed(() => canSave.value && can(PERM.OpsUpload))

async function loadConfig() {
  loading.value = true
  try {
    const res = await getSupportConfig()
    const data = res.data
    form.brandName = data.brandName || ''
    form.brandAvatarUrl = data.brandAvatarUrl || ''
    form.welcomeText = data.welcomeText || ''
    form.offlineText = data.offlineText || ''
    form.enabled = !!data.enabled
  } catch (e) {
    showApiError(e, '加载客服配置失败')
  } finally {
    loading.value = false
  }
}

const onUploadAvatar: UploadRequestHandler = async (options) => {
  uploading.value = true
  try {
    const { data } = await uploadImImage(options.file as File)
    form.brandAvatarUrl = data.url || ''
    options.onSuccess?.(data)
    ElMessage.success('头像已上传，请保存配置')
  } catch (e) {
    options.onError?.(e as Parameters<NonNullable<typeof options.onError>>[0])
    showApiError(e, '上传头像失败')
  } finally {
    uploading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateSupportConfig({
      brandName: form.brandName,
      brandAvatarUrl: form.brandAvatarUrl,
      welcomeText: form.welcomeText,
      offlineText: form.offlineText,
      enabled: form.enabled ? 1 : 0,
    })
    ElMessage.success('保存成功')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.support-config__alert {
  margin-bottom: 16px;
}

.support-config__form {
  max-width: 560px;
}

.support-config__avatar {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  width: 100%;
}

.support-config__avatar-ops {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}
</style>
