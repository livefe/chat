import { reactive } from 'vue'
import { ElMessage } from 'element-plus'

import {
  batchUpdateUserStatus,
  deactivateUser,
  muteUser,
  updateUserStatus,
  type BatchUserResult,
} from '@/api/user'
import { UserStatus } from '@/constants/user'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import { confirmWithReason } from '@/utils/promptReason'

export interface UserActionTarget {
  id: number
  nickname?: string
  mobile?: string
  status?: number
}

export interface UserFormModel {
  id: number
  maskedMobile: string
  nickname: string
}

const muteDialog = reactive({ visible: false, users: [] as UserActionTarget[] })
const formDialog = reactive({
  visible: false,
  model: { id: 0, maskedMobile: '', nickname: '' } as UserFormModel,
})

export function useUserActionDialogs() {
  return { muteDialog, formDialog }
}

export function userLabel(u: UserActionTarget) {
  return u.nickname ? `「${u.nickname}」(ID ${u.id})` : `ID ${u.id}`
}

export function batchResultText(action: string, r: BatchUserResult) {
  return r.skipped > 0
    ? `已${action} ${r.updated} 人，跳过 ${r.skipped} 人`
    : `已${action} ${r.updated} 人`
}

export function useUserActions() {
  const drawer = useUserDrawerStore()
  const changed = () => drawer.refresh()

  function openCreate() {
    formDialog.model = { id: 0, maskedMobile: '', nickname: '' }
    formDialog.visible = true
  }

  function openEdit(u: UserActionTarget) {
    formDialog.model = { id: u.id, maskedMobile: u.mobile ?? '', nickname: u.nickname ?? '' }
    formDialog.visible = true
  }

  function openMute(users: UserActionTarget[]) {
    if (users.length === 0) return
    muteDialog.users = users
    muteDialog.visible = true
  }

  function unmute(u: UserActionTarget) {
    return confirmAndRun(
      `确定解除 ${userLabel(u)} 的禁言吗？`,
      async () => {
        await muteUser(u.id, 0, '')
        changed()
      },
      '已解除禁言',
    )
  }

  async function disable(u: UserActionTarget) {
    const reason = await confirmWithReason(
      `禁用后 ${userLabel(u)} 将被强制下线且无法登录，确定禁用吗？`,
      { title: '禁用账号', required: true, inputPlaceholder: '请输入禁用原因（将记入处罚记录）' },
    )
    if (reason === null) return
    try {
      await updateUserStatus(u.id, UserStatus.Disabled, reason)
      ElMessage.success('已禁用')
      changed()
    } catch (e) {
      showApiError(e, '禁用失败')
    }
  }

  function enable(u: UserActionTarget) {
    return confirmAndRun(
      `确定启用 ${userLabel(u)} 吗？`,
      async () => {
        await updateUserStatus(u.id, UserStatus.Enabled, '')
        changed()
      },
      '已启用',
    )
  }

  async function deactivate(u: UserActionTarget) {
    const reason = await confirmWithReason(
      `注销后 ${userLabel(u)} 的手机号、昵称、头像将被清除，且不可恢复。确定注销吗？`,
      { title: '注销用户', required: true, inputPlaceholder: '请输入注销原因' },
    )
    if (reason === null) return
    try {
      await deactivateUser(u.id, reason)
      ElMessage.success('已注销')
      changed()
    } catch (e) {
      showApiError(e, '注销失败')
    }
  }

  async function batchDisable(users: UserActionTarget[]) {
    if (users.length === 0) return
    const reason = await confirmWithReason(
      `确定禁用选中的 ${users.length} 个用户吗？他们将被强制下线且无法登录。`,
      {
        title: '批量禁用',
        required: true,
        inputPlaceholder: '请输入禁用原因（将记入每个用户的处罚记录）',
      },
    )
    if (reason === null) return
    try {
      const { data } = await batchUpdateUserStatus(
        users.map((u) => u.id),
        UserStatus.Disabled,
        reason,
      )
      ElMessage.success(batchResultText('禁用', data))
      changed()
    } catch (e) {
      showApiError(e, '批量禁用失败')
    }
  }

  return { openCreate, openEdit, openMute, unmute, disable, enable, deactivate, batchDisable }
}
