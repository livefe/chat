package com.nearby.chat.feature.messages

import com.nearby.chat.core.model.ConversationItem

data class MessagesUiState(
    val list: List<ConversationItem> = emptyList(),
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val hasMore: Boolean = false,
)
