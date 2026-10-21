<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<template>
  <div class="im-chat-hd">
    <div>
      <div class="im-chat-hd__title">
        {{ title }}
        <slot name="tags" />
      </div>
      <div class="im-chat-hd__sub">
        <slot name="sub" />
      </div>
    </div>
    <div v-if="$slots.actions" class="im-chat-hd__actions">
      <slot name="actions" />
    </div>
  </div>
</template>
