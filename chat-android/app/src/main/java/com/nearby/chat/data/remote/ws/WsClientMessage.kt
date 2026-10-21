package com.nearby.chat.data.remote.ws

import kotlinx.serialization.Serializable

@Serializable
data class WsClientMessage(
    val type: String,
    val scope: String? = null,
    val roomId: Long? = null,
    val postId: Long? = null,
    val toUserId: Long? = null,
    val lat: Double? = null,
    val lng: Double? = null,
    val contentType: Int? = null,
    val content: String? = null,
    val replyToId: Long? = null,
    val clientMsgId: String? = null,
)
