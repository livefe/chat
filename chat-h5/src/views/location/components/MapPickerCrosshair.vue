<script setup lang="ts">
defineProps<{
  dragging: boolean
  bounce: boolean
}>()
</script>

<template>
  <div class="map-overlay">
    <div
      class="map-crosshair"
      :class="{
        'map-crosshair--dragging': dragging,
        'map-crosshair--bounce': bounce,
      }"
      aria-hidden="true"
    >
      <span class="map-crosshair__v" />
      <span class="map-crosshair__h" />
      <span class="map-crosshair__dot" />
    </div>
  </div>
</template>

<style scoped>
.map-overlay {
  position: absolute;
  inset: 0;
  z-index: 1000;
  pointer-events: none;
}

.map-crosshair {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 28px;
  height: 28px;
  transform: translate(-50%, -50%);
  will-change: transform;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
}

.map-crosshair--dragging {
  transform: translate(-50%, calc(-50% - 4px));
  transition: transform 0.18s ease;
}

.map-crosshair--bounce {
  animation: crosshair-drop 0.36s cubic-bezier(0.34, 1.4, 0.64, 1);
}

@keyframes crosshair-drop {
  0% {
    transform: translate(-50%, calc(-50% - 4px));
  }
  70% {
    transform: translate(-50%, calc(-50% + 1px));
  }
  100% {
    transform: translate(-50%, -50%);
  }
}

.map-crosshair__v,
.map-crosshair__h {
  position: absolute;
  left: 50%;
  top: 50%;
  background: var(--brand-color);
  border-radius: 1px;
}

.map-crosshair__v {
  width: 2px;
  height: 28px;
  transform: translate(-50%, -50%);
}

.map-crosshair__h {
  width: 28px;
  height: 2px;
  transform: translate(-50%, -50%);
}

.map-crosshair__dot {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--brand-color);
  transform: translate(-50%, -50%);
}
</style>
