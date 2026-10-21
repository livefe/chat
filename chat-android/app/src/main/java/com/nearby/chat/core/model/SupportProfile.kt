package com.nearby.chat.core.model

data class SupportProfile(
    val brandName: String = "",
    val brandAvatarUrl: String = "",
    val welcomeText: String = "",
    val offlineText: String = "",
    val enabled: Boolean = true,
)
