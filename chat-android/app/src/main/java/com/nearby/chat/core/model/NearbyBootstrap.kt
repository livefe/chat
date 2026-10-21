package com.nearby.chat.core.model

data class NearbyBootstrap(
    val rooms: List<NearbyRoomItem> = emptyList(),
    val onlineCount: Int = 0,
    val messages: List<MessageItem> = emptyList(),
)
