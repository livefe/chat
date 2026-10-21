<template>
  <div>
    <div class="section-title">{{ isH5Form ? '登记网页版' : '上传安装包' }}</div>

    <el-form label-width="90px" class="upload-form" @submit.prevent>
      <el-form-item label="平台" required>
        <el-select
          v-model="form.platform"
          style="width: 220px"
          :disabled="uploading"
          @visible-change="(v: boolean) => v && ensureValidPlatform()"
        >
          <el-option
            v-for="opt in platformOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="版本号" required>
        <el-input
          v-model="form.version"
          placeholder="如 1.0.0"
          maxlength="32"
          style="width: 220px"
          :disabled="uploading"
        />
      </el-form-item>
      <el-form-item label="更新说明">
        <el-input
          v-model="form.changelog"
          type="textarea"
          :rows="2"
          maxlength="1000"
          show-word-limit
          placeholder="可选，软更新时展示"
          :disabled="uploading"
        />
      </el-form-item>
      <el-form-item v-if="isH5Form" label="入口地址" required>
        <el-input
          v-model="form.downloadUrl"
          placeholder="用户更新时打开的地址，如 https://example.com/"
          maxlength="512"
          style="width: 480px"
          :disabled="uploading"
        />
      </el-form-item>
      <el-form-item v-else label="安装包" required>
        <el-upload
          :auto-upload="false"
          :limit="1"
          :file-list="fileList"
          :on-change="onFileChange"
          :on-remove="onFileRemove"
          :disabled="uploading"
        >
          <el-button :disabled="uploading">选择文件</el-button>
        </el-upload>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="uploading" @click="handleSubmit">
          {{ isH5Form ? '登记' : '上传' }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type UploadUserFile } from 'element-plus'
import { publishAppReleaseMeta, uploadAppRelease } from '@/api/appRelease'
import { usePermission } from '@/composables/usePermission'
import { APP_PLATFORM_OPTIONS, type AppPlatform } from '@/constants/appRelease'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'

const emit = defineEmits<{ published: [] }>()

const { can } = usePermission()
const canUploadPackage = computed(() => can(PERM.OpsAppReleaseUpload))
const canPublishMeta = computed(() => can(PERM.OpsAppReleasePublish))

const platformOptions = computed(() =>
  APP_PLATFORM_OPTIONS.filter((opt) => {
    if (opt.value === 'h5') return canPublishMeta.value
    return canUploadPackage.value
  }),
)

const uploading = ref(false)
const form = reactive({
  platform: 'android' as AppPlatform,
  version: '',
  changelog: '',
  downloadUrl: '',
})
const fileList = ref<UploadUserFile[]>([])
const selectedFile = ref<File | null>(null)

const isH5Form = computed(() => form.platform === 'h5')

function ensureValidPlatform() {
  const allowed = platformOptions.value.map((o) => o.value)
  const first = allowed[0]
  if (first && !allowed.includes(form.platform)) {
    form.platform = first
  }
}

function onFileChange(uploadFile: UploadUserFile, uploadFiles: UploadUserFile[]) {
  selectedFile.value = (uploadFile.raw as File) || null
  fileList.value = uploadFiles.slice(-1)
}

function onFileRemove() {
  selectedFile.value = null
  fileList.value = []
}

async function handleSubmit() {
  const platform = form.platform
  if (!platform) {
    ElMessage.warning('请选择平台')
    return
  }
  if (platform === 'h5' && !canPublishMeta.value) {
    ElMessage.warning('暂无网页版登记权限')
    return
  }
  if (platform !== 'h5' && !canUploadPackage.value) {
    ElMessage.warning('暂无安装包上传权限')
    return
  }
  const version = form.version.trim()
  if (!/^\d+\.\d+\.\d+$/.test(version)) {
    ElMessage.warning('版本号需为 x.y.z，如 1.0.0')
    return
  }
  if (platform === 'h5') {
    const downloadUrl = form.downloadUrl.trim()
    if (!downloadUrl) {
      ElMessage.warning('请填写网页版入口地址')
      return
    }
    await submit('登记失败', async () => {
      await publishAppReleaseMeta({
        platform: 'h5',
        version,
        downloadUrl,
        changelog: form.changelog.trim(),
      })
      ElMessage.success('登记成功')
      form.downloadUrl = ''
    })
    return
  }
  const file = selectedFile.value
  if (!file) {
    ElMessage.warning('请选择安装包文件')
    return
  }
  await submit('上传失败', async () => {
    await uploadAppRelease({
      platform,
      version,
      changelog: form.changelog.trim(),
      file,
    })
    ElMessage.success('上传成功')
    selectedFile.value = null
    fileList.value = []
  })
}

async function submit(errorText: string, action: () => Promise<void>) {
  uploading.value = true
  try {
    await action()
    form.version = ''
    form.changelog = ''
    emit('published')
  } catch (e) {
    showApiError(e, errorText)
  } finally {
    uploading.value = false
  }
}

onMounted(ensureValidPlatform)
</script>

<style scoped>
.section-title {
  margin-bottom: 12px;
}

.upload-form {
  max-width: 640px;
}
</style>
