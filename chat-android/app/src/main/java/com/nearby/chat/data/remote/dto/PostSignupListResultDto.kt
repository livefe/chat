package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostSignupListResultDto(
    val total: Int = 0,
    val list: List<PostSignupItemDto> = emptyList(),
    val needCheckIn: Boolean = false,
    val counts: PostSignupListCountsDto = PostSignupListCountsDto(),
)
