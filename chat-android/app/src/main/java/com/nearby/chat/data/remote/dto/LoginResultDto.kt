package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class LoginResultDto(
    val accessToken: String = "",
    val refreshToken: String = "",
    val isFirstLogin: Boolean = false,
    val user: UserInfoDto = UserInfoDto(),
)
