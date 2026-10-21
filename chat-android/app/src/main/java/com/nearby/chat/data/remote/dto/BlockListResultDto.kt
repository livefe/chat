package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class BlockListResultDto(
    val total: Int = 0,
    val list: List<BlockListItemDto> = emptyList(),
)
