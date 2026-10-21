package com.nearby.chat.feature.forward

import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText

data class ForwardSession(
    val contentType: Int,
    val content: String,
    val pageTitle: UiText = UiText.of(R.string.forward_title),
    val postId: Long = 0L,
    val postTitle: String = "",
    val cover: String = "",
    val shareContent: String = "",
    val showExternalShare: Boolean = false,
)
