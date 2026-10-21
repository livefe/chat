package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostCheckInCodeDto(
    val postId: Long = 0,
    val postTitle: String = "",
    val placeName: String = "",
    val startAt: Long = 0,
    val endAt: Long = 0,
    val payload: String = "",
    val checkedIn: Boolean = false,
    val checkedInAt: Long = 0,
    val checkedInAtText: String = "",
    val userNickname: String = "",
    val signupAt: String = "",
)
