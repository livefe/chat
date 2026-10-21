package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class AdEventItemRequest(
    val requestId: String,
    val campaignId: Long,
    val eventType: String,
    val impressionToken: String,
    val slotIndex: Int = 0,
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
)
