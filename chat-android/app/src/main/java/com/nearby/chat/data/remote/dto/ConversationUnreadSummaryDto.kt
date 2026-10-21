package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class ConversationUnreadSummaryDto(
    val total: Int = 0,
    val items: List<UnreadPeerCountDto> = emptyList(),
)
