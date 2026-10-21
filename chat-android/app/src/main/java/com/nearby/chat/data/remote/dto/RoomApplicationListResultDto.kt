package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RoomApplicationListResultDto(
    val total: Long = 0,
    val list: List<RoomApplicationItemDto> = emptyList(),
)
