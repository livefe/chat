package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class UpdateSignupRemarkRequest(
    val postId: Long,
    val userId: Long,
    val remark: String = "",
)
