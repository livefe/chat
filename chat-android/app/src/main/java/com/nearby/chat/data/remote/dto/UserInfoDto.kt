package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class UserInfoDto(
    val id: Long = 0,
    val mobile: String = "",
    val nickname: String = "",
    val avatarUrl: String = "",
    val status: Int = 0,
)
