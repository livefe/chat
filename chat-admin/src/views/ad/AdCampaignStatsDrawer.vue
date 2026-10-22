<template>
  <el-drawer v-model="visible" :title="title" size="620px" destroy-on-close>
    <div class="stats-toolbar">
      <el-date-picker
        v-model="range"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        unlink-panels
        :clearable="false"
        :shortcuts="DATE_RANGE_SHORTCUTS"
        :disabled-date="(d: Date) => d.getTime() > Date.now()"
        style="width: 260px"
        @change="load"
      />
      <el-button :disabled="!list.length" @click="exportCsv">导出 CSV</el-button>
    </div>

    <div v-loading="loading">
      <div class="stats-chart">
        <div class="stats-chart__head">
          <el-radio-group v-model="metric" size="small">
            <el-radio-button value="impressions">曝光</el-radio-button>
            <el-radio-button value="clicks">点击</el-radio-button>
            <el-radio-button value="ctr">CTR</el-radio-button>
          </el-radio-group>
          <span class="stats-chart__max">峰值 {{ formatMetric(chart.max) }}</span>
        </div>
        <svg
          :viewBox="`0 0 ${CHART_W} ${CHART_H}`"
          class="stats-chart__svg"
          preserveAspectRatio="none"
          role="img"
          aria-label="每日趋势"
        >
          <line
            v-for="y in [0.25, 0.5, 0.75]"
            :key="y"
            x1="0"
            :x2="CHART_W"
            :y1="CHART_H * y"
            :y2="CHART_H * y"
            class="stats-chart__grid"
          />
          <polyline v-if="chart.points" :points="chart.points" class="stats-chart__line" />
          <circle
            v-for="p in chart.dots"
            :key="p.date"
            :cx="p.x"
            :cy="p.y"
            r="2.5"
            class="stats-chart__dot"
          >
            <title>{{ p.date }}：{{ formatMetric(p.value) }}</title>
          </circle>
        </svg>
        <div class="stats-chart__axis">
          <span>{{ days[0] }}</span>
          <span>{{ days[days.length - 1] }}</span>
        </div>
      </div>

      <el-table :data="list" size="small" show-summary :summary-method="summary" max-height="420">
        <el-table-column prop="statDate" label="日期" width="120" />
        <el-table-column prop="impressions" label="曝光" />
        <el-table-column prop="clicks" label="点击" />
        <el-table-column label="CTR">
          <template #default="{ row }">{{ formatAdCtr(row.ctr) }}</template>
        </el-table-column>
        <template #empty>
          <TableEmpty :error="error" text="该时间段暂无数据" @retry="load" />
        </template>
      </el-table>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getAdCampaignDailyStats, type AdDailyStatItem } from '@/api/ad'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { DATE_RANGE_SHORTCUTS } from '@/constants'
import { formatAdCtr } from '@/constants/ad'
import { showApiError } from '@/utils/apiError'

const CHART_W = 560
const CHART_H = 140
const DEFAULT_DAYS = 14

type Metric = 'impressions' | 'clicks' | 'ctr'

const props = defineProps<{ campaignId: number; campaignTitle: string }>()
const visible = defineModel<boolean>({ required: true })

const range = ref<[string, string]>(defaultRange())
const list = ref<AdDailyStatItem[]>([])
const loading = ref(false)
const error = ref(false)
const metric = ref<Metric>('impressions')

const title = computed(() => `每日数据 · ${props.campaignTitle} (#${props.campaignId})`)

function ymd(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function defaultRange(): [string, string] {
  const end = new Date()
  const start = new Date()
  start.setDate(end.getDate() - (DEFAULT_DAYS - 1))
  return [ymd(start), ymd(end)]
}

const days = computed(() => {
  const out: string[] = []
  const [from, to] = range.value
  const cur = new Date(`${from}T00:00:00`)
  const end = new Date(`${to}T00:00:00`)
  while (cur <= end && out.length < 400) {
    out.push(ymd(cur))
    cur.setDate(cur.getDate() + 1)
  }
  return out
})

function metricOf(item: AdDailyStatItem | undefined) {
  if (!item) return 0
  return metric.value === 'ctr' ? item.ctr : item[metric.value]
}

function formatMetric(v: number) {
  return metric.value === 'ctr' ? formatAdCtr(v) : String(v)
}

const chart = computed(() => {
  const byDate = new Map(list.value.map((x) => [x.statDate, x]))
  const values = days.value.map((d) => metricOf(byDate.get(d)))
  const max = Math.max(0, ...values)
  const step = values.length > 1 ? CHART_W / (values.length - 1) : 0
  const pad = 8
  const dots = values.map((value, i) => ({
    date: days.value[i] ?? '',
    value,
    x: values.length > 1 ? i * step : CHART_W / 2,
    y: max > 0 ? CHART_H - pad - (value / max) * (CHART_H - pad * 2) : CHART_H - pad,
  }))
  return { max, dots, points: dots.map((p) => `${p.x},${p.y}`).join(' ') }
})

function summary({ data }: { data: AdDailyStatItem[] }) {
  const impressions = data.reduce((n, x) => n + x.impressions, 0)
  const clicks = data.reduce((n, x) => n + x.clicks, 0)
  return [
    '合计',
    String(impressions),
    String(clicks),
    formatAdCtr(impressions ? clicks / impressions : 0),
  ]
}

async function load() {
  if (!props.campaignId) return
  loading.value = true
  error.value = false
  try {
    const { data } = await getAdCampaignDailyStats({
      campaignId: props.campaignId,
      fromDate: range.value[0],
      toDate: range.value[1],
    })
    list.value = data.list || []
  } catch (e) {
    list.value = []
    error.value = true
    showApiError(e, '加载日报失败')
  } finally {
    loading.value = false
  }
}

function exportCsv() {
  const rows = [['日期', '曝光', '点击', 'CTR']]
  for (const x of list.value) {
    rows.push([x.statDate, String(x.impressions), String(x.clicks), formatAdCtr(x.ctr)])
  }
  rows.push(summary({ data: list.value }))
  const csv = rows.map((r) => r.map((c) => `"${c.replace(/"/g, '""')}"`).join(',')).join('\r\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `广告日报_${props.campaignId}_${range.value[0]}_${range.value[1]}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

watch(visible, (open) => {
  if (!open) return
  range.value = defaultRange()
  metric.value = 'impressions'
  void load()
})
</script>

<style scoped>
.stats-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.stats-chart {
  margin-bottom: 16px;
  padding: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.stats-chart__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.stats-chart__max,
.stats-chart__axis {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.stats-chart__svg {
  width: 100%;
  height: 140px;
  overflow: visible;
}

.stats-chart__grid {
  stroke: var(--el-border-color-lighter);
  stroke-dasharray: 4 4;
}

.stats-chart__line {
  fill: none;
  stroke: var(--el-color-primary);
  stroke-width: 2;
  vector-effect: non-scaling-stroke;
}

.stats-chart__dot {
  fill: var(--el-color-primary);
}

.stats-chart__axis {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
}
</style>
