package com.nearby.chat.core.model

data class RoomBootstrap(
    val room: RoomDetail = RoomDetail(),
    val onlineCount: Int = 0,
    val messages: List<MessageItem> = emptyList(),
)
