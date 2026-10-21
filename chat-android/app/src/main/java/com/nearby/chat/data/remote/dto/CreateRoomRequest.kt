package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class CreateRoomRequest(
    val name: String,
    val placeType: Int,
    val centerLat: Double,
    val centerLng: Double,
    val radius: Int,
    val description: String = "",
    val shapeType: Int = 1,
    val polygon: List<List<Double>>? = null,
)
