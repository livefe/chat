package com.nearby.chat.core.model

data class FeedbackItem(
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
