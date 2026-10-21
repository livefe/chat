package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class ChangeMobileRequest(
    val ticket: String,
    val newMobile: String,
    val newCode: String,
)
