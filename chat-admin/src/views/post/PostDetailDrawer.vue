<template>
  <el-drawer v-model="visible" title="帖子详情" size="720px">
    <template v-if="current">
      <el-alert
        v-if="current.deleted"
        type="info"
        :closable="false"
        show-icon
        title="该帖子已被作者删除，仅后台可见"
        class="section-gap"
      />
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID"><CopyText :text="current.id" /></el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTag(current.status)">{{ statusLabel(current.status) }}</el-tag>
          <el-tag v-if="current.deleted" type="info" class="ml-4">已删除</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="类型">
          {{ postCategoryLabel(current.category) }}
          <el-tag v-if="current.pinned" size="small" type="warning" class="ml-4">置顶</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="归属">
          <template v-if="current.roomId > 0">
            <el-button
              v-if="can(PERM.ChatRoom)"
              type="primary"
              link
              @click="emit('go-room', current.roomId)"
            >
              {{ current.roomName || '聊天区' }} #{{ current.roomId }}
            </el-button>
            <span v-else>{{ current.roomName || '聊天区' }} #{{ current.roomId }}</span>
          </template>
          <span v-else>独立帖</span>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ current.title }}</el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">
          <div class="content-block">{{ current.content || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="发布者" :span="2">
          <div class="publisher-row">
            <UserCell
              :user-id="current.publisherId"
              :nickname="current.publisher?.nickname"
              :mobile="current.publisherMobile"
            />
            <el-button
              v-if="current.publisherId > 0"
              type="primary"
              link
              @click="emit('filter-publisher', current.publisherId)"
            >
              TA 的全部帖子
            </el-button>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="地点">{{ current.placeName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="坐标">
          {{ current.latitude }}, {{ current.longitude }}
        </el-descriptions-item>
        <el-descriptions-item label="活动时间" :span="2">
          <template v-if="current.startAt || current.endAt">
            {{ formatUnixSeconds(current.startAt) }} ~ {{ formatUnixSeconds(current.endAt) }}
          </template>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="发布时间">{{ current.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ current.updatedAt }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions
        v-if="current.auditedAt || current.auditRemark"
        title="审核记录"
        :column="2"
        border
        class="section"
      >
        <el-descriptions-item label="操作人">
          {{ current.auditedByName || '系统' }}
        </el-descriptions-item>
        <el-descriptions-item label="时间">{{ current.auditedAt || '-' }}</el-descriptions-item>
        <el-descriptions-item label="说明" :span="2">
          {{ current.auditRemark || '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="current.images?.length" class="section">
        <div class="section-title">图片</div>
        <el-image
          v-for="(url, idx) in current.images"
          :key="idx"
          :src="url"
          :preview-src-list="current.images"
          fit="cover"
          class="thumb"
        />
      </div>

      <el-descriptions
        v-if="hasExtrasSection(current)"
        title="附属信息"
        :column="2"
        border
        class="section"
      >
        <template v-if="current.category === PostCategory.Idle">
          <el-descriptions-item label="价格" :span="2">
            {{ formatIdlePrice(current) }}
          </el-descriptions-item>
        </template>
        <template v-if="current.extras?.needSignup">
          <el-descriptions-item label="需要报名">是</el-descriptions-item>
          <el-descriptions-item label="需要核验">
            {{ current.extras.needCheckIn ? '是' : '否' }}
          </el-descriptions-item>
          <el-descriptions-item label="报名截止">
            {{ formatUnixSeconds(current.extras.signupDeadlineAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="人数上限">
            <template v-if="(current.extras.maxParticipants ?? 0) > 0">
              限 {{ current.extras.maxParticipants }} 人
              <template v-if="current.extras.allowWaitlist">，可候补</template>
            </template>
            <template v-else>不限</template>
          </el-descriptions-item>
          <el-descriptions-item label="报名统计" :span="2">
            <template v-if="current.signupSummary">
              已报 {{ current.signupSummary.confirmed }}
              <template v-if="current.signupSummary.waitlisted">
                ，候补 {{ current.signupSummary.waitlisted }}
              </template>
              <template v-if="current.signupSummary.cancelled">
                ，取消 {{ current.signupSummary.cancelled }}
              </template>
              <template v-if="current.extras?.needCheckIn">
                ；未核验 {{ current.signupSummary.notCheckedIn ?? 0 }} ，已核验
                {{ current.signupSummary.checkedIn ?? 0 }}
              </template>
            </template>
            <el-button
              v-if="can(PERM.PostPostDetail)"
              type="primary"
              link
              class="ml-8"
              @click="emit('signup', current)"
            >
              查看名单
            </el-button>
          </el-descriptions-item>
        </template>
        <el-descriptions-item
          v-if="current.extras?.feeType === 1 || current.extras?.feeType === 2"
          label="费用/报酬"
          :span="2"
        >
          {{ current.extras.feeType === 1 ? '收费' : '悬赏' }} ¥{{ current.extras.amount }}
          <span v-if="current.extras.feeDescription" class="sub">
            （{{ current.extras.feeDescription }}）
          </span>
        </el-descriptions-item>
      </el-descriptions>

      <div
        v-if="current.status === PostStatus.Published && can(PERM.ChatMessage)"
        class="detail-actions"
      >
        <el-button type="primary" link @click="emit('messages', current.id)">
          查看帖子消息
        </el-button>
      </div>
    </template>

    <template v-if="current && actions.length" #footer>
      <div class="audit-actions">
        <el-button
          v-for="act in actions"
          :key="act.status"
          :type="act.type"
          @click="emit('status', current, act.status)"
        >
          {{ act.label }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { type PostDetail, type PostItem } from '@/api/post'
import CopyText from '@/components/common/CopyText.vue'
import UserCell from '@/components/common/UserCell.vue'
import { formatUnixSeconds } from '@/utils/datetime'
import {
  canChangePostStatus,
  formatIdlePrice,
  hasExtrasSection,
  postCategoryLabel,
  postStatusActionLabel,
  statusLabel,
  statusTag,
} from '@/utils/postDisplay'
import { PostCategory, PostStatus } from '@/constants/post'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const visible = defineModel<boolean>({ required: true })

const props = defineProps<{ current: PostDetail | null }>()

const emit = defineEmits<{
  signup: [row: Pick<PostItem, 'id' | 'title'>]
  messages: [postId: number]
  'go-room': [roomId: number]
  'filter-publisher': [userId: number]
  status: [row: PostItem, status: PostStatus]
}>()

const { can } = usePermission()

const actions = computed(() => {
  const row = props.current
  if (!row || !can(PERM.PostPostStatus)) return []
  const order: { status: PostStatus; type: 'danger' | 'success' }[] = [
    { status: PostStatus.Rejected, type: 'danger' },
    { status: PostStatus.Disabled, type: 'danger' },
    { status: PostStatus.Published, type: 'success' },
  ]
  return order
    .filter((a) => canChangePostStatus(row, a.status))
    .map((a) => ({ ...a, label: postStatusActionLabel(row.status, a.status) }))
})
</script>

<style scoped>
.section {
  margin-top: 16px;
}

.section-gap {
  margin-bottom: 12px;
}

.section-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.ml-4 {
  margin-left: 4px;
}

.ml-8 {
  margin-left: 8px;
}

.publisher-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.thumb {
  width: 72px;
  height: 72px;
  margin-right: 8px;
  margin-bottom: 8px;
  border-radius: 4px;
}

.content-block {
  white-space: pre-wrap;
  word-break: break-word;
}

.detail-actions {
  margin-top: 16px;
}

.audit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
