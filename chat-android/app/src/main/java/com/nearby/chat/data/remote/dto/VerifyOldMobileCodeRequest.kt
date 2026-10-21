package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class VerifyOldMobileCodeRequest(
    val code: String,
)
