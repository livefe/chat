package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class DeviceRegisterRequest(
    val vendor: String,
    val pushToken: String,
    val osVersion: String = "",
)
