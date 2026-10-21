package com.nearby.chat.feature.messages

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.format.sanitizeNickname
import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.DarkActionItem
import com.nearby.chat.core.ui.DarkActionPillPopup
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.appTopAppBarColors

private val ConversationDividerInset = AppSpacing.L + ConversationAvatarSize + AppSpacing.M

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MessagesScreen(
    onOpenChat: (Long, String) -> Unit,
    viewModel: MessagesViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    var actionTarget by remember { mutableStateOf<ConversationItem?>(null) }
    var actionAnchor by remember { mutableStateOf(Rect.Zero) }
    var deleteTarget by remember { mutableStateOf<ConversationItem?>(null) }

    LaunchedEffect(Unit) { viewModel.load() }

    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                viewModel.load()
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = {
            TopAppBar(
                title = {
                    AppTopBarTitle(stringResource(R.string.tab_messages))
                },
                windowInsets = WindowInsets.statusBars,
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            when {
                state.loading && state.list.isEmpty() -> {
                    CircularProgressIndicator(
                        color = AppColors.Brand,
                        strokeWidth = 2.dp,
                        modifier = Modifier.align(Alignment.Center),
                    )
                }

                state.list.isEmpty() -> {
                    EmptyHint(stringResource(R.string.messages_empty), Modifier.fillMaxSize())
                }

                else -> {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                    ) {
                        items(state.list, key = { it.peerId }) { item ->
                            ConversationRow(
                                item = item,
                                onClick = {
                                    onOpenChat(item.peerId, sanitizeNickname(item.peerNickname))
                                },
                                onLongClick = { bounds ->
                                    actionAnchor = bounds
                                    actionTarget = item
                                },
                            )
                            HorizontalDivider(
                                color = AppColors.Divider,
                                modifier = Modifier.padding(start = ConversationDividerInset),
                            )
                        }
                        if (state.hasMore) {
                            item {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .clickable { viewModel.loadMore() }
                                        .padding(AppSpacing.L),
                                    contentAlignment = Alignment.Center,
                                ) {
                                    if (state.loadingMore) {
                                        CircularProgressIndicator(
                                            color = AppColors.Brand,
                                            strokeWidth = 2.dp,
                                            modifier = Modifier.height(AppIconSize.Cell),
                                        )
                                    } else {
                                        Text(
                                            stringResource(R.string.common_load_more),
                                            fontSize = AppType.Meta,
                                            color = AppColors.TextMuted,
                                        )
                                    }
                                }
                            }
                        }
                        item { Spacer(Modifier.height(AppSpacing.S)) }
                    }
                }
            }
        }
    }

    val target = actionTarget
    if (target != null && actionAnchor != Rect.Zero) {
        DarkActionPillPopup(
            anchor = actionAnchor,
            actions = listOf(
                DarkActionItem(
                    stringResource(if (target.pinned) R.string.conversation_unpin else R.string.conversation_pin),
                ) {
                    actionTarget = null
                    viewModel.setPinned(target.peerId, !target.pinned)
                },
                DarkActionItem(stringResource(R.string.common_delete)) {
                    actionTarget = null
                    deleteTarget = target
                },
            ),
            onDismiss = { actionTarget = null },
        )
    }

    deleteTarget?.let { item ->
        AppConfirmDialog(
            title = stringResource(R.string.conversation_delete_title, sanitizeNickname(item.peerNickname)),
            message = stringResource(R.string.conversation_delete_message),
            confirmText = stringResource(R.string.common_delete),
            cancelText = stringResource(R.string.common_cancel),
            danger = true,
            onDismiss = { deleteTarget = null },
            onConfirm = {
                deleteTarget = null
                viewModel.deleteConversation(item.peerId)
            },
        )
    }
}
