package com.nearby.chat.feature.post.ui

data class PostListAction(
    val label: String,
    val danger: Boolean = false,
    val enabled: Boolean = true,
    val onClick: () -> Unit,
)
