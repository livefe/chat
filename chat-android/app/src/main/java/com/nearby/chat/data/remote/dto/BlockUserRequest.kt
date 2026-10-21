package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class BlockUserRequest(
    val userId: Long,
)
