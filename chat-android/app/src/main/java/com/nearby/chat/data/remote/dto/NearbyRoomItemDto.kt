package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class NearbyRoomItemDto(
    val id: Long = 0,
    val name: String = "",
    val placeType: Int = 0,
    val placeTypeText: String = "",
    val shapeType: Int = 1,
    val centerLat: Double = 0.0,
    val centerLng: Double = 0.0,
    val radius: Double = 0.0,
    val distance: Double = 0.0,
    val requireGeofence: Boolean = false,
)
