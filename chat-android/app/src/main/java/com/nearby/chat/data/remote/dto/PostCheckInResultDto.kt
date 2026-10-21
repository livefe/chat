package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostCheckInResultDto(
    val postId: Long = 0,
    val postTitle: String = "",
    val userId: Long = 0,
    val userNickname: String = "",
    val userAvatarUrl: String = "",
    val signupAt: String = "",
    val checkedInAt: Long = 0,
    val checkedInAtText: String = "",
    val alreadyCheckedIn: Boolean = false,
    val operatorId: Long = 0,
    val operatorNickname: String = "",
    val operatorAvatarUrl: String = "",
)
