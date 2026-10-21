package com.nearby.chat.feature.chat.thread

data class ChatRoute(
    val roomId: Long = 0,
    val postId: Long = 0,
    val toUserId: Long = 0,
    val lat: Double? = null,
    val lng: Double? = null,
)
