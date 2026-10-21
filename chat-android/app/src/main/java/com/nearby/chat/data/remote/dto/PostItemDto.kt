package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostItemDto(
    val id: Long = 0,
    val roomId: Long = 0,
    val category: String = "idle",
    val pinned: Boolean = false,
    val title: String = "",
    val content: String = "",
    val images: List<String> = emptyList(),
    val imageIds: List<Long> = emptyList(),
    val placeName: String = "",
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val startAt: Long = 0,
    val endAt: Long = 0,
    val status: Int = 0,
    val publisherId: Long = 0,
    val publisher: PostPublisherDto? = null,
    val discussMessageCount: Int = 0,
    val discussPreview: PostDiscussPreviewDto? = null,
    val createdAt: String = "",
    val updatedAt: String = "",
    val distance: Double = 0.0,
    val extras: PostExtrasDto? = null,
    val signupSummary: PostSignupSummaryDto? = null,
    val mySignupStatus: Int = 0,
    val myCheckedInAt: Long = 0,
)
