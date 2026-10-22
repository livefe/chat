<script setup lang="ts">
defineProps<{ title?: string }>()
</script>

<template>
  <div class="im-chat-hd">
    <div class="im-chat-hd__info">
      <div class="im-chat-hd__title">
        <slot name="title">{{ title }}</slot>
        <slot name="tags" />
      </div>
      <div v-if="$slots.sub" class="im-chat-hd__sub">
        <slot name="sub" />
      </div>
    </div>
    <div v-if="$slots.actions" class="im-chat-hd__actions">
      <slot name="actions" />
    </div>
  </div>
</template>
