package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportOpenRequest(
    val source: String = "android",
)
