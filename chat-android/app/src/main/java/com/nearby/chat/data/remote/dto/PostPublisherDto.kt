package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostPublisherDto(
    val id: Long = 0,
    val nickname: String = "",
    val avatarUrl: String = "",
)
