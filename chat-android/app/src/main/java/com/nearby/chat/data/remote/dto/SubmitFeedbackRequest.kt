package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SubmitFeedbackRequest(
    val type: Int,
    val content: String,
    val contact: String? = null,
    val images: List<Long>? = null,
)
