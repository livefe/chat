<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import PostCategoryGrid from '@/views/post/components/PostCategoryGrid.vue'
import PostPublishCategoryLine from '@/views/post/components/PostPublishCategoryLine.vue'
import PostPublishEventFields from '@/views/post/components/PostPublishEventFields.vue'
import PostPublishHelpFields from '@/views/post/components/PostPublishHelpFields.vue'
import PostPublishIdleFields from '@/views/post/components/PostPublishIdleFields.vue'
import PostPublishImageField from '@/views/post/components/PostPublishImageField.vue'
import PostPublishNearbyRooms from '@/views/post/components/PostPublishNearbyRooms.vue'
import PostPublishSubmitBar from '@/views/post/components/PostPublishSubmitBar.vue'
import { usePostPublishForm } from '@/composables/usePostPublishForm'
import { isLocationPickNavigation, locationPickRoute } from '@/utils/map/locationPick'
import {
  POST_CATEGORY_OPTIONS,
  PostCategory,
  type PostCategoryValue,
} from '@/utils/post/postCategory'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const route = useRoute()
const {
  form,
  imageList,
  uploading,
  onAfterRead,
  loading,
  pageLoading,
  isEdit,
  roomId,
  isRoomPost,
  category,
  categoryMeta,
  hasPendingImages,
  canSubmit,
  postLocation,
  postLocationLabel,
  nearbyLoading,
  displayNearbyRooms,
  showNearbyRooms,
  clearAttachRoom,
  selectAttachRoom,
  attachRoomById,
  loadForEdit,
  submit,
  reset,
} = usePostPublishForm()

const phase = ref<'pick' | 'form'>('pick')
const categorySheetVisible = ref(false)

const pageTitle = computed(() => {
  if (isEdit.value) return '编辑'
  if (phase.value === 'pick') return '选择类型'
  return '发布'
})
const submitLabel = computed(() => {
  if (loading.value) return isEdit.value ? '保存中…' : '提交中…'
  return isEdit.value ? '保存并提交审核' : '发布'
})
const submitHint = computed(() => {
  if (!postLocation.ready) {
    return postLocation.error ? '定位失败，请先选择位置' : '定位中，请稍候…'
  }
  if (uploading.value || hasPendingImages.value) return '图片上传中，请稍候…'
  return ''
})

function resolveInitialCategory() {
  const qCategory = String(route.query.category || '')
  const isKnown = (v: string) => POST_CATEGORY_OPTIONS.some((x) => x.value === v)
  const resolved = isKnown(qCategory) ? qCategory : isKnown(form.category) ? form.category : ''
  if (!resolved) {
    phase.value = 'pick'
    return
  }
  form.category = resolved
  phase.value = 'form'
  if (qCategory !== resolved) syncCategoryQuery(resolved)
}

onMounted(async () => {
  if (isEdit.value) {
    reset()
    phase.value = 'form'
    await loadForEdit()
    return
  }

  resolveInitialCategory()

  const qRoom = Number(route.query.room || 0)
  if (qRoom > 0) {
    try {
      await attachRoomById(qRoom)
      if (!postLocation.ready && phase.value === 'form') {
        await postLocation.initDefault()
      }
    } catch (e: unknown) {
      showCatchError(e, '加载聊天区失败', true)
      router.back()
      return
    }
  } else if (phase.value === 'form' && !postLocation.ready) {
    await postLocation.initDefault()
  }
})

onBeforeRouteLeave((to) => {
  if (isLocationPickNavigation(to)) return
  reset()
})

function syncCategoryQuery(value: string) {
  const query = { ...route.query, category: value }
  router.replace({ query })
}

function selectCategory(value: PostCategoryValue) {
  form.category = value
  syncCategoryQuery(value)
  categorySheetVisible.value = false
  const enteringForm = phase.value === 'pick'
  phase.value = 'form'
  if (enteringForm && !postLocation.ready) {
    void postLocation.initDefault()
  }
}

function openCategorySheet() {
  if (isEdit.value) return
  categorySheetVisible.value = true
}

function onNavBack() {
  if (!isEdit.value && phase.value === 'form') {
    phase.value = 'pick'
    const query = { ...route.query }
    delete query.category
    router.replace({ query })
    return
  }
  router.back()
}

function blurActiveInput() {
  const el = document.activeElement
  if (el instanceof HTMLElement) el.blur()
}

function onReadonlyFieldFocus(e: FocusEvent) {
  const t = e.target
  if (t instanceof HTMLElement) t.blur()
}

function openLocationPicker() {
  blurActiveInput()
  clearAttachRoom()
  router.push(locationPickRoute('publish'))
}
</script>

