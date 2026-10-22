<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Menu } from '@element-plus/icons-vue'
import { useAdminBadges } from '@/composables/useAdminBadges'
import { useEmployeeStore } from '@/stores/employee'
import { buildHomeSections, buildMenus, flattenMenuLeaves } from '@/utils/menu'
import { readRecentVisits } from '@/utils/storage'

const router = useRouter()
const employeeStore = useEmployeeStore()
const { badgeFor, pendingQueryFor } = useAdminBadges()

const name = computed(
  () => employeeStore.employeeInfo?.name || employeeStore.employeeInfo?.mobile || '同事',
)

const sections = computed(() => buildHomeSections(employeeStore.permissions))

const recentVisits = computed(() => {
  const id = employeeStore.employeeInfo?.id
  if (!id) return []
  const titleByPath = new Map(
    flattenMenuLeaves(buildMenus(employeeStore.permissions)).map((m) => [m.path, m.title]),
  )
  return readRecentVisits(id)
    .filter((v) => titleByPath.has(v.path))
    .map((v) => ({ path: v.path, title: titleByPath.get(v.path) || v.title }))
})

function open(path: string) {
  if (path) void router.push({ path, query: pendingQueryFor(path) })
}
</script>

<template>
  <div class="admin-page home-page">
    <div class="home-hero">
      <h1 class="home-hero__title">你好，{{ name }}</h1>
      <p class="home-hero__desc">从下方入口进入常用功能；按 Ctrl+K 可搜索菜单。</p>
    </div>

    <section v-if="recentVisits.length" class="home-section">
      <h2 class="home-section__title">最近访问</h2>
      <div class="home-recent">
        <el-button v-for="item in recentVisits" :key="item.path" @click="open(item.path)">
          {{ item.title }}
        </el-button>
      </div>
    </section>

    <el-empty v-if="sections.length === 0" description="当前账号暂无可用菜单权限，请联系管理员" />

    <section v-for="sec in sections" :key="sec.title" class="home-section">
      <h2 class="home-section__title">{{ sec.title }}</h2>
      <div class="home-grid">
        <button
          v-for="item in sec.items"
          :key="item.path"
          type="button"
          class="home-card"
          @click="open(item.path)"
        >
          <span class="home-card__top">
            <span class="home-card__icon" aria-hidden="true">
              <el-icon v-if="item.icon" :size="22">
                <component :is="item.icon" />
              </el-icon>
              <el-icon v-else :size="22"><Menu /></el-icon>
            </span>
            <el-badge
              v-if="badgeFor(item.path)"
              :value="badgeFor(item.path)"
              :max="99"
              class="home-card__badge"
            />
          </span>
          <span class="home-card__title">{{ item.title }}</span>
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  max-width: 960px;
}

.home-hero {
  margin-bottom: 28px;
}

.home-hero__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.35;
}

.home-hero__desc {
  margin: 8px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.home-section + .home-section {
  margin-top: 28px;
}

.home-section__title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.home-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.home-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 14px;
  min-height: 96px;
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--admin-card-radius, 10px);
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.15s,
    background 0.15s;
}

.home-card:hover {
  border-color: color-mix(in srgb, var(--el-color-primary) 45%, var(--el-border-color-lighter));
  background: color-mix(in srgb, var(--el-color-primary) 6%, #fff);
}

.home-recent {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.home-recent .el-button + .el-button {
  margin-left: 0;
}

.home-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  width: 100%;
  gap: 8px;
}

.home-card__icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  background: color-mix(in srgb, var(--el-color-primary) 10%, transparent);
  color: var(--el-color-primary);
  flex-shrink: 0;
}

.home-card__badge {
  flex-shrink: 0;
  line-height: 1;
}

.home-card__badge :deep(.el-badge__content) {
  position: static;
  transform: none;
}

.home-card__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.35;
}
</style>
