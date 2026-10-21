package com.nearby.chat.core.model

data class PostDraft(
    val id: Long = 0,
    val roomId: Long = 0,
    val category: String,
    val pinned: Boolean = false,
    val title: String,
    val content: String,
    val images: List<Long> = emptyList(),
    val placeName: String = "",
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val startAt: Long = 0,
    val endAt: Long = 0,
    val extras: PostExtras? = null,
)
