package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RoomBootstrapResultDto(
    val room: RoomDetailDto = RoomDetailDto(),
    val onlineCount: Int = 0,
    val messages: List<MessageItemDto> = emptyList(),
)
