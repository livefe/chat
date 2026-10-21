<script setup lang="ts">
import { ref } from 'vue'

import { useTextEditMenu } from '@/composables/useTextEditMenu'

const menuRef = ref<HTMLElement | null>(null)
const { open, menuStyle, actions, runAction } = useTextEditMenu(menuRef)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      ref="menuRef"
      class="text-edit-menu"
      role="menu"
      :style="menuStyle"
      @click.stop
      @pointerdown.stop
    >
      <button
        v-for="action in actions"
        :key="action.name"
        type="button"
        class="text-edit-menu__item"
        role="menuitem"
        @click="runAction(action)"
      >
        {{ action.name }}
      </button>
    </div>
  </Teleport>
</template>

<style scoped>
.text-edit-menu {
  position: fixed;
  z-index: 4000;
  display: flex;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.7);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.16);
  overflow: hidden;
}

.text-edit-menu__item {
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
</style>
