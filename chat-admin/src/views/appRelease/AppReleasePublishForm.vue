<template>
  <div>
    <div class="section-title">{{ isH5Form ? '登记网页版' : '上传安装包' }}</div>

    <el-form label-width="90px" class="upload-form" @submit.prevent>
      <el-form-item label="平台" required>
        <el-select v-model="form.platform" style="width: 220px" :disabled="uploading">
          <el-option
            v-for="opt in platformOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
        <span v-if="currentVersion" class="form-hint">当前对外 {{ currentVersion }}</span>
      </el-form-item>
      <el-form-item label="版本号" required>
        <el-input
          v-model="form.version"
          placeholder="如 1.0.0"
          maxlength="32"
          style="width: 220px"
          :disabled="uploading"
        />
        <span v-if="lowerThanCurrent" class="form-hint is-warn">
          低于当前对外版本 {{ currentVersion }}
        </span>
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
        <div class="upload-field">
          <el-upload
            :auto-upload="false"
            :limit="1"
            :file-list="fileList"
            :accept="packageExt"
            :on-change="onFileChange"
            :on-remove="onFileRemove"
            :on-exceed="onFileExceed"
            :disabled="uploading"
          >
            <el-button :disabled="uploading">选择文件</el-button>
          </el-upload>
          <div class="form-hint form-hint--block">
            仅支持 {{ packageExt }} 文件，不超过 {{ APP_PACKAGE_MAX_MB }}MB
          </div>
          <el-progress v-if="uploading" :percentage="progress" class="upload-progress" />
        </div>
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
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { publishAppReleaseMeta, uploadAppRelease, type AppUpdatePolicyItem } from '@/api/appRelease'
import { usePermission } from '@/composables/usePermission'
import {
  APP_PACKAGE_EXT,
  APP_PACKAGE_MAX_MB,
  APP_PLATFORM_OPTIONS,
  SEMVER_PATTERN,
  compareSemver,
  type AppPlatform,
} from '@/constants/appRelease'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{ policies: AppUpdatePolicyItem[] }>()
const emit = defineEmits<{ published: [] }>()

const { can } = usePermission()
const canUploadPackage = computed(() => can(PERM.OpsAppReleaseUpload))
const canPublishMeta = computed(() => can(PERM.OpsAppReleasePublish))

const platformOptions = computed(() =>
  APP_PLATFORM_OPTIONS.filter((opt) =>
    opt.value === 'h5' ? canPublishMeta.value : canUploadPackage.value,
  ),
)

const uploading = ref(false)
const progress = ref(0)
const form = reactive({
  platform: (platformOptions.value[0]?.value || 'android') as AppPlatform,
  version: '',
  changelog: '',
  downloadUrl: '',
})
const fileList = ref<UploadUserFile[]>([])
const selectedFile = ref<File | null>(null)

const isH5Form = computed(() => form.platform === 'h5')
const packageExt = computed(() => (form.platform === 'h5' ? '' : APP_PACKAGE_EXT[form.platform]))
const currentVersion = computed(
  () => props.policies.find((p) => p.platform === form.platform)?.latestVersion || '',
)
const lowerThanCurrent = computed(() => {
  const v = form.version.trim()
  return (
    !!currentVersion.value && SEMVER_PATTERN.test(v) && compareSemver(v, currentVersion.value) < 0
  )
})

watch(
  () => form.platform,
  () => onFileRemove(),
)

function checkFile(file: File) {
  const ext = packageExt.value
  if (!file.name.toLowerCase().endsWith(ext)) {
    return `所选平台只能上传 ${ext} 文件`
  }
  if (file.size > APP_PACKAGE_MAX_MB * 1024 * 1024) {
    return `安装包不能超过 ${APP_PACKAGE_MAX_MB}MB`
  }
  return ''
}

function onFileChange(uploadFile: UploadUserFile) {
  const file = uploadFile.raw as File | undefined
  const err = file ? checkFile(file) : '请选择安装包文件'
  if (err) {
    ElMessage.warning(err)
    onFileRemove()
    return
  }
  selectedFile.value = file || null
  fileList.value = [uploadFile]
}

function onFileExceed() {
  ElMessage.warning('一次只能上传一个安装包，请先移除已选文件')
}

function onFileRemove() {
  selectedFile.value = null
  fileList.value = []
}

async function confirmLowerVersion(version: string) {
  if (!lowerThanCurrent.value) return true
  try {
    await ElMessageBox.confirm(
      `版本 ${version} 低于当前对外版本 ${currentVersion.value}，确认继续发布？发布后不会自动生效，需手动「设为当前」。`,
      '版本号偏低',
      { type: 'warning', confirmButtonText: '继续' },
    )
    return true
  } catch {
    return false
  }
}

async function handleSubmit() {
  const platform = form.platform
  const version = form.version.trim()
  if (!SEMVER_PATTERN.test(version)) {
    ElMessage.warning('版本号需为 x.y.z，如 1.0.0')
    return
  }
  if (platform === 'h5') {
    const downloadUrl = form.downloadUrl.trim()
    if (!/^https?:\/\//.test(downloadUrl)) {
      ElMessage.warning('请填写以 http:// 或 https:// 开头的入口地址')
      return
    }
    if (!(await confirmLowerVersion(version))) return
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
  if (!(await confirmLowerVersion(version))) return
  progress.value = 0
  await submit('上传失败', async () => {
    await uploadAppRelease(
      { platform, version, changelog: form.changelog.trim(), file },
      (p) => (progress.value = p),
    )
    ElMessage.success('上传成功，如需对外发布请在列表中「设为当前」')
    onFileRemove()
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
</script>

<style scoped>
.section-title {
  margin-bottom: 12px;
}

.upload-form {
  max-width: 720px;
}

.upload-field {
  width: 100%;
}

.form-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.form-hint--block {
  display: block;
  margin-left: 0;
  line-height: 1.6;
}

.form-hint.is-warn {
  color: var(--el-color-warning);
}

.upload-progress {
  max-width: 360px;
  margin-top: 6px;
}
</style>
