package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportMessageListResultDto(
    val list: List<SupportMessageDto> = emptyList(),
)
