package com.nearby.chat.feature.post

import com.nearby.chat.core.model.PostItem

data class MyPostsUiState(
    val mode: MyPostsMode = MyPostsMode.Published,
    val list: List<PostItem> = emptyList(),
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val hasMore: Boolean = true,
    val actingId: Long = 0,
)
