package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class FeedItemDto(
    val type: String = "post",
    val post: PostItemDto? = null,
    val sponsored: SponsoredItemDto? = null,
)
