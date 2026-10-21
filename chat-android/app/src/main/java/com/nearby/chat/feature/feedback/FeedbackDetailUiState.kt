package com.nearby.chat.feature.feedback

import com.nearby.chat.core.model.FeedbackItem

data class FeedbackDetailUiState(
    val loading: Boolean = true,
    val detail: FeedbackItem? = null,
)
