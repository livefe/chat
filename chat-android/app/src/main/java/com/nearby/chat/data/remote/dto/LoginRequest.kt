package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class LoginRequest(
    val mobile: String,
    val code: String,
    val source: String = "app_login",
)
