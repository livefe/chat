package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class ConversationItemDto(
    val peerId: Long = 0,
    val peerNickname: String = "",
    val peerAvatar: String = "",
    val peerMobile: String = "",
    val lastContent: String = "",
    val lastActiveAt: Long = 0,
    val unreadCount: Int = 0,
    val pinned: Boolean = false,
)
