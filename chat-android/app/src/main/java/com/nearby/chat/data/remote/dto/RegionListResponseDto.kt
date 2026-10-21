package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RegionListResponseDto(
    val list: List<RegionItemDto> = emptyList(),
)
