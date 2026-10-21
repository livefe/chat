package com.nearby.chat.core.model

data class SponsoredItem(
    val campaignId: Long = 0,
    val title: String = "",
    val subtitle: String = "",
    val imageUrl: String = "",
    val advertiserName: String = "",
    val badge: String = "",
    val landingType: String = AdLandingType.URL,
    val landingValue: String = "",
    val distance: Double = 0.0,
    val impressionToken: String = "",
    val requestId: String = "",
    val slotIndex: Int = 0,
)
