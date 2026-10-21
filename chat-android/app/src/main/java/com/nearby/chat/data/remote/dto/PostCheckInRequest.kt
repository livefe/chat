package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostCheckInRequest(
    val payload: String = "",
    val postId: Long = 0,
    val userId: Long = 0,
)
