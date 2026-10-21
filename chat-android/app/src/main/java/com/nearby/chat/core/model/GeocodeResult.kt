package com.nearby.chat.core.model

data class GeocodeResult(
    val placeName: String,
    val formattedAddress: String = "",
    val province: String = "",
    val city: String = "",
    val county: String = "",
)
