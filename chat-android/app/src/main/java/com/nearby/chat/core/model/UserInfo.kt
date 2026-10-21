package com.nearby.chat.core.model

data class UserInfo(
    val id: Long = 0,
    val mobile: String = "",
    val nickname: String = "",
    val avatarUrl: String = "",
    val status: Int = 0,
)
