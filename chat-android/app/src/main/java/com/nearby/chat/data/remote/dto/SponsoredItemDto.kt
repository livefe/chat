package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SponsoredItemDto(
    val campaignId: Long = 0,
    val title: String = "",
    val subtitle: String = "",
    val imageUrl: String = "",
    val advertiserName: String = "",
    val badge: String = "",
    val landingType: String = "url",
    val landingValue: String = "",
    val distance: Double = 0.0,
    val impressionToken: String = "",
    val requestId: String = "",
    val slotIndex: Int = 0,
)
