<template>
  <el-dialog
    v-model="formVisible"
    :title="form.id ? '编辑计划' : '新建计划'"
    :width="form.hasGeo ? '960px' : '720px'"
    destroy-on-close
    :close-on-click-modal="false"
    @closed="onFormClosed"
  >
    <div class="dialog-body" :class="{ 'dialog-body--map': form.hasGeo }">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px" class="form-main">
        <el-form-item label="广告主" prop="advertiserName">
          <el-input v-model="form.advertiserName" maxlength="80" show-word-limit />
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="80" show-word-limit />
        </el-form-item>
        <el-form-item label="副标题" prop="subtitle">
          <el-input v-model="form.subtitle" maxlength="120" show-word-limit />
        </el-form-item>
        <el-form-item label="封面图" prop="imageMediaId">
          <div class="upload-row">
            <el-upload
              :show-file-list="false"
              accept="image/*"
              :http-request="handleImageUpload"
              :disabled="uploading || !canUpload"
            >
              <el-button :loading="uploading" :disabled="!canUpload">上传图片</el-button>
            </el-upload>
            <span v-if="form.imageMediaId" class="upload-hint"
              >mediaId: {{ form.imageMediaId }}</span
            >
          </div>
        </el-form-item>
        <el-form-item label="落地类型" prop="landingType">
          <el-radio-group v-model="form.landingType">
            <el-radio v-for="opt in AD_LANDING_OPTIONS" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="落地值" prop="landingValue">
          <el-input v-model="form.landingValue" :placeholder="landingPlaceholder" maxlength="500" />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-input-number v-model="form.priority" :min="1" :max="9999" />
        </el-form-item>
        <el-form-item label="开始时间" prop="startAt">
          <el-date-picker
            v-model="form.startAt"
            type="datetime"
            placeholder="开始时间"
            value-format="X"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束时间" prop="endAt">
          <el-date-picker
            v-model="form.endAt"
            type="datetime"
            placeholder="结束时间"
            value-format="X"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="CPD 单价">
          <el-input-number v-model="form.cpdPrice" :min="0" :precision="2" :step="1" />
        </el-form-item>
        <el-form-item label="地理定向">
          <el-switch v-model="form.hasGeo" active-text="按地图范围投放" inactive-text="不限" />
        </el-form-item>
        <el-form-item v-if="form.hasGeo" label="投放范围" class="geo-map-item">
          <AmapPicker
            :key="`ad-geo-${form.id}-${form.hasGeo}`"
            :shape-type="1"
            circle-only
            :center-lat="form.latitude"
            :center-lng="form.longitude"
            :radius="form.radiusM || 3000"
            :min-radius="100"
            :max-radius="50000"
            @change="onGeoMapChange"
          />
        </el-form-item>
        <el-form-item label="日曝光上限">
          <el-input-number v-model="form.dailyImpressionLimit" :min="0" :step="100" />
          <span class="field-hint">0 表示不限</span>
        </el-form-item>
        <el-form-item label="提交审核">
          <el-checkbox v-model="form.submitAudit">保存后进入待审核</el-checkbox>
        </el-form-item>
      </el-form>

      <AdCampaignPreview
        :image-url="form.imageUrl"
        :title="form.title"
        :advertiser-name="form.advertiserName"
      />
    </div>
    <template #footer>
      <el-button @click="formVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref } from 'vue'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { ElMessage } from 'element-plus'
import {
  getAdCampaignDetail,
  createAdCampaign,
  updateAdCampaign,
  uploadAdImage,
  type AdCampaignItem,
} from '@/api/ad'
import AmapPicker from '@/components/amap/AmapPicker.vue'
import { showApiError } from '@/utils/apiError'
import { hasCoordinates } from '@/utils/geo'
import { AD_LANDING_OPTIONS } from '@/constants/ad'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'
import AdCampaignPreview from './AdCampaignPreview.vue'

const emit = defineEmits<{ success: [] }>()

const { can } = usePermission()
const canUpload = computed(() => can(PERM.OpsUpload))

const formVisible = ref(false)
const saving = ref(false)
const uploading = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  id: 0,
  advertiserName: '',
  title: '',
  subtitle: '',
  imageMediaId: 0,
  imageUrl: '',
  landingType: 'url',
  landingValue: '',
  priority: 100,
  startAt: '' as string | number,
  endAt: '' as string | number,
  cpdPrice: 0,
  hasGeo: false,
  latitude: 0,
  longitude: 0,
  radiusM: 3000,
  dailyImpressionLimit: 0,
  submitAudit: false,
})

const landingPlaceholder = computed(() => {
  switch (form.landingType) {
    case 'post':
      return '帖子 ID'
    case 'room':
      return '聊天区 ID'
    default:
      return 'https://...'
  }
})

