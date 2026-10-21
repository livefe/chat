package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostSignupItemDto(
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
