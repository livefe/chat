<script setup lang="ts">
import { imInitials } from './media'

withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    preview?: string
    time?: string
    avatarUrl?: string
    unread?: number
    active?: boolean
    online?: boolean
    waitText?: string
    waitTone?: 'ok' | 'warn' | 'danger'
    disabled?: boolean
  }>(),
  {
    unread: 0,
    active: false,
    online: false,
    disabled: false,
  },
)

defineEmits<{
  click: []
}>()
</script>

<template>
  <button
    type="button"
    class="im-list-item"
    :class="{ 'is-active': active }"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <div class="im-list-item__avatar">
      <el-avatar :size="40" :src="avatarUrl || undefined">{{ imInitials(title) }}</el-avatar>
      <span v-if="online" class="im-list-item__online" title="在线" />
    </div>
    <div class="im-list-item__body">
      <div class="im-list-item__top">
        <span class="im-list-item__title">{{ title }}</span>
        <span v-if="time" class="im-list-item__time">{{ time }}</span>
        <span v-else-if="waitText" class="im-list-item__wait" :class="`is-${waitTone || 'ok'}`">
          {{ waitText }}
        </span>
      </div>
      <div class="im-list-item__bottom">
        <span class="im-list-item__preview">{{ preview || subtitle || ' ' }}</span>
        <el-badge v-if="unread > 0" :value="unread" :max="99" />
      </div>
    </div>
    <div v-if="$slots.action" class="im-list-item__action" @click.stop>
      <slot name="action" />
    </div>
  </button>
</template>
