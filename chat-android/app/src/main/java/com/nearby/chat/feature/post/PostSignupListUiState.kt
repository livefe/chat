package com.nearby.chat.feature.post

import com.nearby.chat.core.model.PostSignupItem
import com.nearby.chat.core.model.PostSignupListCounts

data class PostSignupListUiState(
    val postId: Long = 0,
    val postTitle: String = "",
    val statusFilter: Int? = null,
    val checkInFilter: Int? = null,
    val list: List<PostSignupItem> = emptyList(),
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val hasMore: Boolean = true,
    val exporting: Boolean = false,
    val needCheckIn: Boolean = false,
    val checkInWindowOpen: Boolean = true,
    val checkingUserId: Long = 0,
    val counts: PostSignupListCounts = PostSignupListCounts(),
    val remarkEditing: PostSignupItem? = null,
    val remarkDraft: String = "",
    val remarkSaving: Boolean = false,
)
