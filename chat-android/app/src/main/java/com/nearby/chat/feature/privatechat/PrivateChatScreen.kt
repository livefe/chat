package com.nearby.chat.feature.privatechat

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.ime
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.InfiniteScrollEffect
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.core.ui.dismissKeyboardOnTap
import com.nearby.chat.feature.chat.ui.ChatLoadMoreHint
import com.nearby.chat.feature.chat.ui.ChatThreadComposer
import com.nearby.chat.feature.chat.ui.chatMessageItems
import com.nearby.chat.feature.chat.ui.scrollToLatestIfFollowing

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PrivateChatScreen(
    peerId: Long,
    peerName: String,
    onBack: () -> Unit,
    onOpenProfile: (Long) -> Unit,
    onOpenPost: (Long) -> Unit = {},
    onPickLocation: () -> Unit = {},
    onReportMessage: (MessageItem) -> Unit = {},
    onOpenForward: (MessageItem) -> Unit = {},
    viewModel: PrivateChatViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val thread by viewModel.threadState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    LaunchedEffect(peerId) {
        viewModel.load(peerId, peerName)
    }

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
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = {
                    AppTopBarTitle(
                        text = state.peerName,
                        modifier = Modifier.clickable(onClick = { onOpenProfile(peerId) }),
                    )
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            Modifier
                .padding(padding)
                .fillMaxSize()
                .imePadding(),
        ) {
            var moreOpen by remember { mutableStateOf(false) }
            val dismissExtras = { moreOpen = false }
            if (thread.messages.isEmpty()) {
                EmptyHint(
                    stringResource(R.string.private_chat_empty),
                    Modifier.weight(1f).dismissKeyboardOnTap(dismissExtras),
                )
            } else {
                LazyColumn(
                    state = listState,
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .dismissKeyboardOnTap(dismissExtras),
                    contentPadding = PaddingValues(vertical = AppSpacing.S),
                ) {
                    if (thread.hasMore) {
                        item { ChatLoadMoreHint() }
                    }
                    chatMessageItems(
                        messages = thread.messages,
                        selfId = thread.selfId,
                        hideName = true,
                        onAvatarClick = { uid ->
                            if (uid > 0) onOpenProfile(uid)
                        },
                        onReply = viewModel.thread::setReplyTo,
                        onForward = onOpenForward,
                        onReport = onReportMessage,
                        onRecall = { viewModel.thread.recall(it.id) },
                        onPostClick = onOpenPost,
                    )
                }
                InfiniteScrollEffect(
                    listState = listState,
                    enabled = thread.hasMore,
                    nearStart = true,
                    onNearEnd = viewModel::loadMore,
                )
            }
            ChatThreadComposer(
                thread = viewModel.thread,
                state = thread,
                onPickLocation = onPickLocation,
                sendEnabled = thread.input.isNotBlank(),
                moreOpen = moreOpen,
                onMoreOpenChange = { moreOpen = it },
                modifier = Modifier
                    .background(AppColors.PageBg)
                    .navigationBarsPadding(),
            )
        }
    }
}
