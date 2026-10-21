package com.nearby.chat.core.model

data class RegionItem(
    val code: String = "",
    val name: String = "",
    val shortName: String = "",
    val parentCode: String = "",
    val level: Int = 0,
    val fullPath: String = "",
    val tiandituGb: String = "",
    val amapAdcode: String = "",
)
