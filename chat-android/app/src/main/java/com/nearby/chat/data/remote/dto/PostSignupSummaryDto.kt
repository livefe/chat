package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostSignupSummaryDto(
    val confirmed: Int = 0,
    val waitlisted: Int = 0,
)
