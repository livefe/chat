package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RefreshTokenResultDto(
    val accessToken: String = "",
    val refreshToken: String = "",
)
