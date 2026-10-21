<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import {
  reportReasonOptions,
  reportTargetTypeText,
  submitReport,
  ReportTargetUser,
} from '@/api/report'
import { useVanImageUpload, type VanUploaderMediaItem } from '@/composables/useVanImageUpload'
import { collectMediaIds } from '@/utils/media'

const route = useRoute()
const router = useRouter()

const reason = ref(1)
const content = ref('')
const imageFiles = ref<VanUploaderMediaItem[]>([])
const submitting = ref(false)

const maxImages = 3
const maxSizeMB = 10
const maxContent = 500

const targetType = computed(() => Number(route.query.targetType || 0))
const targetId = computed(() => Number(route.query.targetId || 0))
const previewTitle = computed(() => String(route.query.title || '').trim())
const previewSubtitle = computed(() => String(route.query.subtitle || '').trim())
const previewAvatar = computed(() => String(route.query.avatar || '').trim())

const targetLabel = computed(() => reportTargetTypeText[targetType.value] || '内容')

const displayTitle = computed(() => {
  if (previewTitle.value) return previewTitle.value
  return targetLabel.value
})

const { uploading, onAfterRead } = useVanImageUpload()

function collectImageIds(): number[] {
  return collectMediaIds(imageFiles.value)
}

async function submit() {
  if (!targetType.value || !targetId.value) {
    showToast('举报对象无效')
    return
  }
  if (uploading.value) {
    showToast('图片上传中，请稍候')
    return
  }

  submitting.value = true
  try {
    await submitReport({
      targetType: targetType.value,
      targetId: targetId.value,
      reason: reason.value,
      content: content.value.trim(),
      images: collectImageIds(),
    })
    showToast({ message: '举报已提交', duration: 2000 })
    setTimeout(() => router.back(), 500)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <SubPageLayout title="举报">
    <div class="form-body">
      <section class="card">
        <div class="card-label muted">举报{{ targetLabel }}</div>
        <div class="target-preview">
          <UserAvatar
            v-if="targetType === ReportTargetUser || previewAvatar"
            :src="previewAvatar"
            :size="40"
            tone="profile"
          />
          <div class="target-meta">
            <div class="target-title">{{ displayTitle }}</div>
            <div v-if="previewSubtitle" class="target-subtitle">{{ previewSubtitle }}</div>
          </div>
        </div>
      </section>

      <section class="card">
        <div class="card-label">举报原因</div>
        <div class="reason-chips">
          <button
            v-for="opt in reportReasonOptions"
            :key="opt.value"
            type="button"
            class="filter-chip"
            :class="{ 'filter-chip--active': reason === opt.value }"
            @click="reason = opt.value"
          >
            {{ opt.text }}
          </button>
        </div>
      </section>

      <section class="card">
        <div class="card-label muted">补充说明</div>
        <textarea
          v-model="content"
          class="content-input"
          :maxlength="maxContent"
          rows="5"
          placeholder="请描述违规情况（选填）"
        />
        <div class="word-count">{{ content.length }}/{{ maxContent }}</div>

        <div class="upload-label muted">截图证据（选填，最多 {{ maxImages }} 张）</div>
        <van-uploader
          v-model="imageFiles"
          :max-count="maxImages"
          :max-size="maxSizeMB * 1024 * 1024"
          accept="image/*"
          :after-read="onAfterRead"
          @oversize="() => showToast(`单张图片不能超过 ${maxSizeMB}MB`)"
        />
      </section>

      <p class="hint">我们会在 72 小时内处理，感谢你的反馈。</p>

      <van-button
        type="primary"
        block
        class="page-btn"
        :loading="submitting"
        :disabled="uploading"
        @click="submit"
      >
        {{ uploading ? '图片上传中…' : '提交举报' }}
      </van-button>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.form-body {
  padding: 12px 16px 24px;
  box-sizing: border-box;
}

.card {
  margin-bottom: 12px;
  padding: 14px;
  background: var(--card-bg);
  border-radius: var(--radius-md);
}

.card-label {
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.card-label.muted,
.upload-label.muted {
  font-weight: 400;
  color: var(--text-secondary);
}

.card-label.muted {
  font-size: 12px;
  color: var(--text-tertiary);
}

.target-preview {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.target-meta {
  flex: 1;
  min-width: 0;
}

.target-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.target-subtitle {
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.reason-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.content-input {
  display: block;
  width: 100%;
  min-height: 120px;
  box-sizing: border-box;
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: transparent;
  font-size: 14px;
  line-height: 1.5;
  color: var(--text-primary);
  resize: vertical;
  outline: none;
  font-family: inherit;
}

.content-input::placeholder {
  color: var(--text-tertiary);
}

.content-input:focus {
  border-color: var(--brand-color);
}

.word-count {
  margin-top: 4px;
  text-align: right;
  font-size: 12px;
  color: var(--text-tertiary);
}

.upload-label {
  margin: 16px 0 10px;
  font-size: 14px;
}

.hint {
  margin: 0 0 4px;
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

.page-btn {
  margin-top: 20px;
}
</style>