<template>
  <SubPageLayout :title="pageTitle" full-height :on-back="onNavBack">
    <van-loading v-if="pageLoading" class="tab-page__status" />

    <div v-else-if="phase === 'pick'" class="pick-page">
      <p class="pick-page__lead">先选一种要发布的内容</p>
      <PostCategoryGrid @select="selectCategory" />
    </div>

    <div v-else class="publish-page">
      <div class="form sub-page__scroll">
        <PostPublishCategoryLine
          class="form-card"
          :label="categoryMeta?.label"
          :hint="categoryMeta?.hint"
          :locked="isEdit"
          @open="openCategorySheet"
        />

        <div class="form-card">
          <van-field
            v-model="form.title"
            label="标题"
            input-align="right"
            placeholder="请输入标题"
          />
          <van-cell v-if="isRoomPost && category === PostCategory.Notice" title="置顶">
            <template #right-icon>
              <van-switch v-model="form.pinned" size="20px" />
            </template>
          </van-cell>
          <van-field
            class="field--location"
            :class="{ 'field--location--with-nearby': showNearbyRooms }"
            label="位置"
            readonly
            is-link
            clickable
            @click="openLocationPicker"
            @focus="onReadonlyFieldFocus"
          >
            <template #input>
              <span class="publish-location-value">{{ postLocationLabel }}</span>
            </template>
          </van-field>
          <PostPublishNearbyRooms
            v-if="showNearbyRooms"
            :rooms="displayNearbyRooms"
            :selected-id="roomId"
            :loading="nearbyLoading"
            @select="selectAttachRoom"
          />
        </div>

        <div class="form-card">
          <van-field
            v-model="form.content"
            class="field--stack"
            label="内容"
            type="textarea"
            rows="4"
            autosize
            placeholder="请输入内容"
          />
        </div>

        <PostPublishEventFields
          v-if="category === PostCategory.Event"
          v-model:form="form"
          @blur-input="blurActiveInput"
        />

        <PostPublishIdleFields
          v-if="category === PostCategory.Idle"
          v-model:form="form"
          @blur-input="blurActiveInput"
        />

        <PostPublishHelpFields
          v-if="category === PostCategory.Help"
          v-model:form="form"
          @blur-input="blurActiveInput"
        />

        <PostPublishImageField v-model="imageList" :after-read="onAfterRead" />
      </div>
      <PostPublishSubmitBar
        :label="submitLabel"
        :disabled="!canSubmit"
        :hint="submitHint"
        @submit="submit"
      />
    </div>

    <van-popup v-model:show="categorySheetVisible" position="bottom" round safe-area-inset-bottom>
      <div class="app-sheet">
        <div class="app-sheet__handle" />
        <div class="app-sheet__title">更换类型</div>
        <PostCategoryGrid sheet :active="form.category" @select="selectCategory" />
      </div>
    </van-popup>
  </SubPageLayout>
</template>

<style scoped>
.pick-page {
  flex: 1;
  min-height: 0;
  padding: 20px 16px 24px;
  overflow-y: auto;
  background: var(--page-bg);
}
.pick-page__lead {
  margin: 0 0 16px;
  font-size: var(--font-body);
  color: var(--text-primary);
}

.publish-page {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--page-bg);
}
.form {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 8px 16px 16px;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-y: none;
  touch-action: pan-y;
}
.form > * {
  flex-shrink: 0;
}
.form-card {
  overflow: clip;
  border-radius: 16px;
  background: var(--card-bg);
}
.form-card :deep(.van-cell) {
  background: transparent;
}
.form-card :deep(.van-cell::after) {
  left: 16px;
  right: 0;
  border-color: var(--border-color);
}
.form-card :deep(.van-cell:last-child::after) {
  display: none;
}
.form-card :deep(.field--location--with-nearby) {
  padding-bottom: 8px;
}
.form-card :deep(.field--location--with-nearby::after) {
  display: none;
}
.form-card :deep(.van-field__label) {
  width: 5.5em;
  flex: none;
  margin-right: 16px;
  color: var(--text-primary);
  white-space: nowrap;
}
.form-card :deep(.van-field__value),
.form-card :deep(.van-field__control) {
  min-width: 0;
}
.form-card :deep(.van-field__control--custom) {
  overflow: hidden;
}
.form-card :deep(.field--stack.van-cell) {
  flex-direction: column;
  align-items: stretch;
}
.form-card :deep(.field--stack .van-field__label) {
  width: auto;
  margin-right: 0;
  margin-bottom: 8px;
}
.form-card :deep(.field--stack .van-field__value) {
  width: 100%;
}
.form-card :deep(.field--stack .van-field__control) {
  min-height: 88px;
  line-height: 1.5;
  text-align: left;
}
.publish-location-value {
  display: block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
  color: var(--text-primary);
}
</style>
