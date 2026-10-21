package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SendNewMobileCodeRequest(
    val mobile: String,
    val ticket: String,
)
