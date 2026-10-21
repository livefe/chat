package com.nearby.chat.feature.post

import com.nearby.chat.core.model.PostItem

data class PostDetailUiState(
    val detail: PostItem? = null,
    val selfId: Long = 0,
    val showChat: Boolean = false,
    val discussionLoaded: Boolean = false,
    val onlineCount: Int = 0,
)
