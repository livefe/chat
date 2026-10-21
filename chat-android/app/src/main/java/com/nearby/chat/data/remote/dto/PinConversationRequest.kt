package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PinConversationRequest(
    val peerId: Long,
    val pinned: Boolean,
)
