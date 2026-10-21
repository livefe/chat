package com.nearby.chat.core.chat

import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.MessageItem
import java.util.UUID

const val RECALLED_REPLY_CONTENT = "原消息已撤回"

fun applyIncomingChat(
    messages: List<MessageItem>,
    event: ChatEvent,
    acceptMessage: (MessageItem) -> Boolean,
): List<MessageItem>? {
    return when (event) {
        is ChatEvent.Message -> {
            if (!acceptMessage(event.item)) return null
            replaceOptimisticOrAppend(messages, event.item, event.clientMsgId)
        }

        is ChatEvent.Ack -> applyAck(messages, event.clientMsgId, event.id)

        is ChatEvent.Recall -> {
            if (!acceptMessage(event.item)) return null
            applyRecall(messages, event.item.id)
        }

        else -> null
    }
}

fun mergeMessages(existing: List<MessageItem>, incoming: List<MessageItem>): List<MessageItem> {
    if (existing.isEmpty()) return incoming.sortedWith(compareBy({ it.createdAt }, { it.id }))
    if (incoming.isEmpty()) return existing.sortedWith(compareBy({ it.createdAt }, { it.id }))
    val byId = LinkedHashMap<Long, MessageItem>()
    for (m in incoming) {
        if (m.id > 0) byId[m.id] = m
    }
    val optimistic = mutableListOf<MessageItem>()
    for (m in existing) {
        if (m.id > 0) {
            if (m.id !in byId) byId[m.id] = m
        } else if (m.sendStatus == SendStatus.SENDING || m.sendStatus == SendStatus.FAILED) {
            optimistic += m
        }
    }
    return (byId.values + optimistic).sortedWith(compareBy({ it.createdAt }, { it.id }))
}

fun trimMessages(messages: List<MessageItem>, keep: Int = ChatLimits.MESSAGE_KEEP): List<MessageItem> {
    if (keep <= 0 || messages.size <= keep) return messages
    return messages.takeLast(keep)
}

fun replaceOptimisticOrAppend(
    messages: List<MessageItem>,
    item: MessageItem,
    clientMsgId: String? = null,
): List<MessageItem> {
    if (!clientMsgId.isNullOrBlank()) {
        val idx = messages.indexOfFirst { it.clientMsgId == clientMsgId }
        if (idx >= 0) {
            val local = messages[idx]
            return messages.toMutableList().also {
                it[idx] = item.copy(
                    clientMsgId = clientMsgId,
                    sendStatus = "",
                    fromName = item.fromName.ifBlank { local.fromName },
                    fromAvatar = item.fromAvatar.ifBlank { local.fromAvatar },
                )
            }
        }
    }
    return if (messages.any { it.id == item.id && item.id > 0 }) messages else messages + item
}

fun applyAck(messages: List<MessageItem>, clientMsgId: String, id: Long): List<MessageItem> {
    if (clientMsgId.isBlank() || id <= 0) return messages
    val idx = messages.indexOfFirst { it.clientMsgId == clientMsgId }
    if (idx < 0) return messages
    return messages.toMutableList().also {
        it[idx] = it[idx].copy(id = id, sendStatus = "")
    }
}

fun applyRecall(messages: List<MessageItem>, id: Long): List<MessageItem> {
    if (id <= 0) return messages
    var changed = false
    val next = messages.map { m ->
        when {
            m.id == id && !m.recalled -> {
                changed = true
                m.copy(recalled = true)
            }

            m.replyToId == id && m.replyToContent != RECALLED_REPLY_CONTENT -> {
                changed = true
                m.copy(replyToContent = RECALLED_REPLY_CONTENT, replyToContentType = 0)
            }

            else -> m
        }
    }
    return if (changed) next else messages
}

fun markSendFailed(messages: List<MessageItem>, clientMsgId: String?): List<MessageItem> {
    if (clientMsgId.isNullOrBlank()) return messages
    return messages.map {
        if (it.clientMsgId == clientMsgId && it.id <= 0) it.copy(sendStatus = SendStatus.FAILED) else it
    }
}

fun newOptimisticMessage(
    selfId: Long,
    selfName: String,
    selfAvatar: String,
    content: String,
    contentType: Int,
    scope: Int = 0,
    roomId: Long = 0,
    postId: Long = 0,
    toUserId: Long = 0,
    lat: Double = 0.0,
    lng: Double = 0.0,
    replyTo: MessageItem? = null,
    replyToContent: String = "",
): Pair<String, MessageItem> {
    val clientMsgId = UUID.randomUUID().toString()
    val quote = replyTo?.takeIf { it.id > 0 }
    return clientMsgId to MessageItem(
        id = 0,
        scope = scope,
        roomId = roomId,
        postId = postId,
        fromUserId = selfId,
        toUserId = toUserId,
        fromName = selfName,
        fromAvatar = selfAvatar,
        contentType = contentType,
        content = content,
        replyToId = quote?.id ?: 0,
        replyToFromName = quote?.fromName.orEmpty(),
        replyToContentType = quote?.contentType ?: 0,
        replyToContent = if (quote != null) replyToContent else "",
        latitude = lat,
        longitude = lng,
        createdAt = System.currentTimeMillis() / 1000,
        clientMsgId = clientMsgId,
        sendStatus = SendStatus.SENDING,
    )
}

fun messageListKey(m: MessageItem): Any = m.clientMsgId.ifBlank { if (m.id > 0) m.id else "t-${m.createdAt}-${m.fromUserId}" }
