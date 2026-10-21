<template>
  <div class="scan-mask" aria-hidden="true">
    <div class="scan-mask__band scan-mask__band--v" />
    <div class="scan-mask__row">
      <div class="scan-mask__band scan-mask__band--h" />
      <div class="scan-finder">
        <span class="scan-finder__corner scan-finder__corner--tl" />
        <span class="scan-finder__corner scan-finder__corner--tr" />
        <span class="scan-finder__corner scan-finder__corner--bl" />
        <span class="scan-finder__corner scan-finder__corner--br" />
        <span class="scan-finder__beam" />
      </div>
      <div class="scan-mask__band scan-mask__band--h" />
    </div>
    <div class="scan-mask__band scan-mask__band--v" />
  </div>
</template>

<style scoped>
.scan-mask {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: flex;
  flex-direction: column;
  pointer-events: none;
}

.scan-mask__band {
  flex: 1 1 auto;
  min-height: 0;
  background: rgba(0, 0, 0, 0.42);
}

.scan-mask__row {
  display: flex;
  flex: 0 0 auto;
  height: min(68vw, 280px);
}

.scan-mask__band--h {
  flex: 1 1 auto;
}

.scan-finder {
  position: relative;
  flex: 0 0 auto;
  width: min(68vw, 280px);
  height: min(68vw, 280px);
  overflow: hidden;
  background: transparent;
}

.scan-finder__corner {
  position: absolute;
  width: 18px;
  height: 18px;
  border-color: rgba(126, 176, 255, 0.72);
  border-style: solid;
}

.scan-finder__corner--tl {
  top: 0;
  left: 0;
  border-width: 2px 0 0 2px;
  border-radius: 3px 0 0 0;
}

.scan-finder__corner--tr {
  top: 0;
  right: 0;
  border-width: 2px 2px 0 0;
  border-radius: 0 3px 0 0;
}

.scan-finder__corner--bl {
  bottom: 0;
  left: 0;
  border-width: 0 0 2px 2px;
  border-radius: 0 0 0 3px;
}

.scan-finder__corner--br {
  right: 0;
  bottom: 0;
  border-width: 0 2px 2px 0;
  border-radius: 0 0 3px 0;
}

.scan-finder__beam {
  position: absolute;
  left: 0;
  width: 100%;
  height: 38%;
  top: -38%;
  border-radius: 0;

  background: linear-gradient(
    180deg,
    rgba(76, 141, 255, 0) 0%,
    rgba(76, 141, 255, 0.05) 40%,
    rgba(76, 141, 255, 0.16) 72%,
    rgba(126, 176, 255, 0.36) 92%,
    rgba(76, 141, 255, 0.08) 100%
  );
  box-shadow: none;
  filter: none;
  animation: scan-beam 2.4s linear infinite;
}

.scan-finder__beam::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  border-radius: 1px;
  background: linear-gradient(
    90deg,
    rgba(126, 176, 255, 0) 0%,
    rgba(126, 176, 255, 0.7) 18%,
    rgba(255, 255, 255, 0.95) 50%,
    rgba(126, 176, 255, 0.7) 82%,
    rgba(126, 176, 255, 0) 100%
  );
  box-shadow: 0 0 10px rgba(76, 141, 255, 0.55);
}

@keyframes scan-beam {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  6% {
    opacity: 1;
  }
  94% {
    opacity: 1;
  }
  100% {
    transform: translateY(363%);
    opacity: 0;
  }
}
</style>
