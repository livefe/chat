package com.nearby.chat.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.ime
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.ExperimentalMaterial3Api
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
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.dismissKeyboardOnTap
import com.nearby.chat.feature.chat.ui.ChatLoadMoreHint
import com.nearby.chat.feature.chat.ui.ChatThreadComposer
import com.nearby.chat.feature.chat.ui.chatMessageItems
import com.nearby.chat.feature.chat.ui.scrollToLatestIfFollowing

private val OnlineDotSize = 6.dp
private val OnlineDotGap = 5.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RoomDetailScreen(
    roomId: Long,
    onBack: () -> Unit,
    onOpenPost: (Long) -> Unit,
    onPickLocation: () -> Unit = {},
    onReportMessage: (MessageItem) -> Unit = {},
    onOpenForward: (MessageItem) -> Unit = {},
    viewModel: RoomDetailViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val thread by viewModel.threadState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var moreOpen by remember { mutableStateOf(false) }

    LaunchedEffect(roomId) { viewModel.load(roomId) }
    LaunchedEffect(thread.messages.lastOrNull()?.id, thread.messages.lastOrNull()?.clientMsgId) {
        val last = thread.messages.lastOrNull() ?: return@LaunchedEffect
        val header = if (thread.hasMore) 1 else 0
        listState.scrollToLatestIfFollowing(
            last = last,
            selfId = thread.selfId,
            reverseLayout = false,
            latestIndex = header + thread.messages.lastIndex,
        )
    }
    val imeOpen = WindowInsets.ime.getBottom(LocalDensity.current) > 0
    LaunchedEffect(imeOpen) {
        if (imeOpen && thread.messages.isNotEmpty()) {
            val header = if (thread.hasMore) 1 else 0
            listState.scrollToItem(header + thread.messages.lastIndex)
        }
    }

    Scaffold(
        containerColor = AppColors.CardBg,
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            state.room?.name ?: stringResource(R.string.room_default_name),
                            fontSize = AppType.PageTitle,
                            fontWeight = FontWeight.SemiBold,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                            modifier = Modifier.weight(1f, fill = false),
                        )
                        Spacer(modifier = Modifier.width(AppSpacing.S))
                        Box(
                            Modifier
                                .size(OnlineDotSize)
                                .clip(CircleShape)
                                .background(AppColors.Online),
                        )
                        Spacer(modifier = Modifier.width(OnlineDotGap))
                        Text(
                            stringResource(R.string.chat_online_count, state.onlineCount),
                            fontSize = AppType.Caption,
                            color = AppColors.TextMuted,
                        )
                    }
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                colors = appTopAppBarColors(AppColors.CardBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            RoomMessageList(
                messages = thread.messages,
                hasMoreMessages = thread.hasMore,
                selfId = thread.selfId,
                listState = listState,
                onLoadMore = viewModel::loadMoreMessages,
                onReplyTo = viewModel.thread::setReplyTo,
                onForward = onOpenForward,
                onReportMessage = onReportMessage,
                onRecall = viewModel.thread::recall,
                onOpenPost = onOpenPost,
                onDismissExtras = { moreOpen = false },
                modifier = Modifier.weight(1f),
            )
            ChatThreadComposer(
                thread = viewModel.thread,
                state = thread,
                onPickLocation = onPickLocation,
                sendEnabled = thread.input.isNotBlank(),
                moreOpen = moreOpen,
                onMoreOpenChange = { moreOpen = it },
                modifier = Modifier
                    .background(AppColors.CardBg)
                    .navigationBarsPadding()
                    .imePadding(),
            )
        }
    }
}

@Composable
private fun RoomMessageList(
    messages: List<MessageItem>,
    hasMoreMessages: Boolean,
    selfId: Long,
    listState: LazyListState,
    onLoadMore: () -> Unit,
    onReplyTo: (MessageItem?) -> Unit,
    onForward: (MessageItem) -> Unit,
    onReportMessage: (MessageItem) -> Unit,
    onRecall: (Long) -> Unit,
    onOpenPost: (Long) -> Unit,
    onDismissExtras: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .background(AppColors.PageBg),
    ) {
        if (messages.isEmpty()) {
            EmptyHint(
                stringResource(R.string.chat_empty_messages),
                Modifier.fillMaxSize().dismissKeyboardOnTap(onDismissExtras),
            )
        } else {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .dismissKeyboardOnTap(onDismissExtras),
                contentPadding = PaddingValues(vertical = AppSpacing.S),
            ) {
                if (hasMoreMessages) {
                    item { ChatLoadMoreHint() }
                }
                chatMessageItems(
                    messages = messages,
                    selfId = selfId,
                    onReply = onReplyTo,
                    onForward = onForward,
                    onReport = onReportMessage,
                    onRecall = { onRecall(it.id) },
                    onPostClick = onOpenPost,
                )
            }
            InfiniteScrollEffect(
                listState = listState,
                enabled = hasMoreMessages,
                nearStart = true,
                onNearEnd = onLoadMore,
            )
        }
    }
}
