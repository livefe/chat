package com.nearby.chat.feature.chat

import com.nearby.chat.core.chat.ChatLimits

data class RoomCreateUiState(
    val name: String = "",
    val description: String = "",
    val placeType: Int? = null,
    val placeName: String = "",
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val locationReady: Boolean = false,
    val locating: Boolean = false,
    val radius: Float = ChatLimits.ROOM_RADIUS_DEFAULT.toFloat(),
    val loading: Boolean = false,
)
