package com.nearby.chat.core.model

sealed interface ChatEvent {
    data class Message(
        val scope: Int,
        val item: MessageItem,
        val clientMsgId: String?,
    ) : ChatEvent

    data class Ack(
        val clientMsgId: String,
        val id: Long,
    ) : ChatEvent

    data class Recall(
        val scope: Int,
        val item: MessageItem,
    ) : ChatEvent

    data class RoomPresence(
        val roomId: Long,
        val onlineCount: Int,
    ) : ChatEvent

    data class PostPresence(
        val postId: Long,
        val onlineCount: Int,
    ) : ChatEvent
}
