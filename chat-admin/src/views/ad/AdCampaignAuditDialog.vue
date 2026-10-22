<template>
  <el-dialog
    :model-value="visible"
    title="审核广告"
    :width="detail?.hasGeo ? '960px' : '720px'"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <div v-loading="loading" class="audit-body">
      <template v-if="detail">
        <el-alert
          v-if="detail.rejectRemark"
          type="warning"
          :closable="false"
          show-icon
          :title="`上次驳回原因：${detail.rejectRemark}`"
          description="请确认本次提交已按驳回意见修改"
        />
        <el-descriptions :column="2" border>
          <el-descriptions-item label="广告主">{{
            detail.advertiserName || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag size="small">{{ detail.statusText || '-' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="标题" :span="2">{{
            detail.title || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="副标题" :span="2">{{
            detail.subtitle || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="内部名称" :span="2">{{
            detail.name || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="落地类型">{{
            landingLabel(detail.landingType)
          }}</el-descriptions-item>
          <el-descriptions-item label="落地值">
            <a
              v-if="detail.landingType === 'url' && detail.landingValue"
              :href="detail.landingValue"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ detail.landingValue }}
            </a>
            <span v-else>{{ detail.landingValue || '-' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="排期" :span="2">
            {{ formatUnixSeconds(detail.startAt) }} ~ {{ formatUnixSeconds(detail.endAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="优先级">{{ detail.priority }}</el-descriptions-item>
          <el-descriptions-item label="投放范围">
            <template v-if="detail.hasGeo">
              {{ detail.latitude.toFixed(5) }}, {{ detail.longitude.toFixed(5) }} · 半径
              {{ detail.radiusM }} 米
            </template>
            <template v-else>不限</template>
          </el-descriptions-item>
        </el-descriptions>

        <div class="cover-block">
          <div class="cover-label">封面</div>
          <el-image
            v-if="detail.imageUrl"
            :src="detail.imageUrl"
            :preview-src-list="[detail.imageUrl]"
            fit="cover"
            class="cover"
            preview-teleported
          />
          <span v-else class="muted">未上传封面</span>
        </div>

        <AmapPicker
          v-if="showMap"
          :key="`ad-audit-${detail.id}`"
          :shape-type="1"
          circle-only
          readonly
          readonly-hint="蓝色圆形为投放范围"
          :center-lat="detail.latitude"
          :center-lng="detail.longitude"
          :radius="detail.radiusM || 3000"
          :min-radius="100"
          :max-radius="50000"
        />
      </template>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:visible', false)">取消</el-button>
        <template v-if="isPending">
          <el-button
            type="danger"
            :disabled="!canDecide"
            :loading="submitting"
            @click="confirmReject"
          >
            驳回
          </el-button>
          <el-button
            type="success"
            :disabled="!canDecide"
            :loading="submitting"
            @click="confirmApprove"
          >
            通过并上架
          </el-button>
        </template>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import AmapPicker from '@/components/amap/AmapPicker.vue'
import { getAdCampaignDetail, type AdCampaignItem } from '@/api/ad'
import { AD_LANDING_OPTIONS, AD_STATUS } from '@/constants/ad'
import { showApiError } from '@/utils/apiError'
import { confirmWithReason } from '@/utils/promptReason'
import { formatUnixSeconds } from '@/utils/datetime'
import { hasCoordinates } from '@/utils/geo'

const props = defineProps<{
  visible: boolean
  campaignId: number
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  approve: []
  reject: [remark: string]
}>()

const loading = ref(false)
const detail = ref<AdCampaignItem | null>(null)

const canDecide = computed(() => !loading.value && !!detail.value)
const isPending = computed(() => detail.value?.status === AD_STATUS.Pending)
const showMap = computed(
  () => !!detail.value?.hasGeo && hasCoordinates(detail.value.latitude, detail.value.longitude),
)

function landingLabel(type: string) {
  return AD_LANDING_OPTIONS.find((opt) => opt.value === type)?.label || type || '-'
}

async function confirmApprove() {
  if (props.submitting || !canDecide.value || !detail.value) return
  try {
    await ElMessageBox.confirm(
      `确认通过「${detail.value.title}」？通过后将立即按排期投放。`,
      '审核通过',
      { type: 'warning' },
    )
  } catch {
    return
  }
  emit('approve')
}

async function confirmReject() {
  if (props.submitting || !canDecide.value || !detail.value) return
  const remark = await confirmWithReason(`驳回「${detail.value.title}」，原因将展示给编辑人员`, {
    title: '驳回计划',
    required: true,
    inputPlaceholder: '请填写驳回原因（200 字以内）',
  })
  if (remark === null) return
  emit('reject', remark.slice(0, 200))
}

async function load() {
  if (!props.campaignId) return
  loading.value = true
  detail.value = null
  try {
    const { data } = await getAdCampaignDetail(props.campaignId)
    detail.value = data
  } catch (e) {
    showApiError(e, '加载审核信息失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.visible, props.campaignId] as const,
  ([visible]) => {
    if (visible && props.campaignId > 0) {
      load()
    } else if (!visible) {
      detail.value = null
    }
  },
)
</script>

<style scoped>
.audit-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 160px;
}
.cover-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.cover-label {
  font-size: 13px;
  font-weight: 600;
}
.cover {
  width: 240px;
  height: 135px;
  border-radius: 6px;
}
.muted {
  color: var(--el-text-color-placeholder);
}
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
