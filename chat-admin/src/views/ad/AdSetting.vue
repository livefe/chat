<template>
  <div class="admin-page">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>投放设置</span>
        </div>
      </template>

      <el-alert
        v-if="!canSave"
        type="info"
        :closable="false"
        show-icon
        title="当前为只读模式，如需修改请联系管理员分配「保存投放设置」权限"
        class="ad-setting__alert"
      />

      <el-form :model="form" label-width="140px" class="ad-setting__form" @submit.prevent>
        <el-form-item label="启用投放">
          <el-switch v-model="form.enabled" :disabled="!canSave" />
        </el-form-item>
        <el-form-item label="首条位置">
          <div>
            <el-input-number
              v-model="form.firstSlotAfter"
              :min="1"
              :max="100"
              :step="1"
              :disabled="!canSave"
            />
            <div class="ad-setting__hint">第几条自然内容之后出现第一条广告</div>
          </div>
        </el-form-item>
        <el-form-item label="插入间隔">
          <div>
            <el-input-number
              v-model="form.slotInterval"
              :min="1"
              :max="100"
              :step="1"
              :disabled="!canSave"
            />
            <div class="ad-setting__hint">之后每隔多少条自然内容再插一条</div>
          </div>
        </el-form-item>
        <el-form-item label="商家同时投放">
          <div>
            <el-input-number
              v-model="form.maxActivePerAdvertiser"
              :min="1"
              :max="50"
              :step="1"
              :disabled="!canSave"
            />
            <div class="ad-setting__hint">同一商家处于投放中的计划上限</div>
          </div>
        </el-form-item>
        <el-form-item label="低点击率阈值">
          <div>
            <el-input-number
              v-model="form.lowCtrThreshold"
              :min="0.0001"
              :max="1"
              :step="0.001"
              :precision="4"
              :disabled="!canSave"
            />
            <div class="ad-setting__hint">
              今日曝光不少于 100 且点击率低于该值时标红。0.01 表示 1%
            </div>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button v-if="canSave" type="primary" :loading="saving" @click="handleSave">
            保存
          </el-button>
          <div class="ad-setting__hint">曝光和点击的签名密钥仍在服务端配置文件中，不在这里修改</div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getAdSetting, updateAdSetting } from '@/api/ad'
import { showApiError } from '@/utils/apiError'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const { can } = usePermission()
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  enabled: true,
  firstSlotAfter: 3,
  slotInterval: 8,
  maxActivePerAdvertiser: 3,
  lowCtrThreshold: 0.01,
})

const canSave = computed(() => can(PERM.AdSettingUpdate))

async function loadSetting() {
  loading.value = true
  try {
    const { data } = await getAdSetting()
    form.enabled = !!data.enabled
    form.firstSlotAfter = data.firstSlotAfter
    form.slotInterval = data.slotInterval
    form.maxActivePerAdvertiser = data.maxActivePerAdvertiser
    form.lowCtrThreshold = data.lowCtrThreshold
  } catch (e) {
    showApiError(e, '加载投放设置失败')
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateAdSetting({
      enabled: form.enabled,
      firstSlotAfter: form.firstSlotAfter,
      slotInterval: form.slotInterval,
      maxActivePerAdvertiser: form.maxActivePerAdvertiser,
      lowCtrThreshold: form.lowCtrThreshold,
    })
    ElMessage.success('保存成功')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadSetting)
</script>

<style scoped>
.ad-setting__alert {
  margin-bottom: 16px;
}

.ad-setting__form {
  max-width: 560px;
}

.ad-setting__hint {
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
}
</style>
