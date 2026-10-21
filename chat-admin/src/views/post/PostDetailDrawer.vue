<template>
  <el-drawer v-model="visible" title="帖子详情" size="720px">
    <template v-if="current">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID">{{ current.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTag(current.status)">{{ statusLabel(current.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="类型">
          {{ postCategoryLabel(current.category) }}
          <el-tag v-if="current.pinned" size="small" type="warning" class="pin-tag">置顶</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="归属">
          <template v-if="current.roomId > 0">
            <el-button
              v-if="can(PERM.ChatRoom)"
              type="primary"
              link
              @click="emit('go-room', current.roomId)"
            >
              聊天区 #{{ current.roomId }}
            </el-button>
            <span v-else>聊天区 #{{ current.roomId }}</span>
          </template>
          <span v-else>独立帖</span>
        </el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ current.title }}</el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">
          <div class="content-block">{{ current.content || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="发布者">
          {{ current.publisher?.nickname || '-' }}（ID {{ current.publisherId }}）
        </el-descriptions-item>
        <el-descriptions-item label="地点">{{ current.placeName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="坐标" :span="2">
          {{ current.latitude }}, {{ current.longitude }}
        </el-descriptions-item>
        <el-descriptions-item label="开始/结束" :span="2">
          <template v-if="current.startAt || current.endAt">
            {{ formatUnixSeconds(current.startAt) }} ~ {{ formatUnixSeconds(current.endAt) }}
          </template>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ current.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ current.updatedAt }}</el-descriptions-item>
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
          <el-descriptions-item v-if="current.signupSummary" label="报名统计" :span="2">
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
            <el-button
              v-if="can(PERM.PostPostDetail)"
              type="primary"
              link
              class="signup-link"
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
          <template v-if="current.extras.feeType === 1">
            收费 ¥{{ current.extras.amount }}
            <span v-if="current.extras.feeDescription" class="sub">
              （{{ current.extras.feeDescription }}）
            </span>
          </template>
          <template v-else>
            悬赏 ¥{{ current.extras.amount }}
            <span v-if="current.extras.feeDescription" class="sub">
              （{{ current.extras.feeDescription }}）
            </span>
          </template>
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

      <div v-if="canAudit" class="detail-actions audit-actions">
        <el-button type="danger" @click="emit('status', current, PostStatus.Rejected, '驳回')">
          驳回
        </el-button>
        <el-button type="success" @click="emit('status', current, PostStatus.Published, '通过')">
          通过
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { type PostDetail, type PostItem } from '@/api/post'
import { formatUnixSeconds } from '@/utils/datetime'
import {
  formatIdlePrice,
  hasExtrasSection,
  postCategoryLabel,
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
  status: [row: PostItem, status: PostStatus, action: string]
}>()

const { can } = usePermission()

const canAudit = computed(
  () => props.current?.status === PostStatus.Pending && can(PERM.PostPostStatus),
)
</script>

<style scoped>
.section {
  margin-top: 16px;
}

.section-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.signup-link {
  margin-left: 8px;
}

.pin-tag {
  margin-left: 4px;
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
