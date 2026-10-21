package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class UpdateProfileRequest(
    val nickname: String? = null,
    val avatarMediaId: Long? = null,
)
