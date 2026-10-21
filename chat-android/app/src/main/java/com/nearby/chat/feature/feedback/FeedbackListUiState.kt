package com.nearby.chat.feature.feedback

import com.nearby.chat.core.model.FeedbackItem

data class FeedbackListUiState(
    val list: List<FeedbackItem> = emptyList(),
    val loading: Boolean = false,
    val hasMore: Boolean = true,
)
