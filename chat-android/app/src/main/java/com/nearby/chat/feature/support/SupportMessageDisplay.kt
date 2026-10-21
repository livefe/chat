package com.nearby.chat.feature.support

import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.SupportMessage

private const val SENDER_KIND_USER = 1
private const val SENDER_KIND_SYSTEM = 3

internal fun SupportMessage.toChatMessage(selfId: Long, brandAvatar: String): MessageItem {
    if (senderKind == SENDER_KIND_SYSTEM) {
        return MessageItem(
            id = id,
            scope = MessageScopes.SYSTEM_HINT,
            contentType = MessageContentTypes.TEXT,
            content = content,
            createdAt = createdAt,
            recalled = recalled,
            clientMsgId = clientMsgId,
        )
    }

    val mine = senderKind == SENDER_KIND_USER
    return MessageItem(
        id = id,
        fromUserId = if (mine) {
            if (selfId > 0) selfId else senderId
        } else {
            0L
        },
        fromName = senderName,
        fromAvatar = senderAvatarUrl.ifBlank { brandAvatar },
        contentType = contentType,
        content = content,
        createdAt = createdAt,
        recalled = recalled,
        clientMsgId = clientMsgId,
    )
}
