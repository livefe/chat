package com.nearby.chat.feature.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.feature.post.ui.PostFeedItem

private val ProgressStroke = 2.dp
private val LoadMoreIndicatorSize = 22.dp
private val PostsTitleVerticalPadding = 10.dp

@Composable
fun UserProfileScreen(
    userId: Long,
    onBack: () -> Unit,
    onChat: (Long, String) -> Unit,
    onReport: (Long, String, String) -> Unit,
    onOpenPost: (Long) -> Unit,
    viewModel: UserProfileViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var menuVisible by remember { mutableStateOf(false) }
    var reportConfirmVisible by remember { mutableStateOf(false) }
    var blockConfirmVisible by remember { mutableStateOf(false) }
    var unblockConfirmVisible by remember { mutableStateOf(false) }

    LaunchedEffect(userId) { viewModel.load(userId) }

    InfiniteScrollEffect(
        listState = listState,
        enabled = state.hasMore && !state.postsLoading,
        onNearEnd = viewModel::loadMorePosts,
    )

    val profile = state.profile
    val canMessage = profile != null && !profile.deactivated && profile.canMessage
    val canReport = profile != null && !profile.isSelf && !profile.deactivated
    val showPosts = profile != null && !profile.deactivated && (profile.isSelf || profile.canMessage)
    val displayName = profile?.nickname.orEmpty()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.PageBg),
    ) {
        UserProfileHeader(
            state = state,
            canReport = canReport,
            canMessage = canMessage,
            onBack = onBack,
            onMenu = { menuVisible = true },
            onChat = { onChat(it.id, it.nickname) },
        )
        HorizontalDivider(color = AppColors.BorderSoft, thickness = 1.dp)

        if (showPosts) {
            Text(
                stringResource(R.string.profile_posts),
                fontSize = AppType.Secondary,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextSecondary,
                modifier = Modifier.padding(horizontal = AppSpacing.L, vertical = PostsTitleVerticalPadding),
            )
            AppPullToRefreshBox(
                isRefreshing = state.refreshing,
                onRefresh = viewModel::refreshPosts,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth(),
            ) {
                UserPostsContent(state = state, listState = listState, onOpenPost = onOpenPost)
            }
        } else {
            Spacer(Modifier.weight(1f))
        }
    }

    if (menuVisible && canReport && profile != null) {
        UserProfileActionSheet(
            blockedByMe = profile.blockedByMe,
            onDismiss = { menuVisible = false },
            onBlock = {
                menuVisible = false
                blockConfirmVisible = true
            },
            onUnblock = {
                menuVisible = false
                unblockConfirmVisible = true
            },
            onReport = {
                menuVisible = false
                reportConfirmVisible = true
            },
        )
    }

    if (reportConfirmVisible && profile != null) {
        AppConfirmDialog(
            title = stringResource(R.string.profile_report_title),
            message = stringResource(R.string.profile_report_confirm, displayName),
            danger = true,
            onConfirm = {
                reportConfirmVisible = false
                onReport(profile.id, displayName, profile.avatarUrl)
            },
            onDismiss = { reportConfirmVisible = false },
        )
    }

    if (blockConfirmVisible && profile != null) {
        AppConfirmDialog(
            title = stringResource(R.string.profile_block_title),
            message = stringResource(R.string.profile_block_confirm, displayName),
            danger = true,
            onConfirm = {
                blockConfirmVisible = false
                viewModel.blockUser()
            },
            onDismiss = { blockConfirmVisible = false },
        )
    }

    if (unblockConfirmVisible && profile != null) {
        AppConfirmDialog(
            title = stringResource(R.string.block_unblock_title),
            message = stringResource(R.string.profile_unblock_confirm, displayName),
            onConfirm = {
                unblockConfirmVisible = false
                viewModel.unblockUser()
            },
            onDismiss = { unblockConfirmVisible = false },
        )
    }
}

@Composable
private fun UserPostsContent(
    state: UserProfileUiState,
    listState: LazyListState,
    onOpenPost: (Long) -> Unit,
) {
    when {
        state.postsLoading && state.posts.isEmpty() && !state.refreshing -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppColors.Brand, strokeWidth = ProgressStroke)
            }
        }

        !state.hasMore && state.posts.isEmpty() -> {
            EmptyHint(stringResource(R.string.my_posts_empty_published), Modifier.fillMaxSize())
        }

        else -> {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppColors.CardBg),
            ) {
                items(state.posts, key = { it.id }) { post ->
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { onOpenPost(post.id) },
                    ) {
                        PostFeedItem(post = post)
                    }
                    HorizontalDivider(
                        color = AppColors.Divider,
                        modifier = Modifier.padding(start = AppSpacing.L),
                    )
                }
                if (state.postsLoading && state.posts.isNotEmpty() && !state.refreshing) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(AppSpacing.L),
                            contentAlignment = Alignment.Center,
                        ) {
                            CircularProgressIndicator(
                                color = AppColors.Brand,
                                strokeWidth = ProgressStroke,
                                modifier = Modifier.size(LoadMoreIndicatorSize),
                            )
                        }
                    }
                }
            }
        }
    }
}
