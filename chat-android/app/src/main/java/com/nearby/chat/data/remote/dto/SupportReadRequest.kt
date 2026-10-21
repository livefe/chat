package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportReadRequest(
    val conversationId: Long,
    val lastReadMsgId: Long = 0,
)
