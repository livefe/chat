package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class CreatePostRequest(
    val id: Long = 0,
    val roomId: Long = 0,
    val category: String = "idle",
    val pinned: Boolean = false,
    val title: String,
    val content: String,
    val images: List<Long> = emptyList(),
    val placeName: String = "",
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val startAt: Long = 0,
    val endAt: Long = 0,
    val extras: PostExtrasDto? = null,
)
