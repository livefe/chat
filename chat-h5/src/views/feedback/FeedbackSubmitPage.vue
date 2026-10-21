<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { submitFeedback, feedbackTypeOptions } from '@/api/feedback'
import { useVanImageUpload, type VanUploaderMediaItem } from '@/composables/useVanImageUpload'
import { collectMediaIds } from '@/utils/media'

const router = useRouter()

const type = ref(1)
const content = ref('')
const contact = ref('')
const imageFiles = ref<VanUploaderMediaItem[]>([])
const submitting = ref(false)

const maxImages = 3
const maxContent = 500
const maxSizeMB = 10

const { uploading, onAfterRead } = useVanImageUpload()

const submitLabel = computed(() => {
  if (submitting.value) return '提交中…'
  if (uploading.value) return '图片上传中…'
  return '提交反馈'
})

function collectImageIds(): number[] {
  return collectMediaIds(imageFiles.value)
}

async function submit() {
  const text = content.value.trim()
  if (text.length < 10) {
    showToast('反馈内容至少 10 个字符')
    return
  }
  if (uploading.value) {
    showToast('图片上传中，请稍候')
    return
  }

  submitting.value = true
  try {
    await submitFeedback({
      type: type.value,
      content: text,
      contact: contact.value.trim(),
      images: collectImageIds(),
    })
    showToast({ message: '提交成功', duration: 2000 })
    setTimeout(() => router.back(), 500)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <SubPageLayout title="意见反馈">
    <template #right>
      <span class="link-records" @click="router.push('/me/feedback/records')">我的反馈</span>
    </template>

    <div class="form-body">
      <section class="card">
        <div class="card-label">反馈类型</div>
        <div class="type-chips">
          <button
            v-for="opt in feedbackTypeOptions"
            :key="opt.value"
            type="button"
            class="filter-chip"
            :class="{ 'filter-chip--active': type === opt.value }"
            @click="type = opt.value"
          >
            {{ opt.text }}
          </button>
        </div>

        <div class="card-label">反馈内容</div>
        <textarea
          v-model="content"
          class="content-input"
          :maxlength="maxContent"
          rows="5"
          placeholder="请描述您的问题或建议（10-500字）"
        />
        <div class="word-count">{{ content.length }}/{{ maxContent }}</div>

        <div class="card-label">联系方式</div>
        <input
          v-model="contact"
          class="contact-input"
          type="text"
          maxlength="64"
          placeholder="选填，便于我们联系您"
        />
      </section>

      <section class="card">
        <div class="card-label muted">截图（选填，最多 {{ maxImages }} 张）</div>
        <van-uploader
          v-model="imageFiles"
          :max-count="maxImages"
          :max-size="maxSizeMB * 1024 * 1024"
          accept="image/*"
          :after-read="onAfterRead"
          @oversize="() => showToast(`单张图片不能超过 ${maxSizeMB}MB`)"
        />
      </section>

      <van-button
        type="primary"
        block
        class="page-btn"
        :loading="submitting"
        :disabled="uploading || submitting"
        @click="submit"
      >
        {{ submitLabel }}
      </van-button>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.link-records {
  font-size: 13px;
  color: var(--text-primary);
}

.form-body {
  padding: 12px 16px 24px;
  box-sizing: border-box;
}

.card {
  margin-bottom: 12px;
  padding: 16px;
  background: var(--card-bg);
  border-radius: var(--radius-md);
}

.card-label {
  margin-bottom: 10px;
  font-size: 14px;
  color: var(--text-secondary);
}

.card-label.muted {
  color: var(--text-secondary);
}

.card-label + .content-input,
.card-label + .contact-input {
  margin-top: 0;
}

.type-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.content-input,
.contact-input {
  display: block;
  width: 100%;
  box-sizing: border-box;
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--card-bg);
  font-size: 15px;
  color: var(--text-primary);
  outline: none;
  resize: none;
}

.content-input {
  min-height: 140px;
  line-height: 1.5;
}

.content-input:focus,
.contact-input:focus {
  border-color: var(--brand-color);
}

.content-input::placeholder,
.contact-input::placeholder {
  color: var(--text-tertiary);
}

.word-count {
  margin-top: 4px;
  margin-bottom: 16px;
  text-align: right;
  font-size: 12px;
  color: var(--text-tertiary);
}

.page-btn {
  margin: 12px 0 0;
  width: 100%;
}
</style>