const formRules: FormRules = {
  advertiserName: [{ required: true, message: '请输入广告主', trigger: 'blur' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  imageMediaId: [
    {
      validator: (_r, v, cb) => (v ? cb() : cb(new Error('请上传封面图'))),
      trigger: 'change',
    },
  ],
  landingType: [{ required: true, message: '请选择落地类型', trigger: 'change' }],
  landingValue: [{ required: true, message: '请输入落地值', trigger: 'blur' }],
  startAt: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  endAt: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

function defaultForm() {
  const now = Math.floor(Date.now() / 1000)
  return {
    id: 0,
    advertiserName: '',
    title: '',
    subtitle: '',
    imageMediaId: 0,
    imageUrl: '',
    landingType: 'url',
    landingValue: '',
    priority: 100,
    startAt: String(now) as string | number,
    endAt: String(now + 7 * 24 * 3600) as string | number,
    cpdPrice: 0,
    hasGeo: false,
    latitude: 0,
    longitude: 0,
    radiusM: 3000,
    dailyImpressionLimit: 0,
    submitAudit: false,
  }
}

async function openForm(data: ReturnType<typeof defaultForm>) {
  Object.assign(form, data)
  formVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

function onFormClosed() {
  Object.assign(form, defaultForm())
}

function openCreate() {
  void openForm(defaultForm())
}

async function openEdit(row: AdCampaignItem) {
  try {
    const { data } = await getAdCampaignDetail(row.id)
    await openForm({
      id: data.id,
      advertiserName: data.advertiserName || '',
      title: data.title || '',
      subtitle: data.subtitle || '',
      imageMediaId: data.imageMediaId || 0,
      imageUrl: data.imageUrl || '',
      landingType: data.landingType || 'url',
      landingValue: data.landingValue || '',
      priority: data.priority || 100,
      startAt: String(data.startAt || ''),
      endAt: String(data.endAt || ''),
      cpdPrice: data.cpdPrice || 0,
      hasGeo: !!data.hasGeo,
      latitude: data.latitude || 0,
      longitude: data.longitude || 0,
      radiusM: data.radiusM || 3000,
      dailyImpressionLimit: data.dailyImpressionLimit || 0,
      submitAudit: false,
    })
  } catch (e) {
    showApiError(e, '加载详情失败')
  }
}

async function handleImageUpload(options: UploadRequestOptions) {
  uploading.value = true
  try {
    const { data } = await uploadAdImage(options.file)
    form.imageMediaId = data.id
    form.imageUrl = data.url
    formRef.value?.validateField('imageMediaId')
    ElMessage.success('上传成功')
    return data
  } catch (e) {
    showApiError(e, '上传失败')
    throw e
  } finally {
    uploading.value = false
  }
}

function toUnix(v: string | number) {
  const n = typeof v === 'number' ? v : Number(v)
  return Number.isFinite(n) ? n : 0
}

function onGeoMapChange(payload: {
  shapeType: 1 | 2
  centerLat?: number
  centerLng?: number
  radius?: number
}) {
  if (payload.shapeType !== 1) return
  if (typeof payload.centerLat === 'number') form.latitude = payload.centerLat
  if (typeof payload.centerLng === 'number') form.longitude = payload.centerLng
  if (typeof payload.radius === 'number') form.radiusM = payload.radius
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  if (form.hasGeo) {
    if (!hasCoordinates(form.latitude, form.longitude)) {
      ElMessage.warning('请在地图上点击选择投放圆心')
      return
    }
    if (!form.radiusM || form.radiusM < 100) {
      ElMessage.warning('请设置有效投放半径')
      return
    }
  }
  saving.value = true
  try {
    const payload = {
      advertiserName: form.advertiserName.trim(),
      title: form.title.trim(),
      subtitle: form.subtitle.trim(),
      imageMediaId: form.imageMediaId,
      landingType: form.landingType,
      landingValue: form.landingValue.trim(),
      priority: form.priority || 100,
      startAt: toUnix(form.startAt),
      endAt: toUnix(form.endAt),
      cpdPrice: form.cpdPrice,
      hasGeo: form.hasGeo,
      latitude: form.hasGeo ? form.latitude : 0,
      longitude: form.hasGeo ? form.longitude : 0,
      radiusM: form.hasGeo ? form.radiusM : 0,
      dailyImpressionLimit: form.dailyImpressionLimit,
      submitAudit: form.submitAudit,
    }
    if (form.id) await updateAdCampaign({ id: form.id, ...payload })
    else await createAdCampaign(payload)
    ElMessage.success('保存成功')
    formVisible.value = false
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
.dialog-body {
  display: grid;
  grid-template-columns: 1fr 220px;
  gap: 16px;
  align-items: start;
}
.dialog-body--map {
  grid-template-columns: minmax(0, 1fr) 200px;
}
.geo-map-item :deep(.el-form-item__content) {
  display: block;
  width: 100%;
}
.upload-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.upload-hint,
.field-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-left: 8px;
}
@media (max-width: 768px) {
  .dialog-body {
    grid-template-columns: 1fr;
  }
}
</style>
