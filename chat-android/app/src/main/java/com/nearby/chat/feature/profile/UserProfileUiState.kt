package com.nearby.chat.feature.profile

import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.model.PublicUserProfile

data class UserProfileUiState(
    val profile: PublicUserProfile? = null,
    val profileLoading: Boolean = false,
    val posts: List<PostItem> = emptyList(),
    val postsLoading: Boolean = false,
    val refreshing: Boolean = false,
    val hasMore: Boolean = true,
    val loadFailed: Boolean = false,
)
