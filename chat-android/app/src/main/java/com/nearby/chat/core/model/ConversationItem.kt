package com.nearby.chat.core.model

data class ConversationItem(
    val peerId: Long = 0,
    val peerNickname: String = "",
    val peerAvatar: String = "",
    val peerMobile: String = "",
    val lastContent: String = "",
    val lastActiveAt: Long = 0,
    val unreadCount: Int = 0,
    val pinned: Boolean = false,
)
