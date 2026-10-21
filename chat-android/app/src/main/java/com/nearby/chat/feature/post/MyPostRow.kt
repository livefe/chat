package com.nearby.chat.feature.post

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.QrCode2
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.post.PostStatus
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.feature.post.ui.PostFeedBadgeMode
import com.nearby.chat.feature.post.ui.PostFeedItem
import com.nearby.chat.feature.post.ui.PostListAction
import com.nearby.chat.feature.post.ui.PostListActionBar

@Composable
internal fun MyPostRow(
    post: PostItem,
    mode: MyPostsMode,
    actingId: Long,
    onOpenPost: (Long) -> Unit,
    onEditPost: (Long) -> Unit,
    onOpenSignupList: (Long) -> Unit,
    onOpenCheckInCode: (Long) -> Unit,
    onConfirm: (MyPostConfirm) -> Unit,
) {
    val acting = actingId == post.id
    val idle = actingId == 0L
    val footer: (@Composable () -> Unit)? = if (mode == MyPostsMode.Published) {
        val actions = publishedActions(post, acting, idle, onEditPost, onOpenSignupList, onConfirm)
        if (actions.isNotEmpty()) {
            { PostListActionBar(actions = actions) }
        } else {
            null
        }
    } else {
        val showQrIcon = PostSignupRules.canShowJoinedCheckInCodeIcon(post)
        val actions = joinedActions(post, acting, idle, onOpenCheckInCode, onConfirm)
        if (showQrIcon || actions.isNotEmpty()) {
            {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    if (showQrIcon) {
                        Box(
                            contentAlignment = Alignment.Center,
                            modifier = Modifier
                                .size(width = 28.dp, height = 18.dp)
                                .clickable { onOpenCheckInCode(post.id) },
                        ) {
                            Icon(
                                imageVector = Icons.Outlined.QrCode2,
                                contentDescription = stringResource(R.string.check_in_code_label),
                                tint = AppColors.TextSecondary,
                                modifier = Modifier.size(AppIconSize.Inline),
                            )
                        }
                    }
                    if (actions.isNotEmpty()) {
                        PostListActionBar(actions = actions)
                    }
                }
            }
        } else {
            null
        }
    }
    Column {
        PostFeedItem(
            post = post,
            badgeMode = if (mode == MyPostsMode.Joined) PostFeedBadgeMode.Joined else PostFeedBadgeMode.Owner,
            onClick = { onOpenPost(post.id) },
            footerActions = footer,
        )
        HorizontalDivider(color = AppColors.Divider)
    }
}

@Composable
private fun publishedActions(
    post: PostItem,
    acting: Boolean,
    idle: Boolean,
    onEditPost: (Long) -> Unit,
    onOpenSignupList: (Long) -> Unit,
    onConfirm: (MyPostConfirm) -> Unit,
): List<PostListAction> {
    val signupList = stringResource(R.string.signup_list_title)
    val edit = stringResource(R.string.post_edit_title)
    val withdraw = stringResource(if (acting) R.string.post_withdrawing else R.string.post_withdraw)
    val delete = stringResource(if (acting) R.string.post_deleting else R.string.common_delete)
    return buildList {
        if (post.extras?.needSignup == true) {
            add(PostListAction(label = signupList) { onOpenSignupList(post.id) })
        }
        if (post.status == PostStatus.PENDING || post.status == PostStatus.REJECTED) {
            add(PostListAction(label = edit) { onEditPost(post.id) })
        }
        if (post.status != PostStatus.OFFLINE) {
            add(PostListAction(label = withdraw, danger = true, enabled = idle) { onConfirm(MyPostConfirm.Withdraw(post)) })
        }
        if (PostSignupRules.canDeletePublishedPost(post)) {
            add(
                PostListAction(label = delete, danger = true, enabled = idle) {
                    onConfirm(MyPostConfirm.DeletePublished(post))
                },
            )
        }
    }
}

@Composable
private fun joinedActions(
    post: PostItem,
    acting: Boolean,
    idle: Boolean,
    onOpenCheckInCode: (Long) -> Unit,
    onConfirm: (MyPostConfirm) -> Unit,
): List<PostListAction> {
    val viewCode = stringResource(R.string.post_view_check_in_code)
    val remove = stringResource(if (acting) R.string.post_removing else R.string.post_remove_joined)
    val showCodeInMenu = PostSignupRules.canShowJoinedCheckInCode(post) &&
        PostSignupRules.hasTimestamp(post.myCheckedInAt)
    return buildList {
        if (showCodeInMenu) {
            add(PostListAction(label = viewCode) { onOpenCheckInCode(post.id) })
        }
        if (PostSignupRules.canDeleteJoinedPost(post)) {
            add(
                PostListAction(label = remove, danger = true, enabled = idle) {
                    onConfirm(MyPostConfirm.DeleteJoined(post))
                },
            )
        }
    }
}
