package com.nearby.chat.feature.account

import com.nearby.chat.core.model.BlockListItem

data class BlockListUiState(
    val list: List<BlockListItem> = emptyList(),
    val loading: Boolean = false,
    val hasMore: Boolean = true,
    val unblockingId: Long = 0,
)
