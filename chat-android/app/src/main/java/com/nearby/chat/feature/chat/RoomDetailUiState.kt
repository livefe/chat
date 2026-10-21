package com.nearby.chat.feature.chat

import com.nearby.chat.core.model.RoomDetail

data class RoomDetailUiState(
    val room: RoomDetail? = null,
    val onlineCount: Int = 0,
)
