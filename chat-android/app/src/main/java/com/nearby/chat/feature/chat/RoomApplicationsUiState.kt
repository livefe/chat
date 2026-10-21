package com.nearby.chat.feature.chat

import com.nearby.chat.core.model.RoomApplicationItem

data class RoomApplicationsUiState(
    val loading: Boolean = false,
    val list: List<RoomApplicationItem> = emptyList(),
)
