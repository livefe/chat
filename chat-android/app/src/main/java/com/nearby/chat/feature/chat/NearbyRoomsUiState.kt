package com.nearby.chat.feature.chat

import com.nearby.chat.core.model.NearbyRoomItem

data class NearbyRoomsUiState(
    val rooms: List<NearbyRoomItem> = emptyList(),
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val locationReady: Boolean = false,
)
