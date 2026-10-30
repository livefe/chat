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
    <el-form :model="form" label-width="96px" class="support-config__form" @submit.prevent>
      <el-form-item label="品牌名称" required>
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
            <el-button
              v-if="form.brandAvatarMediaId && canSave"
              link
              type="danger"
              :disabled="uploading"
              @click="clearAvatar"
            >
              清除
            </el-button>
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
      <el-form-item label="快捷回复">
        <div class="support-config__replies">
          <div class="support-config__hint">
            坐席在回复框点击「快捷回复」或输入 / 即可插入，最多 {{ SUPPORT_QUICK_REPLY_MAX }} 条
          </div>
          <div v-for="(q, i) in form.quickReplies" :key="q.key" class="support-config__reply">
            <div class="support-config__reply-main">
              <el-input
                v-model="q.title"
                placeholder="标题，如：问候语"
                maxlength="20"
                show-word-limit
                :disabled="!canSave"
              />
              <el-input
                v-model="q.content"
                type="textarea"
                :rows="2"
                placeholder="回复内容"
                maxlength="500"
                show-word-limit
                :disabled="!canSave"
              />
            </div>
            <div v-if="canSave" class="support-config__reply-ops">
              <el-button link :disabled="i === 0" title="上移" @click="move(i, -1)">
                <el-icon><ArrowUp /></el-icon>
              </el-button>
              <el-button
                link
                :disabled="i === form.quickReplies.length - 1"
                title="下移"
                @click="move(i, 1)"
              >
                <el-icon><ArrowDown /></el-icon>
              </el-button>
              <el-button link type="danger" title="删除" @click="form.quickReplies.splice(i, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
          <el-button
            v-if="canSave"
            :disabled="form.quickReplies.length >= SUPPORT_QUICK_REPLY_MAX"
            @click="addReply"
          >
            添加快捷回复
          </el-button>
        </div>
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
import { ElMessage, ElMessageBox, type UploadRequestHandler } from 'element-plus'
import { ArrowDown, ArrowUp, Delete } from '@element-plus/icons-vue'
import { showApiError } from '@/utils/apiError'
import { getSupportConfig, updateSupportConfig, type SupportConfig } from '@/api/support'
import { uploadImImage } from '@/api/im'
import { PERM } from '@/constants/permission'
import { SUPPORT_QUICK_REPLY_MAX } from '@/constants/support'
import { usePermission } from '@/composables/usePermission'

interface QuickReplyRow {
  key: number
  title: string
  content: string
}

const { can } = usePermission()

const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)
const savedEnabled = ref(true)
let rowKey = 0

const form = reactive({
  brandName: '',
  brandAvatarMediaId: 0,
  brandAvatarUrl: '',
  welcomeText: '',
  offlineText: '',
  enabled: true,
  quickReplies: [] as QuickReplyRow[],
})

const canSave = computed(() => can(PERM.ImSupportOverviewConfig))
const canUploadAvatar = computed(() => canSave.value && can(PERM.OpsUpload))

function fillForm(data: SupportConfig) {
  form.brandName = data.brandName || ''
  form.brandAvatarMediaId = data.brandAvatarMediaId || 0
  form.brandAvatarUrl = data.brandAvatarUrl || ''
  form.welcomeText = data.welcomeText || ''
  form.offlineText = data.offlineText || ''
  form.enabled = !!data.enabled
  form.quickReplies = (data.quickReplies || []).map((q) => ({ key: ++rowKey, ...q }))
  savedEnabled.value = form.enabled
}

async function loadConfig() {
  loading.value = true
  try {
    const { data } = await getSupportConfig()
    fillForm(data)
  } catch (e) {
    showApiError(e, '加载客服配置失败')
  } finally {
    loading.value = false
  }
}

function addReply() {
  form.quickReplies.push({ key: ++rowKey, title: '', content: '' })
}

function move(index: number, step: number) {
  const list = form.quickReplies
  const target = index + step
  const [item] = list.splice(index, 1)
  if (item) list.splice(target, 0, item)
}

const onUploadAvatar: UploadRequestHandler = async (options) => {
  uploading.value = true
  try {
    const { data } = await uploadImImage(options.file as File)
    form.brandAvatarMediaId = data.id
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

function clearAvatar() {
  form.brandAvatarMediaId = 0
  form.brandAvatarUrl = ''
}

async function handleSave() {
  if (!form.brandName.trim()) {
    ElMessage.warning('请填写品牌名称')
    return
  }
  const replies = form.quickReplies.map((q) => ({
    title: q.title.trim(),
    content: q.content.trim(),
  }))
  const emptyIndex = replies.findIndex((q) => !q.title || !q.content)
  if (emptyIndex >= 0) {
    ElMessage.warning(`第 ${emptyIndex + 1} 条快捷回复的标题和内容不能为空`)
    return
  }
  if (savedEnabled.value && !form.enabled) {
    try {
      await ElMessageBox.confirm(
        '关闭后用户将无法发起新的客服咨询，App 内客服入口会提示暂未开放；排队中和进行中的会话不受影响。确定关闭客服？',
        '关闭客服',
        { type: 'warning', confirmButtonText: '确定关闭' },
      )
    } catch {
      return
    }
  }
  saving.value = true
  try {
    const { data } = await updateSupportConfig({
      brandName: form.brandName.trim(),
      brandAvatarMediaId: form.brandAvatarMediaId,
      welcomeText: form.welcomeText.trim(),
      offlineText: form.offlineText.trim(),
      enabled: form.enabled ? 1 : 0,
      quickReplies: replies,
    })
    fillForm(data)
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

.support-config__replies {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.support-config__hint {
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.support-config__reply {
  display: flex;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.support-config__reply-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.support-config__reply-ops {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.support-config__reply-ops .el-button + .el-button {
  margin-left: 0;
}
</style>
