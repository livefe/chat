package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PublicUserProfileDto(
    val id: Long = 0,
    val nickname: String = "",
    val avatarUrl: String = "",
    val createdAt: String = "",
    val isSelf: Boolean = false,
    val deactivated: Boolean = false,
    val blockedByMe: Boolean = false,
    val canMessage: Boolean = false,
)
