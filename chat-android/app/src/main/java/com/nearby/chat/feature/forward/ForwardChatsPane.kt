package com.nearby.chat.feature.forward

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import com.nearby.chat.R
import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.UserAvatar

@Composable
internal fun ForwardChatsPane(
    state: ForwardUiState,
    chats: List<ConversationItem>,
    onSelect: (ConversationItem) -> Unit,
    onLoadMore: () -> Unit,
) {
    when {
        state.chatsLoading && state.chats.isEmpty() -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(
                    color = AppColors.Brand,
                    strokeWidth = ForwardProgressStroke,
                    modifier = Modifier.size(ForwardProgressSize),
                )
            }
        }

        state.chats.isEmpty() -> EmptyHint(stringResource(R.string.messages_empty), Modifier.fillMaxSize())

        chats.isEmpty() -> EmptyHint(stringResource(R.string.forward_chats_no_match), Modifier.fillMaxSize())

        else -> {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(chats, key = { _, it -> it.peerId }) { index, item ->
                    ForwardListRow(
                        title = item.peerNickname.ifBlank { stringResource(R.string.user_default_name) },
                        enabled = !state.sending,
                        onClick = { onSelect(item) },
                    ) {
                        UserAvatar(url = item.peerAvatar, size = ForwardLeadingSize)
                    }
                    if (index < chats.lastIndex || state.chatsHasMore) ForwardListDivider()
                }
                if (state.chatsHasMore) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable(onClick = onLoadMore)
                                .padding(AppSpacing.L),
                            contentAlignment = Alignment.Center,
                        ) {
                            if (state.chatsLoadingMore) {
                                CircularProgressIndicator(
                                    color = AppColors.Brand,
                                    strokeWidth = ForwardProgressStroke,
                                    modifier = Modifier.size(AppIconSize.Cell),
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
            }
        }
    }
}
