<script setup lang="ts">
import type { LegalDoc } from '../legalDoc'
import LegalBlocks from './LegalBlocks.vue'

defineProps<{ doc: LegalDoc }>()

function scrollToSection(id: string) {
  document.getElementById(`legal-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
</script>

<template>
  <article class="legal-doc">
    <h1 class="legal-title">{{ doc.title }}</h1>
    <p class="legal-dates">
      <span>更新日期：{{ doc.updatedAt }}</span>
      <span>生效日期：{{ doc.effectiveAt }}</span>
    </p>

    <LegalBlocks :blocks="doc.preamble" />

    <nav class="legal-toc" aria-label="目录">
      <div class="legal-toc-title">目录</div>
      <a
        v-for="section in doc.sections"
        :key="section.id"
        :href="`#legal-${section.id}`"
        @click.prevent="scrollToSection(section.id)"
      >
        {{ section.heading }}
      </a>
    </nav>

    <section
      v-for="section in doc.sections"
      :id="`legal-${section.id}`"
      :key="section.id"
      class="legal-section"
    >
      <h2>{{ section.heading }}</h2>
      <LegalBlocks v-if="section.blocks" :blocks="section.blocks" />
      <template v-for="sub in section.subsections ?? []" :key="sub.heading">
        <h3>{{ sub.heading }}</h3>
        <LegalBlocks :blocks="sub.blocks" />
      </template>
    </section>
  </article>
</template>

<style scoped>
.legal-doc {
  color: #323232;
  font-size: 14px;
  line-height: 1.75;
  word-break: break-word;
}
.legal-doc :deep(strong) {
  color: var(--text-primary);
  font-weight: 600;
}
.legal-title {
  margin: 0 0 6px;
  font-size: 18px;
  text-align: center;
  color: var(--text-primary);
}
.legal-dates {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin: 0 0 14px;
  font-size: 12px;
  color: var(--text-tertiary);
}
.legal-toc {
  display: flex;
  flex-direction: column;
  margin: 14px 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f7f7f7;
  font-size: 13px;
}
.legal-toc-title {
  margin-bottom: 4px;
  font-weight: 600;
  color: var(--text-primary);
}
.legal-toc a {
  padding: 2px 0;
  color: var(--text-secondary);
  text-decoration: none;
}
.legal-section {
  scroll-margin-top: 56px;
}
.legal-section h2 {
  margin: 18px 0 8px;
  font-size: 15px;
  color: var(--text-primary);
}
.legal-section h3 {
  margin: 12px 0 6px;
  font-size: 14px;
  color: var(--text-primary);
}
</style>
