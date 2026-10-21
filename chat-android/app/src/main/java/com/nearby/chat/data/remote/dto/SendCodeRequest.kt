package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SendCodeRequest(
    val mobile: String,
    val source: String = "app_login",
)
