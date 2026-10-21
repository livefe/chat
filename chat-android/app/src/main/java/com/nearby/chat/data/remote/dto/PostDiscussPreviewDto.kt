package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostDiscussPreviewDto(
    val fromName: String = "",
    val contentType: Int = 0,
    val content: String = "",
)
