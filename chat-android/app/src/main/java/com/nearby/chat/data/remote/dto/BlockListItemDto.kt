package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class BlockListItemDto(
    val userId: Long = 0,
    val nickname: String = "",
    val avatarUrl: String = "",
    val createdAt: String = "",
)
