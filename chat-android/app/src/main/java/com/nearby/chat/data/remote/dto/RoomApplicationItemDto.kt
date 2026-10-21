package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RoomApplicationItemDto(
    val id: Long = 0,
    val name: String = "",
    val description: String = "",
    val placeType: Int = 0,
    val placeTypeText: String = "",
    val status: Int = 0,
    val statusText: String = "",
    val radius: Double = 0.0,
    val createdAt: String = "",
)
