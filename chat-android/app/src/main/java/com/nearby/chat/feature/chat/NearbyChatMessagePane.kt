package com.nearby.chat.feature.chat

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.dismissKeyboardOnTap
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadState
import com.nearby.chat.feature.chat.ui.ChatLoadMoreHint
import com.nearby.chat.feature.chat.ui.ChatThreadComposer
import com.nearby.chat.feature.chat.ui.chatMessageItems

private val MessagePaneHPad = 10.dp

@Composable
internal fun NearbyChatMessagePane(
    loading: Boolean,
    locationReady: Boolean,
    chat: ChatThread,
    threadState: ChatThreadState,
    listState: LazyListState,
    onPickLocation: () -> Unit,
    onLoadMore: () -> Unit,
    onOpenUser: (Long) -> Unit,
    onOpenPost: (Long) -> Unit,
    onForward: (MessageItem) -> Unit,
    onReportMessage: (MessageItem) -> Unit,
) {
    var moreOpen by remember { mutableStateOf(false) }
    val dismissExtras = { moreOpen = false }
    val messages = threadState.messages
    Column(Modifier.fillMaxSize().padding(horizontal = MessagePaneHPad)) {
        when {
            loading && messages.isEmpty() -> {
                Box(
                    Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .dismissKeyboardOnTap(dismissExtras),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                }
            }

            messages.isEmpty() -> {
                EmptyHint(
                    stringResource(R.string.nearby_empty_messages),
                    Modifier.weight(1f).fillMaxWidth().dismissKeyboardOnTap(dismissExtras),
                )
            }

            else -> {
                val displayMessages = remember(messages) { messages.asReversed() }
                var listHeightPx by remember { mutableIntStateOf(0) }
                var stickAfterShrink by remember { mutableStateOf(false) }
                LaunchedEffect(stickAfterShrink) {
                    if (!stickAfterShrink || messages.isEmpty()) return@LaunchedEffect
                    listState.scrollToItem(0)
                    stickAfterShrink = false
                }
                LazyColumn(
                    state = listState,
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .dismissKeyboardOnTap(dismissExtras)
                        .onSizeChanged { size ->
                            if (listHeightPx > 0 && size.height < listHeightPx) {
                                stickAfterShrink = true
                            }
                            listHeightPx = size.height
                        },
                    contentPadding = PaddingValues(vertical = AppSpacing.Xs),
                    reverseLayout = true,
                ) {
                    chatMessageItems(
                        messages = displayMessages,
                        selfId = threadState.selfId,
                        newestFirst = true,
                        onAvatarClick = onOpenUser,
                        onReply = chat::setReplyTo,
                        onForward = onForward,
                        onReport = onReportMessage,
                        onRecall = { chat.recall(it.id) },
                        onPostClick = onOpenPost,
                    )
                    if (threadState.hasMore) {
                        item { ChatLoadMoreHint() }
                    }
                }
                InfiniteScrollEffect(
                    listState = listState,
                    enabled = threadState.hasMore,
                    onNearEnd = onLoadMore,
                )
            }
        }

        ChatThreadComposer(
            thread = chat,
            state = threadState,
            onPickLocation = onPickLocation,
            sendEnabled = threadState.input.isNotBlank() && locationReady,
            canCompose = locationReady,
            moreOpen = moreOpen,
            onMoreOpenChange = { moreOpen = it },
            modifier = Modifier.padding(bottom = AppSpacing.S, top = AppSpacing.Xs),
        )
    }
}
