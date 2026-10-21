package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PeerIdRequest(
    val peerId: Long,
)
