package com.nearby.chat.core.model

data class SavedChatLocation(
    val lat: Double,
    val lng: Double,
    val placeName: String,
    val browsing: Boolean,
    val updatedAt: Long,
)
