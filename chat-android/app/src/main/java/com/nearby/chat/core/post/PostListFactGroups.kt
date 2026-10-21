package com.nearby.chat.core.post

import com.nearby.chat.core.ui.UiText

data class PostListFactGroups(
    val phase: PostListTag? = null,
    val schedule: String? = null,
    val signup: List<UiText> = emptyList(),
    val signupDeadline: UiText? = null,
    val fee: UiText? = null,
)
