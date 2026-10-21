package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class MessageListResultDto(
    val list: List<MessageItemDto> = emptyList(),
    val clearedBeforeId: Long = 0,
)
