package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportProfileDto(
    val brandName: String = "",
    val brandAvatarUrl: String = "",
    val welcomeText: String = "",
    val offlineText: String = "",
    val enabled: Boolean = true,
)
