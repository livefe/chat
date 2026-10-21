package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class FeedbackItemDto(
    val id: Long = 0,
    val type: Int = 0,
    val typeText: String = "",
    val content: String = "",
    val contact: String = "",
    val images: List<String> = emptyList(),
    val status: Int = 0,
    val statusText: String = "",
    val reply: String = "",
    val handledAt: String = "",
    val createdAt: String = "",
)
