<script setup lang="ts">
defineProps<{
  placeName: string
  dragHint: string
  confirmText: string
  geocoding: boolean
  confirming: boolean
}>()

const emit = defineEmits<{
  useCurrent: []
  confirm: []
}>()
</script>

<template>
  <div class="bottom-card">
    <div class="bottom-card__head">
      <span class="bottom-card__icon">
        <van-icon name="location-o" size="18" />
      </span>
      <div class="bottom-card__text">
        <p class="bottom-card__name">
          <van-loading v-if="geocoding" size="16px" />
          <span v-else>{{ placeName || dragHint }}</span>
        </p>
      </div>
    </div>
    <div class="bottom-card__actions">
      <van-button size="small" plain icon="aim" @click="emit('useCurrent')">当前位置</van-button>
      <van-button
        type="primary"
        class="bottom-card__confirm"
        :loading="confirming"
        @click="emit('confirm')"
      >
        {{ confirmText }}
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.bottom-card {
  flex-shrink: 0;
  padding: 14px 16px calc(14px + env(safe-area-inset-bottom, 0px));
  background: var(--card-bg);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.04);
}

.bottom-card__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.bottom-card__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  color: var(--brand-color);
  background: var(--brand-soft);
}

.bottom-card__text {
  flex: 1;
  min-width: 0;
}

.bottom-card__name {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  color: #1f1f1f;
}

.bottom-card__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bottom-card__confirm {
  flex: 1;
}
</style>
