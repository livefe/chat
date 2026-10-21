package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class RegionItemDto(
    val code: String = "",
    val name: String = "",
    val shortName: String = "",
    val parentCode: String = "",
    val level: Int = 0,
    val fullPath: String = "",
    val tiandituGb: String = "",
    val amapAdcode: String = "",
)
