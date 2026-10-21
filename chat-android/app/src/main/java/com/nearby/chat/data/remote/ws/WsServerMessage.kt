package com.nearby.chat.data.remote.ws

import com.nearby.chat.data.remote.dto.MessageItemDto
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.decodeFromJsonElement

val wsPayloadJson = Json {
    ignoreUnknownKeys = true
    explicitNulls = false
}

@Serializable
data class WsServerMessage(
    val type: String = "",
    val scope: String? = null,
    val data: JsonElement? = null,
    val content: String? = null,
    val clientMsgId: String? = null,
    val id: Long? = null,
    val roomId: Long? = null,
    val postId: Long? = null,
    val onlineCount: Int? = null,
)

fun WsServerMessage.messageItem(): MessageItemDto? = data?.let { runCatching { wsPayloadJson.decodeFromJsonElement<MessageItemDto>(it) }.getOrNull() }

inline fun <reified T> WsServerMessage.decodeData(): T? = data?.let { runCatching { wsPayloadJson.decodeFromJsonElement<T>(it) }.getOrNull() }
