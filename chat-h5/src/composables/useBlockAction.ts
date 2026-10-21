import { showConfirmDialog, showToast } from 'vant'

import { blockUser, unblockUser } from '@/api/user'
import type { NavMoreAction } from '@/types/nav'
import { showCatchError } from '@/utils/toast'

export const ACTION_BLOCK_USER = '拉黑'
export const ACTION_UNBLOCK_USER = '取消拉黑'

export const blockUserAction: NavMoreAction = {
  name: ACTION_BLOCK_USER,
  color: '#d4544a',
}

export const unblockUserAction: NavMoreAction = {
  name: ACTION_UNBLOCK_USER,
}

export async function confirmBlockUser(userId: number, nickname: string): Promise<boolean> {
  try {
    await showConfirmDialog({
      title: '拉黑用户',
      message: `拉黑后将无法与「${nickname}」私信，且不再看到对方发布的内容。确定拉黑吗？`,
    })
  } catch {
    return false
  }
  try {
    await blockUser(userId)
    showToast('已拉黑')
    return true
  } catch (e: unknown) {
    showCatchError(e, '拉黑失败')
    return false
  }
}

export async function confirmUnblockUser(userId: number, nickname: string): Promise<boolean> {
  try {
    await showConfirmDialog({
      title: '取消拉黑',
      message: `确定取消拉黑「${nickname}」吗？`,
    })
  } catch {
    return false
  }
  try {
    await unblockUser(userId)
    showToast('已取消拉黑')
    return true
  } catch (e: unknown) {
    showCatchError(e, '操作失败')
    return false
  }
}
