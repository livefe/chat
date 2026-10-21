package com.nearby.chat.data.mapper

import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.data.remote.ws.WsScopes
import com.nearby.chat.data.remote.ws.WsServerMessage
import com.nearby.chat.data.remote.ws.WsTypes
import com.nearby.chat.data.remote.ws.messageItem

private fun wsScopeToMessageScope(scope: String?): Int = when (scope) {
    WsScopes.NEARBY -> MessageScopes.NEARBY
    WsScopes.ROOM -> MessageScopes.ROOM
    WsScopes.PRIVATE -> MessageScopes.PRIVATE
    WsScopes.POST -> MessageScopes.POST
    else -> 0
}

fun WsServerMessage.toChatEvent(): ChatEvent? = when (type) {
    WsTypes.MESSAGE -> messageItem()?.let {
        ChatEvent.Message(
            scope = wsScopeToMessageScope(scope),
            item = it.toDomain(),
            clientMsgId = clientMsgId,
        )
    }

    WsTypes.ACK -> {
        val cid = clientMsgId
        val ackId = id
        if (cid != null && ackId != null) ChatEvent.Ack(cid, ackId) else null
    }

    WsTypes.RECALL -> messageItem()?.let {
        ChatEvent.Recall(scope = wsScopeToMessageScope(scope), item = it.toDomain())
    }

    WsTypes.ROOM_PRESENCE -> roomId?.let { ChatEvent.RoomPresence(it, onlineCount ?: 0) }

    WsTypes.POST_PRESENCE -> postId?.let { ChatEvent.PostPresence(it, onlineCount ?: 0) }

    else -> null
}
