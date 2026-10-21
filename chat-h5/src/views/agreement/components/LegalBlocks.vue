<script setup lang="ts">
import type { LegalBlock } from '../legalDoc'
import LegalRichText from './LegalRichText.vue'

defineProps<{ blocks: readonly LegalBlock[] }>()
</script>

<template>
  <template v-for="(block, i) in blocks" :key="i">
    <p v-if="block.type === 'p'" class="legal-p">
      <LegalRichText :text="block.text" />
    </p>
    <ul v-else-if="block.type === 'list'" class="legal-list">
      <li v-for="(item, j) in block.items" :key="j">
        <LegalRichText :text="item" />
      </li>
    </ul>
    <div v-else class="legal-table-wrap">
      <table class="legal-table">
        <thead>
          <tr>
            <th v-for="col in block.columns" :key="col">{{ col }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, r) in block.rows" :key="r">
            <td v-for="(cell, c) in row" :key="c">
              <LegalRichText :text="cell" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>
</template>

<style scoped>
.legal-p {
  margin: 0 0 8px;
}
.legal-list {
  margin: 0 0 8px;
  padding-left: 1.4em;
}
.legal-list li + li {
  margin-top: 4px;
}
.legal-table-wrap {
  margin: 4px 0 10px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}
.legal-table {
  width: 100%;
  min-width: 520px;
  border-collapse: collapse;
  font-size: 12px;
  line-height: 1.6;
}
.legal-table th,
.legal-table td {
  padding: 6px 8px;
  border: 1px solid var(--border-color-soft);
  text-align: left;
  vertical-align: top;
}
.legal-table th {
  background: #f7f7f7;
  color: var(--text-primary);
  font-weight: 600;
  white-space: nowrap;
}
</style>
