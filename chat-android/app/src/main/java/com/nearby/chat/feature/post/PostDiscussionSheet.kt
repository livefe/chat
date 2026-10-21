package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.dismissKeyboardOnTap
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadState
import com.nearby.chat.feature.chat.ui.ChatComposerBar
import com.nearby.chat.feature.chat.ui.ChatLoadMoreHint
import com.nearby.chat.feature.chat.ui.chatMessageItems
import com.nearby.chat.feature.chat.ui.scrollToContentEnd
import kotlinx.coroutines.flow.first

private const val SHEET_HEIGHT_FRACTION = 0.52f

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun PostDiscussionSheet(
    thread: ChatThread,
    state: ChatThreadState,
    onlineCount: Int,
    onDismiss: () -> Unit,
    onLoadMore: () -> Unit,
    onPickLocation: () -> Unit,
    onForwardMessage: (MessageItem) -> Unit,
    onReportMessage: (MessageItem) -> Unit,
    onOpenPost: (Long) -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    val listState = rememberLazyListState()
    var composerMoreOpen by remember { mutableStateOf(false) }
    var openedAtLatest by remember { mutableStateOf(false) }

    LaunchedEffect(state.messages.lastOrNull()?.id, state.messages.lastOrNull()?.clientMsgId) {
        val last = state.messages.lastOrNull() ?: return@LaunchedEffect
        snapshotFlow { listState.layoutInfo.viewportSize.height }.first { it > 0 }
        val header = if (state.hasMore) 1 else 0
        val latestIndex = header + state.messages.lastIndex
        val atEnd = listState.layoutInfo.visibleItemsInfo.any { it.index == listState.layoutInfo.totalItemsCount - 1 }
        if (!openedAtLatest || last.fromUserId == state.selfId || atEnd) {
            listState.scrollToContentEnd(latestIndex)
            openedAtLatest = true
        }
    }

    val sheetMax = (LocalConfiguration.current.screenHeightDp * SHEET_HEIGHT_FRACTION).dp
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        modifier = Modifier.height(sheetMax),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = AppSheetTokens.TitleHPad, vertical = AppSpacing.Xxs),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                stringResource(R.string.post_discussion),
                fontSize = AppSheetTokens.TitleSize,
                fontWeight = FontWeight.SemiBold,
            )
            Spacer(modifier = Modifier.width(AppSpacing.S))
            Box(
                Modifier
                    .size(6.dp)
                    .clip(CircleShape)
                    .background(AppColors.Online),
            )
            Spacer(modifier = Modifier.width(5.dp))
            Text(
                stringResource(R.string.chat_online_count, onlineCount),
                fontSize = AppType.Caption,
                color = AppColors.TextMuted,
            )
        }
        HorizontalDivider(color = AppColors.BorderSoft, modifier = Modifier.padding(top = 6.dp))
        if (state.messages.isEmpty()) {
            EmptyHint(
                stringResource(R.string.post_discussion_empty),
                Modifier
                    .weight(1f)
                    .dismissKeyboardOnTap { composerMoreOpen = false },
            )
        } else {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .background(AppColors.PageBg)
                    .dismissKeyboardOnTap { composerMoreOpen = false },
                contentPadding = PaddingValues(vertical = AppSpacing.S),
            ) {
                if (state.hasMore) {
                    item { ChatLoadMoreHint() }
                }
                chatMessageItems(
                    messages = state.messages,
                    selfId = state.selfId,
                    onReply = thread::setReplyTo,
                    onForward = onForwardMessage,
                    onReport = onReportMessage,
                    onRecall = { thread.recall(it.id) },
                    onPostClick = onOpenPost,
                )
            }
            InfiniteScrollEffect(
                listState = listState,
                enabled = state.hasMore,
                nearStart = true,
                onNearEnd = onLoadMore,
            )
        }
        ChatComposerBar(
            input = state.input,
            onInputChange = thread::onInputChange,
            onSend = thread::sendText,
            onImagePicked = thread::sendImage,
            onFilePicked = thread::sendFile,
            onPickLocation = onPickLocation,
            onVoiceRecorded = thread::sendVoice,
            uploadVoice = thread::uploadVoice,
            sendEnabled = state.input.isNotBlank(),
            uploading = state.uploading,
            placeholder = stringResource(R.string.post_discussion_placeholder),
            replyTo = state.replyTo,
            onClearReply = { thread.setReplyTo(null) },
            moreOpen = composerMoreOpen,
            onMoreOpenChange = { composerMoreOpen = it },
        )
    }
}
