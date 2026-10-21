package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class NearbyBootstrapResultDto(
    val rooms: List<NearbyRoomItemDto> = emptyList(),
    val onlineCount: Int = 0,
    val messages: List<MessageItemDto> = emptyList(),
)
