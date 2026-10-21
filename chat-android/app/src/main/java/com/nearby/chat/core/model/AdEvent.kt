package com.nearby.chat.core.model

data class AdEvent(
    val requestId: String,
    val campaignId: Long,
    val eventType: String,
    val impressionToken: String,
    val slotIndex: Int = 0,
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
)
