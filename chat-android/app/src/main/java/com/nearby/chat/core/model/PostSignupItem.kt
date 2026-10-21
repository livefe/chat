package com.nearby.chat.core.model

data class PostSignupItem(
    val id: Long = 0,
    val userId: Long = 0,
    val userNickname: String = "",
    val userMobile: String = "",
    val signupStatus: Int = 0,
    val signupStatusText: String = "",
    val checkedInAt: String = "",
    val createdAt: String = "",
    val cancelledAt: String = "",
    val remark: String = "",
)
