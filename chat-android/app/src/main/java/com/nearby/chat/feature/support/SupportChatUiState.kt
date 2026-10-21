package com.nearby.chat.feature.support

import com.nearby.chat.core.model.SupportMessage
import com.nearby.chat.core.model.SupportSession

data class SupportChatUiState(
    val brandName: String = "",
    val brandAvatar: String = "",
    val session: SupportSession? = null,
    val messages: List<SupportMessage> = emptyList(),
    val input: String = "",
    val loading: Boolean = true,
    val sending: Boolean = false,
    val uploading: Boolean = false,
    val selfId: Long = 0,
)
