package com.nearby.chat.core.model

data class CheckInResultPayload(
    val kind: String = "",
    val message: String = "",
    val postId: Long = 0,
    val postTitle: String = "",
    val userId: Long = 0,
    val userNickname: String = "",
    val userAvatarUrl: String = "",
    val signupAt: String = "",
    val checkedInAtText: String = "",
    val operatorNickname: String = "",
    val operatorAvatarUrl: String = "",
)
