package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class UnreadPeerCountDto(
    val peerId: Long = 0,
    val count: Int = 0,
)
