package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class MessageItemDto(
    val id: Long = 0,
    val scope: Int = 0,
    val roomId: Long = 0,
    val postId: Long = 0,
    val fromUserId: Long = 0,
    val toUserId: Long = 0,
    val fromName: String = "",
    val fromAvatar: String = "",
    val contentType: Int = 0,
    val content: String = "",
    val replyToId: Long = 0,
    val replyToFromName: String = "",
    val replyToContentType: Int = 0,
    val replyToContent: String = "",
    val createdAt: Long = 0,
    val recalled: Boolean = false,
)
