package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportSendRequest(
    val conversationId: Long,
    val contentType: Int = 1,
    val content: String,
    val clientMsgId: String = "",
)
