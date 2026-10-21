package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportMessageDto(
    val id: Long = 0,
    val conversationId: Long = 0,
    val senderKind: Int = 0,
    val senderId: Long = 0,
    val senderName: String = "",
    val senderAvatarUrl: String = "",
    val contentType: Int = 1,
    val content: String = "",
    val createdAt: Long = 0,
    val recalled: Boolean = false,
    val clientMsgId: String = "",
)
