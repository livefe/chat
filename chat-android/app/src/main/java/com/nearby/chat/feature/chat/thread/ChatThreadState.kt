package com.nearby.chat.feature.chat.thread

import com.nearby.chat.core.model.MessageItem

data class ChatThreadState(
    val messages: List<MessageItem> = emptyList(),
    val hasMore: Boolean = false,
    val selfId: Long = 0,
    val input: String = "",
    val uploading: Boolean = false,
    val replyTo: MessageItem? = null,
)
