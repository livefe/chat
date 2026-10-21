package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.MoreHoriz
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.post.PostStatus
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.appTopAppBarColors
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostDetailScreen(
    postId: Long,
    onBack: () -> Unit,
    onOpenPublisher: (Long) -> Unit,
    onOpenPost: (Long) -> Unit = {},
    onReport: (Long, String) -> Unit,
    onReportMessage: (MessageItem) -> Unit = {},
    onShare: (title: String, cover: String, body: String) -> Unit = { _, _, _ -> },
    onForwardMessage: (MessageItem) -> Unit = {},
    onPickLocation: () -> Unit = {},
    viewModel: PostDetailViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val thread by viewModel.threadState.collectAsStateWithLifecycle()
    LaunchedEffect(postId) { viewModel.load(postId) }

    val context = LocalContext.current
    val post = state.detail
    val canDiscuss = post != null && PostSignupRules.canDiscuss(post)
    val messagePeek = remember(thread.messages) { PostDetailFormatter.chatPeek(context, thread.messages) }
    val serverPeek = post?.discussPreview
    val chatPeek = when {
        !canDiscuss -> ""

        messagePeek.isNotBlank() -> messagePeek

        state.discussionLoaded -> ""

        serverPeek != null -> PostDetailFormatter.chatPeek(
            context,
            serverPeek.fromName,
            serverPeek.contentType,
            serverPeek.content,
        )

        else -> ""
    }
    val isOwner = post != null && post.publisherId == state.selfId
    var moreSheetVisible by remember { mutableStateOf(false) }
    var reportConfirmVisible by remember { mutableStateOf(false) }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = {
                    val publisher = post?.publisher
                    if (publisher != null) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier
                                .clickable { onOpenPublisher(publisher.id) }
                                .padding(end = AppSpacing.S),
                        ) {
                            UserAvatar(
                                url = publisher.avatarUrl,
                                size = 28.dp,
                            )
                            Spacer(Modifier.width(AppSpacing.S))
                            Text(
                                publisher.nickname,
                                fontSize = AppType.PageTitle,
                                fontWeight = FontWeight.SemiBold,
                                color = AppColors.TextPrimary,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    } else {
                        AppTopBarTitle(stringResource(R.string.post_detail_title))
                    }
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    if (post != null && !isOwner) {
                        IconButton(onClick = { moreSheetVisible = true }) {
                            Icon(
                                Icons.Outlined.MoreHoriz,
                                contentDescription = stringResource(R.string.common_more),
                                tint = AppColors.TextPrimary,
                                modifier = Modifier.size(AppIconSize.Cell),
                            )
                        }
                    }
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
        bottomBar = {
            if (post != null) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(AppColors.CardBg),
                ) {
                    HorizontalDivider(color = AppColors.BorderSoft)
                    if (post.status == PostStatus.OFFLINE) {
                        Text(
                            stringResource(R.string.post_offline_notice),
                            fontSize = AppType.Meta,
                            color = AppColors.TextTertiary,
                            textAlign = TextAlign.Center,
                            modifier = Modifier
                                .fillMaxWidth()
                                .navigationBarsPadding()
                                .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
                        )
                    } else {
                        if (canDiscuss && chatPeek.isNotBlank() && !state.showChat) {
                            PostChatPeekBar(
                                text = chatPeek,
                                onClick = viewModel::openChat,
                            )
                        }
                        PostDetailBottomBar(
                            post = post,
                            onShare = {
                                onShare(post.title, post.images.firstOrNull().orEmpty(), post.content)
                            },
                            onSignup = viewModel::signup,
                            onCancelSignup = viewModel::cancelSignup,
                            onOpenChat = viewModel::openChat,
                            showDiscussion = canDiscuss,
                        )
                    }
                }
            }
        },
    ) { padding ->
        if (post == null) {
            Box(
                modifier = Modifier
                    .padding(padding)
                    .fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
            }
        } else {
            PostDetailContent(
                post = post,
                modifier = Modifier.padding(padding),
            )
        }
    }

    if (state.showChat) {
        PostDiscussionSheet(
            thread = viewModel.thread,
            state = thread,
            onlineCount = state.onlineCount,
            onDismiss = viewModel::closeChat,
            onLoadMore = viewModel::loadMoreMessages,
            onPickLocation = onPickLocation,
            onForwardMessage = onForwardMessage,
            onReportMessage = onReportMessage,
            onOpenPost = onOpenPost,
        )
    }

    if (moreSheetVisible) {
        PostReportSheet(
            onReport = {
                moreSheetVisible = false
                reportConfirmVisible = true
            },
            onDismiss = { moreSheetVisible = false },
        )
    }

    if (reportConfirmVisible && post != null) {
        val reportTitle = post.title.ifBlank { stringResource(R.string.chat_post_default_title) }
        AppConfirmDialog(
            title = stringResource(R.string.post_report_title),
            message = stringResource(R.string.post_report_confirm, reportTitle),
            danger = true,
            onConfirm = {
                reportConfirmVisible = false
                onReport(postId, reportTitle)
            },
            onDismiss = { reportConfirmVisible = false },
        )
    }
}

@Composable
private fun PostDetailContent(post: PostItem, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = AppSpacing.M)
            .padding(top = AppSpacing.M, bottom = AppSpacing.Xxl),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        PostDetailCard {
            Text(
                post.title,
                fontSize = 20.sp,
                fontWeight = FontWeight.SemiBold,
                color = AppColors.TextPrimary,
                lineHeight = 28.sp,
            )
        }

        PostDetailCard {
            if (post.images.isNotEmpty()) {
                PostDetailImages(post.images)
                Spacer(Modifier.height(AppSpacing.M))
            }
            if (post.content.isNotBlank()) {
                Text(
                    post.content,
                    fontSize = AppType.Subhead,
                    color = AppColors.TextPrimary,
                    lineHeight = 26.sp,
                )
            }
            if (post.placeName.isNotBlank() || post.createdAt.isNotBlank()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 14.dp),
                    horizontalArrangement = Arrangement.spacedBy(AppSpacing.M),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (post.placeName.isNotBlank()) {
                        Text(
                            stringResource(R.string.post_location, post.placeName),
                            fontSize = AppType.Caption,
                            color = AppColors.TextSecondary,
                            modifier = Modifier.weight(1f),
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis,
                        )
                    } else {
                        Spacer(Modifier.weight(1f))
                    }
                    Text(
                        PostDetailFormatter.createdAt(post.createdAt),
                        fontSize = AppType.Caption,
                        color = AppColors.TextTertiary,
                    )
                }
            }
        }

        if (PostDetailFormatter.hasExtras(post)) {
            PostDetailCard {
                PostDetailExtras(post = post)
            }
        }
    }
}

@Composable
private fun PostDetailCard(content: @Composable () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppShapes.Card)
            .background(AppColors.CardBg)
            .padding(AppSpacing.L),
        content = { content() },
    )
}
