package com.nearby.chat.feature.chat

data class NearbyChatUiState(
    val onlineCount: Int = 0,
    val locating: Boolean = false,
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val placeName: String = "",
    val isBrowsing: Boolean = false,
    val locationReady: Boolean = false,
    val locationFailed: Boolean = false,
)
