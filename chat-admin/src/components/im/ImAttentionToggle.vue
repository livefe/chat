<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Bell, MuteNotification } from '@element-plus/icons-vue'
import {
  attentionPrefs,
  desktopNotifySupported,
  setAttentionDesktop,
  setAttentionSound,
} from '@/utils/attention'

const anyOn = computed(() => attentionPrefs.sound || attentionPrefs.desktop)

async function onDesktopChange(on: string | number | boolean) {
  const enabled = await setAttentionDesktop(!!on)
  if (on && !enabled) ElMessage.warning('浏览器未授权通知，请在地址栏左侧的站点设置中允许通知')
}
</script>

<template>
  <el-popover trigger="click" placement="bottom-end" :width="240">
    <template #reference>
      <el-button :icon="anyOn ? Bell : MuteNotification" title="提醒设置">提醒</el-button>
    </template>
    <div class="im-attention">
      <div class="im-attention__row">
        <span>新消息提示音</span>
        <el-switch
          :model-value="attentionPrefs.sound"
          @update:model-value="(v) => setAttentionSound(!!v)"
        />
      </div>
      <div class="im-attention__row">
        <span>桌面通知</span>
        <el-switch
          :model-value="attentionPrefs.desktop"
          :disabled="!desktopNotifySupported()"
          @update:model-value="onDesktopChange"
        />
      </div>
      <div class="im-attention__hint">页面在后台时，标签标题会显示待处理数量</div>
    </div>
  </el-popover>
</template>

<style scoped>
.im-attention {
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;
}

.im-attention__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.im-attention__hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
