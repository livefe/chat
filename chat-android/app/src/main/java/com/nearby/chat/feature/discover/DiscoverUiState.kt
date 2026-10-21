package com.nearby.chat.feature.discover

import com.nearby.chat.core.model.FeedItem

data class DiscoverUiState(
    val list: List<FeedItem> = emptyList(),
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val hasMore: Boolean = true,
)
