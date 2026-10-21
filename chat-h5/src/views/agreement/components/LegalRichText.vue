<script setup lang="ts">
import { computed } from 'vue'

import { segmentsOf, type RichText } from '../legalDoc'

const props = defineProps<{ text: RichText }>()
const segments = computed(() => segmentsOf(props.text))
</script>

<template>
  <template v-for="(seg, i) in segments" :key="i">
    <strong v-if="typeof seg !== 'string'">{{ seg.strong }}</strong>
    <template v-else>{{ seg }}</template>
  </template>
</template>
