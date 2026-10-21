package com.nearby.chat.core.model

data class PublicUserProfile(
    val id: Long = 0,
    val nickname: String = "",
    val avatarUrl: String = "",
    val createdAt: String = "",
    val isSelf: Boolean = false,
    val deactivated: Boolean = false,
    val blockedByMe: Boolean = false,
    val canMessage: Boolean = false,
)
