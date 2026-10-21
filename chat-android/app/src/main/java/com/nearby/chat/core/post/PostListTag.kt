package com.nearby.chat.core.post

import com.nearby.chat.core.ui.UiText

data class PostListTag(
    val text: UiText,
    val tone: PostListTagTone = PostListTagTone.None,
)
