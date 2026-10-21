<script setup lang="ts">
defineProps<{
  open: boolean
  menuStyle: Record<string, string>
  self: boolean
  canRecall: boolean
}>()

const emit = defineEmits<{
  close: []
  copy: []
  forward: []
  quote: []
  recall: []
  report: []
}>()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="msg-menu-mask"
      aria-hidden="true"
      @click="emit('close')"
      @touchstart.passive="emit('close')"
    />
    <div v-if="open" class="msg-menu" role="menu" :style="menuStyle" @click.stop @touchstart.stop>
      <button type="button" class="msg-menu__item" role="menuitem" @click="emit('copy')">
        复制
      </button>
      <button type="button" class="msg-menu__item" role="menuitem" @click="emit('forward')">
        转发
      </button>
      <button type="button" class="msg-menu__item" role="menuitem" @click="emit('quote')">
        引用
      </button>
      <button
        v-if="canRecall"
        type="button"
        class="msg-menu__item"
        role="menuitem"
        @click="emit('recall')"
      >
        撤回
      </button>
      <button
        v-if="!self"
        type="button"
        class="msg-menu__item"
        role="menuitem"
        @click="emit('report')"
      >
        举报
      </button>
    </div>
  </Teleport>
</template>

<style scoped>
.msg-menu {
  position: fixed;
  z-index: 4000;
  display: flex;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.7);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.16);
  overflow: hidden;
}

.msg-menu__item {
  padding: 9px 14px;
  border: none;
  background: transparent;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.2;
  white-space: nowrap;
  cursor: pointer;
}

.msg-menu-mask {
  position: fixed;
  inset: 0;
  z-index: 3990;
}
</style>
