package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class AdEventsRequest(
    val events: List<AdEventItemRequest>,
)
