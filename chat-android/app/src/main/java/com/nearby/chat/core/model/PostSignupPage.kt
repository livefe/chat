package com.nearby.chat.core.model

data class PostSignupPage(
    val total: Int = 0,
    val list: List<PostSignupItem> = emptyList(),
    val needCheckIn: Boolean = false,
    val counts: PostSignupListCounts = PostSignupListCounts(),
)
