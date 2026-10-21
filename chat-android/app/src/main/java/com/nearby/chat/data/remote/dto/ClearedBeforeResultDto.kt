package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class ClearedBeforeResultDto(
    val clearedBeforeId: Long = 0,
)
